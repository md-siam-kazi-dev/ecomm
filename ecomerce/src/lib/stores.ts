import { cookies } from "next/headers";

import {
  fetchListResult,
  fetchOne,
  toProduct,
  type ApiProduct,
  type PaginationMeta,
} from "@/lib/catalog";
import type { Product } from "@/types/product";

/**
 * Public storefront store data layer. Store profile + products are public; the
 * follow state needs the caller's cookies (returns null when signed out).
 */

const BACKEND_API_URL = process.env.BACKEND_API_URL ?? "http://localhost:8080";

export interface Store {
  id: string;
  slug: string;
  name: string;
  description: string;
  logoUrl?: string;
  bannerUrl?: string;
  category?: string;
  productCount: number;
  followerCount: number;
  reviewCount: number;
  /** Share of reviews rated 4–5 stars, 0–100. */
  positiveFeedback: number;
  /** ISO date the store was last changed — drives sitemap `lastmod`. */
  updatedAt?: string;
}

interface ApiStore {
  id: string;
  name: string;
  slug: string;
  description: string;
  logo_url?: string;
  banner_url?: string;
  category?: string;
  product_count: number;
  follower_count: number;
  review_count: number;
  positive_feedback: number;
  updated_at?: string;
}

export interface StoreFollowStatus {
  following: boolean;
  followerCount: number;
}

function toStore(s: ApiStore): Store {
  return {
    id: s.id,
    slug: s.slug,
    name: s.name,
    description: s.description,
    logoUrl: s.logo_url || undefined,
    bannerUrl: s.banner_url || undefined,
    category: s.category || undefined,
    productCount: s.product_count,
    followerCount: s.follower_count,
    reviewCount: s.review_count,
    positiveFeedback: s.positive_feedback,
    updatedAt: s.updated_at || undefined,
  };
}

/**
 * A single public store. Returns `null` **only** for a genuine 404 (unknown or
 * inactive slug); a backend/network failure throws so the route renders a 500
 * instead of a deindexing false "not found" (SEO plan, Part 1 §2).
 */
export async function getStoreBySlug(slug: string): Promise<Store | null> {
  const result = await fetchOne<ApiStore>(
    `/api/v1/stores/${encodeURIComponent(slug)}`
  );
  if (result.status === "ok") return toStore(result.data);
  if (result.status === "not-found") return null;
  throw new Error(`Failed to load store "${slug}": ${result.message}`);
}

/** A page of a store's products. A backend failure throws (never empty). */
export async function getStoreProducts(
  slug: string,
  query: { page?: number; pageSize?: number; sort?: string } = {}
): Promise<{ items: Product[]; pagination: PaginationMeta | null }> {
  const page = Math.max(1, query.page ?? 1);
  const pageSize = Math.min(48, Math.max(1, query.pageSize ?? 12));
  const params = new URLSearchParams({
    page: String(page),
    limit: String(pageSize),
    sort: query.sort ?? "featured",
  });
  const result = await fetchListResult<ApiProduct>(
    `/api/v1/stores/${encodeURIComponent(slug)}/products?${params.toString()}`
  );
  if (result.status !== "ok") {
    throw new Error(
      result.status === "not-found"
        ? `Failed to load products for store "${slug}": not found`
        : `Failed to load products for store "${slug}": ${result.message}`
    );
  }
  return {
    items: result.data.items.map(toProduct),
    pagination: result.data.pagination,
  };
}

/**
 * Follow state for a store. Works for guests too (following=false + follower
 * count), so the count always renders even when signed out.
 */
export async function getStoreFollow(
  slug: string
): Promise<StoreFollowStatus | null> {
  const cookieHeader = (await cookies()).toString();
  try {
    const res = await fetch(
      `${BACKEND_API_URL}/api/v1/stores/${encodeURIComponent(slug)}/follow`,
      {
        headers: cookieHeader ? { cookie: cookieHeader } : undefined,
        cache: "no-store",
      }
    );
    if (!res.ok) return null;
    const payload = (await res.json()) as {
      data?: { following: boolean; follower_count: number };
    };
    if (!payload.data) return null;
    return {
      following: payload.data.following,
      followerCount: payload.data.follower_count,
    };
  } catch {
    return null;
  }
}
