import { post, patch } from './api/client'
import type { ListQuery, Page } from '@/types/api'
import type { Compliance, CreateComplianceRequest, ReviewComplianceRequest } from '@/types/entities'

export const compliancesService = {
  create: (body: CreateComplianceRequest) => post<Compliance>('/api/v1/compliances', body),
  review: (body: ReviewComplianceRequest) => patch<Compliance>('/api/v1/compliances', body),
  list: (q: ListQuery) => post<Page<Compliance>>('/api/v1/compliances/list', q),
  detail: (id: string) => post<Compliance>('/api/v1/compliances/detail', { id }),
}
