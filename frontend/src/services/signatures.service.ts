import { post } from './api/client'
import type { ListQuery, Page } from '@/types/api'
import type { Signature } from '@/types/entities'

export const signaturesService = {
  list: (q: ListQuery) => post<Page<Signature>>('/api/v1/signatures/list', q),
  detail: (id: string) => post<Signature>('/api/v1/signatures/detail', { id }),
  /** Capture a signature (draw | type | upload). */
  capture: (body: { contract_id: string; signer_id: string; type: string; data: string }) =>
    post<Signature>('/api/v1/signatures/capture', body),
}
