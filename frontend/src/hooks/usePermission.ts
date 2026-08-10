import { useAppSelector } from '@/store'
import { can, isSuperAdmin, roleSlugs } from '@/utils/permissions'

/** Permission helpers bound to the authenticated user. */
export function usePermission() {
  const user = useAppSelector((s) => s.auth.user)

  return {
    user,
    isSuperAdmin: isSuperAdmin(user),
    roles: roleSlugs(user),
    can: (method: string, path: string) => can(user, method, path),
  }
}
