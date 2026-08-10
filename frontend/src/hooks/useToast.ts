import { useCallback } from 'react'
import { useAppDispatch } from '@/store'
import { dismissToast, pushToast } from '@/store/slices/uiSlice'

/** Imperative toast API used across the app (no business logic inside UI). */
export function useToast() {
  const dispatch = useAppDispatch()

  const success = useCallback(
    (title: string, description?: string) => {
      dispatch(pushToast({ kind: 'success', title, description }))
    },
    [dispatch],
  )

  const error = useCallback(
    (title: string, description?: string) => {
      dispatch(pushToast({ kind: 'error', title, description }))
    },
    [dispatch],
  )

  const info = useCallback(
    (title: string, description?: string) => {
      dispatch(pushToast({ kind: 'info', title, description }))
    },
    [dispatch],
  )

  const dismiss = useCallback(
    (id: string) => {
      dispatch(dismissToast(id))
    },
    [dispatch],
  )

  return { success, error, info, dismiss }
}
