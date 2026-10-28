"use client";

import { create } from "zustand";

import { apiGet, apiPost, type ApiUser } from "@/lib/api-client";

export type AuthStatus = "loading" | "authenticated" | "unauthenticated";

interface LoginResult {
  ok: boolean;
  message?: string;
}

interface AuthState {
  status: AuthStatus;
  user: ApiUser | null;
  /** Resolve the initial auth state from GET /auth/me. */
  initialize: () => Promise<void>;
  login: (email: string, password: string) => Promise<LoginResult>;
  logout: () => Promise<void>;
  logoutAll: () => Promise<void>;
  setUser: (user: ApiUser | null) => void;
}

export const useAuthStore = create<AuthState>((set) => ({
  status: "loading",
  user: null,

  initialize: async () => {
    const res = await apiGet<{ user: ApiUser }>("/api/v1/auth/me");
    if (res.ok) {
      set({ status: "authenticated", user: res.data.user });
    } else {
      set({ status: "unauthenticated", user: null });
    }
  },

  login: async (email, password) => {
    const res = await apiPost<{ user: ApiUser }>("/api/v1/auth/login", { email, password });
    if (!res.ok) {
      return { ok: false, message: res.error.message };
    }
    set({ status: "authenticated", user: res.data.user });
    return { ok: true };
  },

  logout: async () => {
    await apiPost("/api/v1/auth/logout");
    set({ status: "unauthenticated", user: null });
  },

  logoutAll: async () => {
    await apiPost("/api/v1/auth/logout-all");
    set({ status: "unauthenticated", user: null });
  },

  setUser: (user) => set({ user, status: user ? "authenticated" : "unauthenticated" }),
}));
