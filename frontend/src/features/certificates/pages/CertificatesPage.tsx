import { useListController } from '@/hooks/useListController'
import { useListQuery, useDebouncedValue } from '@/hooks/useListQuery'
import { DataTable, type Column } from '@/components/ui/DataTable'
import { Badge } from '@/components/ui/Badge'
import { toneForStatus } from '@/utils/status'
import { formatDate, formatDateTime, humanize, shortId } from '@/utils/format'
import type { Certificate } from '@/types/entities'

export function CertificatesPage() {
  const ctrl = useListController()
  const debouncedSearch = useDebouncedValue(ctrl.search)

  const { data, isLoading, isError, refetch } = useListQuery<Certificate>(
    ['certificates', 'list'],
    '/api/v1/certificates/list',
    { ...ctrl.query, search: debouncedSearch },
  )

  const columns: Column<Certificate>[] = [
    {
      key: 'subject',
      header: 'Subject',
      sortField: 'subject',
      cell: (c) => (
        <div className="min-w-0">
          <p className="truncate text-sm font-medium text-slate-900">{c.subject || '—'}</p>
          <p className="truncate font-mono text-[11px] text-slate-400">{c.serial_number}</p>
        </div>
      ),
    },
    {
      key: 'issuer',
      header: 'Issuer',
      hideOnMobile: true,
      cell: (c) => <span className="text-xs text-slate-500">{c.issuer || '—'}</span>,
    },
    {
      key: 'validity',
      header: 'Valid until',
      hideOnMobile: true,
      cell: (c) => <span className="text-xs text-slate-500 tabular">{formatDate(c.not_after)}</span>,
    },
    {
      key: 'contract_id',
      header: 'Contract',
      hideOnMobile: true,
      cell: (c) => <span className="font-mono text-xs text-slate-500">{shortId(c.contract_id)}</span>,
    },
    {
      key: 'status',
      header: 'Status',
      sortField: 'status',
      cell: (c) => (
        <Badge tone={toneForStatus(c.status)} dot>{humanize(c.status)}</Badge>
      ),
    },
    {
      key: 'created_at',
      header: 'Issued',
      hideOnMobile: true,
      cell: (c) => <span className="text-xs text-slate-500 tabular">{formatDateTime(c.created_at)}</span>,
    },
  ]

  return (
    <div className="flex flex-col gap-5">
      <div>
        <h1 className="text-xl font-semibold text-slate-900">Certificates</h1>
        <p className="mt-1 text-sm text-slate-500">
          Issued signing certificates. {isLoading ? '' : `${data?.pagination.total_count ?? 0} total`}
        </p>
      </div>

      <DataTable<Certificate>
        columns={columns}
        data={data?.items ?? []}
        loading={isLoading}
        error={isError}
        onRetry={() => void refetch()}
        rowKey={(c) => c.id}
        sort={ctrl.sort}
        onSort={ctrl.onSort}
        searchValue={ctrl.search}
        onSearch={ctrl.setSearch}
        searchPlaceholder="Search certificates…"
        pageInfo={data?.pagination}
        page={ctrl.page}
        hasPrevious={ctrl.hasPrevious}
        onNext={() => ctrl.onNext(data?.pagination.next_cursor)}
        onPrevious={ctrl.onPrevious}
        emptyTitle="No certificates issued"
        emptyDescription="Certificates appear here once they’re issued for contracts."
      />
    </div>
  )
}
