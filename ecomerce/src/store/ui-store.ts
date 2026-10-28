"use client";

import { create } from "zustand";

interface UiState {
  isMobileNavOpen: boolean;
  openMobileNav: () => void;
  closeMobileNav: () => void;
  toggleMobileNav: () => void;
  isCartOpen: boolean;
  openCart: () => void;
  closeCart: () => void;
}

/**
 * Genuinely global UI-only state (PRD §25): mobile navigation and the cart
 * drawer. The search overlay will join when search is built.
 */
export const useUiStore = create<UiState>()((set) => ({
  isMobileNavOpen: false,
  openMobileNav: () => set({ isMobileNavOpen: true }),
  closeMobileNav: () => set({ isMobileNavOpen: false }),
  toggleMobileNav: () =>
    set((state) => ({ isMobileNavOpen: !state.isMobileNavOpen })),
  isCartOpen: false,
  openCart: () => set({ isCartOpen: true }),
  closeCart: () => set({ isCartOpen: false }),
}));
