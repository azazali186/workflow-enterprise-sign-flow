import { post, patch } from './api/client'
import type { ListQuery, Page } from '@/types/api'
import type {
  Certificate,
  CreateCertificateRequest,
  PatchCertificateRequest,
} from '@/types/entities'

export const certificatesService = {
  create: (body: CreateCertificateRequest) => post<Certificate>('/api/v1/certificates', body),
  patch: (body: PatchCertificateRequest) => patch<Certificate>('/api/v1/certificates', body),
  list: (q: ListQuery) => post<Page<Certificate>>('/api/v1/certificates/list', q),
  detail: (id: string) => post<Certificate>('/api/v1/certificates/detail', { id }),
}
