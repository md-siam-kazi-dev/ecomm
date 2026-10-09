# e-go-backend

Production multi-vendor e-commerce backend — Go + PostgreSQL (Neon) + Redis.
Implements the modular-monolith architecture described in [PRD.md](PRD.md) and
the schema in [ARCHITECTURE.md](ARCHITECTURE.md).

## Stack

```text
Language : Go 1.22
HTTP     : net/http + chi
Database : PostgreSQL (Neon) via pgx + pgxpool
Cache    : Redis (optional in development)
Auth     : JWT (access + refresh)
Storage  : S3-compatible object storage
```

Layering: `Handler → Service → Repository → PostgreSQL`. No SQL or business
logic in handlers.

## Layout

```text
cmd/api/            application entrypoint
internal/           domain modules + config, database, server
  config/           env-driven configuration
  database/         Postgres + Redis connections
  server/           router, middleware wiring, probes
pkg/                reusable, domain-agnostic packages
  apperr/ response/ logger/ middleware/ jwt/ password/ validator/
migrations/         ordered SQL migrations
docs/  tests/
```

## Getting started

```bash
cp .env.example .env      # then fill in DATABASE_URL and the JWT secrets
go mod download
make run                  # or: go run ./cmd/api
```

The server starts on `APP_PORT` (default `8080`) and exposes:

```text
GET /health    liveness  — process is up
GET /ready     readiness — Postgres (and Redis if configured) reachable
```

## Configuration

All configuration is read from environment variables (loaded from `.env` in
development via godotenv). See [.env.example](.env.example) for the full list.
`DATABASE_URL` and both JWT secrets are required; startup fails fast if any is
missing.

## Authentication

Cookie-based auth (`internal/auth`). Access + refresh tokens are set as
`HttpOnly` cookies; the access token is a 15-minute HS256 JWT and the refresh
token is an opaque random value stored only as a SHA-256 hash in
`user_sessions`. Refresh rotates on every use.

```text
POST /api/v1/auth/register        public                    201
POST /api/v1/auth/login           public                    { user }  + cookies
POST /api/v1/auth/refresh         refresh cookie            { user }  + rotated cookies
POST /api/v1/auth/logout          refresh cookie            idempotent
POST /api/v1/auth/logout-all      access token
GET  /api/v1/auth/me              access token              { user }
POST /api/v1/auth/verify-email    public  { token }
POST /api/v1/auth/forgot-password public  { email }         generic response
POST /api/v1/auth/reset-password  public  { token, new_password }
```

Authorization uses `auth.Authenticate` then `auth.RequireRole("admin", ...)` on
a route group; a missing/invalid token is `401`, a disallowed role is `403`.

CSRF is covered by `SameSite=Lax` cookies, a strict CORS allow-list and an
`Origin` allow-list check on state-changing requests. Rate limits are applied
per IP/route and fall back to an in-process limiter when `REDIS_URL` is unset.

Email is abstracted behind a `Mailer`; the development implementation logs the
verification / password-reset link instead of sending mail. Wire a real provider
before production.

## Admin API

Admin routes sit behind `auth.Authenticate` + `auth.RequireRole("admin")`.

```text
GET /api/v1/admin/users?page=1&limit=20&role=customer&search=siam
```

- `page` — default `1`
- `limit` — one of `10 | 20 | 50 | 100` (default `20`; anything else falls back)
- `role` — `user` | `vendor` | `admin` (also accepts the `customer` alias); omit for all
- `search` — partial, case-insensitive match on **name, email or user ID**

`data` is the user array; `pagination` carries `page`, `limit`, `total`,
`total_pages`, `has_next`, `has_previous`. Filtering, search, ordering and
pagination all run in PostgreSQL — the client never loads the full list.
Migration `009` adds trigram (pg_trgm) GIN indexes for the partial search plus
supporting indexes for the sort and role join.

### Product moderation

```text
GET  /api/v1/admin/products?status=pending|rejected|changes_requested|all&search=&page=&limit=
GET  /api/v1/admin/products/{id}
POST /api/v1/admin/products/{id}/approve
POST /api/v1/admin/products/{id}/reject     { "reason": "..." }  (required)
POST /api/v1/admin/products/{id}/changes    { "note": "..." }    (required)
```

Approve sets `approval_status=approved`, `status=active` and records
`approved_by`/`approved_at`. Reject and request-changes keep the product as a
draft and store the note in `products.review_note`, which the vendor sees.
Migration `012` adds the `changes_requested` enum value, `review_note`,
`submitted_at`, `products.attributes` and seeds the storefront categories.

## Vendor API

Vendor routes sit behind `auth.Authenticate` + `auth.RequireRole("vendor")`.
The store is then resolved from the authenticated user via `store_memberships`
(store-scoped authorization): an active membership in an **active** store is
required, otherwise the request is rejected (`NO_ACTIVE_STORE` /
`STORE_NOT_ACTIVE`). An optional `X-Store-Id` header (or `?store_id=`) selects
one of several stores an account may manage; omit it to use the caller's active
store.

```text
GET  /api/v1/vendor/profile
GET  /api/v1/vendor/dashboard
GET  /api/v1/vendor/products?page=1&limit=20&search=
GET  /api/v1/vendor/products/{id}
POST /api/v1/vendor/products
PUT  /api/v1/vendor/products/{id}
POST /api/v1/vendor/products/images          (multipart, field "image")
GET  /api/v1/vendor/orders?page=1&limit=20&status=shipped
```

Every query is scoped to the vendor (`products.vendor_id`,
`order_items.vendor_id`); another vendor's product id returns 404. Product
creation writes the core row, a default `product_variants` row (regular/sale
price, SKU), inventory (auto-created per-vendor location) and product images in
one transaction. Approval/publishing fields are set by the server, never the
payload: a new product is `draft` + `pending`, submitted products show in the
admin review queue, and only an approved product becomes `active` (publicly
eligible). Images are hosted on imgbb (`IMGBB_API_KEY`).

Vendor order fulfilment mutations (`/vendor/orders/:id/confirm|pack|ship`) and
review replies are not implemented yet — the vendor UI simulates them (demo).

## Store applications & vendor onboarding

A customer applies for a store from their existing account; an admin reviews it;
the applicant pays an application fee; only then does the store activate and the
`vendor` role get granted. Migration `013` adds `store_applications`, `stores`,
`store_memberships` and `application_payments`, plus new enums
(`application_status`, `store_status`, `store_member_role`, `membership_status`,
`business_type`).

Applicant (auth required):

```text
POST /api/v1/store-applications                (submit; rate-limited)
POST /api/v1/store-applications/images         (multipart logo/banner, "image")
GET  /api/v1/store-applications/me
GET  /api/v1/store-applications/{id}
GET  /api/v1/store-applications/{id}/approval-letter   (printable HTML)
POST /api/v1/store-applications/{id}/payments  (initiate application fee)
GET  /api/v1/store-applications/{id}/payments
GET  /api/v1/my-stores                          (stores the caller can access)
```

Admin (`auth.Authenticate` + `RequireRole("admin")`):

```text
GET  /api/v1/admin/store-applications?status=&payment_status=&search=&sort=&page=&limit=
GET  /api/v1/admin/store-applications/{id}
POST /api/v1/admin/store-applications/{id}/approve
POST /api/v1/admin/store-applications/{id}/reject   { "reason": "..." }  (required)
POST /api/v1/admin/store-applications/{id}/notes    { "note": "..." }
GET  /api/v1/admin/stores
GET  /api/v1/admin/application-payments
```

Public: `POST /api/v1/payments/webhook/{provider}` — signature-verified and
idempotent; only a provider-verified payload can mark a payment `paid`.

Approval creates the store (`pending_activation`) + its `vendors` row
(`pending`) + an owner membership, and records the fee from server config.
Activation — which flips the store to `active`, the vendor to `approved`, grants
the `vendor` role and is idempotent — happens at approval when the fee is
disabled, or after a verified payment when it is enabled. Approval alone never
grants dashboard access.

No payment provider is integrated yet: the platform ships with
`pkg/payment.NotConfigured`, so initiating a fee returns an explicit
`PAYMENT_NOT_CONFIGURED` (never a fake success). Configure with
`APPLICATION_FEE_ENABLED`, `APPLICATION_FEE_AMOUNT`, `APPLICATION_FEE_CURRENCY`,
`PAYMENT_PROVIDER`, `PAYMENT_WEBHOOK_SECRET` and `APP_NAME`.

## Migrations

SQL migrations live in `migrations/` and are applied in lexical order. Tables
are created in dependency order; note the documented circular dependency
between `orders` and `inventory_reservations` (create the reservation table
without the FK, then add it with `ALTER TABLE`).

## Common commands

```bash
make run     # run the API
make build   # compile to bin/api
make test    # go test ./... -race
make vet     # go vet ./...
make fmt     # gofmt
make tidy    # go mod tidy
```
