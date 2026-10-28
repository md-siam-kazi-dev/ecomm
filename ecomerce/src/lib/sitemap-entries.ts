import type { MetadataRoute } from "next";

import { SITE_CONFIG } from "./constants";

/**
 * Pure, I/O-free helpers for building the XML sitemap. Kept separate from the
 * `/sitemap.xml` route so the URL-building, de-duplication and cursor logic can
 * be unit-tested without a running backend.
 */

/** A change-frequency value accepted by Next's sitemap entries. */
type SitemapChangeFrequency = NonNullable<
  MetadataRoute.Sitemap[number]["changeFrequency"]
>;

/** One indexable resource identified by slug, with its optional last change. */
export interface SitemapSource {
  slug: string;
  /** ISO date the resource was last changed; omitted when unknown. */
  updatedAt?: string;
}

/** A static page worth indexing (homepage, /products, support pages, ...). */
export interface StaticSitemapPage {
  path: string;
  priority: number;
  changeFrequency: SitemapChangeFrequency;
}

export interface SitemapInputs {
  products: SitemapSource[];
  categories: SitemapSource[];
  stores: SitemapSource[];
  staticPages: StaticSitemapPage[];
}

/** One row as returned by the backend sitemap endpoints. */
export interface SitemapRow {
  id: string;
  slug: string;
  updated_at: string;
}

const CATEGORY_ENTRY = { changeFrequency: "weekly", priority: 0.7 } as const;
const STORE_ENTRY = { changeFrequency: "weekly", priority: 0.6 } as const;
const PRODUCT_ENTRY = { changeFrequency: "weekly", priority: 0.8 } as const;

/** Parses a timestamp, returning `undefined` when it is missing or invalid. */
function parseTimestamp(value?: string): Date | undefined {
  if (!value) return undefined;
  const parsed = new Date(value);
  return Number.isNaN(parsed.getTime()) ? undefined : parsed;
}

/** Builds one slug-based entry (category, store or product), or `null` when the
 * slug is empty so the caller can skip it. */
function slugEntry(
  basePath: string,
  source: SitemapSource,
  changeFrequency: SitemapChangeFrequency,
  priority: number
): MetadataRoute.Sitemap[number] | null {
  const slug = typeof source.slug === "string" ? source.slug.trim() : "";
  if (!slug) return null;

  const url = new URL(
    `${basePath}/${encodeURIComponent(slug)}`,
    SITE_CONFIG.url
  ).toString();
  const lastModified = parseTimestamp(source.updatedAt);

  // `lastModified` is emitted only for a real timestamp — never a fabricated one.
  return {
    url,
    ...(lastModified ? { lastModified } : {}),
    changeFrequency,
    priority,
  };
}

/**
 * Builds the sitemap entries: static pages first, then categories, stores and
 * products. Entries are de-duplicated by absolute URL (first occurrence wins) so
 * a resource reachable through two paths appears once, and any source with an
 * empty slug is dropped.
 */
export function buildSitemapEntries(
  input: SitemapInputs
): MetadataRoute.Sitemap {
  const entries: MetadataRoute.Sitemap = [];
  const seen = new Set<string>();

  const add = (entry: MetadataRoute.Sitemap[number] | null): void => {
    if (!entry || seen.has(entry.url)) return;
    seen.add(entry.url);
    entries.push(entry);
  };

  for (const page of input.staticPages) {
    add({
      url: new URL(page.path, SITE_CONFIG.url).toString(),
      changeFrequency: page.changeFrequency,
      priority: page.priority,
    });
  }
  for (const category of input.categories) {
    add(
      slugEntry(
        "/category",
        category,
        CATEGORY_ENTRY.changeFrequency,
        CATEGORY_ENTRY.priority
      )
    );
  }
  for (const store of input.stores) {
    add(
      slugEntry(
        "/stores",
        store,
        STORE_ENTRY.changeFrequency,
        STORE_ENTRY.priority
      )
    );
  }
  for (const product of input.products) {
    add(
      slugEntry(
        "/products",
        product,
        PRODUCT_ENTRY.changeFrequency,
        PRODUCT_ENTRY.priority
      )
    );
  }

  return entries;
}

/** base64url encodes a UTF-8 string without padding, matching Go's
 * `base64.RawURLEncoding`. */
function base64UrlEncode(value: string): string {
  const bytes = new TextEncoder().encode(value);
  let binary = "";
  for (const byte of bytes) binary += String.fromCharCode(byte);
  return btoa(binary)
    .replace(/\+/g, "-")
    .replace(/\//g, "_")
    .replace(/=+$/, "");
}

/**
 * Derives the next keyset cursor from the last sitemap row, matching the
 * backend's opaque cursor format: base64url(`<updated_at>|<id>`). Returns `null`
 * when there is no row to continue from.
 */
export function nextSitemapCursor(
  row: SitemapRow | null | undefined
): string | null {
  if (!row) return null;
  const id = typeof row.id === "string" ? row.id.trim() : "";
  const updatedAt =
    typeof row.updated_at === "string" ? row.updated_at.trim() : "";
  if (!id || !updatedAt) return null;
  return base64UrlEncode(`${updatedAt}|${id}`);
}
