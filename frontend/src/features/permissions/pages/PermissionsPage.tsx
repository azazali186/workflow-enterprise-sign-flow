import { useListController } from '@/hooks/useListController'
import { useListQuery, useDebouncedValue } from '@/hooks/useListQuery'
import { DataTable, type Column } from '@/components/ui/DataTable'
import { Badge } from '@/components/ui/Badge'
import { formatDateTime } from '@/utils/format'
import type { Permission } from '@/types/entities'

const methodTone = (m: string) =>
  m === 'POST' ? 'primary' : m === 'PATCH' ? 'warning' : m === 'DELETE' ? 'danger' : 'neutral'

export function PermissionsPage() {
  const ctrl = useListController()
  const debouncedSearch = useDebouncedValue(ctrl.search)

  const { data, isLoading, isError, refetch } = useListQuery<Permission>(
    ['permissions', 'list'],
    '/api/v1/permissions/list',
    { ...ctrl.query, search: debouncedSearch },
  )

  const columns: Column<Permission>[] = [
    {
      key: 'name',
      header: 'Permission',
      cell: (p) => <span className="font-medium text-slate-900">{p.name}</span>,
    },
    {
      key: 'method',
      header: 'Method',
      cell: (p) => (
        <Badge tone={methodTone(p.method)} className="font-mono">
          {p.method}
        </Badge>
      ),
    },
    {
      key: 'path',
      header: 'Route',
      hideOnMobile: true,
      cell: (p) => <span className="font-mono text-xs text-slate-600">{p.path}</span>,
    },
    {
      key: 'service',
      header: 'Service',
      hideOnMobile: true,
      cell: (p) => <span className="text-xs text-slate-500">{p.service}</span>,
    },
    {
      key: 'created_at',
      header: 'Seeded',
      hideOnMobile: true,
      cell: (p) => <span className="text-xs text-slate-500 tabular">{formatDateTime(p.created_at)}</span>,
    },
  ]

  return (
    <div className="flex flex-col gap-5">
      <div>
        <h1 className="text-xl font-semibold text-slate-900">Permissions</h1>
        <p className="mt-1 text-sm text-slate-500">
          The API capability catalog, seeded from registered routes. Roles reference these by id.
        </p>
      </div>

      <DataTable<Permission>
        columns={columns}
        data={data?.items ?? []}
        loading={isLoading}
        error={isError}
        onRetry={() => void refetch()}
        rowKey={(p) => p.id}
        sort={ctrl.sort}
        onSort={ctrl.onSort}
        searchValue={ctrl.search}
        onSearch={ctrl.setSearch}
        searchPlaceholder="Search permissions…"
        pageInfo={data?.pagination}
        page={ctrl.page}
        hasPrevious={ctrl.hasPrevious}
        onNext={() => ctrl.onNext(data?.pagination.next_cursor)}
        onPrevious={ctrl.onPrevious}
        emptyTitle="No permissions yet"
        emptyDescription="Permissions are seeded automatically from the backend routes on startup."
      />
    </div>
  )
}
