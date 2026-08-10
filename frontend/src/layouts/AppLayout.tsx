import { Suspense } from 'react'
import { Navigate, Outlet } from 'react-router-dom'
import { Sidebar } from './Sidebar'
import { Header } from './Header'
import { PageSkeleton } from '@/components/ui/PageSkeleton'
import { useAppSelector } from '@/store'

export function AppLayout() {
  const status = useAppSelector((s) => s.auth.status)
  const token = useAppSelector((s) => s.auth.token)

  // Session bootstrap is in-flight: keep the shell hidden to avoid flashes.
  if (status === 'idle') {
    return (
      <div className="flex h-dvh items-center justify-center bg-surface">
        <PageSkeleton />
      </div>
    )
  }

  if (status !== 'authenticated' || !token) {
    return <Navigate to="/login" replace />
  }

  return (
    <div className="flex h-dvh overflow-hidden bg-surface">
      <Sidebar />
      <div className="flex min-w-0 flex-1 flex-col">
        <Header />
        <main className="flex-1 overflow-y-auto">
          <div className="mx-auto w-full max-w-7xl px-4 py-6 sm:px-6 lg:px-8">
            <Suspense fallback={<PageSkeleton />}>
              <Outlet />
            </Suspense>
          </div>
        </main>
      </div>
    </div>
  )
}
