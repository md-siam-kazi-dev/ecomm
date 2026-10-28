import { CURRENCY } from "@/lib/constants";
import type { Category } from "@/types/category";
import type { Product, ProductImage } from "@/types/product";
import type { Review } from "@/types/review";

/**
 * Shared public-catalog client. Works from server components (fetch the Go API
 * directly) and the browser (same-origin `/api/v1/*`, proxied by next.config).
 * Every call is resilient: on failure it returns empty so pages render their
 * empty states instead of crashing.
 */

const BACKEND_API_URL = process.env.BACKEND_API_URL ?? "http://localhost:8080";

export interface PaginationMeta {
  page: number;
  limit: number;
  total: number;
  total_pages: number;
  has_next: boolean;
  has_previous: boolean;
}

interface Envelope<T> {
  success: boolean;
  data?: T;
  pagination?: PaginationMeta;
}

/**
 * The outcome of a catalogue request, so callers can distinguish a genuine 404
 * from a backend/network failure. A transient outage must never be presented as
 * "not found" (that would 404 every product/category page — SEO plan, Part 1).
 */
export type FetchResult<T> =
  | { status: "ok"; data: T }
  | { status: "not-found" }
  | { status: "error"; message: string };

/** A page of list results plus the API's pagination metadata, when supplied. */
export interface FetchListResult<T> {
  items: T[];
  pagination: PaginationMeta | null;
}

/** API shapes (snake_case) mirroring internal/catalog DTOs. */
export interface ApiProduct {
  id: string;
  slug: string;
  name: string;
  short_description: string;
  description: string;
  price: number;
  compare_at_price?: number;
  sku: string;
  stock: number;
  image_url?: string;
  image_alt?: string;
  category: string;
  category_slug: string;
  brand?: string;
  store_name?: string;
  store_slug?: string;
  rating: number;
  review_count: number;
  is_featured: boolean;
  is_new: boolean;
  is_best_seller: boolean;
  specifications: { label: string; value: string }[];
  seo_title?: string;
  seo_description?: string;
  created_at: string;
  updated_at: string;
}

export interface ApiCategory {
  id: string;
  slug: string;
  name: string;
  description: string;
  image_url?: string;
  image_alt?: string;
  product_count: number;
}

export interface ApiReview {
  id: string;
  product_id: string;
  product_name?: string;
  product_slug?: string;
  author: string;
  avatar_url?: string;
  rating: number;
  title: string;
  comment: string;
  is_verified_purchase: boolean;
  helpful_count: number;
  images?: string[];
  vendor_reply?: string;
  vendor_replied_at?: string;
  created_at: string;
}

/** A product's public rating aggregate (from GET /products/{slug}/rating). */
export interface ProductRatingSummary {
  average: number;
  total: number;
  /** Star value ("1".."5") to count. */
  distribution: Record<string, number>;
}

/** Public answer to a product question. */
export interface ProductAnswer {
  id: string;
  answer: string;
  isOfficialVendorAnswer: boolean;
  author: string;
  createdAt: string;
}

/** Public product question with its answers. */
export interface ProductQuestion {
  id: string;
  question: string;
  author: string;
  createdAt: string;
  answers: ProductAnswer[];
}

interface ApiAnswer {
  id: string;
  answer: string;
  is_official_vendor_answer: boolean;
  author: string;
  created_at: string;
}

export interface ApiQuestion {
  id: string;
  question: string;
  author: string;
  created_at: string;
  answers: ApiAnswer[];
}

function apiBase(): string {
  return typeof window === "undefined" ? BACKEND_API_URL : "";
}

function errorMessage(error: unknown): string {
  return error instanceof Error ? error.message : "Request failed";
}

/**
 * Next.js signals that a route must render dynamically (e.g. because a fetch
 * used `cache: "no-store"`) by throwing an internal error carrying a `digest`.
 * Those must never be swallowed: doing so makes a statically-analysed build
 * fail instead of bailing the route out to dynamic rendering.
 */
function isNextControlFlowError(error: unknown): boolean {
  return typeof error === "object" && error !== null && "digest" in error;
}

/**
 * Fetches a single resource and reports exactly what happened:
 * - `ok`        — 2xx response with data in the envelope;
 * - `not-found` — a 404, or a 2xx envelope with no data (a genuine "unknown");
 * - `error`     — any other non-2xx response, or a network/parse failure.
 */
export async function fetchOne<T>(path: string): Promise<FetchResult<T>> {
  try {
    const res = await fetch(`${apiBase()}${path}`, { cache: "no-store" });
    if (res.status === 404) return { status: "not-found" };
    if (!res.ok) {
      return {
        status: "error",
        message: `Request failed with status ${res.status}`,
      };
    }
    const payload = (await res.json()) as Envelope<T>;
    if (!payload.data) return { status: "not-found" };
    return { status: "ok", data: payload.data };
  } catch (error) {
    if (isNextControlFlowError(error)) throw error;
    return { status: "error", message: errorMessage(error) };
  }
}

/**
 * Fetches a list and reports the outcome. An empty list is a *successful*
 * response (`items: []`), never an error; a 404 or any other failure is
 * surfaced as such so listing callers can throw instead of rendering an
 * empty catalogue.
 */
export async function fetchListResult<T>(
  path: string
): Promise<FetchResult<FetchListResult<T>>> {
  try {
    const res = await fetch(`${apiBase()}${path}`, { cache: "no-store" });
    if (res.status === 404) return { status: "not-found" };
    if (!res.ok) {
      return {
        status: "error",
        message: `Request failed with status ${res.status}`,
      };
    }
    const payload = (await res.json()) as Envelope<T[]>;
    return {
      status: "ok",
      data: {
        items: payload.data ?? [],
        pagination: payload.pagination ?? null,
      },
    };
  } catch (error) {
    if (isNextControlFlowError(error)) throw error;
    return { status: "error", message: errorMessage(error) };
  }
}

/**
 * Convenience single-resource getter: `null` for any failure (404 *or*
 * backend error). Use {@link fetchOne} when a genuine 404 must be told apart
 * from an outage.
 */
export async function fetchJson<T>(path: string): Promise<T | null> {
  const result = await fetchOne<T>(path);
  return result.status === "ok" ? result.data : null;
}

/**
 * Convenience list getter: an empty page for any failure. Use
 * {@link fetchListResult} when the caller must surface backend errors.
 */
export async function fetchList<T>(
  path: string
): Promise<{ items: T[]; pagination: PaginationMeta | null }> {
  const result = await fetchListResult<T>(path);
  return result.status === "ok" ? result.data : { items: [], pagination: null };
}

export function toProduct(p: ApiProduct): Product {
  const imageUrl = p.image_url ?? "";
  const hasImage = imageUrl.length > 0;

  // Images are never fabricated: without a real `image_url` the product keeps an
  // empty gallery and an empty thumbnail, which callers render as a placeholder.
  const thumbnail: ProductImage = hasImage
    ? { url: imageUrl, alt: p.image_alt || `${p.name} — product image` }
    : { url: "", alt: "" };

  return {
    id: p.id,
    slug: p.slug,
    name: p.name,
    shortDescription: p.short_description,
    description: p.description,
    price: p.price,
    compareAtPrice: p.compare_at_price ?? undefined,
    currency: CURRENCY,
    images: hasImage ? [thumbnail] : [],
    thumbnail,
    category: p.category,
    categorySlug: p.category_slug,
    brand: p.brand ?? "",
    storeName: p.store_name || undefined,
    storeSlug: p.store_slug || undefined,
    rating: p.review_count > 0 ? p.rating : undefined,
    reviewCount: p.review_count,
    stock: p.stock,
    sku: p.sku,
    tags: [],
    features: [],
    specifications: p.specifications ?? [],
    variants: [],
    isFeatured: p.is_featured,
    isNew: p.is_new,
    isBestSeller: p.is_best_seller,
    createdAt: p.created_at,
    updatedAt: p.updated_at || p.created_at,
    seo: {
      title: p.seo_title || p.name,
      description: p.seo_description || p.short_description,
      keywords: [],
      imageAlt: thumbnail.alt,
    },
  };
}

export function toApiCategory(c: ApiCategory): Category & { productCount: number } {
  return {
    id: c.id,
    slug: c.slug,
    name: c.name,
    description: c.description,
    image: { url: c.image_url ?? "", alt: c.image_alt || `${c.name} category` },
    productCount: c.product_count,
  };
}

export function toReview(r: ApiReview): Review {
  return {
    id: r.id,
    productId: r.product_id,
    author: r.author,
    avatarUrl: r.avatar_url || undefined,
    rating: r.rating,
    title: r.title,
    comment: r.comment,
    createdAt: r.created_at,
    isVerifiedPurchase: r.is_verified_purchase,
    helpfulCount: r.helpful_count,
    images: r.images,
    productName: r.product_name,
    productSlug: r.product_slug,
    vendorReply: r.vendor_reply || undefined,
    vendorRepliedAt: r.vendor_replied_at || undefined,
  };
}

export function toQuestion(q: ApiQuestion): ProductQuestion {
  return {
    id: q.id,
    question: q.question,
    author: q.author,
    createdAt: q.created_at,
    answers: (q.answers ?? []).map((a) => ({
      id: a.id,
      answer: a.answer,
      isOfficialVendorAnswer: a.is_official_vendor_answer,
      author: a.author,
      createdAt: a.created_at,
    })),
  };
}
