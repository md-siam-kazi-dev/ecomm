# Admin Audit Logs — Implementation Report

## Summary

Added a production-grade, read-only **admin audit trail** to the marketplace backend
(Go/Postgres) and admin dashboard (Next.js), built on the pre-existing `audit_logs`
table which was enriched rather than recreated.

- **Backend:** new `internal/audit` package (recorder, sanitizer, repository, service,
  handler, routes) + schema migration + real business-event wiring.
- **Frontend:** `/admin/audit-logs` page with summary cards, filters, pagination and a
  detail dialog; new sidebar entry.
- **Tests:** new unit tests for the audit package; all existing suites still pass.

## 1. Database

`migrations/020_audit_logs.sql` (idempotent, per-statement, matching 007–019 style):

- Adds columns to the existing table: `actor_role`, `result`, `reason`, `request_id`,
  `metadata JSONB`.
- Adds indexes: `(created_at DESC, id DESC)`, `action`, `actor_role`, `result`, and a
  partial index on `request_id`.

Apply with the same process as prior migrations (lexical order), e.g.:

```bash
psql "$DATABASE_URL" -f migrations/020_audit_logs.sql
```

The table remains append-only — no API ever updates or deletes audit rows.

## 2. Audit package (`internal/audit`)

| File | Responsibility |
| --- | --- |
| `model.go` | `Event`, actor/result/entity constants, stable dotted action names, display labels, action catalogue |
| `sanitize.go` | Redacts sensitive keys (password/token/secret/card/cvv/… ) recursively, caps string length and depth, never mutates input |
| `context.go` | `Middleware` + `RequestInfoFromContext` to attach IP / user-agent / request-id to each request |
| `recorder.go` | `Record(ctx, Querier, Event)` — one INSERT; works with `*pgxpool.Pool` or `pgx.Tx`; nil-safe; validates UUIDs/IPs and demotes invalid entity ids into `metadata.entity_ref` |
| `repository.go` | Read model: filtered/paginated `List`, full `Get`, `Summary` aggregates |
| `service.go` | Query normalization (limits allow-list, role/result/entity allow-lists, 366-day range clamp) |
| `dto.go` | JSON DTOs + category classification |
| `handler.go` / `routes.go` | Admin read endpoints |

Security properties:

- Every payload (`old_values`, `new_values`, `metadata`) is sanitized on write.
- No secrets, tokens, cookies, card data or API keys can be persisted.
- IPs are validated before insertion into the `INET` column; invalid values become `NULL`.
- Read APIs are mounted only inside the authenticated `RequireRole("admin")` group.

## 3. API endpoints (admin-only)

| Method | Path |
| --- | --- |
| GET | `/api/v1/admin/audit-logs` (filters: `search, action, actor_role, entity_type, entity_id, result, from, to, page, limit`) |
| GET | `/api/v1/admin/audit-logs/summary` (totals + action catalogue) |
| GET | `/api/v1/admin/audit-logs/{id}` (full detail incl. payloads + request metadata) |

`audit.Middleware` is registered globally in `internal/server/routes.go` (after
`RequestID`) so every request carries correlation metadata.

## 4. Business events recorded

**Transactional (same transaction as the change):**

- `admin.product_approved` / `admin.product_rejected` / `admin.product_changes_requested`
  — `admin.SetApproval` now reads the prior state `FOR UPDATE`, applies the change and
  writes the audit row atomically. The acting admin id is threaded through
  `Reject`/`RequestChanges` (previously dropped).
- `store_application.submitted|approved|rejected`, `store.activated` — storeapp now
  records through `audit.Record` inside its existing transactions, with actor role and
  rejection reason.

**Best-effort (post-success, handler-level via injected recorder):**

- Vendor: `vendor.product_created`, `vendor.product_updated`,
  `vendor.inventory_adjusted`, `vendor.order_status_changed`,
  `vendor.shipment_updated`, `vendor.review_replied`, `vendor.question_answered`.
- Auth/security: `auth.registered`, `auth.login_succeeded`, `auth.login_failed`,
  `auth.logged_out_all`, `auth.email_verified`, `auth.password_reset`.
- Orders: `order.created`.

Not yet wired: payment webhook capture/failure (`payment.paid`/`payment.failed`). The
`HandleWebhook` service path returns only an error, so auditing it cleanly requires a
service signature change; the action names and labels are already reserved.

## 5. Frontend

- `src/types/audit.ts` — API-mirroring types.
- `src/components/admin/audit-logs-manager.tsx` — server-driven filters (search, action,
  actor role, entity, result, date range), summary cards, desktop table + mobile cards,
  pagination, loading/empty/error states, and an accessible detail dialog showing
  before/after values and request metadata.
- `src/app/admin/audit-logs/page.tsx` — admin page.
- `src/components/admin/admin-shell.tsx` — added “Audit logs” nav item.
- `src/lib/formatters.ts` — added `formatDateTime`.

URL state is shareable; every filter, page and limit is reflected in the query string.

## 6. Verification performed

Backend:

- `gofmt -l .` → clean; `go vet ./...` → clean; `go build ./...` → ok.
- `go test ./... -race -count=1` → **all packages pass**, including the new
  `internal/audit` tests and updated `internal/admin` product-service tests.

Frontend:

- `bun run lint` → clean.
- `bun test` → 28 pass / 0 fail.
- `bun run build` → succeeds; `/admin/audit-logs` is included.

## 7. Known limitations / follow-ups

- Vendor/auth/order events are recorded best-effort after the business operation; a
  process crash in the gap could omit a row. Moving them into per-module transactions
  would require widening repository interfaces and is deferred.
- Payment webhook events are not yet emitted (see §4).
- Summary aggregates are date-range scoped; a very wide range only runs `count(*)`
  predicates (index-backed).
