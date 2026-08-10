import type { ReactNode } from 'react'
import { ArrowDown, ArrowUp, ArrowUpDown, Search } from 'lucide-react'
import type { PageInfo } from '@/types/api'
import { cn } from '@/utils/cn'
import { EmptyState, ErrorState } from './EmptyState'
import { PaginationBar } from './PaginationBar'
import { TableSkeleton } from './Skeleton'

export interface Column<T> {
  key: string
  header: ReactNode
  /** Render the cell; receives the row. Falls back to row[key]. */
  cell?: (row: T) => ReactNode
  /** Field name sent to the backend for sorting. */
  sortField?: string
  align?: 'left' | 'right' | 'center'
  className?: string
  /** Hide on small screens (still visible ≥md). */
  hideOnMobile?: boolean
}

export interface DataTableProps<T> {
  columns: Column<T>[]
  data: T[]
  loading?: boolean
  error?: boolean
  emptyTitle?: string
  emptyDescription?: string
  emptyAction?: ReactNode
  onRetry?: () => void
  /** Row key extractor. */
  rowKey: (row: T) => string
  /** Current sort directive ("field:asc" | "field:desc") or undefined. */
  sort?: string
  onSort?: (field: string) => void
  /** Cursor pagination. */
  pageInfo?: PageInfo
  page?: number
  hasPrevious?: boolean
  onNext?: () => void
  onPrevious?: () => void
  /** Optional search box above the table. */
  searchValue?: string
  onSearch?: (v: string) => void
  searchPlaceholder?: string
  toolbar?: ReactNode
  /** Render an expanded row below the main one (mobile card view). */
  expandedRender?: (row: T) => ReactNode
  /** Optional row click handler (e.g. open a detail view). */
  onRowClick?: (row: T) => void
  rows?: number
}

/**
 * DataTable — the shared workhorse for every list page.
 * Sorting, search, cursor pagination, skeleton/empty/error states.
 */
export function DataTable<T>({
  columns,
  data,
  loading,
  error,
  emptyTitle = 'No records found',
  emptyDescription,
  emptyAction,
  onRetry,
  rowKey,
  sort,
  onSort,
  pageInfo,
  page = 1,
  hasPrevious = false,
  onNext,
  onPrevious,
  searchValue,
  onSearch,
  searchPlaceholder = 'Search…',
  toolbar,
  expandedRender,
  onRowClick,
  rows = 6,
}: DataTableProps<T>) {
  const parseSort = (s?: string) => {
    if (!s) return { field: '', desc: true }
    const [f, dir] = s.split(':')
    return { field: f, desc: dir === 'desc' }
  }
  const active = parseSort(sort)

  const toggleSort = (field: string) => {
    if (!onSort) return
    const cur = active
    if (cur.field === field) {
      onSort(cur.desc ? `${field}:asc` : `${field}:desc`)
    } else {
      onSort(`${field}:desc`)
    }
  }

  const SortIcon = ({ field }: { field?: string }) => {
    if (!field) return null
    const isActive = active.field === field
    if (!isActive) return <ArrowUpDown className="h-3 w-3 text-slate-300" aria-hidden />
    return active.desc ? (
      <ArrowDown className="h-3 w-3 text-primary-600" aria-hidden />
    ) : (
      <ArrowUp className="h-3 w-3 text-primary-600" aria-hidden />
    )
  }

  return (
    <div className="overflow-hidden rounded-xl border border-slate-200/80 bg-white shadow-card">
      {(searchValue !== undefined || toolbar) && (
        <div className="flex flex-wrap items-center justify-between gap-3 border-b border-slate-100 px-4 py-3">
          {searchValue !== undefined && onSearch && (
            <div className="relative w-full max-w-xs">
              <Search className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-slate-400" aria-hidden />
              <input
                value={searchValue}
                onChange={(e) => onSearch(e.target.value)}
                placeholder={searchPlaceholder}
                className="h-9 w-full rounded-lg border border-slate-200 bg-white pl-9 pr-3 text-sm placeholder:text-slate-400 shadow-sm transition-colors focus:border-primary-500 focus:outline-none focus:ring-4 focus:ring-primary-500/10"
                aria-label="Search"
              />
            </div>
          )}
          {toolbar && <div className="flex items-center gap-2">{toolbar}</div>}
        </div>
      )}

      {loading ? (
        <TableSkeleton rows={rows} cells={columns.length} />
      ) : error ? (
        <ErrorState onRetry={onRetry} />
      ) : data.length === 0 ? (
        <EmptyState title={emptyTitle} description={emptyDescription} action={emptyAction} />
      ) : (
        <div className="overflow-x-auto">
          <table className="w-full text-left text-sm">
            <thead>
              <tr className="border-b border-slate-100 bg-slate-50/60">
                {columns.map((col) => (
                  <th
                    key={col.key}
                    scope="col"
                    aria-sort={
                      col.sortField && active.field === col.sortField
                        ? active.desc
                          ? 'descending'
                          : 'ascending'
                        : 'none'
                    }
                    className={cn(
                      'whitespace-nowrap px-4 py-2.5 text-xs font-medium text-slate-500',
                      col.align === 'right' && 'text-right',
                      col.align === 'center' && 'text-center',
                      col.hideOnMobile && 'hidden md:table-cell',
                    )}
                  >
                    {col.sortField ? (
                      <button
                        onClick={() => toggleSort(col.sortField!)}
                        className={cn(
                          'inline-flex items-center gap-1 rounded transition-colors hover:text-slate-800',
                          active.field === col.sortField && 'text-primary-700',
                        )}
                      >
                        {col.header}
                        <SortIcon field={col.sortField} />
                      </button>
                    ) : (
                      col.header
                    )}
                  </th>
                ))}
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-100">
              {data.map((row) => (
                <Row
                  key={rowKey(row)}
                  row={row}
                  columns={columns}
                  expandedRender={expandedRender}
                  onRowClick={onRowClick}
                />
              ))}
            </tbody>
          </table>
        </div>
      )}

      {pageInfo && onNext && (
        <PaginationBar
          pageInfo={pageInfo}
          page={page}
          hasPrevious={hasPrevious}
          onNext={onNext}
          onPrevious={onPrevious ?? (() => undefined)}
        />
      )}
    </div>
  )
}

function Row<T>({
  row,
  columns,
  expandedRender,
  onRowClick,
}: {
  row: T
  columns: Column<T>[]
  expandedRender?: (row: T) => ReactNode
  onRowClick?: (row: T) => void
}) {
  return (
    <>
      <tr
        className="group transition-colors duration-100 hover:bg-slate-50/70"
        onClick={onRowClick ? () => onRowClick(row) : undefined}
        style={onRowClick ? { cursor: 'pointer' } : undefined}
      >
        {columns.map((col) => (
          <td
            key={col.key}
            className={cn(
              'px-4 py-3 text-slate-700',
              col.align === 'right' && 'text-right',
              col.align === 'center' && 'text-center',
              col.hideOnMobile && 'hidden md:table-cell',
              col.className,
            )}
          >
            {col.cell ? col.cell(row) : String((row as Record<string, unknown>)[col.key] ?? '—')}
          </td>
        ))}
      </tr>
      {expandedRender && (
        <tr className="border-t-0 bg-slate-50/40 md:hidden">
          <td colSpan={columns.length} className="px-4 py-3">
            {expandedRender(row)}
          </td>
        </tr>
      )}
    </>
  )
}
