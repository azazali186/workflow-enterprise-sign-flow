import { useEffect, useRef, useState } from 'react'
import { useQuery, type UseQueryOptions, type QueryKey } from '@tanstack/react-query'
import { fetchPage } from '@/services/api/client'
import type { ListQuery, NormalizedError, Page } from '@/types/api'

export interface ListState {
  search?: string
  sort?: string
  filters?: Record<string, unknown>
  cursor?: string
  limit?: number
  date_from?: string | null
  date_to?: string | null
}

/**
 * Shared list hook: builds the cursor-paginated query for any /list endpoint.
 * Keeps URL-friendly state (search, sort, filters, cursor) and the total.
 */
export function useListQuery<T>(
  key: QueryKey,
  path: string,
  state: ListState,
  options?: Omit<UseQueryOptions<Page<T>, NormalizedError, Page<T>, QueryKey>, 'queryKey' | 'queryFn'>,
) {
  const body: ListQuery = {
    limit: state.limit ?? 20,
    cursor: state.cursor,
    search: state.search || undefined,
    sort: state.sort || undefined,
    filters: state.filters,
    date_from: state.date_from,
    date_to: state.date_to,
  }

  return useQuery<Page<T>, NormalizedError>({
    queryKey: [...key, body],
    queryFn: () => fetchPage<T>(path, body),
    placeholderData: (prev) => prev,
    ...options,
  })
}

/** Tiny debounce hook for search inputs (default 250ms). */
export function useDebouncedValue<T>(value: T, delay = 250): T {
  const [debounced, setDebounced] = useState(value)
  const timer = useRef<ReturnType<typeof setTimeout> | null>(null)

  useEffect(() => {
    if (timer.current) clearTimeout(timer.current)
    timer.current = setTimeout(() => setDebounced(value), delay)
    return () => {
      if (timer.current) clearTimeout(timer.current)
    }
  }, [value, delay])

  return debounced
}
