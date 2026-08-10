import { useEffect, useState, type FormEvent } from 'react'
import { Modal } from '@/components/ui/Modal'
import { Button } from '@/components/ui/Button'
import { FieldShell, Input, Select } from '@/components/ui/Input'
import { usersService } from '@/services/users.service'
import { useToast } from '@/hooks/useToast'
import { getErrorMessage } from '@/utils/errors'
import type { Role, User } from '@/types/entities'

export interface UserFormModalProps {
  open: boolean
  onClose: () => void
  onSaved: () => void
  /** When provided the modal edits; otherwise it creates. */
  user?: User | null
  roles: Role[]
}

const empty = { name: '', email: '', password: '', phone: '', status: 'active', role_ids: [] as string[] }

export function UserFormModal({ open, onClose, onSaved, user, roles }: UserFormModalProps) {
  const toast = useToast()
  const [form, setForm] = useState(empty)
  const [saving, setSaving] = useState(false)
  const [errors, setErrors] = useState<Record<string, string>>({})

  useEffect(() => {
    if (open) {
      setErrors({})
      setForm(
        user
          ? {
              name: user.name,
              email: user.email,
              password: '',
              phone: user.phone ?? '',
              status: user.status,
              role_ids: (user.roles ?? []).map((r) => r.id),
            }
          : empty,
      )
    }
  }, [open, user])

  const validate = (): boolean => {
    const e: Record<string, string> = {}
    if (!form.name.trim()) e.name = 'Name is required.'
    if (!form.email.trim()) e.email = 'Email is required.'
    else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(form.email)) e.email = 'Enter a valid email address.'
    if (!user && form.password.length < 8) e.password = 'Password must be at least 8 characters.'
    setErrors(e)
    return Object.keys(e).length === 0
  }

  const onSubmit = async (ev: FormEvent) => {
    ev.preventDefault()
    if (!validate()) return
    setSaving(true)
    try {
      if (user) {
        await usersService.patch({
          id: user.id,
          name: form.name,
          phone: form.phone,
          status: form.status as User['status'],
        })
        toast.success('User updated', `${form.name}’s details were saved.`)
      } else {
        await usersService.create({
          name: form.name,
          email: form.email,
          password: form.password,
          phone: form.phone,
          status: form.status as User['status'],
          role_ids: form.role_ids,
        })
        toast.success('User created', `${form.name} can now sign in.`)
      }
      onSaved()
      onClose()
    } catch (err) {
      toast.error(user ? 'Update failed' : 'Creation failed', getErrorMessage(err))
    } finally {
      setSaving(false)
    }
  }

  return (
    <Modal
      open={open}
      onClose={onClose}
      title={user ? 'Edit user' : 'Create user'}
      description={user ? 'Update profile details and account status.' : 'Invite a new member to the workspace.'}
      size="lg"
      footer={
        <>
          <Button variant="ghost" onClick={onClose} disabled={saving}>
            Cancel
          </Button>
          <Button type="submit" form="user-form" loading={saving}>
            {user ? 'Save changes' : 'Create user'}
          </Button>
        </>
      }
    >
      <form id="user-form" onSubmit={onSubmit} className="grid grid-cols-1 gap-4 sm:grid-cols-2">
        <Input
          id="name"
          label="Full name"
          required
          value={form.name}
          onChange={(e) => setForm({ ...form, name: e.target.value })}
          error={errors.name}
          placeholder="Ada Lovelace"
        />
        <Input
          id="email"
          type="email"
          label="Email address"
          required
          value={form.email}
          onChange={(e) => setForm({ ...form, email: e.target.value })}
          error={errors.email}
          placeholder="ada@company.com"
          disabled={Boolean(user)}
          hint={user ? 'Email cannot be changed.' : undefined}
        />
        {!user && (
          <Input
            id="password"
            type="password"
            label="Temporary password"
            required
            value={form.password}
            onChange={(e) => setForm({ ...form, password: e.target.value })}
            error={errors.password}
            hint="At least 8 characters."
          />
        )}
        <Input
          id="phone"
          label="Phone"
          value={form.phone}
          onChange={(e) => setForm({ ...form, phone: e.target.value })}
          placeholder="+1 555 000 1234"
        />
        <FieldShell label="Status" htmlFor="status">
          <Select
            id="status"
            value={form.status}
            onChange={(e) => setForm({ ...form, status: e.target.value })}
            options={[
              { value: 'active', label: 'Active' },
              { value: 'suspended', label: 'Suspended' },
            ]}
          />
        </FieldShell>
        <FieldShell label="Roles" hint="Assign one or more roles." htmlFor="roles">
          <select
            id="roles"
            multiple
            value={form.role_ids}
            onChange={(e) =>
              setForm({
                ...form,
                role_ids: Array.from(e.target.selectedOptions).map((o) => o.value),
              })
            }
            className="h-24 w-full rounded-lg border border-slate-200 bg-white px-3 py-2 text-sm shadow-sm focus:border-primary-500 focus:outline-none focus:ring-4 focus:ring-primary-500/10"
          >
            {roles.map((r) => (
              <option key={r.id} value={r.id}>
                {r.name}
              </option>
            ))}
          </select>
        </FieldShell>
      </form>
    </Modal>
  )
}
