/**
 * Cart API — the server is the single source of truth for the cart. Every call
 * returns the full cart so callers can replace their local copy atomically.
 */

import { apiDelete, apiGet, apiPost, apiPut, type ApiResult } from "@/lib/api-client";
import type { CartApiResponse } from "@/types/cart";

export function getCart(): Promise<ApiResult<CartApiResponse>> {
  return apiGet<CartApiResponse>("/api/v1/cart");
}

export function addCartItem(input: {
  productId: string;
  variantId?: string;
  quantity: number;
}): Promise<ApiResult<CartApiResponse>> {
  return apiPost<CartApiResponse>("/api/v1/cart/items", {
    product_id: input.productId,
    variant_id: input.variantId ?? "",
    quantity: input.quantity,
  });
}

export function updateCartItem(
  itemId: string,
  quantity: number
): Promise<ApiResult<CartApiResponse>> {
  return apiPut<CartApiResponse>(`/api/v1/cart/items/${itemId}`, { quantity });
}

export function removeCartItem(
  itemId: string
): Promise<ApiResult<CartApiResponse>> {
  return apiDelete<CartApiResponse>(`/api/v1/cart/items/${itemId}`);
}

export function clearCart(): Promise<ApiResult<CartApiResponse>> {
  return apiDelete<CartApiResponse>("/api/v1/cart");
}
