import { useListController } from '@/hooks/useListController'
import { useListQuery, useDebouncedValue } from '@/hooks/useListQuery'
import { DataTable, type Column } from '@/components/ui/DataTable'
import { Badge } from '@/components/ui/Badge'
import { formatDateTime, humanize } from '@/utils/format'
import type { LoginLog } from '@/types/entities'

export function LoginLogsPage() {
  const ctrl = useListController()
  const debouncedSearch = useDebouncedValue(ctrl.search)

  const { data, isLoading, isError, refetch } = useListQuery<LoginLog>(
    ['login_logs', 'list'],
    '/api/v1/login_logs/list',
    { ...ctrl.query, search: debouncedSearch },
  )

  const columns: Column<LoginLog>[] = [
    {
      key: 'username',
      header: 'User',
      sortField: 'username',
      cell: (l) => <span className="font-medium text-slate-900">{l.username || '—'}</span>,
    },
    {
      key: 'success',
      header: 'Result',
      cell: (l) => (
        <Badge tone={l.success ? 'success' : 'danger'} dot>
          {l.success ? 'Success' : 'Failed'}
        </Badge>
      ),
    },
    {
      key: 'message',
      header: 'Message',
      hideOnMobile: true,
      cell: (l) => <span className="text-xs text-slate-500">{humanize(l.message)}</span>,
    },
    {
      key: 'ip',
      header: 'IP address',
      hideOnMobile: true,
      cell: (l) => <span className="font-mono text-xs text-slate-500">{l.ip || '—'}</span>,
    },
    {
      key: 'login_at',
      header: 'When',
      sortField: 'login_at',
      cell: (l) => <span className="text-xs text-slate-500 tabular">{formatDateTime(l.login_at)}</span>,
    },
  ]

  return (
    <div className="flex flex-col gap-5">
      <div>
        <h1 className="text-xl font-semibold text-slate-900">Login Logs</h1>
        <p className="mt-1 text-sm text-slate-500">
          Authentication attempts across the console. {isLoading ? '' : `${data?.pagination.total_count ?? 0} attempts`}
        </p>
      </div>

      <DataTable<LoginLog>
        columns={columns}
        data={data?.items ?? []}
        loading={isLoading}
        error={isError}
        onRetry={() => void refetch()}
        rowKey={(l) => l.id}
        sort={ctrl.sort}
        onSort={ctrl.onSort}
        searchValue={ctrl.search}
        onSearch={ctrl.setSearch}
        searchPlaceholder="Search by username…"
        pageInfo={data?.pagination}
        page={ctrl.page}
        hasPrevious={ctrl.hasPrevious}
        onNext={() => ctrl.onNext(data?.pagination.next_cursor)}
        onPrevious={ctrl.onPrevious}
        emptyTitle="No login attempts"
        emptyDescription="Sign-in activity will appear here."
      />
    </div>
  )
}
