import { post, patch } from './api/client'
import type { ListQuery, Page } from '@/types/api'
import type { CreateStorageRequest, PatchStorageRequest, Storage } from '@/types/entities'

export const storagesService = {
  create: (body: CreateStorageRequest) => post<Storage>('/api/v1/storages', body),
  patch: (body: PatchStorageRequest) => patch<Storage>('/api/v1/storages', body),
  list: (q: ListQuery) => post<Page<Storage>>('/api/v1/storages/list', q),
  detail: (id: string) => post<Storage>('/api/v1/storages/detail', { id }),
}
