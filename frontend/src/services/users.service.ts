import { post, patch, del } from './api/client'
import type { ListQuery, Page } from '@/types/api'
import type {
  AssignRolesRequest,
  CreateUserRequest,
  PatchUserRequest,
  User,
  UserSummary,
} from '@/types/entities'

export const usersService = {
  create: (body: CreateUserRequest) => post<User>('/api/v1/users', body),
  patch: (body: PatchUserRequest) => patch<User>('/api/v1/users', body),
  remove: (id: string) => del<null>('/api/v1/users', { id }),
  list: (q: ListQuery) => post<Page<User & { summary?: UserSummary[] }>>('/api/v1/users/list', q),
  detail: (id: string) => post<User>('/api/v1/users/detail', { id }),
  assignRoles: (body: AssignRolesRequest) => patch<User>('/api/v1/users/assign_roles', body),
}
