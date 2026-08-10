import type { BaseEntity } from '../api'

export interface AuditLog extends BaseEntity {
  action: string
  entity_type: string
  entity_id: string
  actor_user_id: string
  actor_name: string
  before_data?: unknown
  after_data?: unknown
  changed_fields?: unknown
  ip: string
  user_agent: string
  request_id: string
}

export interface LoginLog extends BaseEntity {
  username: string
  ip: string
  user_agent: string
  success: boolean
  message: string
  login_at: string
}
