/**
 * Browser API client for the Go backend.
 *
 * Requests go to same-origin `/api/v1/*`, which Next.js rewrites to the Go API
 * (see next.config.ts), so the HttpOnly auth cookies are same-site and are sent
 * with `credentials: "include"`. On a 401 it performs a single-flight refresh
 * and retries the request once.
 */

export interface ApiUser {
  id: string;
  email: string;
  name: string;
  role: string;
  email_verified: boolean;
  created_at?: string;
}

export interface ApiError {
  code: string;
  message: string;
  fields?: Record<string, string>;
}

/** Pagination metadata returned alongside list endpoints. */
export interface PaginationMeta {
  page: number;
  limit: number;
  total: number;
  total_pages: number;
  has_next: boolean;
  has_previous: boolean;
}

export type ApiResult<T> =
  | { ok: true; data: T; pagination?: PaginationMeta }
  | { ok: false; status: number; error: ApiError };

interface Envelope<T> {
  success: boolean;
  data?: T;
  message?: string;
  pagination?: PaginationMeta;
  error?: ApiError;
}

/** In-flight refresh, shared so concurrent 401s trigger only one refresh. */
let refreshPromise: Promise<boolean> | null = null;

function refreshSession(): Promise<boolean> {
  if (!refreshPromise) {
    refreshPromise = fetch("/api/v1/auth/refresh", {
      method: "POST",
      credentials: "include",
      cache: "no-store",
    })
      .then((res) => res.ok)
      .catch(() => false)
      .finally(() => {
        refreshPromise = null;
      });
  }
  return refreshPromise;
}

async function request<T>(
  path: string,
  init: RequestInit,
  allowRetry: boolean
): Promise<ApiResult<T>> {
  let res: Response;
  try {
    res = await fetch(path, { credentials: "include", cache: "no-store", ...init });
  } catch {
    return {
      ok: false,
      status: 0,
      error: {
        code: "NETWORK_ERROR",
        message: "Could not reach the server. Check your connection and try again.",
      },
    };
  }

  if (res.status === 401 && allowRetry) {
    if (await refreshSession()) {
      return request<T>(path, init, false);
    }
  }

  const payload = (await res.json().catch(() => null)) as Envelope<T> | null;

  if (!res.ok || !payload?.success) {
    return {
      ok: false,
      status: res.status,
      error: payload?.error ?? {
        code: "REQUEST_FAILED",
        message: `The server responded with status ${res.status}.`,
      },
    };
  }

  return { ok: true, data: payload.data as T, pagination: payload.pagination };
}

export function apiGet<T>(path: string): Promise<ApiResult<T>> {
  return request<T>(path, { method: "GET" }, true);
}

export function apiPost<T>(path: string, body?: unknown): Promise<ApiResult<T>> {
  return request<T>(
    path,
    {
      method: "POST",
      headers: body === undefined ? undefined : { "Content-Type": "application/json" },
      body: body === undefined ? undefined : JSON.stringify(body),
    },
    true
  );
}

export function apiPut<T>(path: string, body?: unknown): Promise<ApiResult<T>> {
  return request<T>(
    path,
    {
      method: "PUT",
      headers: body === undefined ? undefined : { "Content-Type": "application/json" },
      body: body === undefined ? undefined : JSON.stringify(body),
    },
    true
  );
}

export function apiDelete<T>(path: string): Promise<ApiResult<T>> {
  return request<T>(path, { method: "DELETE" }, true);
}

/** POSTs multipart form data (e.g. a file upload). */
export function apiUpload<T>(path: string, formData: FormData): Promise<ApiResult<T>> {
  return request<T>(path, { method: "POST", body: formData }, true);
}
