import { act, renderHook } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import { useListController } from './useListController'

describe('useListController', () => {
  it('starts on page 1 with no cursor', () => {
    const { result } = renderHook(() => useListController())
    expect(result.current.page).toBe(1)
    expect(result.current.hasPrevious).toBe(false)
    expect(result.current.query.cursor).toBeUndefined()
    expect(result.current.query.sort).toBe('created_at:desc')
  })

  it('advances the cursor stack on Next and rewinds on Previous', () => {
    const { result } = renderHook(() => useListController())

    act(() => result.current.onNext('cursor-1'))
    expect(result.current.page).toBe(2)
    expect(result.current.hasPrevious).toBe(true)
    expect(result.current.query.cursor).toBe('cursor-1')

    act(() => result.current.onNext('cursor-2'))
    expect(result.current.page).toBe(3)
    expect(result.current.query.cursor).toBe('cursor-2')

    act(() => result.current.onPrevious())
    expect(result.current.page).toBe(2)
    expect(result.current.query.cursor).toBe('cursor-1')

    act(() => result.current.onPrevious())
    expect(result.current.page).toBe(1)
    expect(result.current.hasPrevious).toBe(false)
    expect(result.current.query.cursor).toBeUndefined()
  })

  it('ignores Next without a cursor (no more pages)', () => {
    const { result } = renderHook(() => useListController())
    act(() => result.current.onNext(undefined))
    expect(result.current.page).toBe(1)
    expect(result.current.query.cursor).toBeUndefined()
  })

  it('does not rewind before the first page', () => {
    const { result } = renderHook(() => useListController())
    act(() => result.current.onPrevious())
    expect(result.current.page).toBe(1)
    expect(result.current.hasPrevious).toBe(false)
  })

  it('resets pagination when search, sort or filters change', () => {
    const { result } = renderHook(() => useListController())

    act(() => result.current.onNext('cursor-1'))
    act(() => result.current.setSearch('ada'))
    expect(result.current.page).toBe(1)
    expect(result.current.query.cursor).toBeUndefined()
    expect(result.current.query.search).toBe('ada')

    act(() => result.current.onNext('cursor-2'))
    act(() => result.current.onSort('name:asc'))
    expect(result.current.page).toBe(1)
    expect(result.current.query.sort).toBe('name:asc')

    act(() => result.current.onNext('cursor-3'))
    act(() => result.current.applyFilters({ status: 'active' }))
    expect(result.current.page).toBe(1)
    expect(result.current.query.filters).toEqual({ status: 'active' })
  })

  it('respects custom defaults', () => {
    const { result } = renderHook(() =>
      useListController({ defaultSort: 'name:asc', defaultLimit: 50, defaultFilters: { status: 'active' } }),
    )
    expect(result.current.query.sort).toBe('name:asc')
    expect(result.current.query.limit).toBe(50)
    expect(result.current.query.filters).toEqual({ status: 'active' })
  })
})
