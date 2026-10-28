"use client";

import { useEffect, type ReactNode } from "react";

import { useAuthStore } from "@/store/auth-store";

/**
 * Resolves the client auth state once on mount by calling GET /auth/me. The
 * backend stays authoritative for authorization; this only drives UI.
 */
export function AuthProvider({ children }: { children: ReactNode }) {
  const initialize = useAuthStore((state) => state.initialize);

  useEffect(() => {
    void initialize();
  }, [initialize]);

  return <>{children}</>;
}
