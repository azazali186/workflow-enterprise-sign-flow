import { useEffect, useState, type FormEvent } from 'react'
import { Modal } from '@/components/ui/Modal'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { rolesService } from '@/services/roles.service'
import { useToast } from '@/hooks/useToast'
import { getErrorMessage } from '@/utils/errors'
import type { Role } from '@/types/entities'

export interface RoleFormModalProps {
  open: boolean
  onClose: () => void
  onSaved: () => void
  role?: Role | null
}

export function RoleFormModal({ open, onClose, onSaved, role }: RoleFormModalProps) {
  const toast = useToast()
  const [name, setName] = useState('')
  const [slug, setSlug] = useState('')
  const [description, setDescription] = useState('')
  const [saving, setSaving] = useState(false)
  const [errors, setErrors] = useState<Record<string, string>>({})

  useEffect(() => {
    if (open) {
      setErrors({})
      setName(role?.name ?? '')
      setSlug(role?.slug ?? '')
      setDescription(role?.description ?? '')
    }
  }, [open, role])

  const slugify = (v: string) =>
    v.toLowerCase().trim().replace(/[^a-z0-9]+/g, '_').replace(/^_+|_+$/g, '')

  const validate = () => {
    const e: Record<string, string> = {}
    if (!name.trim()) e.name = 'Name is required.'
    if (!role && !slug.trim()) e.slug = 'Slug is required.'
    if (!role && !/^[a-z0-9_]+$/.test(slug)) e.slug = 'Use lowercase letters, numbers and underscores.'
    setErrors(e)
    return Object.keys(e).length === 0
  }

  const onSubmit = async (ev: FormEvent) => {
    ev.preventDefault()
    if (!validate()) return
    setSaving(true)
    try {
      if (role) {
        await rolesService.patch({ id: role.id, name, description })
        toast.success('Role updated', `“${name}” was saved.`)
      } else {
        await rolesService.create({ name, slug, description })
        toast.success('Role created', `“${name}” is ready to receive permissions.`)
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
      title={role ? 'Edit role' : 'Create role'}
      description={role ? 'Update role metadata.' : 'Define a new access level.'}
      footer={
        <>
          <Button variant="ghost" onClick={onClose} disabled={saving}>Cancel</Button>
          <Button type="submit" form="role-form" loading={saving}>
            {role ? 'Save changes' : 'Create role'}
          </Button>
        </>
      }
    >
      <form id="role-form" onSubmit={onSubmit} className="space-y-4">
        <Input
          id="role-name"
          label="Role name"
          required
          value={name}
          onChange={(e) => {
            setName(e.target.value)
            if (!role) setSlug(slugify(e.target.value))
          }}
          error={errors.name}
          placeholder="Contract Manager"
        />
        {!role && (
          <Input
            id="role-slug"
            label="Slug"
            required
            value={slug}
            onChange={(e) => setSlug(e.target.value)}
            error={errors.slug}
            hint="Unique identifier, e.g. contract_manager."
          />
        )}
        <Input
          id="role-desc"
          label="Description"
          value={description}
          onChange={(e) => setDescription(e.target.value)}
          placeholder="What can this role do?"
        />
      </form>
    </Modal>
  )
}
