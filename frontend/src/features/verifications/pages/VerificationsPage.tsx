import { useListController } from '@/hooks/useListController'
import { useListQuery, useDebouncedValue } from '@/hooks/useListQuery'
import { DataTable, type Column } from '@/components/ui/DataTable'
import { Badge } from '@/components/ui/Badge'
import { toneForStatus } from '@/utils/status'
import { formatDateTime, humanize, shortId } from '@/utils/format'
import type { Verification } from '@/types/entities'

export function VerificationsPage() {
  const ctrl = useListController()
  const debouncedSearch = useDebouncedValue(ctrl.search)

  const { data, isLoading, isError, refetch } = useListQuery<Verification>(
    ['verifications', 'list'],
    '/api/v1/verifications/list',
    { ...ctrl.query, search: debouncedSearch },
  )

  const columns: Column<Verification>[] = [
    {
      key: 'method',
      header: 'Method',
      cell: (v) => <Badge tone="neutral">{humanize(v.method)}</Badge>,
    },
    {
      key: 'status',
      header: 'Status',
      sortField: 'status',
      cell: (v) => (
        <Badge tone={toneForStatus(v.status)} dot>{humanize(v.status)}</Badge>
      ),
    },
    {
      key: 'attempts',
      header: 'Attempts',
      hideOnMobile: true,
      cell: (v) => <span className="text-xs text-slate-500 tabular">{v.attempts}</span>,
    },
    {
      key: 'contract_id',
      header: 'Contract',
      hideOnMobile: true,
      cell: (v) => <span className="font-mono text-xs text-slate-500">{shortId(v.contract_id)}</span>,
    },
    {
      key: 'verified_at',
      header: 'Verified at',
      hideOnMobile: true,
      cell: (v) => <span className="text-xs text-slate-500 tabular">{formatDateTime(v.verified_at)}</span>,
    },
  ]

  return (
    <div className="flex flex-col gap-5">
      <div>
        <h1 className="text-xl font-semibold text-slate-900">Verifications</h1>
        <p className="mt-1 text-sm text-slate-500">
          Identity checks attached to signatures. {isLoading ? '' : `${data?.pagination.total_count ?? 0} total`}
        </p>
      </div>

      <DataTable<Verification>
        columns={columns}
        data={data?.items ?? []}
        loading={isLoading}
        error={isError}
        onRetry={() => void refetch()}
        rowKey={(v) => v.id}
        sort={ctrl.sort}
        onSort={ctrl.onSort}
        searchValue={ctrl.search}
        onSearch={ctrl.setSearch}
        searchPlaceholder="Search verifications…"
        pageInfo={data?.pagination}
        page={ctrl.page}
        hasPrevious={ctrl.hasPrevious}
        onNext={() => ctrl.onNext(data?.pagination.next_cursor)}
        onPrevious={ctrl.onPrevious}
        emptyTitle="No verifications yet"
        emptyDescription="Verification records appear here once signatures are checked."
      />
    </div>
  )
}
