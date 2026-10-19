package follow

import (
	"net/http"

	"github.com/md-siam-kazi/e-go-backend/pkg/apperr"
)

// errStoreNotFound is returned when a slug does not match an active store.
var errStoreNotFound = apperr.New(http.StatusNotFound, "STORE_NOT_FOUND", "Store not found")
