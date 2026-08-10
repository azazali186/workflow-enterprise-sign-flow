import type { LucideIcon } from 'lucide-react'
import { cn } from '@/utils/cn'

const toneStyles = {
  default: 'bg-slate-100 text-slate-500',
  warning: 'bg-warning-50 text-warning-600',
  success: 'bg-success-50 text-success-600',
} as const

export interface StatCardProps {
  label: string
  value: string
  icon: LucideIcon
  detail?: string
  tone?: keyof typeof toneStyles
}

export function StatCard({ label, value, icon: Icon, detail, tone = 'default' }: StatCardProps) {
  return (
    <div className="group rounded-xl border border-slate-200/80 bg-white p-4 shadow-card transition-all duration-200 hover:-translate-y-0.5 hover:shadow-pop">
      <div className="flex items-center justify-between">
        <p className="text-xs font-medium text-slate-500">{label}</p>
        <span className={cn('flex h-7 w-7 items-center justify-center rounded-lg', toneStyles[tone])}>
          <Icon className="h-3.5 w-3.5" aria-hidden />
        </span>
      </div>
      <p className="mt-2 text-2xl font-semibold tracking-tight text-slate-900 tabular">{value}</p>
      {detail && <p className="mt-1 text-[11px] text-slate-400">{detail}</p>}
    </div>
  )
}
