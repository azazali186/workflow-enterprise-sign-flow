import { post } from './api/client'
import type { ListQuery, Page } from '@/types/api'
import type { AuditLog, LoginLog } from '@/types/entities'

export const auditlogsService = {
  list: (q: ListQuery) => post<Page<AuditLog>>('/api/v1/audit_logs/list', q),
  detail: (id: string) => post<AuditLog>('/api/v1/audit_logs/detail', { id }),
}

export const loginlogsService = {
  list: (q: ListQuery) => post<Page<LoginLog>>('/api/v1/login_logs/list', q),
}
