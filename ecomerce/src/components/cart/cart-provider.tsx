"use client";

import { useEffect, type ReactNode } from "react";

import { useAuthStore } from "@/store/auth-store";
import { useCartStore } from "@/store/cart-store";

/**
 * Loads the server-side cart on mount, and again whenever the auth state
 * resolves/changes. The second pass matters because signing in merges the guest
 * cart into the account cart server-side, so the client must re-read it.
 */
export function CartProvider({ children }: { children: ReactNode }) {
  const fetchCart = useCartStore((state) => state.fetchCart);
  const authStatus = useAuthStore((state) => state.status);

  useEffect(() => {
    void fetchCart();
  }, [fetchCart]);

  useEffect(() => {
    if (authStatus === "loading") return;
    void fetchCart();
  }, [authStatus, fetchCart]);

  return <>{children}</>;
}
