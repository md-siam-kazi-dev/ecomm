import { cookies } from "next/headers";

import type { SiteSettings } from "@/types/site-settings";

/**
 * Server-side helpers for storefront settings. The public read needs no auth;
 * the admin read forwards the incoming request cookies to the Go API, matching
 * the pattern in `lib/session.ts` and `lib/store-applications.ts`.
 */

const BACKEND_API_URL = process.env.BACKEND_API_URL ?? "http://localhost:8080";

/** Public storefront settings (the announcement bar). `null` on any failure. */
export async function getSiteSettings(): Promise<SiteSettings | null> {
  try {
    const res = await fetch(`${BACKEND_API_URL}/api/v1/storefront/settings`, {
      cache: "no-store",
    });
    if (!res.ok) return null;
    const payload = (await res.json()) as { data?: SiteSettings };
    return payload.data ?? null;
  } catch {
    return null;
  }
}

/** Admin storefront settings (authenticated). `null` on any failure. */
export async function getAdminSiteSettings(): Promise<SiteSettings | null> {
  const cookieHeader = (await cookies()).toString();
  if (!cookieHeader) return null;

  try {
    const res = await fetch(`${BACKEND_API_URL}/api/v1/admin/settings`, {
      headers: { cookie: cookieHeader },
      cache: "no-store",
    });
    if (!res.ok) return null;
    const payload = (await res.json()) as { data?: SiteSettings };
    return payload.data ?? null;
  } catch {
    return null;
  }
}
