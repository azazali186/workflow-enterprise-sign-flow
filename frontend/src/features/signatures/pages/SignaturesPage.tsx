import { useListController } from '@/hooks/useListController'
import { useListQuery, useDebouncedValue } from '@/hooks/useListQuery'
import { DataTable, type Column } from '@/components/ui/DataTable'
import { Badge } from '@/components/ui/Badge'
import { toneForStatus } from '@/utils/status'
import { formatDateTime, humanize, shortId } from '@/utils/format'
import type { Signature } from '@/types/entities'

export function SignaturesPage() {
  const ctrl = useListController()
  const debouncedSearch = useDebouncedValue(ctrl.search)

  const { data, isLoading, isError, refetch } = useListQuery<Signature>(
    ['signatures', 'list'],
    '/api/v1/signatures/list',
    { ...ctrl.query, search: debouncedSearch },
  )

  const columns: Column<Signature>[] = [
    {
      key: 'type',
      header: 'Type',
      cell: (s) => <Badge tone="neutral">{humanize(s.type)}</Badge>,
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
      key: 'signer_id',
      header: 'Signer',
      hideOnMobile: true,
      cell: (s) => <span className="font-mono text-xs text-slate-500">{shortId(s.signer_id)}</span>,
    },
    {
      key: 'signed_at',
      header: 'Signed at',
      sortField: 'signed_at',
      hideOnMobile: true,
      cell: (s) => <span className="text-xs text-slate-500 tabular">{formatDateTime(s.signed_at)}</span>,
    },
    {
      key: 'ip_address',
      header: 'IP',
      hideOnMobile: true,
      cell: (s) => <span className="font-mono text-xs text-slate-400">{s.ip_address || '—'}</span>,
    },
  ]

  return (
    <div className="flex flex-col gap-5">
      <div>
        <h1 className="text-xl font-semibold text-slate-900">Signatures</h1>
        <p className="mt-1 text-sm text-slate-500">
          Captured signature records. {isLoading ? '' : `${data?.pagination.total_count ?? 0} total`}
        </p>
      </div>

      <DataTable<Signature>
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
        searchPlaceholder="Search signatures…"
        pageInfo={data?.pagination}
        page={ctrl.page}
        hasPrevious={ctrl.hasPrevious}
        onNext={() => ctrl.onNext(data?.pagination.next_cursor)}
        onPrevious={ctrl.onPrevious}
        emptyTitle="No signatures captured yet"
        emptyDescription="Signatures appear here once signers complete their signing flow."
      />
    </div>
  )
}
