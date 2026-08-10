import { Suspense } from 'react'
import { Outlet } from 'react-router-dom'
import { MarketingHeader } from '../components/MarketingHeader'
import { MarketingFooter } from '../components/MarketingFooter'
import { PageSkeleton } from '@/components/ui/PageSkeleton'

/** Public marketing shell — every marketing page shares header + footer. */
export function MarketingLayout() {
  return (
    <div className="flex min-h-dvh flex-col bg-white text-slate-900">
      <MarketingHeader />
      <main className="flex-1">
        <Suspense fallback={<PageSkeleton />}>
          <Outlet />
        </Suspense>
      </main>
      <MarketingFooter />
    </div>
  )
}
