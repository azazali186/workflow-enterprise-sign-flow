import { post } from './api/client'
import type { ListQuery, Page } from '@/types/api'
import type { Verification } from '@/types/entities'

export const verificationsService = {
  list: (q: ListQuery) => post<Page<Verification>>('/api/v1/verifications/list', q),
  detail: (id: string) => post<Verification>('/api/v1/verifications/detail', { id }),
  /** Create a verification request; OTP is delivered separately. */
  create: (body: { signature_id: string; contract_id: string; method: string }) =>
    post<Verification>('/api/v1/verifications', body),
  verify: (body: { id: string; code: string }) =>
    post<Verification>('/api/v1/verifications/verify', body),
}
