import { useState } from 'react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { Ban, Play, Plus, Send, Trash2 } from 'lucide-react'
import { useListController } from '@/hooks/useListController'
import { useListQuery, useDebouncedValue } from '@/hooks/useListQuery'
import { DataTable, type Column } from '@/components/ui/DataTable'
import { Button } from '@/components/ui/Button'
import { Badge } from '@/components/ui/Badge'
import { toneForStatus } from '@/utils/status'
import { ConfirmDialog } from '@/components/ui/Modal'
import { Dropdown } from '@/components/ui/Dropdown'
import { BareSelect } from '@/components/ui/Input'
import { ContractFormModal } from '../components/ContractFormModal'
import { contractsService } from '@/services/contracts.service'
import { useToast } from '@/hooks/useToast'
import { getErrorMessage } from '@/utils/errors'
import { formatDate, humanize } from '@/utils/format'
import type { Contract, StatusSummary } from '@/types/entities'

export function ContractsPage() {
  const ctrl = useListController()
  const debouncedSearch = useDebouncedValue(ctrl.search)
  const queryClient = useQueryClient()
  const toast = useToast()

  const [modalOpen, setModalOpen] = useState(false)
  const [deleting, setDeleting] = useState<Contract | null>(null)

  const { data, isLoading, isError, refetch } = useListQuery<Contract>(
    ['contracts', 'list'],
    '/api/v1/contracts/list',
    { ...ctrl.query, search: debouncedSearch },
  )

  const invalidate = () => {
    void queryClient.invalidateQueries({ queryKey: ['contracts', 'list'] })
    void queryClient.invalidateQueries({ queryKey: ['dashboard', 'summary'] })
  }

  // Explicit top-level hooks (never called from a helper — rules of hooks).
  const sendMut = useMutation({
    mutationFn: (id: string) => contractsService.sendSignatureRequest(id),
    onSuccess: () => {
      toast.success('Signature request sent to signers')
      invalidate()
    },
    onError: (err) => toast.error('Action failed', getErrorMessage(err, 'The action failed.')),
  })

  const executeMut = useMutation({
    mutationFn: (id: string) => contractsService.execute(id),
    onSuccess: () => {
      toast.success('Contract executed')
      invalidate()
    },
    onError: (err) => toast.error('Action failed', getErrorMessage(err, 'The action failed.')),
  })

  const cancelMut = useMutation({
    mutationFn: (id: string) => contractsService.cancel(id),
    onSuccess: () => {
      toast.success('Contract cancelled')
      invalidate()
    },
    onError: (err) => toast.error('Action failed', getErrorMessage(err, 'The action failed.')),
  })

  const deleteMut = useMutation({
    mutationFn: (id: string) => contractsService.remove(id),
    onSuccess: () => {
      toast.success('Contract deleted')
      setDeleting(null)
      invalidate()
    },
    onError: (err) => {
      toast.error('Deletion failed', getErrorMessage(err, 'Deletion failed.'))
    },
  })

  const summary = (data?.summary as StatusSummary[] | undefined) ?? []
  const countFor = (s: string) => summary.find((x) => x.status === s)?.count ?? 0

  const columns: Column<Contract>[] = [
    {
      key: 'title',
      header: 'Contract',
      sortField: 'title',
      cell: (c) => (
        <div className="min-w-0">
          <p className="truncate font-medium text-slate-900">{c.title || '—'}</p>
          <p className="truncate font-mono text-[11px] text-slate-400">{c.reference_no || '—'}</p>
        </div>
      ),
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
      key: 'signers',
      header: 'Signers',
      hideOnMobile: true,
      cell: (c) => {
        const signers = c.signers ?? []
        const done = signers.filter((s) => s.status === 'signed').length
        return (
          <span className="text-xs text-slate-500 tabular">
            {signers.length === 0 ? '—' : `${done}/${signers.length} signed`}
          </span>
        )
      },
    },
    {
      key: 'expires_at',
      header: 'Expires',
      sortField: 'expires_at',
      hideOnMobile: true,
      cell: (c) => <span className="text-xs text-slate-500 tabular">{formatDate(c.expires_at)}</span>,
    },
    {
      key: 'actions',
      header: '',
      align: 'right',
      cell: (c) => (
        <Dropdown
          trigger={
            <Button variant="ghost" size="sm" iconOnly aria-label="Row actions">
              <span className="text-sm">⋯</span>
            </Button>
          }
          items={[
            {
              key: 'send',
              label: 'Send for signature',
              icon: <Send className="h-4 w-4" />,
              disabled: c.status !== 'draft',
              onSelect: () => sendMut.mutate(c.id),
            },
            {
              key: 'execute',
              label: 'Execute',
              icon: <Play className="h-4 w-4" />,
              disabled: !['signed'].includes(c.status),
              onSelect: () => executeMut.mutate(c.id),
            },
            {
              key: 'cancel',
              label: 'Cancel contract',
              icon: <Ban className="h-4 w-4" />,
              disabled: !['draft', 'awaiting_signature', 'partially_signed'].includes(c.status),
              onSelect: () => cancelMut.mutate(c.id),
            },
            {
              key: 'delete',
              label: 'Delete contract',
              icon: <Trash2 className="h-4 w-4" />,
              danger: true,
              onSelect: () => setDeleting(c),
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
          <h1 className="text-xl font-semibold text-slate-900">Contracts</h1>
          <p className="mt-1 text-sm text-slate-500">
            {isLoading ? 'Loading…' : `${data?.pagination.total_count ?? 0} contracts`}
            {summary.length > 0 && (
              <span className="ml-2 text-xs text-slate-400">
                · {countFor('draft')} draft · {countFor('awaiting_signature') + countFor('partially_signed')} in flight · {countFor('signed')} signed
              </span>
            )}
          </p>
        </div>
        <Button onClick={() => setModalOpen(true)}>
          <Plus className="h-4 w-4" aria-hidden /> New contract
        </Button>
      </div>

      <DataTable<Contract>
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
        searchPlaceholder="Search contracts…"
        toolbar={
          <BareSelect
            value={filterStatus ?? ''}
            onChange={(e) =>
              ctrl.applyFilters(e.target.value ? { status: e.target.value } : {})
            }
            options={[
              { value: '', label: 'All statuses' },
              { value: 'draft', label: 'Draft' },
              { value: 'awaiting_signature', label: 'Awaiting signature' },
              { value: 'partially_signed', label: 'Partially signed' },
              { value: 'signed', label: 'Signed' },
              { value: 'executed', label: 'Executed' },
              { value: 'cancelled', label: 'Cancelled' },
              { value: 'expired', label: 'Expired' },
            ]}
          />
        }
        pageInfo={data?.pagination}
        page={ctrl.page}
        hasPrevious={ctrl.hasPrevious}
        onNext={() => ctrl.onNext(data?.pagination.next_cursor)}
        onPrevious={ctrl.onPrevious}
        emptyTitle="No contracts found"
        emptyDescription="Contracts you create appear here. Start with a draft and attach signers."
        emptyAction={
          <Button size="sm" onClick={() => setModalOpen(true)}>
            <Plus className="h-4 w-4" aria-hidden /> New contract
          </Button>
        }
        expandedRender={(c) => (
          <div className="space-y-1 text-xs text-slate-500">
            <p>Status: {humanize(c.status)}</p>
            <p>Signers: {(c.signers ?? []).length}</p>
          </div>
        )}
      />

      <ContractFormModal
        open={modalOpen}
        onClose={() => setModalOpen(false)}
        onSaved={invalidate}
      />

      <ConfirmDialog
        open={Boolean(deleting)}
        onClose={() => setDeleting(null)}
        onConfirm={() => deleting && deleteMut.mutate(deleting.id)}
        title="Delete this contract?"
        description={deleting ? `“${deleting.title}” will be removed.` : undefined}
        loading={deleteMut.isPending}
      />

    </div>
  )
}
