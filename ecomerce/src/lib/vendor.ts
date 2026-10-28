import { cookies } from "next/headers";

import { CURRENCY, LOW_STOCK_THRESHOLD } from "@/lib/constants";
import { calculateDiscountPercentage } from "@/lib/formatters";
import type { Product } from "@/types/product";

/**
 * Vendor data layer — server-side helpers that forward the caller's cookies to
 * the Go API (`/api/v1/vendor/*`). The backend is store-scoped: it resolves the
 * signed-in user's active store and returns only that store's data.
 */

const BACKEND_API_URL = process.env.BACKEND_API_URL ?? "http://localhost:8080";

async function fetchVendor<T>(path: string): Promise<T | null> {
  const cookieHeader = (await cookies()).toString();
  if (!cookieHeader) return null;
  try {
    const res = await fetch(`${BACKEND_API_URL}${path}`, {
      headers: { cookie: cookieHeader },
      cache: "no-store",
    });
    if (!res.ok) return null;
    const payload = (await res.json()) as { data?: T };
    return payload.data ?? null;
  } catch {
    return null;
  }
}

// ---- products -------------------------------------------------------------

interface VendorProductApi {
  id: string;
  name: string;
  slug: string;
  price: number;
  stock: number;
  image_url?: string;
  created_at: string;
}

/** The store's products (real data). */
export async function getVendorProducts(): Promise<Product[]> {
  const items = await fetchVendor<VendorProductApi[]>(
    "/api/v1/vendor/products?limit=100"
  );

  return (items ?? []).map((product) => ({
    id: product.id,
    slug: product.slug,
    name: product.name,
    shortDescription: "",
    description: "",
    price: product.price,
    currency: CURRENCY,
    images: product.image_url
      ? [{ url: product.image_url, alt: `${product.name} image` }]
      : [],
    thumbnail: { url: product.image_url ?? "", alt: `${product.name} image` },
    category: "",
    categorySlug: "",
    brand: "",
    reviewCount: 0,
    stock: product.stock,
    sku: "",
    tags: [],
    features: [],
    specifications: [],
    variants: [],
    isFeatured: false,
    isNew: false,
    isBestSeller: false,
    createdAt: product.created_at,
  }));
}

/** Store products at or below the low-stock threshold. */
export async function getVendorLowStock(): Promise<Product[]> {
  return (await getVendorProducts())
    .filter((product) => product.stock > 0 && product.stock <= LOW_STOCK_THRESHOLD)
    .sort((a, b) => a.stock - b.stock);
}

/** Percentage discount for a store product, or null. */
export function vendorDiscount(product: Product): number | null {
  return calculateDiscountPercentage(product.price, product.compareAtPrice);
}

// ---- orders ---------------------------------------------------------------

/** A row in the store's order list. */
export interface VendorOrderSummary {
  id: string;
  order_number: string;
  status: string;
  item_count: number;
  total: number;
  created_at: string;
}

/** Raw order row from the API (older builds expose `order_id`). */
interface VendorOrderApi extends Omit<VendorOrderSummary, "id"> {
  id?: string;
  order_id?: string;
}

export interface VendorAddress {
  full_name: string;
  phone?: string;
  line1: string;
  line2?: string;
  city: string;
  state?: string;
  postal_code?: string;
  country: string;
}

export interface VendorOrderLine {
  order_item_id: string;
  product_id?: string;
  name: string;
  sku?: string;
  quantity: number;
  unit_price: number;
  line_total: number;
}

export interface VendorShipment {
  id: string;
  carrier?: string;
  tracking_number?: string;
  status: string;
  shipped_at?: string;
  delivered_at?: string;
}

export interface VendorTrackingEvent {
  id: string;
  status: string;
  location?: string;
  description?: string;
  event_time: string;
}

export interface VendorOrderDetail {
  id: string;
  order_number: string;
  status: string;
  currency: string;
  grand_total: number;
  created_at: string;
  shipping_address: VendorAddress;
  items: VendorOrderLine[];
  vendor_revenue: number;
  vendor_item_count: number;
  shipment?: VendorShipment;
  events: VendorTrackingEvent[];
}

/** Orders that contain the store's items (real data). */
export async function getVendorOrders(): Promise<VendorOrderSummary[]> {
  const items = await fetchVendor<VendorOrderApi[]>("/api/v1/vendor/orders?limit=50");
  return (items ?? []).map((order) => ({
    id: order.id ?? order.order_id ?? "",
    order_number: order.order_number,
    status: order.status,
    item_count: order.item_count,
    total: order.total,
    created_at: order.created_at,
  }));
}

/** One store order with its items, address and shipment timeline. */
export async function getVendorOrder(id: string): Promise<VendorOrderDetail | null> {
  return fetchVendor<VendorOrderDetail>(`/api/v1/vendor/orders/${encodeURIComponent(id)}`);
}

// ---- dashboard + revenue --------------------------------------------------

export interface VendorSummary {
  revenue: number;
  order_count: number;
  product_count: number;
  low_stock_count: number;
  pending_fulfilment: number;
}

export async function getVendorSummary(): Promise<VendorSummary | null> {
  return fetchVendor<VendorSummary>("/api/v1/vendor/dashboard");
}

export interface VendorRevenuePoint {
  month: string;
  revenue: number;
  orders: number;
}

export interface VendorTopProduct {
  product_id: string;
  name: string;
  units_sold: number;
  revenue: number;
}

export interface VendorStatusCount {
  status: string;
  count: number;
  revenue: number;
}

export interface VendorRevenue {
  revenue: number;
  order_count: number;
  units_sold: number;
  average_order_value: number;
  series: VendorRevenuePoint[];
  top_products: VendorTopProduct[];
  statuses: VendorStatusCount[];
}

export async function getVendorRevenue(): Promise<VendorRevenue | null> {
  return fetchVendor<VendorRevenue>("/api/v1/vendor/revenue");
}

// ---- reviews --------------------------------------------------------------

/** One review on the store's products (GET /api/v1/vendor/reviews). */
export interface VendorReview {
  id: string;
  product_id: string;
  product_name: string;
  product_slug: string;
  product_image?: string;
  author: string;
  order_number?: string;
  rating: number;
  title: string;
  comment: string;
  images?: string[];
  created_at: string;
  replied: boolean;
  vendor_reply?: string;
  vendor_replied_at?: string;
}

/** The store's aggregate star feedback (GET /api/v1/vendor/reviews/summary). */
export interface VendorReviewSummary {
  average: number;
  total: number;
  distribution: Record<string, number>;
  unreplied: number;
}

// ---- product questions ----------------------------------------------------

/** One question on the store's products (GET /api/v1/vendor/questions). */
export interface VendorQuestion {
  id: string;
  product_id: string;
  product_name: string;
  product_slug: string;
  product_image?: string;
  author: string;
  question: string;
  created_at: string;
  answered: boolean;
  answer?: string;
  answer_created_at?: string;
}
