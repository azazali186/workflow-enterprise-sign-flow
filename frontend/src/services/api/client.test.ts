import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { fetchPage, listBody, post, setUnauthorizedHandler, tokenStore } from './client'

function mockFetchOnce(status: number, body: unknown) {
  return vi.spyOn(globalThis, 'fetch').mockResolvedValueOnce(
    new Response(JSON.stringify(body), {
      status,
      headers: { 'Content-Type': 'application/json' },
    }),
  )
}

describe('API client', () => {
  beforeEach(() => {
    tokenStore.clear()
    setUnauthorizedHandler(null)
  })
  afterEach(() => {
    vi.restoreAllMocks()
  })

  it('unwraps the envelope and returns data', async () => {
    mockFetchOnce(200, { code: 0, message: 'ok', data: { id: 'x' } })
    const data = await post<{ id: string }>('/api/v1/test', {})
    expect(data).toEqual({ id: 'x' })
  })

  it('attaches the bearer token from storage', async () => {
    tokenStore.set('tok-123')
    const spy = mockFetchOnce(200, { code: 0, message: 'ok', data: null })
    await post('/api/v1/test', {})
    const headers = spy.mock.calls[0][1]?.headers as Record<string, string>
    expect(headers.Authorization).toBe('Bearer tok-123')
  })

  it('normalizes error envelopes with backend messages', async () => {
    mockFetchOnce(401, { code: 40100, message: 'invalid credentials' })
    await expect(post('/api/v1/auth/login', {})).rejects.toMatchObject({
      status: 401,
      code: 40100,
      message: 'invalid credentials',
    })
  })

  it('clears the token and fires the 401 handler', async () => {
    tokenStore.set('tok')
    const onUnauth = vi.fn()
    setUnauthorizedHandler(onUnauth)
    mockFetchOnce(401, { code: 40100, message: 'unauthorized' })
    await expect(post('/api/v1/protected', {})).rejects.toMatchObject({ status: 401 })
    expect(tokenStore.get()).toBeNull()
    expect(onUnauth).toHaveBeenCalledOnce()
  })

  it('surfaces a network error when fetch itself fails', async () => {
    vi.spyOn(globalThis, 'fetch').mockRejectedValueOnce(new TypeError('Failed to fetch'))
    await expect(post('/api/v1/test', {})).rejects.toMatchObject({ isNetwork: true, status: 0 })
  })
})

describe('listBody', () => {
  it('compacts empty values and keeps the limit default', () => {
    expect(listBody({})).toEqual({ limit: 20 })
    expect(listBody({ search: '', sort: '', cursor: '', limit: 10, filters: {} })).toEqual({
      limit: 10,
    })
    expect(listBody({ search: 'ada', sort: 'name:asc', cursor: 'abc', filters: { status: 'active' }, date_from: '2026-01-01' })).toEqual({
      limit: 20,
      search: 'ada',
      sort: 'name:asc',
      cursor: 'abc',
      filters: { status: 'active' },
      date_from: '2026-01-01',
    })
  })
})

describe('fetchPage', () => {
  it('posts the compact query to the exact path and returns the page', async () => {
    const spy = mockFetchOnce(200, {
      code: 0,
      message: 'ok',
      data: { items: [], pagination: { next_cursor: '', has_more: false, limit: 20, total_count: 3 }, summary: null },
    })
    const page = await fetchPage<unknown>('/api/v1/users/list', { limit: 20 })
    // Guard against regressions that change method or URL (backend is POST-only).
    const [url, init] = spy.mock.calls[0]
    expect(url).toBe('/api/v1/users/list')
    expect(init?.method).toBe('POST')
    expect(page.pagination.total_count).toBe(3)
    expect(page.items).toEqual([])
  })
})
