import { beforeEach, describe, expect, it } from 'vitest'
import authReducer, { sessionEnd, sessionRestore, sessionStart } from './authSlice'
import { tokenStore } from '@/services/api/client'
import type { User } from '@/types/entities'

const user: User = {
  id: 'u1',
  name: 'Ada',
  email: 'ada@example.com',
  phone: '',
  status: 'active',
  last_login_at: null,
  created_at: '2026-01-01T00:00:00Z',
  updated_at: '2026-01-01T00:00:00Z',
  roles: [],
}

describe('auth slice', () => {
  beforeEach(() => tokenStore.clear())

  it('starts idle with no token', () => {
    const s = authReducer(undefined, { type: '@@init' })
    expect(s.status).toBe('idle')
    expect(s.token).toBeNull()
    expect(s.user).toBeNull()
  })

  it('sessionStart stores the token and user', () => {
    const s = authReducer(undefined, sessionStart({ token: 'tok', user }))
    expect(s.status).toBe('authenticated')
    expect(s.token).toBe('tok')
    expect(s.user?.email).toBe('ada@example.com')
  })

  it('sessionRestore restores the persisted token (refresh regression)', () => {
    // Simulate a hard reload: token is in storage, slice is idle.
    tokenStore.set('persisted-token')
    const s = authReducer(undefined, sessionRestore(user))
    // Critical: without this, AppLayout redirects authenticated users to
    // /login on refresh. See the review that caught the original bug.
    expect(s.status).toBe('authenticated')
    expect(s.token).toBe('persisted-token')
    expect(s.user).toEqual(user)
  })

  it('sessionEnd clears everything', () => {
    const started = authReducer(undefined, sessionStart({ token: 'tok', user }))
    const s = authReducer(started, sessionEnd())
    expect(s.status).toBe('anonymous')
    expect(s.token).toBeNull()
    expect(s.user).toBeNull()
  })
})
