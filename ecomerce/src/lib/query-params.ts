import type { ProductSortOption } from "@/lib/products";

/**
 * URL-as-state helpers for the catalog listing pages (PRD §12/§45):
 * filters, sort and page live in search params so every view is shareable
 * and server-renderable.
 */

export type CatalogSearchParams = Record<string, string | string[] | undefined>;

export interface CatalogFilters {
  category?: string;
  query?: string;
  minPrice?: number;
  maxPrice?: number;
  /** Minimum average rating (2–4). */
  rating?: number;
  inStock: boolean;
  onSale: boolean;
  sort: ProductSortOption;
  /** 1-based "page" — load-more multiplies the page size. */
  page: number;
}

const SORT_OPTIONS: ProductSortOption[] = [
  "featured",
  "newest",
  "price-asc",
  "price-desc",
  "rating",
  "discount",
];

function firstValue(sp: CatalogSearchParams, key: string): string | undefined {
  const value = sp[key];
  return (Array.isArray(value) ? value[0] : value) || undefined;
}

function numberValue(sp: CatalogSearchParams, key: string): number | undefined {
  const raw = firstValue(sp, key);
  if (!raw) return undefined;
  const parsed = Number(raw);
  return Number.isFinite(parsed) ? parsed : undefined;
}

/** Parses and validates raw search params into a typed filter set. */
export function parseCatalogFilters(sp: CatalogSearchParams): CatalogFilters {
  const sortRaw = firstValue(sp, "sort") as ProductSortOption | undefined;

  return {
    category: firstValue(sp, "category"),
    query: firstValue(sp, "q"),
    minPrice: numberValue(sp, "min"),
    maxPrice: numberValue(sp, "max"),
    rating: numberValue(sp, "rating"),
    inStock: firstValue(sp, "stock") === "1",
    onSale: firstValue(sp, "sale") === "1",
    sort: sortRaw && SORT_OPTIONS.includes(sortRaw) ? sortRaw : "featured",
    page: Math.max(1, Math.trunc(numberValue(sp, "page") ?? 1)),
  };
}

/** Inverse of parse — used to rebuild links from a filter set. */
export function serializeCatalogFilters(
  filters: CatalogFilters
): URLSearchParams {
  const params = new URLSearchParams();

  if (filters.category) params.set("category", filters.category);
  if (filters.query) params.set("q", filters.query);
  if (filters.minPrice !== undefined) params.set("min", String(filters.minPrice));
  if (filters.maxPrice !== undefined) params.set("max", String(filters.maxPrice));
  if (filters.rating !== undefined) params.set("rating", String(filters.rating));
  if (filters.inStock) params.set("stock", "1");
  if (filters.onSale) params.set("sale", "1");
  if (filters.sort !== "featured") params.set("sort", filters.sort);
  if (filters.page > 1) params.set("page", String(filters.page));

  return params;
}

export function hasActiveFilters(
  filters: CatalogFilters,
  categoryIsFilter: boolean
): boolean {
  return Boolean(
    (categoryIsFilter && filters.category) ||
      filters.query ||
      filters.minPrice !== undefined ||
      filters.maxPrice !== undefined ||
      filters.rating !== undefined ||
      filters.inStock ||
      filters.onSale
  );
}

/**
 * Builds a catalog URL from the current params with incremental updates.
 * `null` removes a param; default values (sort=featured, page=1) are
 * normalized away to keep URLs clean (PRD §45).
 */
export function buildCatalogPath(
  pathname: string,
  current: URLSearchParams,
  updates: Record<string, string | number | null>
): string {
  const next = new URLSearchParams(current);

  for (const [key, value] of Object.entries(updates)) {
    if (value === null) next.delete(key);
    else next.set(key, String(value));
  }

  if (next.get("sort") === "featured") next.delete("sort");
  if (next.get("page") === "1") next.delete("page");

  const queryString = next.toString();
  return queryString ? `${pathname}?${queryString}` : pathname;
}
