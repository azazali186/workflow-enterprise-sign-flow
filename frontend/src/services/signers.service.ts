import { post, patch, del } from './api/client'
import type { ListQuery, Page } from '@/types/api'
import type { CreateSignerRequest, PatchSignerRequest, Signer } from '@/types/entities'

export const signersService = {
  create: (body: CreateSignerRequest) => post<Signer>('/api/v1/signers', body),
  patch: (body: PatchSignerRequest) => patch<Signer>('/api/v1/signers', body),
  remove: (id: string) => del<null>('/api/v1/signers', { id }),
  list: (q: ListQuery) => post<Page<Signer>>('/api/v1/signers/list', q),
}
