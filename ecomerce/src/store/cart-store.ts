"use client";

import { create } from "zustand";

import * as cartApi from "@/lib/cart";
import type { CartApiResponse } from "@/types/cart";

/**
 * A display-ready cart line. The server returns fully hydrated lines (name,
 * image, price, stock), so the client no longer reconstructs them locally and
 * nothing is persisted to localStorage — the server is the source of truth.
 */
export interface CartLine {
  /** Server-side line id, used to update/remove the line. */
  id: string;
  productId: string;
  variantId: string;
  name: string;
  slug: string;
  imageUrl: string;
  imageAlt: string;
  variantLabel?: string;
  unitPrice: number;
  quantity: number;
  lineTotal: number;
  /** Max purchasable quantity for this line. */
  maxQuantity: number;
  inStock: boolean;
}

export type CartStatus = "idle" | "loading" | "ready" | "error";

interface AddItemInput {
  productId: string;
  variantId?: string;
  quantity?: number;
}

interface CartState {
  items: CartLine[];
  itemCount: number;
  subtotal: number;
  currency: string;
  status: CartStatus;
  /** Loads the caller's cart (guest or user) from the server. */
  fetchCart: () => Promise<void>;
  /** Adds a product; returns false when the server rejects it. */
  addItem: (input: AddItemInput) => Promise<boolean>;
  /** Sets a line quantity (0 removes the line). */
  updateQuantity: (itemId: string, quantity: number) => Promise<void>;
  removeItem: (itemId: string) => Promise<void>;
  clearCart: () => Promise<void>;
  /** Clears local state without a server call (e.g. after checkout). */
  reset: () => void;
}

function mapItems(resp: CartApiResponse): CartLine[] {
  return resp.items.map((item) => ({
    id: item.id,
    productId: item.product_id,
    variantId: item.variant_id,
    name: item.name,
    slug: item.slug,
    imageUrl: item.image_url,
    imageAlt: item.image_alt,
    variantLabel: item.variant_label,
    unitPrice: item.unit_price,
    quantity: item.quantity,
    lineTotal: item.line_total,
    maxQuantity: item.max_quantity,
    inStock: item.in_stock,
  }));
}

function summarize(items: CartLine[]): { itemCount: number; subtotal: number } {
  let subtotal = 0;
  let itemCount = 0;
  for (const line of items) {
    subtotal += line.unitPrice * line.quantity;
    itemCount += line.quantity;
  }
  return { itemCount, subtotal: Math.round(subtotal * 100) / 100 };
}

function apply(resp: CartApiResponse) {
  const items = mapItems(resp);
  return { items, currency: resp.currency, status: "ready" as const, ...summarize(items) };
}

/** Dedupes concurrent cart loads (mount + auth change can both trigger one). */
let inFlight: Promise<void> | null = null;

export const useCartStore = create<CartState>((set, get) => ({
  items: [],
  itemCount: 0,
  subtotal: 0,
  currency: "usd",
  status: "idle",

  fetchCart: () => {
    if (inFlight) return inFlight;
    if (get().status !== "ready") set({ status: "loading" });
    inFlight = (async () => {
      const res = await cartApi.getCart();
      if (res.ok) {
        set(apply(res.data));
      } else {
        set({ status: "error" });
      }
    })().finally(() => {
      inFlight = null;
    });
    return inFlight;
  },

  addItem: async ({ productId, variantId, quantity = 1 }) => {
    const prev = get();
    // Optimistic count bump so the header badge reacts immediately.
    set({ itemCount: prev.itemCount + quantity });

    const res = await cartApi.addCartItem({ productId, variantId, quantity });
    if (!res.ok) {
      set({ itemCount: prev.itemCount });
      return false;
    }
    set(apply(res.data));
    return true;
  },

  updateQuantity: async (itemId, quantity) => {
    const prev = get();
    const nextItems =
      quantity <= 0
        ? prev.items.filter((line) => line.id !== itemId)
        : prev.items.map((line) =>
            line.id === itemId
              ? {
                  ...line,
                  quantity,
                  lineTotal: Math.round(line.unitPrice * quantity * 100) / 100,
                }
              : line
          );
    set({ items: nextItems, ...summarize(nextItems) });

    const res = await cartApi.updateCartItem(itemId, quantity);
    if (!res.ok) {
      set({ items: prev.items, itemCount: prev.itemCount, subtotal: prev.subtotal });
      return;
    }
    set(apply(res.data));
  },

  removeItem: async (itemId) => {
    const prev = get();
    const nextItems = prev.items.filter((line) => line.id !== itemId);
    set({ items: nextItems, ...summarize(nextItems) });

    const res = await cartApi.removeCartItem(itemId);
    if (!res.ok) {
      set({ items: prev.items, itemCount: prev.itemCount, subtotal: prev.subtotal });
      return;
    }
    set(apply(res.data));
  },

  clearCart: async () => {
    const prev = get();
    set({ items: [], itemCount: 0, subtotal: 0 });

    const res = await cartApi.clearCart();
    if (!res.ok) {
      set({ items: prev.items, itemCount: prev.itemCount, subtotal: prev.subtotal });
      return;
    }
    set(apply(res.data));
  },

  reset: () => set({ items: [], itemCount: 0, subtotal: 0, status: "ready" }),
}));

/** Total number of units across all lines. */
export const selectCartItemCount = (state: CartState): number => state.itemCount;

/** Sum of price × quantity across all lines. */
export const selectCartSubtotal = (state: CartState): number => state.subtotal;
