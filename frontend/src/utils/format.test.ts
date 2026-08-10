import { describe, expect, it } from 'vitest'
import { formatBytes, formatDate, formatNumber, humanize, pageWindow, shortId, timeAgo } from './format'

describe('pageWindow', () => {
  it('describes an empty dataset', () => {
    expect(pageWindow(1, 20, 0)).toBe('Showing 0 results')
  })

  it('shows the correct window for the first page', () => {
    expect(pageWindow(1, 10, 120)).toBe('Showing 1–10 of 120')
  })

  it('shows a partial window on the last page', () => {
    expect(pageWindow(13, 10, 124)).toBe('Showing 121–124 of 124')
  })
})

describe('formatBytes', () => {
  it('formats bytes, KB and GB', () => {
    expect(formatBytes(0)).toBe('0 B')
    expect(formatBytes(512)).toBe('512 B')
    expect(formatBytes(1536)).toBe('1.5 KB')
    expect(formatBytes(5 * 1024 ** 3)).toBe('5 GB')
  })

  it('never renders trailing .0 for whole values', () => {
    expect(formatBytes(8 * 1024)).toBe('8 KB')
    expect(formatBytes(2 * 1024 ** 2)).toBe('2 MB')
    expect(formatBytes(3 * 1024 ** 4)).toBe('3 TB')
  })

  it('is null-safe', () => {
    expect(formatBytes(null)).toBe('—')
    expect(formatBytes(undefined)).toBe('—')
  })
})

describe('formatNumber', () => {
  it('formats with locale separators', () => {
    expect(formatNumber(1234567)).toBe('1,234,567')
  })
  it('is null-safe', () => {
    expect(formatNumber(null)).toBe('—')
  })
})

describe('formatDate', () => {
  it('is null-safe and rejects garbage input', () => {
    expect(formatDate(null)).toBe('—')
    expect(formatDate('not-a-date')).toBe('—')
  })
  it('formats valid dates', () => {
    expect(formatDate('2026-01-15T10:00:00Z')).toMatch(/2026/)
  })
})

describe('humanize', () => {
  it('turns snake_case into title case', () => {
    expect(humanize('awaiting_signature')).toBe('Awaiting Signature')
    expect(humanize('')).toBe('—')
    expect(humanize(null)).toBe('—')
  })
})

describe('shortId', () => {
  it('shortens long ids keeping the tail', () => {
    expect(shortId('0123456789abcdef0123456789abcdef')).toBe('01234567…cdef')
  })
  it('returns short ids untouched', () => {
    expect(shortId('abc123')).toBe('abc123')
  })
})

describe('timeAgo', () => {
  it('is null-safe', () => {
    expect(timeAgo(null)).toBe('—')
    expect(timeAgo('garbage')).toBe('—')
  })
  it('labels recent and older timestamps', () => {
    // Offset from now so the assertion never races a clock boundary.
    expect(timeAgo(new Date(Date.now() - 30_000).toISOString())).toBe('just now')
    const fiveMinAgo = new Date(Date.now() - 5 * 60 * 1000).toISOString()
    expect(timeAgo(fiveMinAgo)).toBe('5m ago')
  })
})
