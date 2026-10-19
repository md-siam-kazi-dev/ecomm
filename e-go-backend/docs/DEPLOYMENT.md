# Deploying to Vercel

The API is a standard `net/http` + chi server, so it deploys with Vercel's
**Go Framework Preset** (`framework: "go"`). Vercel detects `go.mod` at the
project root and the `cmd/api/main.go` entrypoint, runs `go build`, and runs the
compiled server against the injected `PORT` (already honored by
`config.Load`). No serverless `api/` shim is needed.

`vercel.json`:

```json
{
  "$schema": "https://openapi.vercel.sh/vercel.json",
  "framework": "go",
  "buildCommand": "go build -o bin/api ./cmd/api"
}
```

## 1. Prerequisites

- Vercel account (the Go runtime is in **Beta**, available on all plans).
- A Postgres database reachable from the internet — Neon here. Use the
  **pooled** connection string (`...-pooler...`) for serverless.
- Optional: Redis (e.g. Upstash) for a shared rate limiter. Without it the app
  falls back to a per-instance in-process limiter.

## 2. Project root

The Go module lives in `e-go-backend/`, which is also where `.vercel/` is
linked. If you deploy from a repository whose root is *above* this folder, set
Vercel's **Root Directory** to `e-go-backend`.

Deploy with the CLI:

```bash
cd e-go-backend
vercel          # preview deployment
vercel --prod   # production deployment
```

Or connect the Git repo in the Vercel dashboard and set Root Directory to
`e-go-backend`.

## 3. Environment variables

`.env` is gitignored and is **not** uploaded, so add every value in
Vercel → Project → Settings → Environment Variables (Production, and Preview if
you want previews to work).

Required:

| Variable             | Notes                                                        |
| -------------------- | ------------------------------------------------------------ |
| `DATABASE_URL`       | Neon pooled URL, `?sslmode=require`.                          |
| `JWT_ACCESS_SECRET`  | 32+ chars in production (`openssl rand -base64 32`).          |
| `JWT_REFRESH_SECRET` | 32+ chars in production.                                      |

Strongly recommended:

| Variable                | Notes                                                                                     |
| ----------------------- | ----------------------------------------------------------------------------------------- |
| `APP_ENV`               | `production` — enables strict validation.                                                  |
| `FRONTEND_URL`          | e.g. `https://your-frontend`. Merged into the CORS allow-list and used in email links.     |
| `CORS_ALLOWED_ORIGINS`  | Comma-separated browser origins, e.g. `https://your-frontend`.                            |
| `COOKIE_SECURE`         | `true` when served over HTTPS.                                                             |
| `COOKIE_SAMESITE`       | `lax` when frontend and API share a site; `none` when they are on different sites.        |
| `DB_MAX_CONNS`          | `5` (keep each instance small; Neon's pooler multiplexes).                                 |
| `DB_MIN_CONNS`          | `0` (don't hold idle connections from an ephemeral instance).                              |
| `IMGBB_API_KEY`         | Required for image uploads.                                                                |
| `PAYMENT_*`             | Stripe keys/secret if online payment is enabled.                                           |
| `LOG_LEVEL`             | `info`.                                                                                    |

`COOKIE_SAMESITE=none` requires `COOKIE_SECURE=true` — the app fails fast at
startup otherwise, because browsers reject a `SameSite=None` cookie without
`Secure`.

### Cookies across domains

`*.vercel.app` subdomains are **distinct sites**, so a frontend and API on
separate Vercel projects are cross-site. In that topology:

- Set `COOKIE_SAMESITE=none` and `COOKIE_SECURE=true`.
- Call the API with `fetch(url, { credentials: 'include' })`.

Modern browsers may still block third-party cookies (Safari ITP, Chrome). The
most reliable setup is to serve the frontend and API from the **same site** —
custom domains under one parent domain, or a frontend rewrite that proxies
`/api/*` to the API deployment.

## 4. Run migrations

Vercel does not run SQL on deploy. Apply `migrations/*.sql` to Neon once in
lexical order:

```bash
psql "$DATABASE_URL" -f migrations/001_initial_schema.sql
# ... repeat for every migration, in order
```

## 5. Region

Match the function region to the database to avoid round trips. If your Neon
project is in `ap-southeast-1`, use Vercel's `sin1` (Singapore) region — set it
in Project Settings → Functions, or via `regions` in `vercel.json` on plans that
allow custom regions.

## 6. Verify

```bash
curl https://<your-project>.vercel.app/health
curl https://<your-project>.vercel.app/ready
```

`/health` is a liveness check. `/ready` returns `200` only when Postgres (and
Redis, if configured) is reachable.

## Serverless caveats

- Request bodies are capped by the platform (Vercel ~4.5 MB). Image uploads go
  to imgbb; keep them small.
- Each instance owns its own pgx pool and (without Redis) its own rate limiter,
  so limits are per instance. Use Redis for a shared limiter.
- Cold starts open a fresh pool — the Neon pooler keeps that cheap.
