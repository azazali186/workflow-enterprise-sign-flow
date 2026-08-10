import type { ReactNode } from 'react'
import { Navigate, useLocation } from 'react-router-dom'
import { useAppSelector } from '@/store'
import { usePermission } from '@/hooks/usePermission'
import { EmptyState } from '@/components/ui/EmptyState'
import { PageSkeleton } from '@/components/ui/PageSkeleton'
import { ShieldX } from 'lucide-react'

/** Redirects unauthenticated visitors to /login, preserving intent. */
export function RequireAuth({ children }: { children: ReactNode }) {
  const status = useAppSelector((s) => s.auth.status)
  const location = useLocation()

  // Session bootstrap in-flight: show a loader, never a blank flash.
  if (status === 'idle') {
    return (
      <div className="flex h-dvh items-center justify-center bg-surface">
        <PageSkeleton />
      </div>
    )
  }
  if (status !== 'authenticated') {
    return <Navigate to="/login" state={{ from: location.pathname }} replace />
  }
  return <>{children}</>
}

/** Redirects authenticated users away from the login screen. */
export function GuestOnly({ children }: { children: ReactNode }) {
  const status = useAppSelector((s) => s.auth.status)
  if (status === 'authenticated') return <Navigate to="/app/dashboard" replace />
  return <>{children}</>
}

/** Hides the page content when the user lacks the given permission. */
export function RequirePermission({
  permission,
  children,
}: {
  permission: string
  children: ReactNode
}) {
  const { can } = usePermission()
  if (!can(permission.split(' ')[0], permission.split(' ')[1])) {
    return (
      <div className="flex h-full items-center justify-center py-24">
        <EmptyState
          icon={ShieldX}
          title="You don’t have access to this page"
          description="Your role doesn’t include this permission. Ask an administrator if you believe this is a mistake."
        />
      </div>
    )
  }
  return <>{children}</>
}
