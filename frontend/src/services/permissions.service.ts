import { post } from './api/client'
import type { ListQuery, Page } from '@/types/api'
import type { Permission } from '@/types/entities'

export const permissionsService = {
  list: (q: ListQuery) => post<Page<Permission>>('/api/v1/permissions/list', q),
  detail: (id: string) => post<Permission>('/api/v1/permissions/detail', { id }),
}
