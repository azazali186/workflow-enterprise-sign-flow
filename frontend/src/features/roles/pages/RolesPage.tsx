import { useMemo, useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { KeyRound, Pencil, Plus, ShieldCheck, Trash2 } from 'lucide-react'
import { useListController } from '@/hooks/useListController'
import { useListQuery, useDebouncedValue } from '@/hooks/useListQuery'
import { DataTable, type Column } from '@/components/ui/DataTable'
import { Button } from '@/components/ui/Button'
import { Badge } from '@/components/ui/Badge'
import { ConfirmDialog, Modal } from '@/components/ui/Modal'
import { Dropdown } from '@/components/ui/Dropdown'
import { RoleFormModal } from '../components/RoleFormModal'
import { rolesService } from '@/services/roles.service'
import { permissionsService } from '@/services/permissions.service'
import { useToast } from '@/hooks/useToast'
import { getErrorMessage } from '@/utils/errors'
import { formatDateTime } from '@/utils/format'
import type { Permission, Role, RoleSummary } from '@/types/entities'

export function RolesPage() {
  const ctrl = useListController()
  const debouncedSearch = useDebouncedValue(ctrl.search)
  const queryClient = useQueryClient()
  const toast = useToast()

  const [modalOpen, setModalOpen] = useState(false)
  const [editing, setEditing] = useState<Role | null>(null)
  const [deleting, setDeleting] = useState<Role | null>(null)
  const [permRole, setPermRole] = useState<Role | null>(null)
  const [selectedPerms, setSelectedPerms] = useState<string[]>([])

  const { data, isLoading, isError, refetch } = useListQuery<Role>(
    ['roles', 'list'],
    '/api/v1/roles/list',
    { ...ctrl.query, search: debouncedSearch },
  )

  const { data: permsData } = useQuery({
    queryKey: ['permissions', 'options'],
    queryFn: () => permissionsService.list({ limit: 100 }),
  })
  const allPerms = useMemo(() => permsData?.items ?? [], [permsData])

  const invalidate = () => {
    void queryClient.invalidateQueries({ queryKey: ['roles', 'list'] })
  }

  const deleteMutation = useMutation({
    mutationFn: (id: string) => rolesService.remove(id),
    onSuccess: () => {
      toast.success('Role deleted', 'The role was removed.')
      invalidate()
      setDeleting(null)
    },
    onError: (err) => {
      toast.error('Deletion failed', getErrorMessage(err, 'Deletion failed.'))
    },
  })

  const savePerms = useMutation({
    mutationFn: (body: { id: string; permission_ids: string[] }) =>
      rolesService.assignPermissions(body),
    onSuccess: () => {
      toast.success('Permissions saved', 'Role access was updated.')
      invalidate()
      setPermRole(null)
    },
    onError: (err) => {
      toast.error('Save failed', getErrorMessage(err, 'Saving permissions failed.'))
    },
  })

  const summary = data?.summary as RoleSummary | undefined

  const columns: Column<Role>[] = [
    {
      key: 'name',
      header: 'Role',
      sortField: 'name',
      cell: (r) => (
        <div className="flex items-center gap-3">
          <span className="flex h-8 w-8 shrink-0 items-center justify-center rounded-lg bg-slate-100 text-slate-500">
            <ShieldCheck className="h-4 w-4" aria-hidden />
          </span>
          <div>
            <p className="font-medium text-slate-900">
              {r.name}
              {r.is_system && (
                <Badge tone="primary" className="ml-2">system</Badge>
              )}
            </p>
            <p className="text-xs text-slate-500">{r.slug}</p>
          </div>
        </div>
      ),
    },
    {
      key: 'description',
      header: 'Description',
      hideOnMobile: true,
      cell: (r) => <span className="text-sm text-slate-500">{r.description || '—'}</span>,
    },
    {
      key: 'permissions',
      header: 'Permissions',
      cell: (r) => (
        <span className="text-xs text-slate-500 tabular">
          {(r.permissions ?? []).length} assigned
        </span>
      ),
    },
    {
      key: 'created_at',
      header: 'Created',
      sortField: 'created_at',
      hideOnMobile: true,
      cell: (r) => <span className="text-xs text-slate-500 tabular">{formatDateTime(r.created_at)}</span>,
    },
    {
      key: 'actions',
      header: '',
      align: 'right',
      cell: (r) => (
        <Dropdown
          trigger={
            <Button variant="ghost" size="sm" iconOnly aria-label="Row actions">
              <span className="text-sm">⋯</span>
            </Button>
          }
          items={[
            {
              key: 'perms',
              label: 'Manage permissions',
              icon: <KeyRound className="h-4 w-4" />,
              disabled: r.is_system,
              onSelect: () => {
                setPermRole(r)
                setSelectedPerms((r.permissions ?? []).map((p) => p.id))
              },
            },
            {
              key: 'edit',
              label: 'Edit role',
              icon: <Pencil className="h-4 w-4" />,
              disabled: r.is_system,
              onSelect: () => { setEditing(r); setModalOpen(true) },
            },
            {
              key: 'delete',
              label: 'Delete role',
              icon: <Trash2 className="h-4 w-4" />,
              danger: true,
              disabled: r.is_system,
              onSelect: () => setDeleting(r),
            },
          ]}
        />
      ),
    },
  ]

  return (
    <div className="flex flex-col gap-5">
      <div className="flex flex-wrap items-end justify-between gap-4">
        <div>
          <h1 className="text-xl font-semibold text-slate-900">Roles</h1>
          <p className="mt-1 text-sm text-slate-500">
            {isLoading
              ? 'Loading…'
              : `${summary?.total_roles ?? data?.pagination.total_count ?? 0} roles · ${summary?.system_roles ?? 0} system-protected`}
          </p>
        </div>
        <Button onClick={() => { setEditing(null); setModalOpen(true) }}>
          <Plus className="h-4 w-4" aria-hidden /> New role
        </Button>
      </div>

      <DataTable<Role>
        columns={columns}
        data={data?.items ?? []}
        loading={isLoading}
        error={isError}
        onRetry={() => void refetch()}
        rowKey={(r) => r.id}
        sort={ctrl.sort}
        onSort={ctrl.onSort}
        searchValue={ctrl.search}
        onSearch={ctrl.setSearch}
        searchPlaceholder="Search roles…"
        pageInfo={data?.pagination}
        page={ctrl.page}
        hasPrevious={ctrl.hasPrevious}
        onNext={() => ctrl.onNext(data?.pagination.next_cursor)}
        onPrevious={ctrl.onPrevious}
        emptyTitle="No roles found"
        emptyDescription="Roles define what each member can do. Create one to start."
      />

      <RoleFormModal
        open={modalOpen}
        onClose={() => setModalOpen(false)}
        onSaved={invalidate}
        role={editing}
      />

      <Modal
        open={Boolean(permRole)}
        onClose={() => setPermRole(null)}
        title={`Permissions for ${permRole?.name ?? ''}`}
        description="Select the API capabilities this role can perform."
        size="xl"
        footer={
          <>
            <Button variant="ghost" onClick={() => setPermRole(null)}>Cancel</Button>
            <Button
              loading={savePerms.isPending}
              onClick={() =>
                permRole && savePerms.mutate({ id: permRole.id, permission_ids: selectedPerms })
              }
            >
              Save permissions
            </Button>
          </>
        }
      >
        <div className="grid max-h-[52vh] grid-cols-1 gap-1 overflow-y-auto pr-1 sm:grid-cols-2">
          {allPerms.map((p) => (
            <PermCheckbox
              key={p.id}
              permission={p}
              checked={selectedPerms.includes(p.id)}
              onToggle={() =>
                setSelectedPerms((prev) =>
                  prev.includes(p.id) ? prev.filter((x) => x !== p.id) : [...prev, p.id],
                )
              }
            />
          ))}
        </div>
      </Modal>

      <ConfirmDialog
        open={Boolean(deleting)}
        onClose={() => setDeleting(null)}
        onConfirm={() => deleting && deleteMutation.mutate(deleting.id)}
        title="Delete this role?"
        description={deleting ? `${deleting.name} members will lose its permissions.` : undefined}
        loading={deleteMutation.isPending}
      />
    </div>
  )
}

function PermCheckbox({
  permission,
  checked,
  onToggle,
}: {
  permission: Permission
  checked: boolean
  onToggle: () => void
}) {
  return (
    <label
      className={`flex cursor-pointer items-start gap-2.5 rounded-lg border p-2.5 transition-all duration-150 ${
        checked
          ? 'border-primary-300 bg-primary-50/60'
          : 'border-slate-200 hover:border-slate-300 hover:bg-slate-50'
      }`}
    >
      <input type="checkbox" checked={checked} onChange={onToggle} className="mt-0.5 accent-primary-600" />
      <span className="min-w-0">
        <span className="block truncate text-sm font-medium text-slate-800">{permission.name}</span>
        <span className="block truncate text-[11px] text-slate-400 tabular">
          {permission.method} {permission.path}
        </span>
      </span>
    </label>
  )
}
