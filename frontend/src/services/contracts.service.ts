import { post, patch, del } from './api/client'
import type { ListQuery, Page } from '@/types/api'
import type {
  Contract,
  CreateContractRequest,
  PatchContractRequest,
  StatusSummary,
} from '@/types/entities'

export const contractsService = {
  create: (body: CreateContractRequest) => post<Contract>('/api/v1/contracts', body),
  patch: (body: PatchContractRequest) => patch<Contract>('/api/v1/contracts', body),
  remove: (id: string) => del<null>('/api/v1/contracts', { id }),
  list: (q: ListQuery) =>
    post<Page<Contract & { summary?: StatusSummary[] }>>('/api/v1/contracts/list', q),
  detail: (id: string) => post<Contract>('/api/v1/contracts/detail', { id }),
  sendSignatureRequest: (id: string) =>
    post<Contract>('/api/v1/contracts/send_signature_request', { id }),
  execute: (id: string) => post<Contract>('/api/v1/contracts/execute', { id }),
  cancel: (id: string) => post<Contract>('/api/v1/contracts/cancel', { id }),
}
