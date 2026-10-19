package cart

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"strings"

	"github.com/md-siam-kazi/e-go-backend/pkg/apperr"
)

// Service owns the cart business logic (Handler → Service → Repository).
type Service struct {
	repo     CartRepository
	currency string
}

func NewService(repo CartRepository, currency string) *Service {
	currency = strings.ToLower(strings.TrimSpace(currency))
	if currency == "" {
		currency = "usd"
	}
	return &Service{repo: repo, currency: currency}
}

// Get returns the caller's cart, creating one for a new guest visitor.
func (s *Service) Get(ctx context.Context, userID, guestSession string) (CartResponse, Resolution, error) {
	c, res, err := s.resolve(ctx, userID, guestSession)
	if err != nil {
		return CartResponse{}, res, err
	}
	out, err := s.view(ctx, c)
	return out, res, err
}

// Add adds a product to the cart, resolving the variant, price and stock from
// the database.
func (s *Service) Add(ctx context.Context, userID, guestSession string, req AddItemRequest) (CartResponse, Resolution, error) {
	productID := strings.TrimSpace(req.ProductID)
	if productID == "" {
		return CartResponse{}, Resolution{}, errProductUnavailable
	}
	quantity := req.Quantity
	if quantity < 1 {
		return CartResponse{}, Resolution{}, errInvalidQuantity
	}
	if quantity > MaxLineQuantity {
		quantity = MaxLineQuantity
	}

	c, res, err := s.resolve(ctx, userID, guestSession)
	if err != nil {
		return CartResponse{}, res, err
	}

	variant, err := s.repo.ResolveVariant(ctx, productID, strings.TrimSpace(req.VariantID))
	if err != nil {
		return CartResponse{}, res, err
	}
	if variant.Available <= 0 {
		return CartResponse{}, res, errInsufficientStock
	}

	maxAllowed := int(variant.Available)
	if maxAllowed > MaxLineQuantity {
		maxAllowed = MaxLineQuantity
	}
	if err := s.repo.UpsertItem(ctx, c.ID, variant.VariantID, quantity, variant.UnitPrice, maxAllowed); err != nil {
		return CartResponse{}, res, err
	}

	out, err := s.view(ctx, c)
	return out, res, err
}

// SetQuantity sets a line's quantity (0 removes the line), clamped to stock.
func (s *Service) SetQuantity(ctx context.Context, userID, guestSession, itemID string, quantity *int) (CartResponse, Resolution, error) {
	if quantity == nil {
		return CartResponse{}, Resolution{}, errInvalidBody
	}
	qty := *quantity
	if qty < 0 {
		return CartResponse{}, Resolution{}, errInvalidQuantity
	}
	if qty > MaxLineQuantity {
		qty = MaxLineQuantity
	}

	c, res, err := s.resolve(ctx, userID, guestSession)
	if err != nil {
		return CartResponse{}, res, err
	}

	item, err := s.repo.FindItem(ctx, c.ID, itemID)
	if err != nil {
		return CartResponse{}, res, err
	}
	if qty > 0 && int64(qty) > item.Available {
		return CartResponse{}, res, errInsufficientStock
	}

	if err := s.repo.SetItemQuantity(ctx, c.ID, itemID, qty); err != nil {
		return CartResponse{}, res, err
	}

	out, err := s.view(ctx, c)
	return out, res, err
}

// Remove deletes a single line from the cart.
func (s *Service) Remove(ctx context.Context, userID, guestSession, itemID string) (CartResponse, Resolution, error) {
	c, res, err := s.resolve(ctx, userID, guestSession)
	if err != nil {
		return CartResponse{}, res, err
	}
	if err := s.repo.RemoveItem(ctx, c.ID, itemID); err != nil {
		return CartResponse{}, res, err
	}
	out, err := s.view(ctx, c)
	return out, res, err
}

// Clear removes every line from the cart.
func (s *Service) Clear(ctx context.Context, userID, guestSession string) (CartResponse, Resolution, error) {
	c, res, err := s.resolve(ctx, userID, guestSession)
	if err != nil {
		return CartResponse{}, res, err
	}
	if err := s.repo.ClearItems(ctx, c.ID); err != nil {
		return CartResponse{}, res, err
	}
	out, err := s.view(ctx, c)
	return out, res, err
}

// resolve finds (or creates) the right cart for the caller. An authenticated
// caller gets their user cart; a guest cart referenced by the session cookie is
// merged into it and the cookie is then cleared. A guest with no usable session
// gets a fresh session id so the handler can set the cookie.
func (s *Service) resolve(ctx context.Context, userID, guestSession string) (Cart, Resolution, error) {
	var res Resolution

	if userID != "" {
		c, found, err := s.repo.FindUserCart(ctx, userID)
		if err != nil {
			return Cart{}, res, err
		}
		if !found {
			c, err = s.repo.CreateCart(ctx, userID, "", s.currency)
			if err != nil {
				return Cart{}, res, err
			}
		}

		if guestSession != "" {
			guest, gFound, err := s.repo.FindSessionCart(ctx, guestSession)
			if err != nil {
				return Cart{}, res, err
			}
			if gFound && guest.ID != c.ID {
				if err := s.repo.MergeCarts(ctx, guest.ID, c.ID); err != nil {
					return Cart{}, res, err
				}
			}
			// Any guest session ends here: merged or not, the cookie is dropped.
			res.ClearSession = true
		}
		return c, res, nil
	}

	if guestSession != "" {
		c, found, err := s.repo.FindSessionCart(ctx, guestSession)
		if err != nil {
			return Cart{}, res, err
		}
		if found {
			return c, res, nil
		}
	}

	session, err := newSessionID()
	if err != nil {
		return Cart{}, res, apperr.ErrInternal.WithCause(err)
	}
	c, err := s.repo.CreateCart(ctx, "", session, s.currency)
	if err != nil {
		return Cart{}, res, err
	}
	res.NewSession = session
	return c, res, nil
}

// view loads a cart's lines and projects them into the API shape.
func (s *Service) view(ctx context.Context, c Cart) (CartResponse, error) {
	records, err := s.repo.ListItems(ctx, c.ID)
	if err != nil {
		return CartResponse{}, err
	}
	return toResponse(c, records), nil
}

func newSessionID() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}
