# Production E-commerce Backend — PRD

## 1. Project Overview

Build a production-ready, scalable e-commerce backend using **Go** and **PostgreSQL**.

The backend will expose a secure REST API consumed by a Next.js web application.

The platform supports three primary roles:

1. **User** — browses products, manages cart, purchases products, tracks orders, reviews products, and interacts with vendors.
2. **Vendor** — manages their products, inventory, orders, shipping, reviews, revenue, payments, and profit.
3. **Admin** — manages the entire marketplace, approves vendors/products, manages users, monitors orders, revenue, vendor revenue, profit, inventory, payments, and platform operations.

The backend must be designed as a **modular monolith** initially, with clean boundaries so individual modules can later be extracted into services if required.

---

# 2. Goals

The backend must provide:

* Secure JWT authentication.
* Role-based authorization.
* User registration/login.
* Vendor registration and approval.
* Admin management.
* Product management.
* Product approval workflow.
* Categories and brands.
* Product variants.
* Product images.
* Inventory management.
* Cart management.
* Wishlist.
* Checkout.
* Orders.
* Payments.
* Payment status tracking.
* Shipping.
* Order tracking.
* Product reviews.
* Vendor replies to reviews/questions.
* Coupons and discounts.
* Vendor revenue dashboard.
* Vendor profit calculation.
* Admin revenue dashboard.
* Admin profit dashboard.
* Vendor commission/platform revenue tracking.
* Refund support.
* Returns support.
* Notifications.
* Audit logging.
* Rate limiting.
* API validation.
* Database transactions.
* Structured error handling.
* API documentation.

---

# 3. Technology Stack

## Backend

```text
Language: Go
HTTP: net/http
Router: chi or Go standard ServeMux
Architecture: Modular Monolith
API: REST
Authentication: JWT
Password hashing: Argon2id or bcrypt
```

## Database

```text
PostgreSQL
Provider: Neon PostgreSQL
```

Recommended database access:

```text
database/sql
+
pgx
+
sqlc
```

Do not put SQL directly inside HTTP handlers.

Recommended architecture:

```text
Handler
   ↓
Service
   ↓
Repository
   ↓
PostgreSQL
```

## Cache

```text
Redis
```

Use Redis for:

* Rate limiting
* Short-lived cache
* Session/token infrastructure if required
* Product caching
* Frequently accessed category data
* Distributed locks
* Background-job coordination

PostgreSQL remains the source of truth.

## Object Storage

Product and review images must not be stored directly in PostgreSQL.

Use S3-compatible object storage.

Store only:

```text
object_key
url
metadata
```

in PostgreSQL.

---

# 4. Architecture

Use a modular architecture.

```text
cmd/
└── api/
    └── main.go

internal/
├── auth/
├── users/
├── roles/
├── vendors/
├── admins/
├── products/
├── categories/
├── brands/
├── inventory/
├── carts/
├── wishlists/
├── orders/
├── payments/
├── shipping/
├── reviews/
├── coupons/
├── notifications/
├── reports/
├── audit/
└── common/

pkg/
├── jwt/
├── password/
├── logger/
├── validator/
├── response/
└── middleware/

migrations/

docs/

tests/
```

Every business domain should contain:

```text
handler.go
service.go
repository.go
model.go
dto.go
routes.go
```

For larger modules:

```text
products/
├── handler/
├── service/
├── repository/
├── model/
├── dto/
└── routes.go
```

---

# 5. Core Roles

## 5.1 User

A user can:

* Register.
* Login.
* Logout.
* View products.
* Search products.
* Filter products.
* View product details.
* Add products to cart.
* Update cart.
* Remove cart items.
* Add products to wishlist.
* Checkout.
* Make payments.
* View orders.
* Cancel eligible orders.
* Request returns.
* Request refunds where applicable.
* Track shipments.
* Review purchased products.
* Upload review images.
* Ask product questions.
* Read vendor replies.
* Manage profile.
* Manage addresses.
* View notifications.

---

# 6. Vendor

Vendor functionality requires admin approval.

A vendor can:

* Register as vendor.
* Submit business information.
* Upload required documents where applicable.
* View vendor approval status.
* Create products.
* Edit products.
* Submit products for approval.
* View product approval status.
* Manage product variants.
* Manage product images.
* Manage stock.
* View inventory.
* Receive orders.
* View customer order information required for fulfillment.
* Process orders.
* Mark orders as packed.
* Mark orders as shipped.
* Add tracking information.
* View payment status.
* View sales.
* View revenue.
* View profit.
* View commissions/platform fees.
* View payouts.
* Respond to product reviews.
* Answer product questions.
* View product performance.
* View sales reports.

Vendor must not be able to:

* Access another vendor's products.
* Access another vendor's inventory.
* Modify another vendor's orders.
* See private customer information unrelated to their own orders.
* Modify payment records.
* Approve their own products.
* Approve themselves as vendors.
* Modify admin settings.
* Modify platform commission configuration.

---

# 7. Admin

Admin has platform-wide access.

Admin can:

* Manage users.
* Manage vendors.
* Approve vendors.
* Reject vendors.
* Suspend vendors.
* Activate vendors.
* Approve products.
* Reject products.
* Suspend products.
* Manage categories.
* Manage brands.
* Manage platform commission.
* View all orders.
* View total orders.
* View total revenue.
* View vendor revenue.
* View platform revenue.
* View total profit.
* View vendor profit.
* View inventory.
* View payment status.
* View refunds.
* Manage coupons.
* Manage shipping configuration.
* View reviews.
* Moderate reviews.
* Moderate product questions.
* View audit logs.
* Manage admin users.
* View system analytics.

---

# 8. Authentication

All protected API endpoints must require JWT authentication.

Authentication flow:

```text
Client
   ↓
POST /api/v1/auth/login
   ↓
Validate credentials
   ↓
Generate JWT
   ↓
Return access token
```

Recommended token structure:

```json
{
  "sub": "user_uuid",
  "role": "vendor",
  "iat": 123456789,
  "exp": 123456789
}
```

Do not trust role information supplied by the client outside the signed JWT.

The backend must retrieve the authenticated identity from the verified JWT.

---

# 9. JWT Security

JWT middleware must:

1. Read Authorization header.
2. Validate `Bearer` format.
3. Verify signature.
4. Validate expiration.
5. Validate issuer.
6. Validate audience.
7. Extract user ID.
8. Load current user state when necessary.
9. Reject disabled/suspended users.
10. Add authenticated user context to the request.

Example:

```http
Authorization: Bearer <JWT>
```

Invalid token:

```http
401 Unauthorized
```

Valid token but insufficient permission:

```http
403 Forbidden
```

---

# 10. JWT Token Strategy

Use:

```text
Short-lived access token
+
Refresh token
```

Recommended:

```text
Access token: 15–30 minutes
Refresh token: 7–30 days
```

Refresh tokens should be stored securely and revocable.

Never store raw refresh tokens if avoidable.

Store a hash/reference.

---

# 11. Authentication APIs

## Register

```http
POST /api/v1/auth/register
```

Request:

```json
{
  "name": "John Doe",
  "email": "john@example.com",
  "password": "StrongPassword123",
  "phone": "+8801XXXXXXXXX"
}
```

Response:

```json
{
  "success": true,
  "data": {
    "user": {},
    "access_token": "...",
    "refresh_token": "..."
  }
}
```

---

## Login

```http
POST /api/v1/auth/login
```

---

## Refresh Token

```http
POST /api/v1/auth/refresh
```

---

## Logout

```http
POST /api/v1/auth/logout
```

---

## Current User

```http
GET /api/v1/auth/me
```

---

# 12. Database Schema

The initial production database should contain approximately 40–50 tables.

Major domains:

```text
Users
Roles
Vendors
Products
Categories
Brands
Inventory
Cart
Wishlist
Orders
Payments
Shipping
Reviews
Coupons
Returns
Notifications
Analytics
Audit Logs
```

---

# 13. Users

## users

```text
id UUID PRIMARY KEY
email
phone
password_hash
status
email_verified
phone_verified
created_at
updated_at
last_login_at
```

Possible status:

```text
active
suspended
blocked
pending
```

---

# 14. User Profiles

```text
user_profiles
```

Fields:

```text
id
user_id
first_name
last_name
avatar_url
date_of_birth
created_at
updated_at
```

---

# 15. Vendors

Vendor registration must be separate from the normal user account.

```text
vendors
```

Fields:

```text
id
user_id
business_name
business_slug
business_email
business_phone
description
logo_url
status
approval_status
approved_by
approved_at
created_at
updated_at
```

Vendor status:

```text
pending
approved
rejected
suspended
```

Vendor cannot create publicly visible products until approved.

---

# 16. Vendor Documents

Optional depending on marketplace requirements.

```text
vendor_documents
```

Fields:

```text
id
vendor_id
document_type
document_url
status
reviewed_by
reviewed_at
created_at
```

---

# 17. Vendor Approval API

```http
POST /api/v1/vendors/apply
```

Authenticated users can submit a vendor application.

Admin:

```http
GET /api/v1/admin/vendors/pending
```

Approve:

```http
POST /api/v1/admin/vendors/:vendorId/approve
```

Reject:

```http
POST /api/v1/admin/vendors/:vendorId/reject
```

Suspend:

```http
POST /api/v1/admin/vendors/:vendorId/suspend
```

---

# 18. Products

```text
products
```

Fields:

```text
id
vendor_id
brand_id
name
slug
description
short_description
status
approval_status
seo_title
seo_description
created_at
updated_at
```

Product status:

```text
draft
active
inactive
archived
suspended
```

Approval:

```text
pending
approved
rejected
```

Only approved + active products can appear in the public storefront.

---

# 19. Product Variants

```text
product_variants
```

Fields:

```text
id
product_id
sku
barcode
price
compare_at_price
cost_price
weight
status
created_at
updated_at
```

SKU must be unique.

---

# 20. Categories

```text
categories
```

Fields:

```text
id
parent_id
name
slug
description
image_url
sort_order
status
created_at
updated_at
```

Supports hierarchical categories.

Example:

```text
Electronics
 ├── Phones
 │    ├── Android
 │    └── iPhone
 └── Laptops
```

---

# 21. Product Categories

Many-to-many relationship:

```text
product_categories
```

```text
product_id
category_id
```

---

# 22. Brands

```text
brands
```

Fields:

```text
id
name
slug
description
logo_url
status
created_at
updated_at
```

---

# 23. Product Images

```text
product_images
```

Fields:

```text
id
product_id
variant_id
url
alt_text
sort_order
is_primary
created_at
```

---

# 24. Product Attributes

Support:

```text
Color
Size
RAM
Storage
Material
Weight
```

Tables:

```text
product_attributes
attribute_values
variant_attribute_values
```

---

# 25. Product Approval

Vendor:

```http
POST /api/v1/vendor/products
```

Creates a draft.

Vendor submits:

```http
POST /api/v1/vendor/products/:id/submit
```

Admin sees:

```http
GET /api/v1/admin/products/pending
```

Approve:

```http
POST /api/v1/admin/products/:id/approve
```

Reject:

```http
POST /api/v1/admin/products/:id/reject
```

Suspend:

```http
POST /api/v1/admin/products/:id/suspend
```

---

# 26. Public Product APIs

```http
GET /api/v1/products
GET /api/v1/products/:slug
GET /api/v1/products/:id/variants
GET /api/v1/products/:id/reviews
GET /api/v1/products/:id/questions
```

Filtering:

```text
category
brand
min_price
max_price
rating
availability
vendor
sort
search
page
limit
```

Example:

```http
GET /api/v1/products?category=phones&min_price=10000&max_price=50000&sort=price_asc&page=1&limit=20
```

---

# 27. Inventory

Tables:

```text
inventory_locations
inventory
inventory_movements
inventory_reservations
```

Inventory must support:

```text
available quantity
reserved quantity
sold quantity
damaged quantity
returned quantity
```

Never allow stock to become negative.

---

# 28. Inventory Movement

Every stock modification must create an inventory movement.

Types:

```text
purchase
sale
reservation
release
return
damage
adjustment
restock
```

Example:

```text
+100 restock
-2 sale
+1 return
-1 damage
```

This creates an auditable stock history.

---

# 29. Vendor Inventory APIs

```http
GET /api/v1/vendor/inventory
GET /api/v1/vendor/inventory/:variantId
POST /api/v1/vendor/inventory/:variantId/adjust
GET /api/v1/vendor/inventory/movements
```

Vendor can only access their own inventory.

---

# 30. Cart

Tables:

```text
carts
cart_items
```

APIs:

```http
GET    /api/v1/cart
POST   /api/v1/cart/items
PATCH  /api/v1/cart/items/:id
DELETE /api/v1/cart/items/:id
DELETE /api/v1/cart
```

Adding an item must verify:

* Product exists.
* Product is approved.
* Product is active.
* Variant exists.
* Variant belongs to product.
* Stock is available.
* Quantity is valid.

---

# 31. Wishlist

```http
GET    /api/v1/wishlist
POST   /api/v1/wishlist/items
DELETE /api/v1/wishlist/items/:productId
```

---

# 32. Checkout

```http
POST /api/v1/checkout
```

Checkout must:

1. Load cart.
2. Validate products.
3. Validate prices.
4. Validate inventory.
5. Validate coupon.
6. Calculate subtotal.
7. Calculate discount.
8. Calculate tax.
9. Calculate shipping.
10. Calculate grand total.
11. Reserve inventory.
12. Create order.
13. Initialize payment.

Never trust:

```text
price
discount
tax
total
```

from the frontend.

All totals must be recalculated on the backend.

---

# 33. Orders

Tables:

```text
orders
order_items
order_addresses
order_status_history
```

Order status:

```text
pending
confirmed
processing
packed
shipped
delivered
cancelled
returned
refunded
```

---

# 34. Order APIs — User

```http
GET /api/v1/orders
GET /api/v1/orders/:id
POST /api/v1/orders/:id/cancel
POST /api/v1/orders/:id/return
```

Users can only access their own orders.

---

# 35. Vendor Order APIs

Vendor:

```http
GET /api/v1/vendor/orders
GET /api/v1/vendor/orders/:id
POST /api/v1/vendor/orders/:id/confirm
POST /api/v1/vendor/orders/:id/pack
POST /api/v1/vendor/orders/:id/ship
```

A vendor must only see order items belonging to that vendor.

If one customer order contains products from three vendors:

```text
Order #1001

Vendor A
 ├── Product A
 └── Product B

Vendor B
 └── Product C

Vendor C
 └── Product D
```

The vendor API must not expose the other vendors' items.

---

# 36. Shipping

Tables:

```text
shipping_methods
shipments
shipment_items
tracking_events
```

Vendor can create shipment:

```http
POST /api/v1/vendor/orders/:orderId/shipment
```

Request:

```json
{
  "carrier": "Example Courier",
  "tracking_number": "TRK123456"
}
```

Tracking:

```http
GET /api/v1/orders/:orderId/tracking
```

---

# 37. Payments

Payment records must be immutable where possible.

Tables:

```text
payments
payment_transactions
refunds
```

Payment statuses:

```text
pending
processing
paid
failed
cancelled
refunded
partially_refunded
```

Payment methods can include:

```text
cash_on_delivery
card
mobile_payment
bank_transfer
```

Actual supported providers should be implemented behind a payment interface.

Example:

```go
type PaymentProvider interface {
    CreatePayment(...)
    VerifyPayment(...)
    RefundPayment(...)
}
```

---

# 38. Payment Webhooks

Payment providers must communicate through webhook endpoints.

Example:

```http
POST /api/v1/payments/webhook/:provider
```

Webhook handling must:

1. Verify provider signature.
2. Validate payload.
3. Check transaction ID.
4. Check idempotency.
5. Update payment.
6. Update order.
7. Confirm inventory.
8. Create audit event.
9. Trigger notification.

Do not trust frontend payment success.

---

# 39. Payment Idempotency

Payment and checkout APIs must support idempotency.

Client:

```http
Idempotency-Key: unique-request-id
```

Repeated requests must not create duplicate orders or payments.

---

# 40. Vendor Revenue

Vendor dashboard:

```http
GET /api/v1/vendor/dashboard
```

Response:

```json
{
  "total_orders": 1250,
  "total_sales": 1500000,
  "gross_revenue": 1500000,
  "platform_fee": 150000,
  "refunds": 30000,
  "net_revenue": 1320000,
  "cost_of_goods": 850000,
  "profit": 470000
}
```

Exact financial definitions must be consistent throughout the system.

---

# 41. Profit Calculation

Recommended:

```text
Gross Sales
- Discounts
- Refunds
- Platform Fees
- Cost of Goods
- Other applicable costs
=
Net Profit
```

The system must clearly distinguish:

```text
Revenue
Gross Sales
Net Sales
Platform Revenue
Vendor Revenue
Profit
```

Do not use one field called `revenue` for everything.

---

# 42. Admin Dashboard

```http
GET /api/v1/admin/dashboard
```

Must provide:

```text
total users
total vendors
pending vendors
active vendors

total products
pending products
approved products

total orders
pending orders
completed orders
cancelled orders

gross sales
net sales
platform revenue
vendor revenue
total profit

total refunds
pending refunds

inventory statistics
```

Support date filtering:

```text
today
yesterday
7 days
30 days
this month
last month
custom range
```

---

# 43. Admin Analytics

```http
GET /api/v1/admin/analytics/sales
GET /api/v1/admin/analytics/orders
GET /api/v1/admin/analytics/revenue
GET /api/v1/admin/analytics/vendors
GET /api/v1/admin/analytics/products
```

Example:

```http
GET /api/v1/admin/analytics/sales?from=2026-01-01&to=2026-10-01
```

---

# 44. Vendor Analytics

```http
GET /api/v1/vendor/analytics/sales
GET /api/v1/vendor/analytics/orders
GET /api/v1/vendor/analytics/products
GET /api/v1/vendor/analytics/revenue
```

Vendor must only receive data belonging to themselves.

---

# 45. Reviews

Only customers who purchased a product should be allowed to submit a verified review.

```http
POST /api/v1/products/:productId/reviews
```

Backend must verify:

```text
user owns order
AND
order contains product
AND
order is delivered
```

Review:

```text
rating
title
content
images
```

---

# 46. Review APIs

User:

```http
POST   /api/v1/products/:id/reviews
GET    /api/v1/products/:id/reviews
PATCH  /api/v1/reviews/:id
DELETE /api/v1/reviews/:id
```

Vendor:

```http
POST /api/v1/vendor/reviews/:id/reply
PATCH /api/v1/vendor/reviews/:id/reply
```

Admin:

```http
GET /api/v1/admin/reviews
POST /api/v1/admin/reviews/:id/approve
POST /api/v1/admin/reviews/:id/reject
POST /api/v1/admin/reviews/:id/hide
```

---

# 47. Product Questions

Customers can ask questions:

```http
POST /api/v1/products/:productId/questions
```

Vendor can reply:

```http
POST /api/v1/vendor/questions/:questionId/reply
```

Admin can moderate questions.

Tables:

```text
product_questions
product_question_answers
```

---

# 48. Coupons

Tables:

```text
coupons
coupon_products
coupon_categories
```

Coupon types:

```text
percentage
fixed_amount
free_shipping
```

Rules:

```text
minimum_order_value
maximum_discount
start_at
expires_at
usage_limit
per_user_limit
```

Coupon validation must happen server-side.

---

# 49. Returns

Production V2 should support returns.

Tables:

```text
returns
return_items
```

Flow:

```text
Delivered
   ↓
Return Request
   ↓
Vendor/Admin Review
   ↓
Approved
   ↓
Customer Ships Product
   ↓
Received
   ↓
Inspection
   ↓
Approved
   ↓
Refund
```

---

# 50. Notifications

Tables:

```text
notifications
notification_preferences
```

Notification types:

```text
order_created
order_confirmed
order_shipped
order_delivered
payment_success
payment_failed
refund_created
product_approved
vendor_approved
review_reply
```

Channels:

```text
in_app
email
SMS
```

Notifications should eventually be processed asynchronously.

---

# 51. Audit Logs

All important admin actions must be logged.

```text
audit_logs
```

Example:

```json
{
  "user_id": "...",
  "action": "approve_product",
  "entity_type": "product",
  "entity_id": "...",
  "old_values": {},
  "new_values": {}
}
```

Admin should have:

```http
GET /api/v1/admin/audit-logs
```

---

# 52. API Response Format

All APIs should use a consistent response structure.

Success:

```json
{
  "success": true,
  "data": {},
  "message": "Product created successfully"
}
```

List:

```json
{
  "success": true,
  "data": [],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 250,
    "total_pages": 13
  }
}
```

Error:

```json
{
  "success": false,
  "error": {
    "code": "PRODUCT_NOT_FOUND",
    "message": "Product not found"
  }
}
```

Never expose internal stack traces to clients.

---

# 53. HTTP Status Codes

Use standard HTTP semantics.

```text
200 OK
201 Created
204 No Content

400 Bad Request
401 Unauthorized
403 Forbidden
404 Not Found
409 Conflict
422 Unprocessable Entity
429 Too Many Requests

500 Internal Server Error
503 Service Unavailable
```

---

# 54. API Versioning

All APIs must be versioned.

```text
/api/v1/...
```

Example:

```text
/api/v1/products
/api/v1/orders
/api/v1/auth/login
```

Future breaking changes can use:

```text
/api/v2/...
```

---

# 55. Authorization

Implement RBAC middleware.

Example:

```text
RequireAuth()
RequireRole("admin")
RequireRole("vendor")
RequireRole("user")
```

For resources, additionally enforce ownership.

Example:

```text
Vendor A
   ↓
GET /vendor/products
   ↓
Only Vendor A products
```

Never rely solely on:

```text
role == vendor
```

Ownership must also be checked.

---

# 56. Product Ownership

Every vendor-owned product must contain:

```text
vendor_id
```

Every vendor operation must verify:

```text
product.vendor_id == authenticatedVendorID
```

This rule must be enforced in the service/repository layer, not only the frontend.

---

# 57. Database Transactions

Use PostgreSQL transactions for operations such as:

### Checkout

```text
BEGIN

Validate inventory
Reserve inventory
Create order
Create order items
Create payment

COMMIT
```

### Payment confirmation

```text
BEGIN

Update payment
Update order
Update inventory
Create inventory movement
Create notification

COMMIT
```

If any operation fails:

```text
ROLLBACK
```

---

# 58. Concurrency Control

Inventory must protect against overselling.

Example:

```text
Stock = 1

Customer A → checkout
Customer B → checkout
```

Only one should successfully reserve the final unit.

Use PostgreSQL transaction locking such as:

```sql
SELECT ...
FOR UPDATE
```

where appropriate.

Redis locks may be used for distributed coordination, but PostgreSQL must remain authoritative for inventory consistency.

---

# 59. Database Indexing

Important indexes:

```text
users.email
users.phone

vendors.user_id
vendors.status

products.slug
products.vendor_id
products.status
products.approval_status

product_variants.sku
product_variants.product_id

categories.slug
categories.parent_id

orders.order_number
orders.user_id
orders.status
orders.created_at

order_items.order_id
order_items.variant_id

payments.order_id
payments.status

reviews.product_id
reviews.user_id

inventory.variant_id
inventory.location_id
```

Create indexes based on actual query patterns and monitor query performance.

---

# 60. Database Constraints

Use PostgreSQL constraints wherever possible.

Examples:

```text
users.email UNIQUE

product_variants.sku UNIQUE

products.slug UNIQUE

quantity >= 0

price >= 0

rating BETWEEN 1 AND 5

foreign keys

NOT NULL
```

Never depend solely on application-level validation for data integrity.

---

# 61. Money Handling

Never use floating point for money.

Do not use:

```go
float64
```

for financial calculations.

Use PostgreSQL:

```text
NUMERIC(19,4)
```

or integer minor units where appropriate.

Example:

```text
1000.50 BDT
```

must not become:

```text
1000.499999
```

All monetary calculations must be deterministic.

---

# 62. Currency

Every financial record should contain currency when required.

Example:

```text
currency = BDT
```

Do not assume every future transaction will always use BDT.

---

# 63. Security

Implement:

* JWT verification.
* Password hashing.
* Request validation.
* SQL injection protection.
* CORS.
* Rate limiting.
* Request size limits.
* Secure HTTP headers.
* Refresh-token rotation/revocation.
* Authorization checks.
* Ownership checks.
* Webhook signature verification.
* Audit logging.
* Secret management.
* Structured logging without sensitive data.

Never log:

```text
password
JWT
refresh token
payment secrets
card details
```

---

# 64. Password Security

Use Argon2id or bcrypt.

Never store plaintext passwords.

Passwords must never appear in:

```text
logs
database
API responses
errors
```

---

# 65. Rate Limiting

Apply stricter limits to:

```text
login
register
refresh token
password reset
payment
review submission
```

Example:

```text
Login:
5 attempts/minute/IP
```

Exact production limits should be configurable.

---

# 66. CORS

Only configured frontend origins may access the API.

Example environments:

```text
Development:
http://localhost:3000

Production:
https://your-store.com
```

Do not use unrestricted:

```text
Access-Control-Allow-Origin: *
```

for authenticated production APIs.

---

# 67. Logging

Use structured JSON logs.

Example:

```json
{
  "level": "info",
  "request_id": "abc123",
  "method": "POST",
  "path": "/api/v1/orders",
  "status": 201,
  "duration_ms": 84
}
```

Every request should have a request ID.

---

# 68. Error Handling

Create centralized application errors.

Example:

```go
ErrUserNotFound
ErrProductNotFound
ErrUnauthorized
ErrForbidden
ErrInsufficientStock
ErrInvalidPayment
ErrOrderNotCancellable
```

Map internal errors to HTTP responses centrally.

---

# 69. Background Jobs

Use asynchronous workers for:

```text
emails
notifications
image processing
analytics aggregation
expired cart cleanup
expired reservation cleanup
payment reconciliation
refund processing
```

Example:

```text
API
 ↓
Queue
 ↓
Worker
 ↓
Job
```

Do not make the user wait for unnecessary background operations.

---

# 70. Caching

Cache read-heavy resources:

```text
product details
categories
brands
popular products
```

Do not blindly cache financial information.

Invalidate cache after product changes.

---

# 71. Idempotency

The following APIs should support idempotency:

```text
POST /checkout
POST /payments
POST /refunds
POST /shipments
```

This prevents duplicate operations caused by retries.

---

# 72. Pagination

All potentially large collections must support pagination.

Use:

```text
page
limit
```

Initially.

For very large datasets, introduce cursor-based pagination.

Example:

```http
GET /api/v1/products?limit=20&cursor=...
```

---

# 73. Search

V1:

```text
PostgreSQL full-text search
```

Future:

```text
Meilisearch
OpenSearch
Elasticsearch
```

Do not introduce a separate search engine until required.

---

# 74. API Endpoint Summary

## Authentication

```text
POST /api/v1/auth/register
POST /api/v1/auth/login
POST /api/v1/auth/refresh
POST /api/v1/auth/logout
GET  /api/v1/auth/me
```

## Products

```text
GET    /api/v1/products
GET    /api/v1/products/:slug
GET    /api/v1/products/:id/reviews
GET    /api/v1/products/:id/questions
```

## User

```text
GET    /api/v1/users/me
PATCH  /api/v1/users/me
GET    /api/v1/users/me/addresses
POST   /api/v1/users/me/addresses
PATCH  /api/v1/users/me/addresses/:id
DELETE /api/v1/users/me/addresses/:id
```

## Cart

```text
GET    /api/v1/cart
POST   /api/v1/cart/items
PATCH  /api/v1/cart/items/:id
DELETE /api/v1/cart/items/:id
DELETE /api/v1/cart
```

## Wishlist

```text
GET    /api/v1/wishlist
POST   /api/v1/wishlist/items
DELETE /api/v1/wishlist/items/:productId
```

## Orders

```text
POST   /api/v1/checkout
GET    /api/v1/orders
GET    /api/v1/orders/:id
POST   /api/v1/orders/:id/cancel
POST   /api/v1/orders/:id/return
```

## Reviews

```text
POST   /api/v1/products/:id/reviews
PATCH  /api/v1/reviews/:id
DELETE /api/v1/reviews/:id
POST   /api/v1/reviews/:id/helpful
```

## Payments

```text
POST /api/v1/payments/create
GET  /api/v1/payments/:id
POST /api/v1/payments/webhook/:provider
```

---

# 75. Vendor API

```text
GET    /api/v1/vendor/dashboard

GET    /api/v1/vendor/products
POST   /api/v1/vendor/products
GET    /api/v1/vendor/products/:id
PATCH  /api/v1/vendor/products/:id
DELETE /api/v1/vendor/products/:id
POST   /api/v1/vendor/products/:id/submit

GET    /api/v1/vendor/inventory
POST   /api/v1/vendor/inventory/:variantId/adjust

GET    /api/v1/vendor/orders
GET    /api/v1/vendor/orders/:id
POST   /api/v1/vendor/orders/:id/confirm
POST   /api/v1/vendor/orders/:id/pack
POST   /api/v1/vendor/orders/:id/ship

GET    /api/v1/vendor/reviews
POST   /api/v1/vendor/reviews/:id/reply

GET    /api/v1/vendor/questions
POST   /api/v1/vendor/questions/:id/reply

GET    /api/v1/vendor/analytics/sales
GET    /api/v1/vendor/analytics/revenue
GET    /api/v1/vendor/analytics/profit
GET    /api/v1/vendor/analytics/orders
```

---

# 76. Admin API

```text
GET  /api/v1/admin/dashboard

GET  /api/v1/admin/users
GET  /api/v1/admin/vendors
GET  /api/v1/admin/vendors/pending

POST /api/v1/admin/vendors/:id/approve
POST /api/v1/admin/vendors/:id/reject
POST /api/v1/admin/vendors/:id/suspend

GET  /api/v1/admin/products
GET  /api/v1/admin/products/pending

POST /api/v1/admin/products/:id/approve
POST /api/v1/admin/products/:id/reject
POST /api/v1/admin/products/:id/suspend

GET  /api/v1/admin/orders
GET  /api/v1/admin/payments
GET  /api/v1/admin/refunds

GET  /api/v1/admin/reviews
POST /api/v1/admin/reviews/:id/approve
POST /api/v1/admin/reviews/:id/reject

GET  /api/v1/admin/analytics/sales
GET  /api/v1/admin/analytics/revenue
GET  /api/v1/admin/analytics/profit
GET  /api/v1/admin/analytics/vendors
GET  /api/v1/admin/analytics/products

GET  /api/v1/admin/audit-logs
```

---

# 77. API Documentation

Use OpenAPI/Swagger.

Generate documentation from the API definitions.

Documentation should describe:

* Endpoint.
* HTTP method.
* Authentication.
* Roles.
* Request body.
* Query parameters.
* Response.
* Error responses.
* Examples.

Expose Swagger only according to the deployment/security policy.

---

# 78. Testing

The backend must include:

## Unit tests

Test:

```text
services
business rules
pricing
discounts
inventory
authorization
profit calculations
```

## Integration tests

Test:

```text
PostgreSQL
Redis
authentication
orders
payments
inventory
```

## API tests

Test complete flows:

```text
Register
 ↓
Login
 ↓
Browse
 ↓
Cart
 ↓
Checkout
 ↓
Payment
 ↓
Order
 ↓
Shipping
 ↓
Delivery
 ↓
Review
```

---

# 79. Critical Test Scenarios

The following must be tested.

### Overselling

```text
Stock = 1

Customer A buys
Customer B buys simultaneously
```

Only one order succeeds.

### Duplicate checkout

Same idempotency key must not create two orders.

### Vendor isolation

Vendor A must never access Vendor B's resources.

### Admin approval

Unapproved products must not appear publicly.

### Review verification

User who did not purchase a product cannot submit a verified review.

### Payment verification

Frontend cannot mark an order as paid.

### Refund

Refund cannot exceed the paid amount.

### Coupon

Expired coupon cannot be used.

### Authorization

User cannot access another user's order.

---

# 80. Environment Variables

Example:

```env
APP_ENV=production
APP_PORT=8080

DATABASE_URL=

JWT_ACCESS_SECRET=
JWT_REFRESH_SECRET=

REDIS_URL=

STORAGE_ENDPOINT=
STORAGE_BUCKET=
STORAGE_ACCESS_KEY=
STORAGE_SECRET_KEY=

PAYMENT_PROVIDER_KEY=
PAYMENT_WEBHOOK_SECRET=

CORS_ALLOWED_ORIGINS=

LOG_LEVEL=info
```

Never commit secrets to Git.

---

# 81. Database Migration

All schema changes must use migrations.

Example:

```text
migrations/
001_create_users.sql
002_create_roles.sql
003_create_vendors.sql
004_create_products.sql
005_create_categories.sql
...
```

Production database changes must never be made manually.

---

# 82. Deployment

Recommended architecture:

```text
                    Internet
                       │
                       ▼
                  CDN / Proxy
                       │
                       ▼
                    Next.js
                       │
                       ▼
                    Go API
                       │
          ┌────────────┼────────────┐
          ▼            ▼            ▼
      PostgreSQL     Redis      Object Storage
         Neon
```

Go API should be stateless so multiple instances can run simultaneously.

---

# 83. Health Checks

Implement:

```http
GET /health
GET /ready
```

`/health` checks whether the application process is alive.

`/ready` checks required dependencies such as PostgreSQL and Redis.

---

# 84. Graceful Shutdown

The Go server must gracefully handle:

```text
SIGINT
SIGTERM
```

Before shutdown:

1. Stop accepting requests.
2. Finish active requests.
3. Stop background workers.
4. Close database connections.
5. Close Redis connection.
6. Exit cleanly.

---

# 85. Production Requirements

The implementation is not considered complete until:

* All protected APIs require authentication.
* Role-based authorization is implemented.
* Ownership checks are implemented.
* Vendor isolation is enforced.
* Admin approval workflows work.
* PostgreSQL transactions protect critical operations.
* Inventory cannot be oversold.
* Payments are verified through provider/webhooks.
* Duplicate payment/order operations are prevented.
* Money calculations are safe.
* Database migrations exist.
* API documentation exists.
* Tests exist.
* Logging exists.
* Health checks exist.
* Graceful shutdown exists.
* Secrets are externalized.
* CORS is restricted.
* Rate limiting is implemented.
* Audit logging is implemented.
* Error responses are standardized.

---

# 86. Implementation Rules for AI Coding Agent

The coding agent MUST follow these rules.

### Rule 1 — No business logic in handlers

Bad:

```text
HTTP Handler
 ├── SQL
 ├── payment calculation
 ├── inventory logic
 └── response
```

Good:

```text
Handler
   ↓
Service
   ↓
Repository
   ↓
Database
```

### Rule 2 — Never trust frontend calculations

The backend must independently calculate:

```text
price
discount
tax
shipping
subtotal
total
profit
commission
```

### Rule 3 — Never trust frontend role

Role must come from verified authentication data/database.

### Rule 4 — Enforce ownership server-side

Never assume a vendor owns a product because the frontend says so.

### Rule 5 — Use transactions for financial/inventory operations.

### Rule 6 — Never expose sensitive information.

### Rule 7 — Do not use floating-point numbers for money.

### Rule 8 — Do not create duplicate business logic.

### Rule 9 — Use interfaces around external services.

Examples:

```go
type PaymentProvider interface {}

type StorageProvider interface {}

type NotificationProvider interface {}
```

### Rule 10 — Keep external integrations replaceable.

---

# 87. Definition of Done

The backend is considered production-ready when:

```text
[ ] Go project initialized
[ ] Configuration system implemented
[ ] PostgreSQL connection implemented
[ ] Database migrations implemented
[ ] Redis connection implemented
[ ] JWT authentication implemented
[ ] Refresh tokens implemented
[ ] RBAC implemented
[ ] User management implemented
[ ] Vendor registration implemented
[ ] Vendor approval implemented
[ ] Product CRUD implemented
[ ] Product approval implemented
[ ] Category management implemented
[ ] Brand management implemented
[ ] Product variants implemented
[ ] Product images implemented
[ ] Inventory implemented
[ ] Inventory reservations implemented
[ ] Cart implemented
[ ] Wishlist implemented
[ ] Checkout implemented
[ ] Orders implemented
[ ] Payments implemented
[ ] Payment webhook implemented
[ ] Shipping implemented
[ ] Tracking implemented
[ ] Reviews implemented
[ ] Vendor review replies implemented
[ ] Product questions implemented
[ ] Coupons implemented
[ ] Returns implemented
[ ] Refunds implemented
[ ] Notifications implemented
[ ] Vendor dashboard implemented
[ ] Admin dashboard implemented
[ ] Revenue analytics implemented
[ ] Profit analytics implemented
[ ] Audit logs implemented
[ ] Rate limiting implemented
[ ] Request validation implemented
[ ] Error handling implemented
[ ] API versioning implemented
[ ] OpenAPI documentation implemented
[ ] Unit tests implemented
[ ] Integration tests implemented
[ ] API tests implemented
[ ] Health checks implemented
[ ] Graceful shutdown implemented
[ ] Production logging implemented
[ ] Security review completed
```

---

# 88. Final Architecture

The final system should follow:

```text
                         Next.js
                            │
                            │ REST
                            ▼
                    ┌───────────────┐
                    │    Go API     │
                    │               │
                    │ ┌───────────┐ │
                    │ │   Auth    │ │
                    │ ├───────────┤ │
                    │ │  Product  │ │
                    │ ├───────────┤ │
                    │ │  Vendor   │ │
                    │ ├───────────┤ │
                    │ │   Cart    │ │
                    │ ├───────────┤ │
                    │ │  Orders   │ │
                    │ ├───────────┤ │
                    │ │ Payments  │ │
                    │ ├───────────┤ │
                    │ │ Inventory │ │
                    │ ├───────────┤ │
                    │ │ Shipping  │ │
                    │ ├───────────┤ │
                    │ │  Reviews  │ │
                    │ └───────────┘ │
                    └───────┬───────┘
                            │
             ┌──────────────┼──────────────┐
             ▼              ▼              ▼
        PostgreSQL        Redis       Object Storage
           Neon
```

The key architectural principle is:

> **Next.js handles presentation. Go handles business logic. PostgreSQL is the source of truth. Redis handles caching/coordination. Object storage handles files. External providers handle payments/email/shipping where appropriate.**

This architecture should support a serious multi-vendor marketplace while remaining simple enough to operate as a modular monolith before introducing microservices.
