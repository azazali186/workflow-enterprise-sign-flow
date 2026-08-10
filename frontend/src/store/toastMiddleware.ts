import type { Middleware } from '@reduxjs/toolkit'
import { dismissToast, pushToast } from './slices/uiSlice'

interface ToastAction {
  type: string
  payload: { id: string; kind: string }
}

/**
 * Auto-dismisses toasts after a few seconds. Success toasts live briefly;
 * errors stay a little longer so the message is actually read.
 */
export const toastMiddleware: Middleware = (store) => (next) => (action: unknown) => {
  const result = next(action)
  if (typeof action === 'object' && action !== null && (action as { type?: string }).type === pushToast.type) {
    const toast = (action as ToastAction).payload
    const ms = toast.kind === 'error' ? 6000 : 3500
    setTimeout(() => store.dispatch(dismissToast(toast.id)), ms)
  }
  return result
}
