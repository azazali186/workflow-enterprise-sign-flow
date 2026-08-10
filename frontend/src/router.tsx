import { lazy } from 'react'
import { createBrowserRouter, Navigate } from 'react-router-dom'
import { AppLayout } from '@/layouts/AppLayout'
import { GuestOnly, RequireAuth, RequirePermission } from '@/components/RouteGuards'

// Lazy-loaded feature pages (code splitting per route).
const LoginPage = lazy(() =>
  import('@/features/auth/pages/LoginPage').then((m) => ({ default: m.LoginPage })),
)
const DashboardPage = lazy(() =>
  import('@/features/dashboard/pages/DashboardPage').then((m) => ({ default: m.DashboardPage })),
)
const UsersPage = lazy(() =>
  import('@/features/users/pages/UsersPage').then((m) => ({ default: m.UsersPage })),
)
const RolesPage = lazy(() =>
  import('@/features/roles/pages/RolesPage').then((m) => ({ default: m.RolesPage })),
)
const PermissionsPage = lazy(() =>
  import('@/features/permissions/pages/PermissionsPage').then((m) => ({
    default: m.PermissionsPage,
  })),
)
const ContractsPage = lazy(() =>
  import('@/features/contracts/pages/ContractsPage').then((m) => ({ default: m.ContractsPage })),
)
const TemplatesPage = lazy(() =>
  import('@/features/templates/pages/TemplatesPage').then((m) => ({ default: m.TemplatesPage })),
)
const SignersPage = lazy(() =>
  import('@/features/signers/pages/SignersPage').then((m) => ({ default: m.SignersPage })),
)
const SignaturesPage = lazy(() =>
  import('@/features/signatures/pages/SignaturesPage').then((m) => ({
    default: m.SignaturesPage,
  })),
)
const VerificationsPage = lazy(() =>
  import('@/features/verifications/pages/VerificationsPage').then((m) => ({
    default: m.VerificationsPage,
  })),
)
const StoragesPage = lazy(() =>
  import('@/features/storages/pages/StoragesPage').then((m) => ({ default: m.StoragesPage })),
)
const CompliancesPage = lazy(() =>
  import('@/features/compliances/pages/CompliancesPage').then((m) => ({
    default: m.CompliancesPage,
  })),
)
const CertificatesPage = lazy(() =>
  import('@/features/certificates/pages/CertificatesPage').then((m) => ({
    default: m.CertificatesPage,
  })),
)
const AuditLogsPage = lazy(() =>
  import('@/features/auditlogs/pages/AuditLogsPage').then((m) => ({ default: m.AuditLogsPage })),
)
const LoginLogsPage = lazy(() =>
  import('@/features/loginlogs/pages/LoginLogsPage').then((m) => ({ default: m.LoginLogsPage })),
)

export const router = createBrowserRouter([
  {
    path: '/login',
    element: (
      <GuestOnly>
        <LoginPage />
      </GuestOnly>
    ),
  },
  {
    path: '/app',
    element: (
      <RequireAuth>
        <AppLayout />
      </RequireAuth>
    ),
    children: [
      { index: true, element: <Navigate to="/app/dashboard" replace /> },
      { path: 'dashboard', element: <DashboardPage /> },
      {
        path: 'users',
        element: (
          <RequirePermission permission="POST /api/v1/users/list">
            <UsersPage />
          </RequirePermission>
        ),
      },
      {
        path: 'roles',
        element: (
          <RequirePermission permission="POST /api/v1/roles/list">
            <RolesPage />
          </RequirePermission>
        ),
      },
      {
        path: 'permissions',
        element: (
          <RequirePermission permission="POST /api/v1/permissions/list">
            <PermissionsPage />
          </RequirePermission>
        ),
      },
      {
        path: 'contracts',
        element: (
          <RequirePermission permission="POST /api/v1/contracts/list">
            <ContractsPage />
          </RequirePermission>
        ),
      },
      {
        path: 'templates',
        element: (
          <RequirePermission permission="POST /api/v1/templates/list">
            <TemplatesPage />
          </RequirePermission>
        ),
      },
      {
        path: 'signers',
        element: (
          <RequirePermission permission="POST /api/v1/signers/list">
            <SignersPage />
          </RequirePermission>
        ),
      },
      {
        path: 'signatures',
        element: (
          <RequirePermission permission="POST /api/v1/signatures/list">
            <SignaturesPage />
          </RequirePermission>
        ),
      },
      {
        path: 'verifications',
        element: (
          <RequirePermission permission="POST /api/v1/verifications/list">
            <VerificationsPage />
          </RequirePermission>
        ),
      },
      {
        path: 'storages',
        element: (
          <RequirePermission permission="POST /api/v1/storages/list">
            <StoragesPage />
          </RequirePermission>
        ),
      },
      {
        path: 'compliances',
        element: (
          <RequirePermission permission="POST /api/v1/compliances/list">
            <CompliancesPage />
          </RequirePermission>
        ),
      },
      {
        path: 'certificates',
        element: (
          <RequirePermission permission="POST /api/v1/certificates/list">
            <CertificatesPage />
          </RequirePermission>
        ),
      },
      {
        path: 'audit-logs',
        element: (
          <RequirePermission permission="POST /api/v1/audit_logs/list">
            <AuditLogsPage />
          </RequirePermission>
        ),
      },
      {
        path: 'login-logs',
        element: (
          <RequirePermission permission="POST /api/v1/login_logs/list">
            <LoginLogsPage />
          </RequirePermission>
        ),
      },
    ],
  },
  { path: '*', element: <Navigate to="/app/dashboard" replace /> },
])
