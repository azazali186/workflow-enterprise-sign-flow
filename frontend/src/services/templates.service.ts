import { post, patch, del } from './api/client'
import type { ListQuery, Page } from '@/types/api'
import type { CreateTemplateRequest, PatchTemplateRequest, Template } from '@/types/entities'

export const templatesService = {
  create: (body: CreateTemplateRequest) => post<Template>('/api/v1/templates', body),
  patch: (body: PatchTemplateRequest) => patch<Template>('/api/v1/templates', body),
  remove: (id: string) => del<null>('/api/v1/templates', { id }),
  list: (q: ListQuery) => post<Page<Template>>('/api/v1/templates/list', q),
  detail: (id: string) => post<Template>('/api/v1/templates/detail', { id }),
}
