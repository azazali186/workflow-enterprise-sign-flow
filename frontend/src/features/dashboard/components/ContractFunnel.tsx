import type { DashboardSummary } from '@/types/entities'
import { cn } from '@/utils/cn'

const steps = [
  { key: 'draft', label: 'Draft' },
  { key: 'sent', label: 'Awaiting signature' },
  { key: 'signed', label: 'Signed' },
  { key: 'executed', label: 'Executed' },
] as const

/**
 * Contract pipeline funnel. Uses proportional widths + numeric labels so
 * the data is legible without relying on color alone.
 */
export function ContractFunnel({ data }: { data: DashboardSummary }) {
  const total = Math.max(data.contracts.total, 1)
  const get = (key: (typeof steps)[number]['key']) =>
    key === 'draft' ? data.contracts.draft
    : key === 'sent' ? data.contracts.sent
    : key === 'signed' ? data.contracts.signed
    : data.contracts.executed

  return (
    <div className="rounded-xl border border-slate-200/80 bg-white p-5 shadow-card">
      <h2 className="text-sm font-semibold text-slate-900">Contract pipeline</h2>
      <p className="mt-0.5 text-xs text-slate-500">
        {data.contracts.total} contracts · {data.contracts.draft} still in draft
      </p>
      <div className="mt-6 flex flex-col gap-4">
        {steps.map((s, i) => {
          const value = get(s.key)
          const pct = Math.round((value / total) * 100)
          return (
            <div key={s.key} className="flex items-center gap-3">
              <span className="w-32 shrink-0 text-xs font-medium text-slate-600">{s.label}</span>
              <div className="h-7 flex-1 overflow-hidden rounded-md bg-slate-100">
                <div
                  className={cn(
                    'flex h-full items-center rounded-md transition-all duration-500',
                    i === steps.length - 1 ? 'bg-success-500' : 'bg-primary-500',
                  )}
                  style={{ width: `${Math.max(pct, value > 0 ? 6 : 0)}%` }}
                >
                  <span className="ml-2 text-[11px] font-semibold text-white tabular">
                    {value}
                  </span>
                </div>
              </div>
              <span className="w-10 shrink-0 text-right text-xs text-slate-400 tabular">{pct}%</span>
            </div>
          )
        })}
      </div>
    </div>
  )
}
