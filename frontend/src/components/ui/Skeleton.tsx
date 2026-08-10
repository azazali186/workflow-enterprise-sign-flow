import { cn } from '@/utils/cn'

/** Shimmer skeleton block. */
export function Skeleton({ className }: { className?: string }) {
  return (
    <div
      className={cn(
        'animate-shimmer rounded-md bg-[linear-gradient(110deg,#f1f5f9_8%,#e2e8f0_18%,#f1f5f9_33%)] bg-[length:200%_100%]',
        className,
      )}
      aria-hidden
    />
  )
}

/** Row-shaped skeleton matching a table's cell height. */
export function SkeletonRow({ cells = 5 }: { cells?: number }) {
  return (
    <div className="flex items-center gap-4 px-5 py-3.5">
      {Array.from({ length: cells }).map((_, i) => (
        <Skeleton key={i} className={cn('h-4', i === 0 ? 'w-1/4' : 'flex-1')} />
      ))}
    </div>
  )
}

/** Full table skeleton: header bar + rows. */
export function TableSkeleton({ rows = 6, cells = 5 }: { rows?: number; cells?: number }) {
  return (
    <div className="divide-y divide-slate-100">
      <div className="flex items-center gap-4 px-5 py-3">
        {Array.from({ length: cells }).map((_, i) => (
          <Skeleton key={i} className={cn('h-3', i === 0 ? 'w-1/4' : 'flex-1')} />
        ))}
      </div>
      {Array.from({ length: rows }).map((_, i) => (
        <SkeletonRow key={i} cells={cells} />
      ))}
    </div>
  )
}

/** Stat-card skeleton. */
export function StatSkeleton() {
  return (
    <div className="flex flex-col gap-3 rounded-xl border border-slate-200/80 bg-white p-5 shadow-card">
      <Skeleton className="h-3 w-20" />
      <Skeleton className="h-8 w-24" />
      <Skeleton className="h-3 w-32" />
    </div>
  )
}
