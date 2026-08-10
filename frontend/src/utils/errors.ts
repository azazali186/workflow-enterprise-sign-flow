import type { NormalizedError } from '@/types/api'

/**
 * Extracts a human-safe message from any thrown value. The API client throws
 * NormalizedError; other throwables fall back to a generic message so UI
 * code never shows raw exception strings.
 */
export function getErrorMessage(err: unknown, fallback = 'The request failed. Please try again.'): string {
  if (err && typeof err === 'object' && 'message' in err) {
    const msg = (err as { message?: unknown }).message
    if (typeof msg === 'string' && msg.trim()) return msg
  }
  return fallback
}

/** True when the failure was a network-level problem (server unreachable). */
export function isNetworkError(err: unknown): boolean {
  return Boolean(
    err && typeof err === 'object' && (err as Partial<NormalizedError>).isNetwork === true,
  )
}
