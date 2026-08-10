import { useMutation, useQueryClient } from '@tanstack/react-query'
import { Check, X } from 'lucide-react'
import { useListController } from '@/hooks/useListController'
import { useListQuery, useDebouncedValue } from '@/hooks/useListQuery'
import { DataTable, type Column } from '@/components/ui/DataTable'
import { Badge } from '@/components/ui/Badge'
import { toneForStatus } from '@/utils/status'
import { BareSelect } from '@/components/ui/Input'
import { Button } from '@/components/ui/Button'
import { compliancesService } from '@/services/compliances.service'
import { useToast } from '@/hooks/useToast'
import { getErrorMessage } from '@/utils/errors'
import { formatDateTime, humanize, shortId } from '@/utils/format'
import type { Compliance } from '@/types/entities'

export function CompliancesPage() {
  const ctrl = useListController()
  const debouncedSearch = useDebouncedValue(ctrl.search)
  const queryClient = useQueryClient()
  const toast = useToast()

  const { data, isLoading, isError, refetch } = useListQuery<Compliance>(
    ['compliances', 'list'],
    '/api/v1/compliances/list',
    { ...ctrl.query, search: debouncedSearch },
  )

  const review = useMutation({
    mutationFn: (body: { id: string; status: 'approved' | 'rejected' }) =>
      compliancesService.review(body),
    onSuccess: (_d, vars) => {
      toast.success(vars.status === 'approved' ? 'Compliance approved' : 'Compliance rejected')
      void queryClient.invalidateQueries({ queryKey: ['compliances', 'list'] })
    },
    onError: (err) => {
      toast.error('Review failed', getErrorMessage(err, 'Review failed.'))
    },
  })

  const columns: Column<Compliance>[] = [
    {
      key: 'type',
      header: 'Check',
      cell: (c) => (
        <div className="min-w-0">
          <p className="truncate text-sm font-medium text-slate-900">{humanize(c.type)}</p>
          <p className="truncate font-mono text-[11px] text-slate-400">{shortId(c.contract_id)}</p>
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
      key: 'reviewed_at',
      header: 'Reviewed at',
      hideOnMobile: true,
      cell: (c) => <span className="text-xs text-slate-500 tabular">{formatDateTime(c.reviewed_at)}</span>,
    },
    {
      key: 'actions',
      header: '',
      align: 'right',
      cell: (c) =>
        c.status === 'pending' ? (
          <div className="flex justify-end gap-1.5">
            <Button
              size="sm"
              variant="outline"
              loading={review.isPending}
              onClick={() => review.mutate({ id: c.id, status: 'approved' })}
              className="text-success-700 hover:bg-success-50"
            >
              <Check className="h-3.5 w-3.5" aria-hidden /> Approve
            </Button>
            <Button
              size="sm"
              variant="outline"
              onClick={() => review.mutate({ id: c.id, status: 'rejected' })}
              className="text-danger-600 hover:bg-danger-50"
            >
              <X className="h-3.5 w-3.5" aria-hidden /> Reject
            </Button>
          </div>
        ) : (
          <span className="text-xs text-slate-400">{humanize(c.status)}</span>
        ),
    },
  ]

  const filterStatus = ctrl.filters.status as string | undefined

  return (
    <div className="flex flex-col gap-5">
      <div>
        <h1 className="text-xl font-semibold text-slate-900">Compliances</h1>
        <p className="mt-1 text-sm text-slate-500">
          Contract compliance checks awaiting review. {isLoading ? '' : `${data?.pagination.total_count ?? 0} total`}
        </p>
      </div>

      <DataTable<Compliance>
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
        searchPlaceholder="Search compliances…"
        toolbar={
          <BareSelect
            value={filterStatus ?? ''}
            onChange={(e) =>
              ctrl.applyFilters(e.target.value ? { status: e.target.value } : {})
            }
            options={[
              { value: '', label: 'All statuses' },
              { value: 'pending', label: 'Pending' },
              { value: 'approved', label: 'Approved' },
              { value: 'rejected', label: 'Rejected' },
            ]}
          />
        }
        pageInfo={data?.pagination}
        page={ctrl.page}
        hasPrevious={ctrl.hasPrevious}
        onNext={() => ctrl.onNext(data?.pagination.next_cursor)}
        onPrevious={ctrl.onPrevious}
        emptyTitle="No compliance checks"
        emptyDescription="Compliance checks appear here when contracts require review."
      />
    </div>
  )
}
