import { cookies } from "next/headers";

import type {
  AdminApplicationDetail,
  ApplicationDetail,
  ApplicationItem,
  StoreItem,
} from "@/types/store-application";

/**
 * Server-side helpers for store applications. They forward the incoming request
 * cookies to the Go API (the access token lives in an HttpOnly cookie), matching
 * the pattern in `lib/session.ts`.
 */

const BACKEND_API_URL = process.env.BACKEND_API_URL ?? "http://localhost:8080";

async function fetchApi<T>(path: string): Promise<T | null> {
  const cookieHeader = (await cookies()).toString();
  if (!cookieHeader) return null;

  try {
    const res = await fetch(`${BACKEND_API_URL}${path}`, {
      headers: { cookie: cookieHeader },
      cache: "no-store",
    });
    if (!res.ok) return null;
    const payload = (await res.json()) as { data?: T };
    return payload.data ?? null;
  } catch {
    return null;
  }
}

export function getMyApplications(): Promise<ApplicationItem[] | null> {
  return fetchApi<ApplicationItem[]>("/api/v1/store-applications/me");
}

export function getApplicationDetail(id: string): Promise<ApplicationDetail | null> {
  return fetchApi<ApplicationDetail>(`/api/v1/store-applications/${id}`);
}

export function getMyStores(): Promise<StoreItem[] | null> {
  return fetchApi<StoreItem[]>("/api/v1/my-stores");
}

export function getAdminApplicationDetail(
  id: string
): Promise<AdminApplicationDetail | null> {
  return fetchApi<AdminApplicationDetail>(`/api/v1/admin/store-applications/${id}`);
}

/** The most relevant application to show: an active one, else the newest. */
export function pickPrimaryApplication(
  applications: ApplicationItem[]
): ApplicationItem | null {
  if (applications.length === 0) return null;
  const active = applications.find(
    (a) => a.status === "pending" || a.status === "approved"
  );
  return active ?? applications[0];
}
