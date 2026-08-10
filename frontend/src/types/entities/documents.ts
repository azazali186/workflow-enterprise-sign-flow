import type { BaseEntity } from '../api'

export interface Template extends BaseEntity {
  name: string
  slug: string
  description: string
  content: string
  version: number
  is_active: boolean
  created_by: string
}

export interface CreateTemplateRequest {
  name: string
  slug: string
  description?: string
  content?: string
  is_active?: boolean
}

export interface PatchTemplateRequest {
  id: string
  name?: string
  description?: string
  content?: string
  is_active?: boolean
}

export type StorageStatus = 'pending' | 'stored' | 'failed'

export interface Storage extends BaseEntity {
  entity_type: string
  entity_id: string
  bucket: string
  content_type: string
  size_bytes: number
  checksum: string
  status: StorageStatus
  uploaded_at: string | null
}

export interface CreateStorageRequest {
  entity_type: string
  entity_id: string
  bucket?: string
  content_type?: string
  size_bytes?: number
  checksum?: string
}

export interface PatchStorageRequest {
  id: string
  status?: StorageStatus
  uploaded_at?: string | null
}

export type ComplianceStatus = 'pending' | 'approved' | 'rejected'

export interface Compliance extends BaseEntity {
  contract_id: string
  type: string
  status: ComplianceStatus
  evidence?: unknown
  reviewed_by: string
  reviewed_at: string | null
}

export interface CreateComplianceRequest {
  contract_id: string
  type: string
  evidence?: unknown
}

export interface ReviewComplianceRequest {
  id: string
  status: 'approved' | 'rejected'
  evidence?: unknown
}

export type CertificateStatus = 'valid' | 'revoked' | 'expired'

export interface Certificate extends BaseEntity {
  contract_id: string
  subject: string
  issuer: string
  serial_number: string
  not_before: string | null
  not_after: string | null
  status: CertificateStatus
}

export interface CreateCertificateRequest {
  contract_id: string
  subject: string
  issuer?: string
  serial_number?: string
  not_before?: string | null
  not_after?: string | null
}

export interface PatchCertificateRequest {
  id: string
  status?: CertificateStatus
  not_after?: string | null
}

export interface DashboardSummary {
  contracts: {
    total: number
    draft: number
    sent: number
    signed: number
    executed: number
    created_today: number
  }
  signatures: {
    total: number
    captured: number
    verified: number
    captured_today: number
  }
  signers: {
    pending: number
    signed: number
  }
  templates: number
  active_users: number
  storage_bytes: number
}

export interface SummaryRequest {
  date_from?: string | null
  date_to?: string | null
}
