import { cookies } from "next/headers";

import type { Order, OrderItem, OrderShippingAddress, OrderStatus } from "@/types/order";

/**
 * Server-side order helpers. They forward the incoming request cookies to the
 * Go API (`/api/v1/orders`), matching the pattern in `lib/session.ts`. The
 * backend scopes every query to the authenticated user.
 */

const BACKEND_API_URL = process.env.BACKEND_API_URL ?? "http://localhost:8080";

interface ApiOrderItemReview {
  id: string;
  rating: number;
  title: string;
  comment: string;
  images?: string[];
  created_at: string;
}

interface ApiOrderItem {
  order_item_id?: string;
  product_id: string;
  variant_id?: string;
  product_name: string;
  product_sku?: string;
  quantity: number;
  unit_price: number;
  line_total: number;
  image_url?: string;
  reviewed?: boolean;
  review?: ApiOrderItemReview;
}

interface ApiShippingAddress {
  full_name: string;
  line1: string;
  line2?: string;
  city: string;
  state?: string;
  postal_code?: string;
  country: string;
}

export interface ApiOrder {
  id: string;
  order_number: string;
  status: string;
  currency: string;
  subtotal: number;
  shipping_total: number;
  tax_total: number;
  grand_total: number;
  created_at: string;
  items: ApiOrderItem[];
  shipping_address?: ApiShippingAddress;
}

function toShippingAddress(a: ApiShippingAddress | undefined): OrderShippingAddress | undefined {
  if (!a) return undefined;
  return {
    fullName: a.full_name,
    line1: a.line1,
    line2: a.line2,
    city: a.city,
    state: a.state ?? "",
    postalCode: a.postal_code ?? "",
    country: a.country,
  };
}

export function toOrder(a: ApiOrder): Order {
  const items: OrderItem[] = (a.items ?? []).map((item) => ({
    productId: item.product_id,
    variantId: item.variant_id,
    orderItemId: item.order_item_id,
    name: item.product_name,
    sku: item.product_sku,
    quantity: item.quantity,
    price: item.unit_price,
    imageUrl: item.image_url,
    reviewed: item.reviewed,
    review: item.review
      ? {
          id: item.review.id,
          rating: item.review.rating,
          title: item.review.title,
          comment: item.review.comment,
          images: item.review.images,
          createdAt: item.review.created_at,
        }
      : undefined,
  }));

  return {
    id: a.id,
    orderNumber: a.order_number,
    customerId: "",
    items,
    subtotal: a.subtotal,
    shipping: a.shipping_total,
    tax: a.tax_total,
    total: a.grand_total,
    status: a.status as OrderStatus,
    createdAt: a.created_at,
    shippingAddress: toShippingAddress(a.shipping_address),
  };
}

async function fetchApi<T>(path: string): Promise<T | null> {
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

export async function getMyOrders(): Promise<Order[]> {
  const data = await fetchApi<ApiOrder[]>("/api/v1/orders");
  return (data ?? []).map(toOrder);
}

export async function getMyOrder(id: string): Promise<Order | null> {
  const data = await fetchApi<ApiOrder>(`/api/v1/orders/${encodeURIComponent(id)}`);
  return data ? toOrder(data) : null;
}
