import type { MetadataRoute } from "next";

import { getCategories } from "@/lib/products";
import {
  buildSitemapEntries,
  nextSitemapCursor,
  type SitemapRow,
  type SitemapSource,
  type StaticSitemapPage,
} from "@/lib/sitemap-entries";

/**
 * XML sitemap (PRD §30): the homepage, the listing page, every category, every
 * active storefront and every published/approved product. Generated from the
 * public backend keyset endpoint, so unpublished, rejected, draft and deleted
 * products never appear here.
 */

const BACKEND_API_URL = process.env.BACKEND_API_URL ?? "http://localhost:8080";

// Render per request: the sitemap must reflect live catalog state and must not
// be baked at build time (a build-time snapshot could list stale rows).
export const dynamic = "force-dynamic";

// Next still evaluates metadata routes once while collecting page data at build
// time, when the API is typically unreachable. During that pass we degrade to a
// static-only sitemap instead of failing the build; at runtime the route is
// dynamic and a backend outage rethrows (a 5xx is safer for crawlers than a
// truncated sitemap).
const IS_BUILD = process.env.NEXT_PHASE === "phase-production-build";

// Google's per-file ceiling is 50,000 URLs; stay under it and leave headroom for
// the static pages, categories and stores emitted alongside the products.
const MAX_URLS = 45_000;
const PAGE_SIZE = 1_000;

/** Static support/marketing pages worth indexing. */
const STATIC_PAGES: StaticSitemapPage[] = [
  { path: "/", priority: 1, changeFrequency: "weekly" },
  { path: "/products", priority: 0.9, changeFrequency: "daily" },
  { path: "/contact", priority: 0.5, changeFrequency: "monthly" },
  { path: "/shipping-returns", priority: 0.5, changeFrequency: "monthly" },
  { path: "/faq", priority: 0.5, changeFrequency: "monthly" },
  { path: "/privacy-policy", priority: 0.3, changeFrequency: "yearly" },
  { path: "/terms-of-service", priority: 0.3, changeFrequency: "yearly" },
];

interface SitemapEnvelope {
  data?: SitemapRow[];
  pagination?: { total?: number };
}

/** Fetches one backend sitemap page, throwing on any non-2xx so the route
 * surfaces an error rather than emitting a truncated sitemap. */
async function fetchSitemapPage(
  kind: "products" | "stores",
  params: URLSearchParams
): Promise<SitemapEnvelope> {
  const res = await fetch(
    `${BACKEND_API_URL}/api/v1/sitemap/${kind}?${params.toString()}`,
    { cache: "no-store" }
  );
  if (!res.ok) {
    throw new Error(`Sitemap ${kind} request failed with status ${res.status}`);
  }
  return (await res.json()) as SitemapEnvelope;
}

/** Walks the keyset cursor until `limit` rows are collected or the set ends. */
async function fetchSitemapRows(
  kind: "products" | "stores",
  limit: number
): Promise<SitemapSource[]> {
  const collected: SitemapSource[] = [];
  let cursor: string | null = null;

  while (collected.length < limit) {
    const requested = Math.max(1, Math.min(PAGE_SIZE, limit - collected.length));
    const params = new URLSearchParams({ limit: String(requested) });
    if (cursor) params.set("cursor", cursor);

    const payload = await fetchSitemapPage(kind, params);
    const rows = payload.data ?? [];
    if (rows.length === 0) break;

    for (const row of rows) {
      collected.push({ slug: row.slug, updatedAt: row.updated_at });
    }

    cursor = nextSitemapCursor(rows[rows.length - 1]);
    if (!cursor || rows.length < requested) break;
  }

  return collected;
}

export default async function sitemap(): Promise<MetadataRoute.Sitemap> {
  try {
    const [categories, stores, products] = await Promise.all([
      getCategories(),
      fetchSitemapRows("stores", MAX_URLS),
      fetchSitemapRows("products", MAX_URLS),
    ]);

    return buildSitemapEntries({
      staticPages: STATIC_PAGES,
      categories: categories.map((category) => ({
        slug: category.slug,
        updatedAt: category.updatedAt,
      })),
      stores,
      products,
    });
  } catch (error) {
    if (!IS_BUILD) throw error;
    // Build-time fallback: a valid, static-only sitemap so the build succeeds
    // without a reachable API. At runtime the fresh, full sitemap is served.
    return buildSitemapEntries({
      staticPages: STATIC_PAGES,
      categories: [],
      stores: [],
      products: [],
    });
  }
}
