package order

import (
	"context"
	"crypto/rand"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/md-siam-kazi/e-go-backend/pkg/apperr"
	"github.com/md-siam-kazi/e-go-backend/pkg/payment"
	"github.com/md-siam-kazi/e-go-backend/pkg/response"
)

const (
	defaultLimit  = 10
	maxLimit      = 50
	maxOrderItems = 50
)

var allowedLimits = map[int]struct{}{10: {}, 20: {}, 50: {}}

// Config carries order pricing + payment wiring.
type Config struct {
	Currency              string
	StandardShipping      float64
	ExpressShipping       float64
	FreeShippingThreshold float64
	ProviderName          string
	SuccessURL            string
	CancelURL             string
}

// Service owns the order business logic (Handler → Service → Repository).
type Service struct {
	repo     OrderRepository
	registry *payment.Registry
	cfg      Config
}

func NewService(repo OrderRepository, registry *payment.Registry, cfg Config) *Service {
	return &Service{repo: repo, registry: registry, cfg: cfg}
}

// provider resolves the configured payment provider, reporting whether a real
// one is available.
func (s *Service) provider() (payment.Provider, bool) {
	if s.registry == nil {
		return payment.NotConfigured{}, false
	}
	p, err := s.registry.Get(s.cfg.ProviderName)
	if err != nil || p.Name() == "" {
		return payment.NotConfigured{}, false
	}
	return p, true
}

// Create validates and persists an order, then starts payment when required.
func (s *Service) Create(ctx context.Context, userID string, req CreateOrderRequest) (CreateOrderResponse, error) {
	// The checkout is built from the server-side cart unless the client sends
	// explicit items. This keeps prices and line items authoritative.
	usedCart := false
	if len(req.Items) == 0 {
		cartItems, err := s.repo.CartItemsForUser(ctx, userID)
		if err != nil {
			return CreateOrderResponse{}, err
		}
		if len(cartItems) == 0 {
			return CreateOrderResponse{}, errEmptyItems
		}
		req.Items = cartItems
		usedCart = true
	}
	if len(req.Items) > maxOrderItems {
		return CreateOrderResponse{}, apperr.ErrValidation
	}

	method := strings.ToLower(strings.TrimSpace(req.PaymentMethod))
	if method != MethodCard && method != MethodCOD {
		return CreateOrderResponse{}, errBadPayment
	}
	delivery := strings.ToLower(strings.TrimSpace(req.DeliveryMethod))
	if delivery != DeliveryStandard && delivery != DeliveryExpress {
		return CreateOrderResponse{}, errBadDelivery
	}

	address, err := validateAddress(req.ShippingAddress)
	if err != nil {
		return CreateOrderResponse{}, err
	}

	online := method != MethodCOD
	var provider payment.Provider
	if online {
		p, ok := s.provider()
		if !ok {
			return CreateOrderResponse{}, errPaymentNotConfigured
		}
		provider = p
	}

	inputs := make([]OrderItemInput, 0, len(req.Items))
	for _, item := range req.Items {
		if item.Quantity < 1 {
			return CreateOrderResponse{}, apperr.ErrValidation
		}
		if strings.TrimSpace(item.ProductID) == "" {
			return CreateOrderResponse{}, errUnavailable
		}
		inputs = append(inputs, OrderItemInput{
			ProductID: strings.TrimSpace(item.ProductID),
			VariantID: strings.TrimSpace(item.VariantID),
			Quantity:  item.Quantity,
		})
	}

	// Prices and stock come from the DB — never the client.
	resolved, err := s.repo.ResolveVariants(ctx, inputs)
	if err != nil {
		return CreateOrderResponse{}, err
	}

	quantities := make([]int, len(resolved))
	subtotal := 0.0
	for i, line := range resolved {
		quantity := inputs[i].Quantity
		if int64(quantity) > line.Available {
			return CreateOrderResponse{}, errInsufficientStock
		}
		quantities[i] = quantity
		subtotal += round2(line.UnitPrice * float64(quantity))
	}
	subtotal = round2(subtotal)
	shipping := s.shippingCost(delivery, subtotal)
	const tax = 0.0
	grandTotal := round2(subtotal + shipping + tax)

	currency := strings.ToLower(strings.TrimSpace(s.cfg.Currency))
	if currency == "" {
		currency = "usd"
	}

	var created Order
	for attempt := 0; attempt < 5; attempt++ {
		created, err = s.repo.CreateOrder(ctx, OrderParams{
			UserID:        userID,
			OrderNumber:   generateOrderNumber(),
			Currency:      currency,
			Items:         resolved,
			Quantities:    quantities,
			Subtotal:      subtotal,
			ShippingTotal: shipping,
			TaxTotal:      tax,
			GrandTotal:    grandTotal,
			Shipping:      address,
		})
		if err == nil {
			break
		}
		if errors.Is(err, errOrderNumberTaken) {
			continue
		}
		return CreateOrderResponse{}, err
	}
	if err != nil {
		return CreateOrderResponse{}, apperr.ErrInternal.WithCause(err)
	}

	out := CreateOrderResponse{Order: toOrderResponse(created)}
	if !online {
		s.convertCart(ctx, userID, usedCart)
		return out, nil
	}

	if _, err := s.repo.CreatePayment(ctx, created.ID, provider.Name(), grandTotal, currency); err != nil {
		// Roll the order back so a failed payment never leaves an orphan order.
		_ = s.repo.DeleteOrder(ctx, created.ID)
		return CreateOrderResponse{}, err
	}

	init, err := provider.Initiate(ctx, payment.InitRequest{
		Reference:  created.OrderNumber,
		Amount:     formatAmount(grandTotal),
		Currency:   currency,
		Metadata:   map[string]string{"description": "Order " + created.OrderNumber},
		SuccessURL: s.cfg.SuccessURL + "?order=" + created.ID,
		CancelURL:  s.cfg.CancelURL,
	})
	if err != nil {
		_ = s.repo.DeleteOrder(ctx, created.ID)
		if errors.Is(err, payment.ErrNotConfigured) {
			return CreateOrderResponse{}, errPaymentNotConfigured
		}
		return CreateOrderResponse{}, apperr.ErrInternal.WithCause(err)
	}

	out.CheckoutURL = init.RedirectURL
	s.convertCart(ctx, userID, usedCart)
	return out, nil
}

// convertCart empties the user's server-side cart after a successful checkout.
// Best-effort: the order already exists, so a cart hiccup must not fail it.
func (s *Service) convertCart(ctx context.Context, userID string, used bool) {
	if !used {
		return
	}
	_ = s.repo.ConvertUserCart(ctx, userID)
}

// List returns a page of the caller's orders.
func (s *Service) List(ctx context.Context, userID string, page, limit int) ([]OrderResponse, response.Pagination, error) {
	if page < 1 {
		page = 1
	}
	if _, ok := allowedLimits[limit]; !ok {
		limit = defaultLimit
	}

	orders, total, err := s.repo.OrdersForUser(ctx, userID, page, limit)
	if err != nil {
		return nil, response.Pagination{}, err
	}

	items := make([]OrderResponse, 0, len(orders))
	for _, o := range orders {
		items = append(items, toOrderResponse(o))
	}
	return items, response.Pagination{Page: page, Limit: limit, Total: total}, nil
}

// Get returns one of the caller's orders.
func (s *Service) Get(ctx context.Context, userID, id string) (OrderResponse, error) {
	o, err := s.repo.OrderByID(ctx, userID, id)
	if err != nil {
		return OrderResponse{}, err
	}
	return toOrderResponse(o), nil
}

// HandleWebhook verifies a provider callback and, on a paid event, marks the
// payment paid and confirms the order. Idempotent for repeated deliveries.
func (s *Service) HandleWebhook(ctx context.Context, providerName string, req payment.VerifyRequest) error {
	if s.registry == nil {
		return errPaymentNotConfigured
	}
	provider, err := s.registry.Get(providerName)
	if err != nil || provider.Name() == "" {
		return errPaymentNotConfigured
	}

	result, err := provider.Verify(ctx, req)
	if err != nil {
		// Bad signature / payload — reject so the provider retries or alerts.
		return apperr.ErrValidation.WithCause(err)
	}
	if !result.Paid {
		// Nothing to do (we only act on a completed, paid checkout).
		return nil
	}
	if strings.TrimSpace(result.Reference) == "" {
		return apperr.ErrValidation
	}

	record, err := s.repo.PaymentByReference(ctx, result.Reference)
	if err != nil {
		return err
	}
	if _, _, err := s.repo.MarkPaid(ctx, result.Reference, result.ProviderTransactionID); err != nil {
		return err
	}
	_ = s.repo.InsertPaymentTransaction(ctx, record.ID, result.ProviderTransactionID, record.Amount, record.Currency)
	if err := s.repo.ConfirmPaidOrder(ctx, result.Reference); err != nil {
		return err
	}
	return nil
}

func (s *Service) shippingCost(delivery string, subtotal float64) float64 {
	if delivery == DeliveryExpress {
		return round2(s.cfg.ExpressShipping)
	}
	if subtotal >= s.cfg.FreeShippingThreshold {
		return 0
	}
	return round2(s.cfg.StandardShipping)
}

func validateAddress(a AddressRequest) (AddressParams, error) {
	addr := AddressParams{
		FullName:   strings.TrimSpace(a.FullName),
		Phone:      strings.TrimSpace(a.Phone),
		Line1:      strings.TrimSpace(a.Line1),
		Line2:      strings.TrimSpace(a.Line2),
		City:       strings.TrimSpace(a.City),
		State:      strings.TrimSpace(a.State),
		PostalCode: strings.TrimSpace(a.PostalCode),
		Country:    strings.TrimSpace(a.Country),
	}
	if addr.FullName == "" || addr.Line1 == "" || addr.City == "" || addr.Country == "" {
		return AddressParams{}, errNoAddress
	}
	return addr, nil
}

const orderCodeAlphabet = "ABCDEFGHJKMNPQRSTUVWXYZ23456789"

func generateOrderNumber() string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		b = []byte(strconv.FormatInt(time.Now().UnixNano(), 36))
	}
	for i := range b {
		b[i] = orderCodeAlphabet[int(b[i])%len(orderCodeAlphabet)]
	}
	return "SK-" + string(b)
}

func formatAmount(total float64) string {
	return strconv.FormatFloat(total, 'f', 2, 64)
}
