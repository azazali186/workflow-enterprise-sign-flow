import { useMemo, useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Pencil, Plus, Trash2, UserRound } from 'lucide-react'
import { useListController } from '@/hooks/useListController'
import { useListQuery, useDebouncedValue } from '@/hooks/useListQuery'
import { DataTable, type Column } from '@/components/ui/DataTable'
import { Button } from '@/components/ui/Button'
import { Badge } from '@/components/ui/Badge'
import { toneForStatus } from '@/utils/status'
import { ConfirmDialog } from '@/components/ui/Modal'
import { Dropdown } from '@/components/ui/Dropdown'
import { BareSelect } from '@/components/ui/Input'
import { UserFormModal } from '../components/UserFormModal'
import { usersService } from '@/services/users.service'
import { rolesService } from '@/services/roles.service'
import { useToast } from '@/hooks/useToast'
import { getErrorMessage } from '@/utils/errors'
import { formatDateTime, timeAgo } from '@/utils/format'
import type { User, UserSummary } from '@/types/entities'

export function UsersPage() {
  const ctrl = useListController({ defaultSort: 'created_at:desc' })
  const debouncedSearch = useDebouncedValue(ctrl.search)
  const queryClient = useQueryClient()
  const toast = useToast()

  const [modalOpen, setModalOpen] = useState(false)
  const [editing, setEditing] = useState<User | null>(null)
  const [deleting, setDeleting] = useState<User | null>(null)

  const { data, isLoading, isError, refetch } = useListQuery<User>(
    ['users', 'list'],
    '/api/v1/users/list',
    { ...ctrl.query, search: debouncedSearch },
  )

  const { data: rolesData } = useQuery({
    queryKey: ['roles', 'options'],
    queryFn: () => rolesService.list({ limit: 100 }),
  })
  const roles = useMemo(() => rolesData?.items ?? [], [rolesData])

  const invalidate = () => {
    void queryClient.invalidateQueries({ queryKey: ['users', 'list'] })
    void queryClient.invalidateQueries({ queryKey: ['dashboard', 'summary'] })
  }

  const deleteMutation = useMutation({
    mutationFn: (id: string) => usersService.remove(id),
    onSuccess: () => {
      toast.success('User deleted', 'The account was removed.')
      invalidate()
      setDeleting(null)
    },
    onError: (err) => {
      toast.error('Deletion failed', getErrorMessage(err, 'Deletion failed.'))
    },
  })

  const summary = (data?.summary as UserSummary[] | undefined) ?? []
  const statusCount = (s: string) => summary.find((x) => x.status === s)?.count ?? 0

  const columns: Column<User>[] = [
    {
      key: 'name',
      header: 'User',
      sortField: 'name',
      cell: (u) => (
        <div className="flex items-center gap-3">
          <span className="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-primary-50 text-primary-700">
            <UserRound className="h-4 w-4" aria-hidden />
          </span>
          <div className="min-w-0">
            <p className="truncate font-medium text-slate-900">{u.name || '—'}</p>
            <p className="truncate text-xs text-slate-500">{u.email || '—'}</p>
          </div>
        </div>
      ),
    },
    {
      key: 'roles',
      header: 'Roles',
      cell: (u) => (
        <div className="flex flex-wrap gap-1">
          {(u.roles ?? []).length === 0 ? (
            <span className="text-xs text-slate-400">None</span>
          ) : (
            (u.roles ?? []).map((r) => (
              <Badge key={r.id} tone={r.slug === 'super_admin' ? 'primary' : 'neutral'}>
                {r.name}
              </Badge>
            ))
          )}
        </div>
      ),
    },
    {
      key: 'status',
      header: 'Status',
      sortField: 'status',
      cell: (u) => (
        <Badge tone={toneForStatus(u.status)} dot>
          {u.status}
        </Badge>
      ),
    },
    {
      key: 'last_login_at',
      header: 'Last login',
      sortField: 'last_login_at',
      hideOnMobile: true,
      cell: (u) => (
        <span className="text-xs text-slate-500 tabular">{timeAgo(u.last_login_at)}</span>
      ),
    },
    {
      key: 'created_at',
      header: 'Created',
      sortField: 'created_at',
      hideOnMobile: true,
      cell: (u) => <span className="text-xs text-slate-500 tabular">{formatDateTime(u.created_at)}</span>,
    },
    {
      key: 'actions',
      header: '',
      align: 'right',
      cell: (u) => (
        <Dropdown
          trigger={
            <Button variant="ghost" size="sm" iconOnly aria-label="Row actions">
              <span className="text-sm">⋯</span>
            </Button>
          }
          items={[
            {
              key: 'edit',
              label: 'Edit user',
              icon: <Pencil className="h-4 w-4" />,
              onSelect: () => {
                setEditing(u)
                setModalOpen(true)
              },
            },
            {
              key: 'delete',
              label: 'Delete user',
              icon: <Trash2 className="h-4 w-4" />,
              danger: true,
              disabled: u.status === 'active' && (u.roles ?? []).some((r) => r.slug === 'super_admin'),
              onSelect: () => setDeleting(u),
            },
          ]}
        />
      ),
    },
  ]

  const filterStatus = ctrl.filters.status as string | undefined

  return (
    <div className="flex flex-col gap-5">
      <div className="flex flex-wrap items-end justify-between gap-4">
        <div>
          <h1 className="text-xl font-semibold text-slate-900">Users</h1>
          <p className="mt-1 text-sm text-slate-500">
            {isLoading ? 'Loading…' : `${data?.pagination.total_count ?? 0} accounts in total`}
            {summary.length > 0 && (
              <span className="ml-2 text-xs text-slate-400">
                · {statusCount('active')} active · {statusCount('suspended')} suspended
              </span>
            )}
          </p>
        </div>
        <Button onClick={() => { setEditing(null); setModalOpen(true) }}>
          <Plus className="h-4 w-4" aria-hidden /> New user
        </Button>
      </div>

      <DataTable<User>
        columns={columns}
        data={data?.items ?? []}
        loading={isLoading}
        error={isError}
        onRetry={() => void refetch()}
        rowKey={(u) => u.id}
        sort={ctrl.sort}
        onSort={ctrl.onSort}
        searchValue={ctrl.search}
        onSearch={ctrl.setSearch}
        searchPlaceholder="Search by name or email…"
        toolbar={
          <BareSelect
            value={filterStatus ?? ''}
            onChange={(e) =>
              ctrl.applyFilters(e.target.value ? { status: e.target.value } : {})
            }
            options={[
              { value: '', label: 'All statuses' },
              { value: 'active', label: 'Active' },
              { value: 'suspended', label: 'Suspended' },
            ]}
          />
        }
        pageInfo={data?.pagination}
        page={ctrl.page}
        hasPrevious={ctrl.hasPrevious}
        onNext={() => ctrl.onNext(data?.pagination.next_cursor)}
        onPrevious={ctrl.onPrevious}
        emptyTitle="No users found"
        emptyDescription="Try adjusting your search, or invite the first member of your workspace."
        emptyAction={
          <Button size="sm" onClick={() => setModalOpen(true)}>
            <Plus className="h-4 w-4" aria-hidden /> New user
          </Button>
        }
        expandedRender={(u) => (
          <div className="space-y-1 text-xs text-slate-500">
            <p>Email: {u.email || '—'}</p>
            <p>Last login: {timeAgo(u.last_login_at)}</p>
          </div>
        )}
      />

      <UserFormModal
        open={modalOpen}
        onClose={() => setModalOpen(false)}
        onSaved={invalidate}
        user={editing}
        roles={roles}
      />

      <ConfirmDialog
        open={Boolean(deleting)}
        onClose={() => setDeleting(null)}
        onConfirm={() => deleting && deleteMutation.mutate(deleting.id)}
        title="Delete this user?"
        description={deleting ? `${deleting.name} will lose access immediately.` : undefined}
        loading={deleteMutation.isPending}
      />
    </div>
  )
}
