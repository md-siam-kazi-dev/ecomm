"use client";

import { create } from "zustand";
import { persist } from "zustand/middleware";

interface WishlistState {
  /** Stored as product IDs; display data is reconstructed from the repository. */
  items: string[];
  addToWishlist: (productId: string) => void;
  removeFromWishlist: (productId: string) => void;
  /** Toggles a product and returns the new wishlisted state (handy for toasts). */
  toggleWishlist: (productId: string) => boolean;
  clearWishlist: () => void;
}

export const useWishlistStore = create<WishlistState>()(
  persist(
    (set, get) => ({
      items: [],
      addToWishlist: (productId) =>
        set((state) =>
          state.items.includes(productId)
            ? state
            : { items: [...state.items, productId] }
        ),
      removeFromWishlist: (productId) =>
        set((state) => ({
          items: state.items.filter((id) => id !== productId),
        })),
      toggleWishlist: (productId) => {
        const isWishlisted = get().items.includes(productId);

        if (isWishlisted) {
          set((state) => ({
            items: state.items.filter((id) => id !== productId),
          }));
        } else {
          set((state) => ({ items: [...state.items, productId] }));
        }

        return !isWishlisted;
      },
      clearWishlist: () => set({ items: [] }),
    }),
    {
      name: "shopkart-wishlist",
      version: 1,
      partialize: (state) => ({ items: state.items }),
    }
  )
);

/** Number of saved products. */
export const selectWishlistCount = (state: WishlistState): number =>
  state.items.length;

/** Curried selector — `useWishlistStore(selectIsWishlisted(productId))`. */
export const selectIsWishlisted =
  (productId: string) =>
  (state: WishlistState): boolean =>
    state.items.includes(productId);
