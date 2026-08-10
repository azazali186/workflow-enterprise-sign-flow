import type { BaseEntity } from '../api'
import type { Role } from './rbac'

export interface LoginRequest {
  email: string
  password: string
}

export interface LoginResult {
  access_token: string
  token_type: string
  expires_in: number
  user: User
}

export interface User extends BaseEntity {
  name: string
  email: string
  phone: string
  status: UserStatus
  last_login_at: string | null
  roles?: Role[]
}

export type UserStatus = 'active' | 'suspended'

export interface CreateUserRequest {
  name: string
  email: string
  password: string
  phone?: string
  status?: UserStatus
  role_ids?: string[]
}

export interface PatchUserRequest {
  id: string
  name?: string
  phone?: string
  status?: UserStatus
}

export interface AssignRolesRequest {
  id: string
  role_ids: string[]
}

/** Shape returned in Page.summary for users. */
export interface UserSummary {
  status: string
  count: number
}
