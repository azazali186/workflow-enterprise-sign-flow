import {
  ClipboardList,
  FileCheck2,
  FileClock,
  FileSignature,
  Files,
  FolderLock,
  KeyRound,
  LayoutDashboard,
  LogIn,
  ShieldCheck,
  Users,
  type LucideIcon,
} from 'lucide-react'

export interface NavItem {
  label: string
  path: string
  icon: LucideIcon
  /** "METHOD /api/v1/path" the user must hold (super_admin bypasses). */
  permission?: string
  children?: NavItem[]
}

export const navigation: NavItem[] = [
  { label: 'Dashboard', path: '/app/dashboard', icon: LayoutDashboard },
  {
    label: 'Contracts',
    path: '/app/contracts',
    icon: FileSignature,
    permission: 'POST /api/v1/contracts/list',
  },
  {
    label: 'Templates',
    path: '/app/templates',
    icon: Files,
    permission: 'POST /api/v1/templates/list',
  },
  {
    label: 'Signers',
    path: '/app/signers',
    icon: ClipboardList,
    permission: 'POST /api/v1/signers/list',
  },
  {
    label: 'Signatures',
    path: '/app/signatures',
    icon: FileCheck2,
    permission: 'POST /api/v1/signatures/list',
  },
  {
    label: 'Compliance & Trust',
    path: '/app/compliance',
    icon: ShieldCheck,
    permission: 'POST /api/v1/compliances/list',
    children: [
      {
        label: 'Verifications',
        path: '/app/verifications',
        icon: LogIn,
        permission: 'POST /api/v1/verifications/list',
      },
      {
        label: 'Storages',
        path: '/app/storages',
        icon: FolderLock,
        permission: 'POST /api/v1/storages/list',
      },
      {
        label: 'Compliances',
        path: '/app/compliances',
        icon: ShieldCheck,
        permission: 'POST /api/v1/compliances/list',
      },
      {
        label: 'Certificates',
        path: '/app/certificates',
        icon: FileCheck2,
        permission: 'POST /api/v1/certificates/list',
      },
    ],
  },
  {
    label: 'Administration',
    path: '/app/admin',
    icon: KeyRound,
    permission: 'POST /api/v1/users/list',
    children: [
      {
        label: 'Users',
        path: '/app/users',
        icon: Users,
        permission: 'POST /api/v1/users/list',
      },
      {
        label: 'Roles',
        path: '/app/roles',
        icon: ShieldCheck,
        permission: 'POST /api/v1/roles/list',
      },
      {
        label: 'Permissions',
        path: '/app/permissions',
        icon: KeyRound,
        permission: 'POST /api/v1/permissions/list',
      },
      {
        label: 'Audit Logs',
        path: '/app/audit-logs',
        icon: FileClock,
        permission: 'POST /api/v1/audit_logs/list',
      },
      {
        label: 'Login Logs',
        path: '/app/login-logs',
        icon: LogIn,
        permission: 'POST /api/v1/login_logs/list',
      },
    ],
  },
]
