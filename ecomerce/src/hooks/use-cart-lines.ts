"use client";

import { useCartStore, type CartLine } from "@/store/cart-store";

export type { CartLine };

/**
 * Cart lines from the server-backed cart store. The server already returns
 * display-ready lines (name, image, price, stock), so no reconstruction or
 * product lookups are needed here.
 */
export function useCartLines(): CartLine[] {
  return useCartStore((state) => state.items);
}
