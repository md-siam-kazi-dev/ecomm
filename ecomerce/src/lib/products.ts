import {
  fetchJson,
  fetchList,
  fetchListResult,
  fetchOne,
  toApiCategory,
  toProduct,
  toQuestion,
  toReview,
  type ApiCategory,
  type ApiProduct,
  type ApiQuestion,
  type ApiReview,
  type FetchListResult,
  type FetchResult,
  type PaginationMeta,
  type ProductQuestion,
  type ProductRatingSummary,
} from "@/lib/catalog";
import { LOW_STOCK_THRESHOLD } from "@/lib/constants";
import type { Category } from "@/types/category";
import type { Product } from "@/types/product";
import type { Review } from "@/types/review";

export type ProductSortOption =
  | "featured"
  | "newest"
  | "price-asc"
  | "price-desc"
  | "rating"
  | "discount";

export interface ProductQuery {
  categorySlug?: string;
  /** Free-text search across name, brand, category and description. */
  query?: string;
  minPrice?: number;
  maxPrice?: number;
  /** Inclusive minimum average rating (1–5). */
  minRating?: number;
  /** Only products with stock > 0. */
  inStockOnly?: boolean;
  /** Only products with a compare-at price. */
  onSaleOnly?: boolean;
  sort?: ProductSortOption;
  /** 1-based page number. */
  page?: number;
  pageSize?: number;
}

export interface ProductListResult {
  items: Product[];
  total: number;
  page: number;
  pageSize: number;
  pageCount: number;
}

export interface CategoryWithCount extends Category {
  productCount: number;
}

const DEFAULT_PAGE_SIZE = 12;
const MAX_PAGE_SIZE = 48;

/**
 * Unwraps a successful list response, throwing for a backend/network failure.
 * An outage must never be presented as "no results" — the page should surface a
 * 500 instead (SEO plan, Part 1 §2).
 */
function requireList<T>(
  result: FetchResult<FetchListResult<T>>,
  subject: string
): FetchListResult<T> {
  if (result.status === "ok") return result.data;
  throw new Error(
    result.status === "not-found"
      ? `Failed to load ${subject}: not found`
      : `Failed to load ${subject}: ${result.message}`
  );
}

/**
 * Catalog repository — the single source of product data for the UI, backed by
 * the Go API (`internal/catalog`). Same signatures as the previous mock layer.
 */

export async function getProducts(
  query: ProductQuery = {}
): Promise<ProductListResult> {
  const page = Math.max(1, query.page ?? 1);
  const pageSize = Math.min(
    MAX_PAGE_SIZE,
    Math.max(1, query.pageSize ?? DEFAULT_PAGE_SIZE)
  );

  const params = new URLSearchParams({
    page: String(page),
    limit: String(pageSize),
    sort: query.sort ?? "featured",
  });
  if (query.categorySlug) params.set("category", query.categorySlug);
  if (query.query) params.set("q", query.query);
  if (query.minPrice !== undefined) params.set("min_price", String(query.minPrice));
  if (query.maxPrice !== undefined) params.set("max_price", String(query.maxPrice));
  if (query.minRating !== undefined) params.set("min_rating", String(query.minRating));
  if (query.inStockOnly) params.set("in_stock", "1");
  if (query.onSaleOnly) params.set("on_sale", "1");

  const { items, pagination } = requireList(
    await fetchListResult<ApiProduct>(`/api/v1/products?${params.toString()}`),
    "products"
  );

  return {
    items: items.map(toProduct),
    total: pagination?.total ?? items.length,
    page: pagination?.page ?? page,
    pageSize,
    pageCount: Math.max(1, pagination?.total_pages ?? 1),
  };
}

/**
 * A single published product. Returns `null` **only** for a genuine 404
 * (unknown or unpublished slug); a backend/network failure throws so the route
 * renders a 500 rather than deindexing the page as "not found".
 */
export async function getProductBySlug(slug: string): Promise<Product | null> {
  const result = await fetchOne<ApiProduct>(
    `/api/v1/products/${encodeURIComponent(slug)}`
  );
  if (result.status === "ok") return toProduct(result.data);
  if (result.status === "not-found") return null;
  throw new Error(`Failed to load product "${slug}": ${result.message}`);
}

export async function getProductById(id: string): Promise<Product | null> {
  const products = await getProductsByIds([id]);
  return products[0] ?? null;
}

/** Resolves a list of product IDs, preserving the input order (cart lines). */
export async function getProductsByIds(ids: string[]): Promise<Product[]> {
  const wanted = ids.filter(Boolean);
  if (wanted.length === 0) return [];

  const params = new URLSearchParams({
    ids: wanted.join(","),
    limit: String(Math.min(60, wanted.length)),
  });
  const { items } = await fetchList<ApiProduct>(`/api/v1/products?${params.toString()}`);

  const byId = new Map(items.map((product) => [product.id, toProduct(product)]));
  return wanted
    .map((id) => byId.get(id))
    .filter((product): product is Product => product !== undefined);
}

/**
 * All public categories. A failed fetch throws (it used to return `[]`, which
 * 404'd every category page during a backend blip).
 */
export async function getCategories(): Promise<CategoryWithCount[]> {
  const { items } = requireList(
    await fetchListResult<ApiCategory>("/api/v1/categories"),
    "categories"
  );
  return items.map(toApiCategory);
}

/**
 * A single category by slug. `null` means the slug genuinely does not exist; a
 * backend/network failure throws via {@link getCategories}.
 */
export async function getCategoryBySlug(
  slug: string
): Promise<CategoryWithCount | null> {
  const categories = await getCategories();
  return categories.find((category) => category.slug === slug) ?? null;
}

export async function getFeaturedProducts(limit = 8): Promise<Product[]> {
  const { items } = await getProducts({ sort: "featured", pageSize: limit });
  return items;
}

export async function getBestSellers(limit = 8): Promise<Product[]> {
  const { items } = await getProducts({ sort: "rating", pageSize: limit });
  return items;
}

export async function getRelatedProducts(
  product: Product,
  limit = 4
): Promise<Product[]> {
  const { items } = requireList(
    await fetchListResult<ApiProduct>(
      `/api/v1/products/${encodeURIComponent(product.slug)}/related?limit=${limit}`
    ),
    `related products for "${product.slug}"`
  );
  return items.map(toProduct);
}

export async function searchProducts(
  query: string,
  limit = 10
): Promise<Product[]> {
  const trimmed = query.trim();
  if (!trimmed) return [];

  const result = await getProducts({ query: trimmed, sort: "featured", pageSize: limit });
  return result.items;
}

/** A page of published reviews for a product, plus pagination metadata. */
export async function getProductReviews(
  slug: string,
  query: { page?: number; pageSize?: number } = {}
): Promise<{ reviews: Review[]; pagination: PaginationMeta | null }> {
  const page = Math.max(1, query.page ?? 1);
  const pageSize = Math.min(50, Math.max(1, query.pageSize ?? 10));
  const params = new URLSearchParams({ page: String(page), limit: String(pageSize) });
  const { items, pagination } = await fetchList<ApiReview>(
    `/api/v1/products/${encodeURIComponent(slug)}/reviews?${params.toString()}`
  );
  return { reviews: items.map(toReview), pagination };
}

/** A product's published rating summary and 1–5 star distribution. */
export async function getProductRating(
  slug: string
): Promise<ProductRatingSummary | null> {
  return fetchJson<ProductRatingSummary>(
    `/api/v1/products/${encodeURIComponent(slug)}/rating`
  );
}

/** A page of published questions (with answers) for a product. */
export async function getProductQuestions(
  slug: string,
  query: { page?: number; pageSize?: number } = {}
): Promise<{ questions: ProductQuestion[]; pagination: PaginationMeta | null }> {
  const page = Math.max(1, query.page ?? 1);
  const pageSize = Math.min(50, Math.max(1, query.pageSize ?? 10));
  const params = new URLSearchParams({ page: String(page), limit: String(pageSize) });
  const { items, pagination } = await fetchList<ApiQuestion>(
    `/api/v1/products/${encodeURIComponent(slug)}/questions?${params.toString()}`
  );
  return { questions: items.map(toQuestion), pagination };
}

export async function searchCategories(
  query: string,
  limit = 4
): Promise<CategoryWithCount[]> {
  const trimmed = query.trim().toLowerCase();
  if (!trimmed) return [];

  const categories = await getCategories();
  return categories
    .filter((category) =>
      `${category.name} ${category.description}`.toLowerCase().includes(trimmed)
    )
    .slice(0, limit);
}

/** In-stock live products at or below the threshold — admin low-stock panel. */
export async function getLowStockProducts(
  threshold: number = LOW_STOCK_THRESHOLD
): Promise<Product[]> {
  const { items } = await getProducts({ pageSize: MAX_PAGE_SIZE });
  return items
    .filter((product) => product.stock > 0 && product.stock <= threshold)
    .sort((a, b) => a.stock - b.stock);
}

export interface ReviewWithProduct extends Review {
  productName: string;
  productSlug: string;
}

/** Top 5-star reviews by helpful votes — homepage testimonials (PRD §9). */
export async function getTopReviews(limit = 3): Promise<ReviewWithProduct[]> {
  const { items } = requireList(
    await fetchListResult<ApiReview>(`/api/v1/reviews/top?limit=${limit}`),
    "top reviews"
  );
  return items.map((review) => ({
    ...toReview(review),
    productName: review.product_name ?? "a ShopKart product",
    productSlug: review.product_slug ?? "",
  }));
}
