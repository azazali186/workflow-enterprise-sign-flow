import { post } from './api/client'
import type { ListQuery, Page } from '@/types/api'
import type { DashboardSummary, SummaryRequest } from '@/types/entities'

export const dashboardService = {
  summary: (body: SummaryRequest = {}) => post<DashboardSummary>('/api/v1/dashboard/summary', body),
}

/** Report endpoints return the same Page shape with DB summaries. */
export const reportsService = {
  contracts: (q: ListQuery) => post<Page<unknown>>('/api/v1/reports/contracts', q),
  signatures: (q: ListQuery) => post<Page<unknown>>('/api/v1/reports/signatures', q),
  signers: (q: ListQuery) => post<Page<unknown>>('/api/v1/reports/signers', q),
  auditLogs: (q: ListQuery) => post<Page<unknown>>('/api/v1/reports/audit_logs', q),
}
