import { useEffect } from 'react'
import { useAppDispatch, useAppSelector } from '@/store'
import { sessionEnd, sessionRestore, sessionStart } from '@/store/slices/authSlice'
import { authService } from '@/services/auth.service'
import { setUnauthorizedHandler, tokenStore } from '@/services/api/client'

/**
 * Bootstraps the session on first mount:
 *  - no token        -> anonymous (login screen)
 *  - token present   -> POST /me to restore the user; 401 clears the token
 * Also registers the global 401 handler (single sign-out everywhere).
 */
export function useSession() {
  const dispatch = useAppDispatch()
  const status = useAppSelector((s) => s.auth.status)

  useEffect(() => {
    setUnauthorizedHandler(() => dispatch(sessionEnd()))

    const token = tokenStore.get()
    if (!token) {
      dispatch(sessionEnd())
      return
    }

    let cancelled = false
    authService
      .me()
      .then((user) => {
        if (!cancelled) dispatch(sessionRestore(user))
      })
      .catch(() => {
        if (!cancelled) {
          tokenStore.clear()
          dispatch(sessionEnd())
        }
      })
    return () => {
      cancelled = true
    }
  }, [dispatch])

  // Expose a programmatic login helper for the login page.
  const login = async (email: string, password: string) => {
    const res = await authService.login({ email, password })
    tokenStore.set(res.access_token)
    dispatch(sessionStart({ token: res.access_token, user: res.user }))
    return res.user
  }

  return { status, login }
}
