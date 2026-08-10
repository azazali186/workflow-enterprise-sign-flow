import { useListController } from '@/hooks/useListController'
import { useListQuery, useDebouncedValue } from '@/hooks/useListQuery'
import { DataTable, type Column } from '@/components/ui/DataTable'
import { Badge } from '@/components/ui/Badge'
import { toneForStatus } from '@/utils/status'
import { formatBytes, formatDateTime, humanize, shortId } from '@/utils/format'
import type { Storage } from '@/types/entities'

export function StoragesPage() {
  const ctrl = useListController()
  const debouncedSearch = useDebouncedValue(ctrl.search)

  const { data, isLoading, isError, refetch } = useListQuery<Storage>(
    ['storages', 'list'],
    '/api/v1/storages/list',
    { ...ctrl.query, search: debouncedSearch },
  )

  const columns: Column<Storage>[] = [
    {
      key: 'entity_type',
      header: 'Entity',
      cell: (s) => (
        <div className="min-w-0">
          <p className="truncate text-sm font-medium text-slate-900">{humanize(s.entity_type)}</p>
          <p className="truncate font-mono text-[11px] text-slate-400">{shortId(s.entity_id)}</p>
        </div>
      ),
    },
    {
      key: 'bucket',
      header: 'Bucket',
      hideOnMobile: true,
      cell: (s) => <span className="text-xs text-slate-500">{s.bucket || '—'}</span>,
    },
    {
      key: 'content_type',
      header: 'Type',
      hideOnMobile: true,
      cell: (s) => <span className="font-mono text-xs text-slate-500">{s.content_type || '—'}</span>,
    },
    {
      key: 'size_bytes',
      header: 'Size',
      sortField: 'size_bytes',
      align: 'right',
      cell: (s) => <span className="text-xs text-slate-600 tabular">{formatBytes(s.size_bytes)}</span>,
    },
    {
      key: 'status',
      header: 'Status',
      cell: (s) => (
        <Badge tone={toneForStatus(s.status)} dot>{humanize(s.status)}</Badge>
      ),
    },
    {
      key: 'uploaded_at',
      header: 'Uploaded',
      hideOnMobile: true,
      cell: (s) => <span className="text-xs text-slate-500 tabular">{formatDateTime(s.uploaded_at)}</span>,
    },
  ]

  return (
    <div className="flex flex-col gap-5">
      <div>
        <h1 className="text-xl font-semibold text-slate-900">Storages</h1>
        <p className="mt-1 text-sm text-slate-500">
          Objects stored for documents and evidence. {isLoading ? '' : `${data?.pagination.total_count ?? 0} total`}
        </p>
      </div>

      <DataTable<Storage>
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
        searchPlaceholder="Search storages…"
        pageInfo={data?.pagination}
        page={ctrl.page}
        hasPrevious={ctrl.hasPrevious}
        onNext={() => ctrl.onNext(data?.pagination.next_cursor)}
        onPrevious={ctrl.onPrevious}
        emptyTitle="No stored objects"
        emptyDescription="Uploaded documents and evidence appear here."
      />
    </div>
  )
}
