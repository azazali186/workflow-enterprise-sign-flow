/**
 * Shared API contract types. These mirror the backend exactly:
 *   - envelope: { code, message, data }
 *   - list:     POST /list with pagination.Query, returns Page<T>
 */

/** Unified envelope returned by every endpoint. */
export interface ApiEnvelope<T = unknown> {
  code: number
  message: string
  data?: T
}

/** Error envelope (HTTP != 2xx). */
export interface ApiErrorBody {
  code: number
  message: string
}

/** Request body accepted by every list/report endpoint (cursor pagination). */
export interface ListQuery {
  limit?: number
  cursor?: string
  filters?: Record<string, unknown>
  search?: string
  /** "field", "field:asc" | "field:desc", or "-field" */
  sort?: string
  date_from?: string | null
  date_to?: string | null
  date_field?: string
}

/** Pagination summary included in every Page response. */
export interface PageInfo {
  next_cursor: string
  has_more: boolean
  limit: number
  total_count: number
}

/** Unified list response: items + pagination + optional db summary. */
export interface Page<T> {
  items: T[]
  pagination: PageInfo
  summary?: unknown
}

/** Error normalized by the API client for callers and toasts. */
export interface NormalizedError {
  code: number
  message: string
  status: number
  isNetwork: boolean
}

/** Base fields present on every entity (UUID v7 id + timestamps). */
export interface BaseEntity {
  id: string
  created_at: string
  updated_at: string
}
