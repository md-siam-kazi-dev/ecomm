package payment

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"
)

func signStripe(secret string, payload []byte, ts int64) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(strconv.FormatInt(ts, 10)))
	mac.Write([]byte("."))
	mac.Write(payload)
	return fmt.Sprintf("t=%d,v1=%s", ts, hex.EncodeToString(mac.Sum(nil)))
}

func TestMajorToMinor(t *testing.T) {
	cases := []struct {
		amount   string
		currency string
		want     int64
	}{
		{"39.99", "usd", 3999},
		{"39.9", "usd", 3990},
		{"40", "usd", 4000},
		{"0.05", "usd", 5},
		{"1000", "jpy", 1000},
	}
	for _, c := range cases {
		got, err := majorToMinor(c.amount, c.currency)
		if err != nil || got != c.want {
			t.Errorf("majorToMinor(%q,%q) = %d,%v want %d", c.amount, c.currency, got, err, c.want)
		}
	}
	if _, err := majorToMinor("abc", "usd"); err == nil {
		t.Error("expected error for invalid amount")
	}
}

func TestMinorToMajor(t *testing.T) {
	if got := minorToMajor(3999, "usd"); got != "39.99" {
		t.Errorf("got %q want 39.99", got)
	}
	if got := minorToMajor(1000, "jpy"); got != "1000" {
		t.Errorf("got %q want 1000", got)
	}
}

func TestVerifyStripeSignature(t *testing.T) {
	const secret = "whsec_test"
	payload := []byte(`{"type":"checkout.session.completed"}`)

	valid := signStripe(secret, payload, time.Now().Unix())
	if err := verifyStripeSignature(payload, valid, secret); err != nil {
		t.Fatalf("valid signature rejected: %v", err)
	}
	if err := verifyStripeSignature(payload, valid, "wrong-secret"); err == nil {
		t.Error("wrong secret should fail")
	}
	stale := signStripe(secret, payload, time.Now().Add(-time.Hour).Unix())
	if err := verifyStripeSignature(payload, stale, secret); err == nil {
		t.Error("stale signature should fail")
	}
	if err := verifyStripeSignature(payload, "", secret); err == nil {
		t.Error("missing signature should fail")
	}
}

func TestStripeInitiate_FormAndRedirect(t *testing.T) {
	var gotAuth, gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
		_, _ = w.Write([]byte(`{"id":"cs_1","url":"https://checkout.stripe.com/pay/cs_1"}`))
	}))
	defer srv.Close()

	s := NewStripe("sk_test_123", "whsec", "usd", "https://shop.test/order/success", "https://shop.test/checkout")
	s.apiBase = srv.URL

	res, err := s.Initiate(context.Background(), InitRequest{Reference: "SK-ABC123", Amount: "39.99", Currency: "usd"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.RedirectURL != "https://checkout.stripe.com/pay/cs_1" {
		t.Errorf("unexpected redirect: %q", res.RedirectURL)
	}
	if gotAuth != "Bearer sk_test_123" {
		t.Errorf("unexpected auth header: %q", gotAuth)
	}
	for _, want := range []string{"mode=payment", "=3999", "SK-ABC123"} {
		if !strings.Contains(gotBody, want) {
			t.Errorf("form body missing %q: %s", want, gotBody)
		}
	}
}

func TestStripeInitiate_NotConfigured(t *testing.T) {
	s := NewStripe("", "whsec", "usd", "", "")
	if _, err := s.Initiate(context.Background(), InitRequest{Reference: "SK-1", Amount: "1.00"}); err != ErrNotConfigured {
		t.Fatalf("expected ErrNotConfigured, got %v", err)
	}
}

func TestStripeVerify_PaidEvent(t *testing.T) {
	const secret = "whsec_test"
	payload := []byte(`{"type":"checkout.session.completed","data":{"object":{"id":"cs_1","payment_status":"paid","payment_intent":"pi_1","amount_total":3999,"currency":"usd","metadata":{"reference":"SK-ABC"}}}}`)
	header := signStripe(secret, payload, time.Now().Unix())

	s := NewStripe("sk_test", secret, "usd", "", "")
	res, err := s.Verify(context.Background(), VerifyRequest{RawBody: payload, Signature: header})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.Paid || res.Reference != "SK-ABC" || res.ProviderTransactionID != "pi_1" || res.Amount != "39.99" {
		t.Fatalf("unexpected verify result: %+v", res)
	}
}

func TestStripeVerify_UnpaidIsNotPaid(t *testing.T) {
	const secret = "whsec_test"
	payload := []byte(`{"type":"checkout.session.completed","data":{"object":{"payment_status":"unpaid","metadata":{"reference":"SK-ABC"}}}}`)
	header := signStripe(secret, payload, time.Now().Unix())

	s := NewStripe("sk_test", secret, "usd", "", "")
	res, err := s.Verify(context.Background(), VerifyRequest{RawBody: payload, Signature: header})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Paid {
		t.Error("unpaid session must not be Paid")
	}
}

func TestRegistry_RegistersStripe(t *testing.T) {
	r := NewRegistryFromConfig(Config{Provider: "stripe", SecretKey: "sk_test"})
	p, err := r.Get("stripe")
	if err != nil || p.Name() != "stripe" {
		t.Fatalf("stripe not registered: %v", err)
	}

	empty := NewRegistryFromConfig(Config{Provider: "stripe"})
	if _, err := empty.Get("stripe"); err == nil {
		t.Error("no key should not register stripe")
	}
}
