import { describe, expect, it } from 'vitest'
import { can, hasRoutePermission, isSuperAdmin, roleSlugs } from './permissions'
import type { User } from '@/types/entities'

function makeUser(overrides: Partial<User> = {}): User {
  return {
    id: 'u1',
    name: 'Test',
    email: 't@example.com',
    phone: '',
    status: 'active',
    last_login_at: null,
    created_at: '2026-01-01T00:00:00Z',
    updated_at: '2026-01-01T00:00:00Z',
    roles: [],
    ...overrides,
  }
}

describe('permission checks', () => {
  it('denies everything for anonymous users', () => {
    expect(isSuperAdmin(null)).toBe(false)
    expect(hasRoutePermission(null, 'POST', '/api/v1/users/list')).toBe(false)
  })

  it('grants everything to super_admin', () => {
    const user = makeUser({
      roles: [{ id: 'r1', name: 'Super Admin', slug: 'super_admin', description: '', is_system: true, created_at: '', updated_at: '' }],
    })
    expect(isSuperAdmin(user)).toBe(true)
    expect(hasRoutePermission(user, 'DELETE', '/api/v1/users')).toBe(true)
    expect(can(user, 'POST', '/anything')).toBe(true)
  })

  it('grants only when a role holds the exact route permission', () => {
    const user = makeUser({
      roles: [
        {
          id: 'r1',
          name: 'Viewer',
          slug: 'viewer',
          description: '',
          is_system: true,
          created_at: '',
          updated_at: '',
          permissions: [{ id: 'p1', name: 'List Users', route: 'POST /api/v1/users/list', path: '/api/v1/users/list', method: 'POST', service: 'api-gateway', created_at: '', updated_at: '' }],
        },
      ],
    })
    expect(hasRoutePermission(user, 'POST', '/api/v1/users/list')).toBe(true)
    // Different method on the same path is NOT granted.
    expect(hasRoutePermission(user, 'PATCH', '/api/v1/users/list')).toBe(false)
    expect(hasRoutePermission(user, 'POST', '/api/v1/users')).toBe(false)
  })

  it('exposes role slugs', () => {
    const user = makeUser({ roles: [{ id: 'r1', name: 'A', slug: 'viewer', description: '', is_system: true, created_at: '', updated_at: '' }] })
    expect(roleSlugs(user)).toEqual(['viewer'])
    expect(roleSlugs(null)).toEqual([])
  })
})
