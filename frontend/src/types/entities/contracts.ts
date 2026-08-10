import type { BaseEntity } from '../api'

export type ContractStatus =
  | 'draft'
  | 'awaiting_signature'
  | 'partially_signed'
  | 'signed'
  | 'executed'
  | 'cancelled'
  | 'expired'

export type SignerStatus = 'pending' | 'signed' | 'declined'

export interface Signer extends BaseEntity {
  contract_id: string
  name: string
  email: string
  phone: string
  role: string
  status: SignerStatus
  order: number
  signed_at: string | null
  signature_id: string
}

export interface Contract extends BaseEntity {
  title: string
  reference_no: string
  description: string
  status: ContractStatus
  template_id: string
  created_by: string
  document_storage_id: string
  sent_at: string | null
  executed_at: string | null
  expires_at: string | null
  metadata?: unknown
  signers?: Signer[]
}

export interface SignerInput {
  name: string
  email: string
  phone?: string
  role?: string
  order?: number
}

export interface CreateContractRequest {
  title: string
  reference_no?: string
  description?: string
  template_id?: string
  expires_at?: string | null
  metadata?: unknown
  signers?: SignerInput[]
}

export interface PatchContractRequest {
  id: string
  title?: string
  description?: string
  expires_at?: string | null
  metadata?: unknown
}

export interface CreateSignerRequest {
  contract_id: string
  name: string
  email: string
  phone?: string
  role?: string
  order?: number
}

export interface PatchSignerRequest {
  id: string
  name?: string
  phone?: string
  role?: string
  order?: number
}

export type SignatureStatus = 'pending' | 'captured' | 'verified'

export interface Signature extends BaseEntity {
  contract_id: string
  signer_id: string
  status: SignatureStatus
  type: string
  ip_address: string
  user_agent: string
  signed_at: string | null
  verification_id: string
}

export type VerificationStatus = 'pending' | 'verified' | 'failed'

export interface Verification extends BaseEntity {
  signature_id: string
  contract_id: string
  method: string
  status: VerificationStatus
  attempts: number
  otp_expires_at: string | null
  verified_by: string
  verified_at: string | null
}
