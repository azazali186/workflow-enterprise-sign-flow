import { Skeleton } from './Skeleton'

/** Page-level skeleton shown while a lazy route chunk loads. */
export function PageSkeleton() {
  return (
    <div className="flex flex-col gap-6 animate-fade-in" aria-busy="true" aria-label="Loading page">
      <div className="flex items-end justify-between">
        <div className="flex flex-col gap-2">
          <Skeleton className="h-6 w-48" />
          <Skeleton className="h-3 w-72" />
        </div>
        <Skeleton className="h-9 w-28" />
      </div>
      <div className="grid grid-cols-2 gap-4 md:grid-cols-4">
        {Array.from({ length: 4 }).map((_, i) => (
          <Skeleton key={i} className="h-24 rounded-xl" />
        ))}
      </div>
      <Skeleton className="h-96 rounded-xl" />
    </div>
  )
}
