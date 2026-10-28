# ShopKart — Dropshipping E-commerce Storefront

A modern, production-quality dropshipping e-commerce frontend built with **Next.js, React and TypeScript**.

> **The catalog runs on mock/demo data** — no real payments or supplier integrations. Catalog data flows through a repository layer so a future **Go REST API + PostgreSQL** backend can replace it without a UI rewrite. **Authentication is real**, backed by Better Auth and PostgreSQL.

## Tech Stack

| Layer | Choice |
|---|---|
| Framework | Next.js 16 (App Router, Server Components, Turbopack, React Compiler) |
| UI | React 19 + TypeScript (strict) |
| Styling | Tailwind CSS v4 — CSS-first design tokens (emerald primary `#10B981`, slate neutrals) |
| Components | shadcn/ui (radix base) |
| Animation | Motion (Framer Motion), reduced-motion aware |
| State | Zustand (cart / wishlist / UI state, persisted) |
| Auth | Better Auth — email/password sessions stored in PostgreSQL |
| Database | PostgreSQL + Drizzle ORM (migrations in `drizzle/`) |
| Icons | Lucide |

## Getting Started

```bash
bun install
cp .env.example .env   # set DATABASE_URL and BETTER_AUTH_SECRET
bun run db:migrate     # create the auth tables
bun dev                # http://localhost:3000
bun run build
bun run lint
```

## Project Structure

```text
src/
├── app/          # App Router pages, layouts, loading/error states
├── components/
│   ├── layout/   # Header, footer, announcement bar, drawers
│   ├── products/ # Product card, gallery, filters, variants
│   ├── cart/     # Cart lines, summary
│   ├── checkout/ # Checkout sections and forms
│   ├── account/  # Account pages UI
│   ├── auth/     # Sign-in / sign-up forms, sign-out button
│   ├── admin/    # Admin dashboard UI
│   ├── hero/     # Homepage hero
│   ├── navigation/ # Nav menus, breadcrumbs
│   ├── forms/    # Reusable form controls
│   ├── animations/ # FadeIn, FadeUp, Stagger primitives
│   └── ui/       # shadcn/ui primitives
├── data/         # Mock catalog data (products, categories, reviews…)
├── store/        # Zustand stores (cart, wishlist, ui)
├── lib/          # Repository layer, auth, db client, formatters, SEO helpers
├── types/        # Shared TypeScript models
├── hooks/        # Shared React hooks
└── proxy.ts      # Session-cookie route protection (Next.js 16 "proxy")
```

## Authentication

Email/password auth via **Better Auth**, with sessions persisted in PostgreSQL.

| Path | Purpose |
|---|---|
| `src/lib/auth.ts` | Server instance (Drizzle adapter, email/password, session cookie cache) |
| `src/lib/auth-client.ts` | Browser client (`signIn`, `signUp`, `signOut`, `useSession`) |
| `src/lib/session.ts` | Server helpers — `getSession`, `getCurrentUser`, `requireCurrentUser` |
| `src/app/api/auth/[...all]/route.ts` | Catch-all `/api/auth/*` handler |
| `src/app/auth/{sign-in,sign-up}` | Auth pages |
| `src/lib/db/` + `drizzle/` | Drizzle schema and SQL migrations |

`/account/*` and `/admin/*` are protected by `src/proxy.ts` (optimistic cookie check)
and re-checked server-side in each page. Environment variables are documented in
`.env.example`; apply schema changes with `bun run db:generate && bun run db:migrate`.

## Architecture Principle

```text
NOW:    Mock data → Repository → Next.js UI
FUTURE: Go REST API → API repository → Next.js UI
```

Presentation never imports mock arrays directly — see PRD §44/§62 in `AGENTS.md`.

## Roadmap

Built phase-by-phase (PRD §56): foundation → design system → layout → data layer → homepage → listing → product detail → SEO → cart/wishlist → search → checkout → account → admin → polish.

