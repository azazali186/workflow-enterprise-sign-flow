import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { History } from 'lucide-react'
import { useListController } from '@/hooks/useListController'
import { useListQuery, useDebouncedValue } from '@/hooks/useListQuery'
import { DataTable, type Column } from '@/components/ui/DataTable'
import { Badge } from '@/components/ui/Badge'
import { Modal } from '@/components/ui/Modal'
import { auditlogsService } from '@/services/auditlogs.service'
import { formatDateTime, humanize, shortId } from '@/utils/format'
import type { AuditLog } from '@/types/entities'

export function AuditLogsPage() {
  const ctrl = useListController()
  const debouncedSearch = useDebouncedValue(ctrl.search)
  const [detailId, setDetailId] = useState<string | null>(null)

  const { data, isLoading, isError, refetch } = useListQuery<AuditLog>(
    ['audit_logs', 'list'],
    '/api/v1/audit_logs/list',
    { ...ctrl.query, search: debouncedSearch },
  )

  const { data: detail } = useQuery({
    queryKey: ['audit_logs', 'detail', detailId],
    queryFn: () => auditlogsService.detail(detailId!),
    enabled: Boolean(detailId),
  })

  const columns: Column<AuditLog>[] = [
    {
      key: 'action',
      header: 'Action',
      cell: (l) => (
        <div className="flex items-center gap-2.5">
          <span className="flex h-7 w-7 shrink-0 items-center justify-center rounded-lg bg-slate-100 text-slate-500">
            <History className="h-3.5 w-3.5" aria-hidden />
          </span>
          <Badge tone="neutral">{l.action}</Badge>
        </div>
      ),
    },
    {
      key: 'actor',
      header: 'Actor',
      hideOnMobile: true,
      cell: (l) => (
        <div className="min-w-0">
          <p className="truncate text-sm text-slate-700">{l.actor_name || '—'}</p>
          <p className="truncate font-mono text-[11px] text-slate-400">{shortId(l.actor_user_id)}</p>
        </div>
      ),
    },
    {
      key: 'entity',
      header: 'Entity',
      hideOnMobile: true,
      cell: (l) => (
        <span className="text-xs text-slate-500">
          {humanize(l.entity_type)} · {shortId(l.entity_id)}
        </span>
      ),
    },
    {
      key: 'request_id',
      header: 'Request ID',
      hideOnMobile: true,
      cell: (l) => <span className="font-mono text-[11px] text-slate-400">{shortId(l.request_id, 6)}</span>,
    },
    {
      key: 'created_at',
      header: 'When',
      sortField: 'created_at',
      cell: (l) => <span className="text-xs text-slate-500 tabular">{formatDateTime(l.created_at)}</span>,
    },
  ]

  return (
    <div className="flex flex-col gap-5">
      <div>
        <h1 className="text-xl font-semibold text-slate-900">Audit Logs</h1>
        <p className="mt-1 text-sm text-slate-500">
          Every state change, who did it and when. {isLoading ? '' : `${data?.pagination.total_count ?? 0} events`}
        </p>
      </div>

      <DataTable<AuditLog>
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
        searchPlaceholder="Search actions, actors…"
        pageInfo={data?.pagination}
        page={ctrl.page}
        hasPrevious={ctrl.hasPrevious}
        onNext={() => ctrl.onNext(data?.pagination.next_cursor)}
        onPrevious={ctrl.onPrevious}
        emptyTitle="No audit events yet"
        emptyDescription="Changes across the workspace will appear here."
        expandedRender={(l) => (
          <div className="space-y-1 text-xs text-slate-500">
            <p>Actor: {l.actor_name || '—'} · {l.ip || '—'}</p>
            <p>Entity: {humanize(l.entity_type)}</p>
          </div>
        )}
        onRowClick={(l) => setDetailId(l.id)}
      />

      <Modal
        open={Boolean(detail)}
        onClose={() => setDetailId(null)}
        title={detail?.action ?? 'Audit event'}
        description={`${detail ? humanize(detail.entity_type) : ''} · ${detail ? shortId(detail.entity_id) : ''}`}
        size="lg"
      >
        {detail && <AuditDetail log={detail} />}
      </Modal>

    </div>
  )
}

function AuditDetail({ log }: { log: AuditLog }) {
  const rows: [string, string][] = [
    ['Request ID', log.request_id || '—'],
    ['Actor', log.actor_name || '—'],
    ['IP address', log.ip || '—'],
    ['User agent', log.user_agent || '—'],
    ['Timestamp', formatDateTime(log.created_at)],
  ]
  return (
    <div className="space-y-5">
      <dl className="divide-y divide-slate-100 rounded-lg border border-slate-200">
        {rows.map(([k, v]) => (
          <div key={k} className="grid grid-cols-[110px_1fr] gap-3 px-3.5 py-2.5">
            <dt className="text-xs font-medium text-slate-500">{k}</dt>
            <dd className="truncate text-xs text-slate-700">{v}</dd>
          </div>
        ))}
      </dl>
      <JsonBlock title="Changed fields" data={log.changed_fields} />
      <JsonBlock title="Before" data={log.before_data} />
      <JsonBlock title="After" data={log.after_data} />
    </div>
  )
}

function JsonBlock({ title, data }: { title: string; data?: unknown }) {
  if (data === undefined || data === null) return null
  let text = ''
  try {
    text = typeof data === 'string' ? JSON.stringify(JSON.parse(data), null, 2) : JSON.stringify(data, null, 2)
  } catch {
    text = String(data)
  }
  return (
    <div>
      <p className="mb-1.5 text-xs font-medium text-slate-600">{title}</p>
      <pre className="max-h-56 overflow-auto rounded-lg bg-slate-900 p-3.5 text-[11px] leading-relaxed text-slate-200">
        {text}
      </pre>
    </div>
  )
}
