import { useListController } from '@/hooks/useListController'
import { useListQuery, useDebouncedValue } from '@/hooks/useListQuery'
import { DataTable, type Column } from '@/components/ui/DataTable'
import { Badge } from '@/components/ui/Badge'
import { toneForStatus } from '@/utils/status'
import { formatDateTime, humanize, shortId } from '@/utils/format'
import type { Signer } from '@/types/entities'

export function SignersPage() {
  const ctrl = useListController()
  const debouncedSearch = useDebouncedValue(ctrl.search)

  const { data, isLoading, isError, refetch } = useListQuery<Signer>(
    ['signers', 'list'],
    '/api/v1/signers/list',
    { ...ctrl.query, search: debouncedSearch },
  )

  const columns: Column<Signer>[] = [
    {
      key: 'name',
      header: 'Signer',
      sortField: 'name',
      cell: (s) => (
        <div className="min-w-0">
          <p className="truncate font-medium text-slate-900">{s.name || '—'}</p>
          <p className="truncate text-xs text-slate-500">{s.email || '—'}</p>
        </div>
      ),
    },
    {
      key: 'role',
      header: 'Role',
      cell: (s) => <span className="text-sm capitalize text-slate-600">{s.role || 'signer'}</span>,
    },
    {
      key: 'status',
      header: 'Status',
      sortField: 'status',
      cell: (s) => (
        <Badge tone={toneForStatus(s.status)} dot>{humanize(s.status)}</Badge>
      ),
    },
    {
      key: 'contract_id',
      header: 'Contract',
      hideOnMobile: true,
      cell: (s) => <span className="font-mono text-xs text-slate-500">{shortId(s.contract_id)}</span>,
    },
    {
      key: 'signed_at',
      header: 'Signed at',
      hideOnMobile: true,
      cell: (s) => <span className="text-xs text-slate-500 tabular">{formatDateTime(s.signed_at)}</span>,
    },
  ]

  return (
    <div className="flex flex-col gap-5">
      <div>
        <h1 className="text-xl font-semibold text-slate-900">Signers</h1>
        <p className="mt-1 text-sm text-slate-500">
          Every party attached to a contract. {isLoading ? '' : `${data?.pagination.total_count ?? 0} total`}
        </p>
      </div>

      <DataTable<Signer>
        columns={columns}
        data={data?.items ?? []}
        loading={isLoading}
        error={isError}
        onRetry={() => void refetch()}
        rowKey={(s) => s.id}
        sort={ctrl.sort}
        onSort={ctrl.onSort}
        searchValue={ctrl.search}
        onSearch={ctrl.setSearch}
        searchPlaceholder="Search signers…"
        pageInfo={data?.pagination}
        page={ctrl.page}
        hasPrevious={ctrl.hasPrevious}
        onNext={() => ctrl.onNext(data?.pagination.next_cursor)}
        onPrevious={ctrl.onPrevious}
        emptyTitle="No signers yet"
        emptyDescription="Signers are attached when you create a contract."
      />
    </div>
  )
}
