import { ChevronLeft, ChevronRight } from 'lucide-react'
import type { PageInfo } from '@/types/api'
import { cn } from '@/utils/cn'
import { pageWindow } from '@/utils/format'

export interface PaginationBarProps {
  pageInfo: PageInfo
  /** 1-based current page — derived from the cursor stack by the caller. */
  page: number
  hasPrevious: boolean
  onNext: () => void
  onPrevious: () => void
  className?: string
}

/**
 * Cursor pagination footer. The backend uses opaque cursors (no page
 * numbers), so navigation is Previous / Next with a friendly summary.
 */
export function PaginationBar({
  pageInfo,
  page,
  hasPrevious,
  onNext,
  onPrevious,
  className,
}: PaginationBarProps) {
  const total = pageInfo.total_count ?? 0
  const limit = pageInfo.limit ?? 20

  return (
    <div
      className={cn(
        'flex flex-wrap items-center justify-between gap-3 border-t border-slate-100 px-5 py-3',
        className,
      )}
    >
      <p className="text-xs text-slate-500 tabular">
        {pageWindow(page, limit, total)}
      </p>
      <div className="flex items-center gap-1.5">
        <button
          onClick={onPrevious}
          disabled={!hasPrevious}
          className="inline-flex h-8 items-center gap-1 rounded-lg border border-slate-200 bg-white px-2.5 text-xs font-medium text-slate-600 transition-colors hover:bg-slate-50 disabled:pointer-events-none disabled:opacity-40"
          aria-label="Previous page"
        >
          <ChevronLeft className="h-3.5 w-3.5" /> Previous
        </button>
        <span className="px-1 text-xs text-slate-400 tabular">Page {page}</span>
        <button
          onClick={onNext}
          disabled={!pageInfo.has_more}
          className="inline-flex h-8 items-center gap-1 rounded-lg border border-slate-200 bg-white px-2.5 text-xs font-medium text-slate-600 transition-colors hover:bg-slate-50 disabled:pointer-events-none disabled:opacity-40"
          aria-label="Next page"
        >
          Next <ChevronRight className="h-3.5 w-3.5" />
        </button>
      </div>
    </div>
  )
}
