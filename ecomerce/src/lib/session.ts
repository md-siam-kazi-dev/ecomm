import { cookies } from "next/headers";
import { redirect } from "next/navigation";

import type { User, UserRole } from "@/types/user";

/**
 * Server-side session helpers. They call the Go API's GET /auth/me, forwarding
 * the incoming request cookies (the access token lives in an HttpOnly cookie).
 * The backend is the authority; this only adapts its user shape to the app's
 * `User` model.
 */

const BACKEND_API_URL = process.env.BACKEND_API_URL ?? "http://localhost:8080";

interface ApiUser {
  id: string;
  email: string;
  name: string;
  phone?: string;
  avatar_url?: string;
  role: string;
  email_verified: boolean;
  created_at?: string;
}

/** The current user, or `null` when signed out. */
export async function getCurrentUser(): Promise<User | null> {
  const cookieHeader = (await cookies()).toString();
  if (!cookieHeader) return null;

  try {
    const res = await fetch(`${BACKEND_API_URL}/api/v1/auth/me`, {
      headers: { cookie: cookieHeader },
      cache: "no-store",
    });
    if (!res.ok) return null;

    const payload = (await res.json()) as { data?: { user?: ApiUser } };
    const user = payload.data?.user;
    if (!user) return null;

    return {
      id: user.id,
      name: user.name,
      email: user.email,
      phone: user.phone,
      avatarUrl: user.avatar_url,
      addresses: [],
      memberSince: user.created_at ?? "",
      role: user.role as UserRole,
    };
  } catch {
    return null;
  }
}

/** Session user, or a redirect to sign-in — for routes that require auth. */
export async function requireCurrentUser(): Promise<User> {
  const user = await getCurrentUser();
  if (!user) redirect("/auth/sign-in");
  return user;
}
