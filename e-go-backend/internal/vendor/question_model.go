package vendor

import (
	"net/http"
	"time"

	"github.com/md-siam-kazi/e-go-backend/pkg/apperr"
)

// errQuestionNotFound is returned when a question does not exist or is not on
// one of the vendor's products.
var errQuestionNotFound = apperr.New(http.StatusNotFound, "QUESTION_NOT_FOUND", "Question not found")

// errInvalidAnswer is returned for an empty or oversized vendor answer.
var errInvalidAnswer = apperr.New(
	http.StatusUnprocessableEntity, "VALIDATION_ERROR",
	"Enter an answer between 1 and 1000 characters",
)

// VendorQuestionRecord is one question on the vendor's products, with the
// vendor's current official answer if any.
type VendorQuestionRecord struct {
	ID              string
	ProductID       string
	ProductName     string
	ProductSlug     string
	ProductImage    string
	Author          string
	Question        string
	CreatedAt       time.Time
	Answered        bool
	Answer          string
	AnswerCreatedAt *time.Time
}
