package payment

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	stripeAPIBase         = "https://api.stripe.com/v1"
	stripeToleranceWindow = 5 * time.Minute
	stripeRequestTimeout  = 20 * time.Second
)

// zeroDecimalCurrencies are billed without a minor unit.
var zeroDecimalCurrencies = map[string]bool{
	"bif": true, "clp": true, "djf": true, "gnf": true, "jpy": true, "kmf": true,
	"krw": true, "mga": true, "pyg": true, "rwf": true, "ugx": true, "vnd": true,
	"vuv": true, "xaf": true, "xof": true, "xpf": true,
}

// Stripe implements Provider using Stripe Checkout (hosted) over the REST API.
// Webhook authenticity is verified with the endpoint's signing secret.
type Stripe struct {
	secretKey       string
	webhookSecret   string
	defaultCurrency string
	successURL      string
	cancelURL       string
	apiBase         string
	client          *http.Client
}

// NewStripe builds a Stripe provider. An empty secretKey still yields a usable
// provider whose Initiate/Verify return ErrNotConfigured.
func NewStripe(secretKey, webhookSecret, currency, successURL, cancelURL string) *Stripe {
	currency = strings.ToLower(strings.TrimSpace(currency))
	if currency == "" {
		currency = "usd"
	}
	return &Stripe{
		secretKey:       secretKey,
		webhookSecret:   webhookSecret,
		defaultCurrency: currency,
		successURL:      successURL,
		cancelURL:       cancelURL,
		apiBase:         stripeAPIBase,
		client:          &http.Client{Timeout: stripeRequestTimeout},
	}
}

func (s *Stripe) Name() string { return "stripe" }

// Initiate creates a Stripe Checkout Session and returns its hosted URL.
func (s *Stripe) Initiate(ctx context.Context, req InitRequest) (InitResult, error) {
	if s.secretKey == "" {
		return InitResult{}, ErrNotConfigured
	}

	currency := strings.ToLower(strings.TrimSpace(req.Currency))
	if currency == "" {
		currency = s.defaultCurrency
	}
	minor, err := majorToMinor(req.Amount, currency)
	if err != nil {
		return InitResult{}, err
	}

	name := "Order " + req.Reference
	if req.Metadata["description"] != "" {
		name = req.Metadata["description"]
	}

	form := url.Values{}
	form.Set("mode", "payment")
	form.Set("success_url", firstNonEmpty(req.SuccessURL, s.successURL))
	form.Set("cancel_url", firstNonEmpty(req.CancelURL, s.cancelURL))
	form.Set("client_reference_id", req.Reference)
	form.Set("metadata[reference]", req.Reference)
	form.Set("line_items[0][quantity]", "1")
	form.Set("line_items[0][price_data][currency]", currency)
	form.Set("line_items[0][price_data][unit_amount]", strconv.FormatInt(minor, 10))
	form.Set("line_items[0][price_data][product_data][name]", name)
	for key, value := range req.Metadata {
		if key == "" || key == "reference" {
			continue
		}
		form.Set("metadata["+key+"]", value)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, s.apiBase+"/checkout/sessions", strings.NewReader(form.Encode()))
	if err != nil {
		return InitResult{}, err
	}
	httpReq.Header.Set("Authorization", "Bearer "+s.secretKey)
	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := s.client.Do(httpReq)
	if err != nil {
		return InitResult{}, err
	}
	defer func() { _ = resp.Body.Close() }()

	raw, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return InitResult{}, err
	}

	var parsed struct {
		ID    string `json:"id"`
		URL   string `json:"url"`
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return InitResult{}, fmt.Errorf("stripe: unexpected response (status %d)", resp.StatusCode)
	}
	if resp.StatusCode >= 300 || parsed.URL == "" {
		msg := parsed.Error.Message
		if msg == "" {
			msg = fmt.Sprintf("checkout session failed (status %d)", resp.StatusCode)
		}
		return InitResult{}, fmt.Errorf("stripe: %s", msg)
	}

	return InitResult{Provider: s.Name(), RedirectURL: parsed.URL}, nil
}

// Verify checks the webhook signature and returns the verified outcome for a
// completed, paid Checkout Session.
func (s *Stripe) Verify(_ context.Context, req VerifyRequest) (VerifyResult, error) {
	if s.webhookSecret == "" {
		return VerifyResult{}, ErrNotConfigured
	}
	if err := verifyStripeSignature(req.RawBody, req.Signature, s.webhookSecret); err != nil {
		return VerifyResult{}, err
	}

	var event struct {
		Type string `json:"type"`
		Data struct {
			Object struct {
				ID                string            `json:"id"`
				PaymentStatus     string            `json:"payment_status"`
				PaymentIntent     string            `json:"payment_intent"`
				AmountTotal       int64             `json:"amount_total"`
				Currency          string            `json:"currency"`
				ClientReferenceID string            `json:"client_reference_id"`
				Metadata          map[string]string `json:"metadata"`
			} `json:"object"`
		} `json:"data"`
	}
	if err := json.Unmarshal(req.RawBody, &event); err != nil {
		return VerifyResult{}, fmt.Errorf("stripe: invalid event payload")
	}

	obj := event.Data.Object
	reference := obj.Metadata["reference"]
	if reference == "" {
		reference = obj.ClientReferenceID
	}

	result := VerifyResult{
		Reference:             reference,
		ProviderTransactionID: obj.PaymentIntent,
		Currency:              obj.Currency,
	}
	if event.Type == "checkout.session.completed" && obj.PaymentStatus == "paid" {
		result.Paid = true
		result.Amount = minorToMajor(obj.AmountTotal, obj.Currency)
	}
	return result, nil
}

// verifyStripeSignature validates the `Stripe-Signature` header (t=…,v1=…).
func verifyStripeSignature(payload []byte, header, secret string) error {
	if strings.TrimSpace(header) == "" {
		return errors.New("stripe: missing webhook signature")
	}

	var timestamp string
	var signatures []string
	for _, part := range strings.Split(header, ",") {
		kv := strings.SplitN(strings.TrimSpace(part), "=", 2)
		if len(kv) != 2 {
			continue
		}
		switch kv[0] {
		case "t":
			timestamp = kv[1]
		case "v1":
			signatures = append(signatures, kv[1])
		}
	}
	if timestamp == "" || len(signatures) == 0 {
		return errors.New("stripe: malformed webhook signature")
	}

	ts, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return errors.New("stripe: invalid webhook timestamp")
	}
	if delta := time.Since(time.Unix(ts, 0)); delta > stripeToleranceWindow || delta < -stripeToleranceWindow {
		return errors.New("stripe: webhook timestamp outside tolerance")
	}

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(timestamp))
	mac.Write([]byte("."))
	mac.Write(payload)
	expected := mac.Sum(nil)

	for _, signature := range signatures {
		got, err := hex.DecodeString(signature)
		if err != nil {
			continue
		}
		if hmac.Equal(expected, got) {
			return nil
		}
	}
	return errors.New("stripe: webhook signature verification failed")
}

// majorToMinor converts a decimal amount string ("39.99") into integer minor
// units (3999). Never uses floating point.
func majorToMinor(amount, currency string) (int64, error) {
	amount = strings.TrimSpace(amount)
	if amount == "" {
		return 0, errors.New("stripe: empty amount")
	}

	negative := strings.HasPrefix(amount, "-")
	amount = strings.TrimPrefix(amount, "-")

	wholePart := amount
	fracPart := ""
	if dot := strings.IndexByte(amount, '.'); dot >= 0 {
		wholePart = amount[:dot]
		fracPart = amount[dot+1:]
	}
	if wholePart == "" {
		wholePart = "0"
	}

	whole, err := strconv.ParseInt(wholePart, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("stripe: invalid amount %q", amount)
	}

	var minor int64
	if zeroDecimalCurrencies[strings.ToLower(currency)] {
		minor = whole
	} else {
		if len(fracPart) > 2 {
			fracPart = fracPart[:2]
		}
		for len(fracPart) < 2 {
			fracPart += "0"
		}
		frac, err := strconv.ParseInt(fracPart, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("stripe: invalid amount %q", amount)
		}
		minor = whole*100 + frac
	}

	if negative {
		minor = -minor
	}
	return minor, nil
}

// minorToMajor converts integer minor units back to a decimal string.
func minorToMajor(minor int64, currency string) string {
	if zeroDecimalCurrencies[strings.ToLower(currency)] {
		return strconv.FormatInt(minor, 10)
	}
	whole := minor / 100
	frac := minor % 100
	if frac < 0 {
		frac = -frac
	}
	return fmt.Sprintf("%d.%02d", whole, frac)
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}
