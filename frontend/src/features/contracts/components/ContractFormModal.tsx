import { useState, type FormEvent } from 'react'
import { Plus, Trash2 } from 'lucide-react'
import { Modal } from '@/components/ui/Modal'
import { Button } from '@/components/ui/Button'
import { FieldShell, Input } from '@/components/ui/Input'
import { contractsService } from '@/services/contracts.service'
import { useToast } from '@/hooks/useToast'
import { getErrorMessage } from '@/utils/errors'
import type { SignerInput } from '@/types/entities'

interface SignerDraft extends SignerInput {
  key: number
}

export interface ContractFormModalProps {
  open: boolean
  onClose: () => void
  onSaved: () => void
}

export function ContractFormModal({ open, onClose, onSaved }: ContractFormModalProps) {
  const toast = useToast()
  const [title, setTitle] = useState('')
  const [referenceNo, setReferenceNo] = useState('')
  const [description, setDescription] = useState('')
  const [expiresAt, setExpiresAt] = useState('')
  const [signers, setSigners] = useState<SignerDraft[]>([])
  const [saving, setSaving] = useState(false)
  const [errors, setErrors] = useState<Record<string, string>>({})

  const addSigner = () =>
    setSigners((s) => [...s, { key: Date.now(), name: '', email: '', role: 'signer', order: s.length + 1 }])

  const removeSigner = (key: number) => setSigners((s) => s.filter((x) => x.key !== key))

  const updateSigner = (key: number, patch: Partial<SignerDraft>) =>
    setSigners((s) => s.map((x) => (x.key === key ? { ...x, ...patch } : x)))

  const validate = () => {
    const e: Record<string, string> = {}
    if (!title.trim()) e.title = 'Title is required.'
    const invalid = signers.find((s) => !s.name.trim() || !s.email.trim())
    if (invalid) e.signers = 'Every signer needs a name and email.'
    setErrors(e)
    return Object.keys(e).length === 0
  }

  const onSubmit = async (ev: FormEvent) => {
    ev.preventDefault()
    if (!validate()) return
    setSaving(true)
    try {
      await contractsService.create({
        title: title.trim(),
        reference_no: referenceNo.trim() || undefined,
        description: description.trim() || undefined,
        expires_at: expiresAt ? new Date(expiresAt).toISOString() : null,
        signers: signers.map(({ key: _k, ...rest }) => ({
          ...rest,
          phone: rest.phone || undefined,
        })),
      })
      toast.success('Contract created', `“${title}” is now a draft.`)
      onSaved()
      onClose()
    } catch (err) {
      toast.error('Creation failed', getErrorMessage(err))
    } finally {
      setSaving(false)
    }
  }

  return (
    <Modal
      open={open}
      onClose={onClose}
      title="Create contract"
      description="Draft a contract and attach the parties who must sign."
      size="xl"
      footer={
        <>
          <Button variant="ghost" onClick={onClose} disabled={saving}>Cancel</Button>
          <Button type="submit" form="contract-form" loading={saving}>Create contract</Button>
        </>
      }
    >
      <form id="contract-form" onSubmit={onSubmit} className="space-y-4">
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
          <Input
            id="c-title"
            label="Title"
            required
            value={title}
            onChange={(e) => setTitle(e.target.value)}
            error={errors.title}
            placeholder="Master Services Agreement"
          />
          <Input
            id="c-ref"
            label="Reference number"
            value={referenceNo}
            onChange={(e) => setReferenceNo(e.target.value)}
            placeholder="MSA-2026-001"
          />
        </div>
        <FieldShell label="Description" htmlFor="c-desc">
          <textarea
            id="c-desc"
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            rows={3}
            placeholder="What does this agreement cover?"
            className="w-full rounded-lg border border-slate-200 bg-white px-3 py-2 text-sm shadow-sm transition-colors placeholder:text-slate-400 focus:border-primary-500 focus:outline-none focus:ring-4 focus:ring-primary-500/10"
          />
        </FieldShell>
        <Input
          id="c-expires"
          type="date"
          label="Expires at"
          value={expiresAt}
          onChange={(e) => setExpiresAt(e.target.value)}
        />

        <div>
          <div className="mb-2 flex items-center justify-between">
            <p className="text-xs font-medium text-slate-600">Signers</p>
            <button
              type="button"
              onClick={addSigner}
              className="inline-flex items-center gap-1 text-xs font-medium text-primary-600 hover:text-primary-700"
            >
              <Plus className="h-3.5 w-3.5" aria-hidden /> Add signer
            </button>
          </div>
          {errors.signers && <p className="mb-2 text-xs text-danger-600">{errors.signers}</p>}
          <div className="space-y-2">
            {signers.map((s, i) => (
              <div key={s.key} className="flex items-start gap-2 rounded-lg border border-slate-200 p-2.5">
                <span className="mt-2 w-5 text-center text-xs font-semibold text-slate-400 tabular">{i + 1}</span>
                <div className="grid flex-1 grid-cols-1 gap-2 sm:grid-cols-2">
                  <input
                    value={s.name}
                    onChange={(e) => updateSigner(s.key, { name: e.target.value })}
                    placeholder="Full name"
                    aria-label={`Signer ${i + 1} name`}
                    className="h-8 rounded-md border border-slate-200 bg-white px-2.5 text-sm focus:border-primary-500 focus:outline-none"
                  />
                  <input
                    value={s.email}
                    onChange={(e) => updateSigner(s.key, { email: e.target.value })}
                    placeholder="Email"
                    type="email"
                    aria-label={`Signer ${i + 1} email`}
                    className="h-8 rounded-md border border-slate-200 bg-white px-2.5 text-sm focus:border-primary-500 focus:outline-none"
                  />
                </div>
                <button
                  type="button"
                  onClick={() => removeSigner(s.key)}
                  disabled={signers.length === 1}
                  className="mt-1 rounded-md p-1.5 text-slate-400 transition-colors hover:bg-danger-50 hover:text-danger-600 disabled:pointer-events-none disabled:opacity-40"
                  aria-label={`Remove signer ${i + 1}`}
                >
                  <Trash2 className="h-4 w-4" />
                </button>
              </div>
            ))}
          </div>
        </div>
      </form>
    </Modal>
  )
}
