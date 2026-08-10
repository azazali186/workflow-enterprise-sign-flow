import type { User } from '@/types/entities'

/**
 * Backend permissions are seeded from routes as "METHOD /api/v1/...".
 * The backend grants are available on the user's roles' permissions.
 * super_admin bypasses every check (mirrors the backend middleware).
 */

export const SUPER_ADMIN = 'super_admin'

export function isSuperAdmin(user?: User | null): boolean {
  return Boolean(user?.roles?.some((r) => r.slug === SUPER_ADMIN))
}

export function roleSlugs(user?: User | null): string[] {
  return (user?.roles ?? []).map((r) => r.slug)
}

export function hasRoutePermission(user: User | null, method: string, path: string): boolean {
  if (!user) return false
  if (isSuperAdmin(user)) return true
  const route = `${method} ${path}`
  return Boolean(
    user.roles?.some((r) =>
      (r.permissions ?? []).some((p) => p.route === route),
    ),
  )
}

/** UI guard: renders children only when the user holds the permission. */
export function can(
  user: User | null,
  method: string,
  path: string,
): boolean {
  return hasRoutePermission(user, method, path)
}
