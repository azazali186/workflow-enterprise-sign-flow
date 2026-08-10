import { render, screen } from '@testing-library/react'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { Provider } from 'react-redux'
import { configureStore } from '@reduxjs/toolkit'
import { describe, expect, it } from 'vitest'
import { GuestOnly, RequireAuth, RequirePermission } from './RouteGuards'
import authReducer from '@/store/slices/authSlice'
import type { User } from '@/types/entities'

function makeStore(status: 'idle' | 'authenticated' | 'anonymous', token: string | null = null) {
  return configureStore({
    reducer: { auth: authReducer, ui: () => ({ sidebarCollapsed: false, mobileNavOpen: false, toasts: [] }) },
    preloadedState: {
      auth: { status, token, user: status === 'authenticated' ? user : null },
    } as const,
  })
}

const user: User = {
  id: 'u1',
  name: 'Ada',
  email: 'ada@example.com',
  phone: '',
  status: 'active',
  last_login_at: null,
  created_at: '',
  updated_at: '',
  roles: [
    {
      id: 'r1',
      name: 'Super Admin',
      slug: 'super_admin',
      description: '',
      is_system: true,
      created_at: '',
      updated_at: '',
    },
  ],
}

function renderAt(ui: React.ReactNode, status: 'idle' | 'authenticated' | 'anonymous') {
  return render(
    <Provider store={makeStore(status, status === 'authenticated' ? 'tok' : null)}>
      <MemoryRouter initialEntries={['/app/secret']}>
        <Routes>
          <Route path="/app/secret" element={ui} />
          <Route path="/login" element={<div>login-screen</div>} />
        </Routes>
      </MemoryRouter>
    </Provider>,
  )
}

describe('RequireAuth', () => {
  it('renders children when authenticated', () => {
    renderAt(<RequireAuth>secret content</RequireAuth>, 'authenticated')
    expect(screen.getByText('secret content')).toBeInTheDocument()
  })

  it('redirects anonymous users to /login', () => {
    renderAt(<RequireAuth>secret content</RequireAuth>, 'anonymous')
    expect(screen.getByText('login-screen')).toBeInTheDocument()
  })
})

describe('GuestOnly', () => {
  it('redirects authenticated users away from the login screen', () => {
    render(
      <Provider store={makeStore('authenticated', 'tok')}>
        <MemoryRouter initialEntries={['/login']}>
          <Routes>
            <Route path="/login" element={<GuestOnly>login-form</GuestOnly>} />
            <Route path="/app/dashboard" element={<div>dashboard</div>} />
          </Routes>
        </MemoryRouter>
      </Provider>,
    )
    expect(screen.getByText('dashboard')).toBeInTheDocument()
  })

  it('renders the login screen for guests', () => {
    renderAt(<GuestOnly>login-form</GuestOnly>, 'anonymous')
    expect(screen.getByText('login-form')).toBeInTheDocument()
  })
})

describe('RequirePermission', () => {
  it('renders children when the user holds the permission', () => {
    renderAt(
      <RequirePermission permission="POST /api/v1/users/list">users page</RequirePermission>,
      'authenticated',
    )
    expect(screen.getByText('users page')).toBeInTheDocument()
  })

  it('shows an access-denied state without the permission', () => {
    const viewer = { ...user, roles: [{ ...user.roles![0], slug: 'viewer' }] }
    const store = configureStore({
      reducer: { auth: authReducer, ui: () => ({ sidebarCollapsed: false, mobileNavOpen: false, toasts: [] }) },
      preloadedState: { auth: { status: 'authenticated', token: 'tok', user: viewer } } as const,
    })
    render(
      <Provider store={store}>
        <MemoryRouter>
          <RequirePermission permission="POST /api/v1/users/list">users page</RequirePermission>
        </MemoryRouter>
      </Provider>,
    )
    expect(screen.queryByText('users page')).not.toBeInTheDocument()
    expect(screen.getByText(/don’t have access/i)).toBeInTheDocument()
  })
})
