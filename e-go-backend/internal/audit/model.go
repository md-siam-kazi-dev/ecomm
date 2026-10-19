// Package audit records and serves the administrative audit trail. It is the
// single writer for audit_logs, replacing the ad-hoc inserts that previously
// lived in individual modules.
package audit

// Actor roles describe who performed an action. They are a small, closed set so
// the admin UI can filter reliably.
const (
	RoleAdmin    = "admin"
	RoleVendor   = "vendor"
	RoleCustomer = "customer"
	RoleSystem   = "system"
)

// Result is the outcome of the action being audited.
const (
	ResultSuccess = "success"
	ResultFailure = "failure"
)

// Entity types are the kinds of resources an event can target.
const (
	EntityUser             = "user"
	EntityStore            = "store"
	EntityStoreApplication = "store_application"
	EntityProduct          = "product"
	EntityVariant          = "variant"
	EntityOrder            = "order"
	EntityPayment          = "payment"
	EntityReview           = "review"
	EntityQuestion         = "question"
	EntityInventory        = "inventory"
	EntitySettings         = "settings"
)

// Stable, dotted action identifiers. These are part of the public filtering
// contract and must not be renamed without a migration.
const (
	ActionAuthRegistered    = "auth.registered"
	ActionAuthLoginSuccess  = "auth.login_succeeded"
	ActionAuthLoginFailed   = "auth.login_failed"
	ActionAuthLoggedOutAll  = "auth.logged_out_all"
	ActionAuthPasswordReset = "auth.password_reset"
	ActionAuthEmailVerified = "auth.email_verified"

	ActionAdminProductApproved         = "admin.product_approved"
	ActionAdminProductRejected         = "admin.product_rejected"
	ActionAdminProductChangesRequested = "admin.product_changes_requested"
	ActionAdminStoreApproved           = "admin.store_approved"
	ActionAdminStoreRejected           = "admin.store_rejected"
	ActionAdminSettingsUpdated         = "admin.settings_updated"

	ActionVendorProductCreated     = "vendor.product_created"
	ActionVendorProductUpdated     = "vendor.product_updated"
	ActionVendorInventoryAdjusted  = "vendor.inventory_adjusted"
	ActionVendorOrderStatusChanged = "vendor.order_status_changed"
	ActionVendorShipmentUpdated    = "vendor.shipment_updated"
	ActionVendorReviewReplied      = "vendor.review_replied"
	ActionVendorQuestionAnswered   = "vendor.question_answered"

	ActionStoreApplicationSubmitted = "store_application.submitted"
	ActionStoreApplicationApproved  = "store_application.approved"
	ActionStoreApplicationRejected  = "store_application.rejected"
	ActionStoreActivated            = "store.activated"

	ActionOrderCreated  = "order.created"
	ActionPaymentPaid   = "payment.paid"
	ActionPaymentFailed = "payment.failed"
)

// Event is a single audit record. ActorID and EntityID may be nil for
// system-generated events or actions against an entity that has since been
// deleted.
type Event struct {
	ActorID    *string
	ActorRole  string
	Action     string
	EntityType string
	EntityID   *string
	OldValues  map[string]any
	NewValues  map[string]any
	Result     string
	Reason     string
	Metadata   map[string]any

	// Request metadata. Normally left empty by callers and filled from the
	// request context by Record/Log.
	IP        string
	UserAgent string
	RequestID string
}

// labels maps action identifiers to human-readable display labels. Unknown
// actions fall back to a de-slugged form so newly added actions still render.
var labels = map[string]string{
	ActionAuthRegistered:    "Account registered",
	ActionAuthLoginSuccess:  "Signed in",
	ActionAuthLoginFailed:   "Failed sign-in",
	ActionAuthLoggedOutAll:  "Signed out of all sessions",
	ActionAuthPasswordReset: "Password reset",
	ActionAuthEmailVerified: "Email verified",

	ActionAdminProductApproved:         "Approved product",
	ActionAdminProductRejected:         "Rejected product",
	ActionAdminProductChangesRequested: "Requested product changes",
	ActionAdminStoreApproved:           "Approved store",
	ActionAdminStoreRejected:           "Rejected store",
	ActionAdminSettingsUpdated:         "Updated storefront settings",

	ActionVendorProductCreated:     "Created product",
	ActionVendorProductUpdated:     "Updated product",
	ActionVendorInventoryAdjusted:  "Adjusted stock",
	ActionVendorOrderStatusChanged: "Updated order status",
	ActionVendorShipmentUpdated:    "Updated shipment",
	ActionVendorReviewReplied:      "Replied to review",
	ActionVendorQuestionAnswered:   "Answered question",

	ActionStoreApplicationSubmitted: "Submitted store application",
	ActionStoreApplicationApproved:  "Approved store application",
	ActionStoreApplicationRejected:  "Rejected store application",
	ActionStoreActivated:            "Activated store",

	ActionOrderCreated:  "Placed order",
	ActionPaymentPaid:   "Payment captured",
	ActionPaymentFailed: "Payment failed",
}

// Label returns the display label for an action.
func Label(action string) string {
	if l, ok := labels[action]; ok {
		return l
	}
	return humanize(action)
}

// LabeledAction pairs an action identifier with its label so the frontend can
// build filter menus without hard-coding the catalogue.
type LabeledAction struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

// Actions returns the full catalogue of known actions, ordered for display.
func Actions() []LabeledAction {
	ordered := []string{
		ActionAuthRegistered,
		ActionAuthLoginSuccess,
		ActionAuthLoginFailed,
		ActionAuthLoggedOutAll,
		ActionAuthPasswordReset,
		ActionAuthEmailVerified,
		ActionAdminProductApproved,
		ActionAdminProductRejected,
		ActionAdminProductChangesRequested,
		ActionAdminStoreApproved,
		ActionAdminStoreRejected,
		ActionAdminSettingsUpdated,
		ActionStoreApplicationSubmitted,
		ActionStoreApplicationApproved,
		ActionStoreApplicationRejected,
		ActionStoreActivated,
		ActionVendorProductCreated,
		ActionVendorProductUpdated,
		ActionVendorInventoryAdjusted,
		ActionVendorOrderStatusChanged,
		ActionVendorShipmentUpdated,
		ActionVendorReviewReplied,
		ActionVendorQuestionAnswered,
		ActionOrderCreated,
		ActionPaymentPaid,
		ActionPaymentFailed,
	}
	out := make([]LabeledAction, 0, len(ordered))
	for _, a := range ordered {
		out = append(out, LabeledAction{Value: a, Label: Label(a)})
	}
	return out
}

// humanize turns "vendor.order_status_changed" into "Vendor order status
// changed" as a fallback label.
func humanize(action string) string {
	if action == "" {
		return "Unknown action"
	}
	runes := []rune(action)
	for i := 0; i < len(runes); i++ {
		switch runes[i] {
		case '.', '_':
			runes[i] = ' '
		}
	}
	for i := 0; i < len(runes); i++ {
		if i == 0 && runes[i] >= 'a' && runes[i] <= 'z' {
			runes[i] -= 'a' - 'A'
		}
	}
	return string(runes)
}
