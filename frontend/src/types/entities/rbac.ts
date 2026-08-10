import type { BaseEntity } from '../api'

export interface Role extends BaseEntity {
  name: string
  slug: string
  description: string
  is_system: boolean
  permissions?: Permission[]
}

export interface CreateRoleRequest {
  name: string
  slug: string
  description?: string
}

export interface PatchRoleRequest {
  id: string
  name?: string
  description?: string
}

export interface AssignPermissionsRequest {
  id: string
  permission_ids: string[]
}

export interface Permission extends BaseEntity {
  name: string
  route: string
  path: string
  method: string
  service: string
}

/** Shape returned in Page.summary for roles. */
export interface RoleSummary {
  total_roles: number
  system_roles: number
}

/** Common by-status summary used by several lists. */
export interface StatusSummary {
  status: string
  count: number
}
