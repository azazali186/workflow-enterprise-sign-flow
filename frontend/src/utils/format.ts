/** Formatting helpers — always null-safe so missing fields never crash a row. */

export function formatDateTime(value?: string | null): string {
  if (!value) return '—'
  const d = new Date(value)
  if (Number.isNaN(d.getTime())) return '—'
  return d.toLocaleString(undefined, {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}

export function formatDate(value?: string | null): string {
  if (!value) return '—'
  const d = new Date(value)
  if (Number.isNaN(d.getTime())) return '—'
  return d.toLocaleDateString(undefined, { year: 'numeric', month: 'short', day: 'numeric' })
}

export function formatBytes(value?: number | null): string {
  if (value === undefined || value === null) return '—'
  if (value === 0) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.min(Math.floor(Math.log(value) / Math.log(1024)), units.length - 1)
  const n = value / 1024 ** i
  // Whole numbers never get a trailing .0; small values keep one decimal.
  const rendered = Number.isInteger(n) ? String(Math.round(n)) : n.toFixed(1)
  return `${rendered} ${units[i]}`
}

export function formatNumber(value?: number | null): string {
  if (value === undefined || value === null) return '—'
  return new Intl.NumberFormat().format(value)
}

/** "Showing 11–20 of 245" style summary for a page. */
export function pageWindow(page: number, pageSize: number, total: number): string {
  if (total === 0) return 'Showing 0 results'
  const from = (page - 1) * pageSize + 1
  const to = Math.min(page * pageSize, total)
  return `Showing ${from}–${to} of ${new Intl.NumberFormat().format(total)}`
}

/** Relative time like "3m ago" / "2d ago". */
export function timeAgo(value?: string | null): string {
  if (!value) return '—'
  const d = new Date(value)
  if (Number.isNaN(d.getTime())) return '—'
  const diff = Date.now() - d.getTime()
  const s = Math.floor(diff / 1000)
  if (s < 60) return 'just now'
  const m = Math.floor(s / 60)
  if (m < 60) return `${m}m ago`
  const h = Math.floor(m / 60)
  if (h < 24) return `${h}h ago`
  const days = Math.floor(h / 24)
  if (days < 30) return `${days}d ago`
  return formatDate(value)
}

/** Title-case a snake_case label into friendly display text. */
export function humanize(value?: string | null): string {
  if (!value) return '—'
  return value
    .replace(/[_-]+/g, ' ')
    .replace(/\b\w/g, (c) => c.toUpperCase())
}

/** Shorten a long id for display, keeping the tail visible. */
export function shortId(id?: string | null, head = 8): string {
  if (!id) return '—'
  if (id.length <= head + 4) return id
  return `${id.slice(0, head)}…${id.slice(-4)}`
}
