import type { ReactNode } from 'react'
import { Inbox, TriangleAlert, type LucideIcon } from 'lucide-react'
import { Button } from './Button'

export interface EmptyStateProps {
  icon?: LucideIcon
  title: string
  description?: string
  action?: ReactNode
}

/** Friendly empty state — invites action instead of looking broken. */
export function EmptyState({ icon: Icon = Inbox, title, description, action }: EmptyStateProps) {
  return (
    <div className="flex flex-col items-center justify-center gap-2 px-6 py-16 text-center animate-fade-in">
      <div className="flex h-12 w-12 items-center justify-center rounded-xl bg-slate-100 text-slate-400">
        <Icon className="h-6 w-6" aria-hidden />
      </div>
      <h3 className="mt-2 text-sm font-semibold text-slate-900">{title}</h3>
      {description && <p className="max-w-sm text-sm text-slate-500">{description}</p>}
      {action && <div className="mt-3">{action}</div>}
    </div>
  )
}

export interface ErrorStateProps {
  message?: string
  onRetry?: () => void
}

/** Error state with a recovery path — never a dead end. */
export function ErrorState({ message = 'We couldn’t load this data.', onRetry }: ErrorStateProps) {
  return (
    <div className="flex flex-col items-center justify-center gap-2 px-6 py-16 text-center animate-fade-in">
      <div className="flex h-12 w-12 items-center justify-center rounded-xl bg-danger-50 text-danger-500">
        <TriangleAlert className="h-6 w-6" aria-hidden />
      </div>
      <h3 className="mt-2 text-sm font-semibold text-slate-900">Something went wrong</h3>
      <p className="max-w-sm text-sm text-slate-500">{message}</p>
      {onRetry && (
        <Button variant="outline" size="sm" className="mt-3" onClick={onRetry}>
          Try again
        </Button>
      )}
    </div>
  )
}
