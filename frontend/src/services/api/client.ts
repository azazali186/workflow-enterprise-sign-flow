import type { ApiEnvelope, ApiErrorBody, ListQuery, NormalizedError, Page } from '@/types/api'

/**
 * API client with request/response interceptors.
 *  - request:  attaches the bearer token
 *  - response: unwraps the { code, message, data } envelope
 *  - error:    normalizes into a NormalizedError; 401 triggers session expiry
 */

const BASE = import.meta.env.VITE_API_BASE ?? ''

const TOKEN_KEY = 'sf.access_token'

export const tokenStore = {
  get: () => localStorage.getItem(TOKEN_KEY),
  set: (t: string) => localStorage.setItem(TOKEN_KEY, t),
  clear: () => localStorage.removeItem(TOKEN_KEY),
}

/** Called once when any request returns 401 — set by the app bootstrap. */
let onUnauthorized: (() => void) | null = null
/** Register (or clear, with null) the global 401 handler. */
export function setUnauthorizedHandler(fn: (() => void) | null): void {
  onUnauthorized = fn
}

function normalizeError(status: number, body: ApiErrorBody | null): NormalizedError {
  return {
    status,
    code: body?.code ?? status * 100,
    message: body?.message ?? 'Something went wrong. Please try again.',
    isNetwork: false,
  }
}

const networkError: NormalizedError = {
  status: 0,
  code: 0,
  message: 'Cannot reach the server. Check your connection and try again.',
  isNetwork: true,
}

async function request<T>(method: string, path: string, body?: unknown): Promise<T> {
  const headers: Record<string, string> = { Accept: 'application/json' }
  const token = tokenStore.get()
  if (token) headers.Authorization = `Bearer ${token}`
  if (body !== undefined) headers['Content-Type'] = 'application/json'

  let res: Response
  try {
    res = await fetch(`${BASE}${path}`, {
      method,
      headers,
      body: body !== undefined ? JSON.stringify(body) : undefined,
      signal: AbortSignal.timeout(30_000),
    })
  } catch (err) {
    // Network failure / timeout — surface a friendly normalized error.
    void err
    throw networkError
  }

  if (res.status === 401) {
    tokenStore.clear()
    onUnauthorized?.()
  }

  let parsed: ApiEnvelope<T> | null = null
  try {
    parsed = (await res.json()) as ApiEnvelope<T>
  } catch {
    /* non-JSON body */
  }

  if (!res.ok || (parsed && parsed.code !== 0)) {
    throw normalizeError(res.status, parsed ?? null)
  }
  return parsed?.data as T
}

/* ---------------- typed helpers for each verb ---------------- */

export function post<T>(path: string, body?: unknown): Promise<T> {
  return request<T>('POST', path, body)
}

export function patch<T>(path: string, body: unknown): Promise<T> {
  return request<T>('PATCH', path, body)
}

export function del<T>(path: string, body: unknown): Promise<T> {
  return request<T>('DELETE', path, body)
}

/** Compact a ListQuery, dropping empty values. */
export function listBody(q: ListQuery): Record<string, unknown> {
  const out: Record<string, unknown> = { limit: q.limit ?? 20 }
  if (q.cursor) out.cursor = q.cursor
  if (q.search) out.search = q.search
  if (q.sort) out.sort = q.sort
  if (q.filters && Object.keys(q.filters).length > 0) out.filters = q.filters
  if (q.date_from) out.date_from = q.date_from
  if (q.date_to) out.date_to = q.date_to
  if (q.date_field) out.date_field = q.date_field
  return out
}

/** Generic list fetch for any entity endpoint. */
export async function fetchPage<T>(path: string, q: ListQuery): Promise<Page<T>> {
  return post<Page<T>>(path, listBody(q))
}
