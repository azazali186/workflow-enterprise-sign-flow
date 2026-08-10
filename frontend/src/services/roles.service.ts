import { post, patch, del } from './api/client'
import type { ListQuery, Page } from '@/types/api'
import type {
  AssignPermissionsRequest,
  CreateRoleRequest,
  PatchRoleRequest,
  Role,
} from '@/types/entities'

export const rolesService = {
  create: (body: CreateRoleRequest) => post<Role>('/api/v1/roles', body),
  patch: (body: PatchRoleRequest) => patch<Role>('/api/v1/roles', body),
  remove: (id: string) => del<null>('/api/v1/roles', { id }),
  list: (q: ListQuery) => post<Page<Role>>('/api/v1/roles/list', q),
  detail: (id: string) => post<Role>('/api/v1/roles/detail', { id }),
  assignPermissions: (body: AssignPermissionsRequest) =>
    patch<Role>('/api/v1/roles/assign_permissions', body),
}
