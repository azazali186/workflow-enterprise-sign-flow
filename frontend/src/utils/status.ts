/** Semantic badge tones, shared by the Badge component and status mappers. */
export type BadgeTone = 'neutral' | 'primary' | 'success' | 'warning' | 'danger' | 'info'

/**
 * Maps backend status strings to badge tones. Unknown values fall back to
 * neutral so new statuses never render broken.
 */
export function toneForStatus(status?: string | null): BadgeTone {
  switch (status) {
    case 'active':
    case 'signed':
    case 'executed':
    case 'approved':
    case 'verified':
    case 'valid':
    case 'stored':
    case 'success':
      return 'success'
    case 'draft':
    case 'pending':
    case 'awaiting_signature':
    case 'partially_signed':
      return 'warning'
    case 'cancelled':
    case 'expired':
    case 'declined':
    case 'rejected':
    case 'failed':
    case 'suspended':
    case 'revoked':
      return 'danger'
    default:
      return 'neutral'
  }
}
