import { post } from './api/client'
import type { LoginRequest, LoginResult, User } from '@/types/entities'

/** Auth endpoints. Login is PUBLIC; logout and me require the bearer token. */
export const authService = {
  login: (body: LoginRequest) => post<LoginResult>('/api/v1/auth/login', body),
  logout: () => post<null>('/api/v1/auth/logout', {}),
  me: () => post<User>('/api/v1/auth/me', {}),
}
