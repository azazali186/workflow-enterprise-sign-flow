import { useQuery } from '@tanstack/react-query'
import {
  FileSignature,
  Fingerprint,
  FolderOpen,
  HardDrive,
  Layers,
  Users,
} from 'lucide-react'
import { dashboardService } from '@/services/dashboard.service'
import type { DashboardSummary } from '@/types/entities'
import { StatCard } from '../components/StatCard'
import { ContractFunnel } from '../components/ContractFunnel'
import { StatSkeleton } from '@/components/ui/Skeleton'
import { ErrorState } from '@/components/ui/EmptyState'
import { formatBytes, formatNumber } from '@/utils/format'

export function DashboardPage() {
  const { data, isLoading, isError, refetch } = useQuery({
    queryKey: ['dashboard', 'summary'],
    queryFn: () => dashboardService.summary(),
  })

  if (isLoading) {
    return (
      <div className="flex flex-col gap-6">
        <div>
          <h1 className="text-xl font-semibold text-slate-900">Dashboard</h1>
          <p className="mt-1 text-sm text-slate-500">Live summary across your signing workspace.</p>
        </div>
        <div className="grid grid-cols-2 gap-4 md:grid-cols-3 xl:grid-cols-6">
          {Array.from({ length: 6 }).map((_, i) => (
            <StatSkeleton key={i} />
          ))}
        </div>
        <div className="grid gap-6 lg:grid-cols-2">
          <div className="h-72 rounded-xl border border-slate-200/80 bg-white shadow-card" />
          <div className="h-72 rounded-xl border border-slate-200/80 bg-white shadow-card" />
        </div>
      </div>
    )
  }

  if (isError || !data) {
    return (
      <div className="flex h-[70vh] items-center justify-center">
        <ErrorState
          message="The dashboard summary couldn’t be loaded."
          onRetry={() => void refetch()}
        />
      </div>
    )
  }

  const c = data.contracts
  const s = data.signatures

  return (
    <div className="flex flex-col gap-6 animate-fade-in">
      <div>
        <h1 className="text-xl font-semibold text-slate-900">Dashboard</h1>
        <p className="mt-1 text-sm text-slate-500">
          Live summary across your signing workspace.
        </p>
      </div>

      <div className="grid grid-cols-2 gap-4 md:grid-cols-3 xl:grid-cols-6">
        <StatCard
          label="Contracts"
          value={formatNumber(c.total)}
          icon={FileSignature}
          detail={`${formatNumber(c.created_today)} created today`}
        />
        <StatCard
          label="Awaiting signature"
          value={formatNumber(c.sent)}
          icon={Layers}
          detail="sent or partially signed"
          tone="warning"
        />
        <StatCard
          label="Signed"
          value={formatNumber(c.signed)}
          icon={Fingerprint}
          detail="contracts completed"
          tone="success"
        />
        <StatCard
          label="Signers pending"
          value={formatNumber(data.signers.pending)}
          icon={FolderOpen}
          detail={`${formatNumber(data.signers.signed)} already signed`}
          tone="warning"
        />
        <StatCard
          label="Active users"
          value={formatNumber(data.active_users)}
          icon={Users}
          detail={`${formatNumber(data.templates)} templates`}
        />
        <StatCard
          label="Storage"
          value={formatBytes(data.storage_bytes)}
          icon={HardDrive}
          detail={`${formatNumber(s.total)} signatures captured`}
        />
      </div>

      <div className="grid gap-6 lg:grid-cols-2">
        <ContractFunnel data={data} />
        <SignatureSummary data={data} />
      </div>
    </div>
  )
}

function SignatureSummary({ data }: { data: DashboardSummary }) {
  const s = data.signatures
  const rows = [
    { label: 'Total signatures', value: s.total },
    { label: 'Captured', value: s.captured },
    { label: 'Verified', value: s.verified },
    { label: 'Captured today', value: s.captured_today },
    { label: 'Pending signers', value: data.signers.pending },
  ]
  return (
    <div className="rounded-xl border border-slate-200/80 bg-white p-5 shadow-card">
      <h2 className="text-sm font-semibold text-slate-900">Signature activity</h2>
      <p className="mt-0.5 text-xs text-slate-500">Where your signing pipeline stands right now.</p>
      <dl className="mt-5 space-y-3">
        {rows.map((r) => (
          <div key={r.label} className="flex items-center justify-between border-b border-slate-50 pb-3 last:border-0 last:pb-0">
            <dt className="text-sm text-slate-500">{r.label}</dt>
            <dd className="text-sm font-semibold text-slate-900 tabular">{formatNumber(r.value)}</dd>
          </div>
        ))}
      </dl>
    </div>
  )
}
