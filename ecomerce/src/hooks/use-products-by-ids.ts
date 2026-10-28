"use client";

import { use } from "react";

import { getProductsByIds } from "@/lib/products";
import type { Product } from "@/types/product";

const cache = new Map<string, Promise<Product[]>>();

function resolveProducts(ids: string[]): Promise<Product[]> {
  const key = ids.join(",");
  let promise = cache.get(key);
  if (!promise) {
    promise = getProductsByIds(ids);
    cache.set(key, promise);
  }
  return promise;
}

/**
 * Resolves product IDs into products through the repository (PRD §16/§44 —
 * cart/wishlist state persists only IDs; display data is reconstructed).
 *
 * Uses React 19's `use()` with a promise cache: calling components suspend
 * while data resolves, so wrap them in a Suspense boundary. The cache
 * dedupes concurrent and repeat lookups per id set.
 */
export function useProductsByIds(ids: string[]): Product[] {
  return use(resolveProducts(ids));
}
