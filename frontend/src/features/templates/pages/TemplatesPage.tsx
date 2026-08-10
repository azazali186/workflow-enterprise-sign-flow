import { useState } from 'react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { FileText, Pencil, Plus, Trash2 } from 'lucide-react'
import { useListController } from '@/hooks/useListController'
import { useListQuery, useDebouncedValue } from '@/hooks/useListQuery'
import { DataTable, type Column } from '@/components/ui/DataTable'
import { Button } from '@/components/ui/Button'
import { Badge } from '@/components/ui/Badge'
import { ConfirmDialog, Modal } from '@/components/ui/Modal'
import { Dropdown } from '@/components/ui/Dropdown'
import { Input } from '@/components/ui/Input'
import { templatesService } from '@/services/templates.service'
import { useToast } from '@/hooks/useToast'
import { getErrorMessage } from '@/utils/errors'
import { formatDateTime } from '@/utils/format'
import type { Template } from '@/types/entities'

export function TemplatesPage() {
  const ctrl = useListController()
  const debouncedSearch = useDebouncedValue(ctrl.search)
  const queryClient = useQueryClient()
  const toast = useToast()

  const [modalOpen, setModalOpen] = useState(false)
  const [editing, setEditing] = useState<Template | null>(null)
  const [deleting, setDeleting] = useState<Template | null>(null)

  const { data, isLoading, isError, refetch } = useListQuery<Template>(
    ['templates', 'list'],
    '/api/v1/templates/list',
    { ...ctrl.query, search: debouncedSearch },
  )

  const invalidate = () => void queryClient.invalidateQueries({ queryKey: ['templates', 'list'] })

  const deleteMut = useMutation({
    mutationFn: (id: string) => templatesService.remove(id),
    onSuccess: () => {
      toast.success('Template deleted')
      setDeleting(null)
      invalidate()
    },
    onError: (err) => {
      toast.error('Deletion failed', getErrorMessage(err, 'Deletion failed.'))
    },
  })

  const columns: Column<Template>[] = [
    {
      key: 'name',
      header: 'Template',
      sortField: 'name',
      cell: (t) => (
        <div className="flex items-center gap-3">
          <span className="flex h-8 w-8 shrink-0 items-center justify-center rounded-lg bg-primary-50 text-primary-600">
            <FileText className="h-4 w-4" aria-hidden />
          </span>
          <div className="min-w-0">
            <p className="truncate font-medium text-slate-900">{t.name || '—'}</p>
            <p className="truncate font-mono text-[11px] text-slate-400">{t.slug}</p>
          </div>
        </div>
      ),
    },
    {
      key: 'version',
      header: 'Version',
      cell: (t) => <span className="text-xs text-slate-500 tabular">v{t.version}</span>,
    },
    {
      key: 'is_active',
      header: 'Status',
      cell: (t) => (
        <Badge tone={t.is_active ? 'success' : 'neutral'} dot>
          {t.is_active ? 'Active' : 'Inactive'}
        </Badge>
      ),
    },
    {
      key: 'created_at',
      header: 'Created',
      sortField: 'created_at',
      hideOnMobile: true,
      cell: (t) => <span className="text-xs text-slate-500 tabular">{formatDateTime(t.created_at)}</span>,
    },
    {
      key: 'actions',
      header: '',
      align: 'right',
      cell: (t) => (
        <Dropdown
          trigger={
            <Button variant="ghost" size="sm" iconOnly aria-label="Row actions">
              <span className="text-sm">⋯</span>
            </Button>
          }
          items={[
            {
              key: 'edit',
              label: 'Edit template',
              icon: <Pencil className="h-4 w-4" />,
              onSelect: () => { setEditing(t); setModalOpen(true) },
            },
            {
              key: 'delete',
              label: 'Delete template',
              icon: <Trash2 className="h-4 w-4" />,
              danger: true,
              onSelect: () => setDeleting(t),
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
          <h1 className="text-xl font-semibold text-slate-900">Templates</h1>
          <p className="mt-1 text-sm text-slate-500">
            Reusable contract starting points. {isLoading ? '' : `${data?.pagination.total_count ?? 0} total`}
          </p>
        </div>
        <Button onClick={() => { setEditing(null); setModalOpen(true) }}>
          <Plus className="h-4 w-4" aria-hidden /> New template
        </Button>
      </div>

      <DataTable<Template>
        columns={columns}
        data={data?.items ?? []}
        loading={isLoading}
        error={isError}
        onRetry={() => void refetch()}
        rowKey={(t) => t.id}
        sort={ctrl.sort}
        onSort={ctrl.onSort}
        searchValue={ctrl.search}
        onSearch={ctrl.setSearch}
        searchPlaceholder="Search templates…"
        pageInfo={data?.pagination}
        page={ctrl.page}
        hasPrevious={ctrl.hasPrevious}
        onNext={() => ctrl.onNext(data?.pagination.next_cursor)}
        onPrevious={ctrl.onPrevious}
        emptyTitle="No templates yet"
        emptyDescription="Create a template to speed up contract creation."
      />

      <TemplateFormModal
        open={modalOpen}
        onClose={() => setModalOpen(false)}
        onSaved={invalidate}
        template={editing}
      />

      <ConfirmDialog
        open={Boolean(deleting)}
        onClose={() => setDeleting(null)}
        onConfirm={() => deleting && deleteMut.mutate(deleting.id)}
        title="Delete this template?"
        description={deleting ? `“${deleting.name}” will be removed.` : undefined}
        loading={deleteMut.isPending}
      />
    </div>
  )
}

function TemplateFormModal({
  open,
  onClose,
  onSaved,
  template,
}: {
  open: boolean
  onClose: () => void
  onSaved: () => void
  template?: Template | null
}) {
  const toast = useToast()
  const [name, setName] = useState('')
  const [slug, setSlug] = useState('')
  const [description, setDescription] = useState('')
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState('')

  const submit = async () => {
    if (!name.trim()) {
      setError('Name is required.')
      return
    }
    setSaving(true)
    try {
      if (template) {
        await templatesService.patch({ id: template.id, name, description })
        toast.success('Template updated')
      } else {
        await templatesService.create({
          name,
          slug: slug || name.toLowerCase().replace(/[^a-z0-9]+/g, '_'),
          description,
        })
        toast.success('Template created')
      }
      onSaved()
      onClose()
    } catch (err) {
      toast.error('Save failed', getErrorMessage(err))
    } finally {
      setSaving(false)
    }
  }

  return (
    <Modal
      open={open}
      onClose={onClose}
      title={template ? 'Edit template' : 'Create template'}
      size="lg"
      footer={
        <>
          <Button variant="ghost" onClick={onClose}>Cancel</Button>
          <Button loading={saving} onClick={submit}>{template ? 'Save changes' : 'Create template'}</Button>
        </>
      }
    >
      <div className="space-y-4">
        <Input label="Name" required value={name} onChange={(e) => setName(e.target.value)} error={error} />
        {!template && (
          <Input label="Slug" value={slug} onChange={(e) => setSlug(e.target.value)} hint="Leave blank to auto-generate." />
        )}
        <Input label="Description" value={description} onChange={(e) => setDescription(e.target.value)} />
      </div>
    </Modal>
  )
}
