import type { ReactNode } from 'react'
import { cn } from '@/utils/cn'
import type { BadgeTone } from '@/utils/status'

const tones: Record<BadgeTone, string> = {
  neutral: 'bg-slate-100 text-slate-600 ring-slate-200',
  primary: 'bg-primary-50 text-primary-700 ring-primary-200',
  success: 'bg-success-50 text-success-700 ring-success-100',
  warning: 'bg-warning-50 text-warning-700 ring-warning-100',
  danger: 'bg-danger-50 text-danger-700 ring-danger-100',
  info: 'bg-sky-50 text-sky-700 ring-sky-100',
}

export interface BadgeProps {
  tone?: BadgeTone
  children: ReactNode
  className?: string
  /** Dot indicator for at-a-glance status. */
  dot?: boolean
}

export function Badge({ tone = 'neutral', children, className, dot }: BadgeProps) {
  return (
    <span
      className={cn(
        'inline-flex items-center gap-1.5 rounded-full px-2 py-0.5 text-xs font-medium ring-1 ring-inset',
        tones[tone],
        className,
      )}
    >
      {dot && <span className={cn('h-1.5 w-1.5 rounded-full', dotTone[tone])} aria-hidden />}
      {children}
    </span>
  )
}

const dotTone: Record<BadgeTone, string> = {
  neutral: 'bg-slate-400',
  primary: 'bg-primary-500',
  success: 'bg-success-500',
  warning: 'bg-warning-500',
  danger: 'bg-danger-500',
  info: 'bg-sky-500',
}


