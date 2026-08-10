import { useCallback, useMemo, useState } from 'react'

export interface ListControllerOptions {
  defaultSort?: string
  defaultLimit?: number
  defaultFilters?: Record<string, unknown>
}

/**
 * Encapsulates list-page state so every entity page behaves identically:
 *   - search (debounced by the page)
 *   - sort directive ("field:asc" | "field:desc")
 *   - cursor stack for Previous/Next navigation
 *   - filters map
 */
export function useListController(opts: ListControllerOptions = {}) {
  const { defaultSort = 'created_at:desc', defaultLimit = 20, defaultFilters } = opts
  const [search, setSearch] = useState('')
  const [sort, setSort] = useState(defaultSort)
  /** Stack of cursors: last element is the current page's cursor. */
  const [cursorStack, setCursorStack] = useState<string[]>([])
  const [filters, setFilters] = useState<Record<string, unknown>>(defaultFilters ?? {})

  const page = cursorStack.length + 1
  const cursor = cursorStack.length > 0 ? cursorStack[cursorStack.length - 1] : undefined

  const reset = useCallback(() => setCursorStack([]), [])

  const onSearch = useCallback(
    (v: string) => {
      setSearch(v)
      reset()
    },
    [reset],
  )

  const onSort = useCallback(
    (s: string) => {
      setSort(s)
      reset()
    },
    [reset],
  )

  const onNext = useCallback(
    (nextCursor?: string) => {
      if (!nextCursor) return
      setCursorStack((stack) => [...stack, nextCursor])
    },
    [],
  )

  const onPrevious = useCallback(() => {
    setCursorStack((stack) => stack.slice(0, -1))
  }, [])

  const applyFilters = useCallback(
    (next: Record<string, unknown>) => {
      setFilters(next)
      reset()
    },
    [reset],
  )

  const hasPrevious = page > 1

  const query = useMemo(
    () => ({ search, sort, cursor, filters, limit: defaultLimit }),
    [search, sort, cursor, filters, defaultLimit],
  )

  return {
    query,
    search,
    sort,
    page,
    hasPrevious,
    setSearch: onSearch,
    onSort,
    onNext,
    onPrevious,
    applyFilters,
    filters,
  }
}
