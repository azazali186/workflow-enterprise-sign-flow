import { createSlice, type PayloadAction } from '@reduxjs/toolkit'
import { tokenStore } from '@/services/api/client'
import type { User } from '@/types/entities'

export type AuthStatus = 'idle' | 'authenticating' | 'authenticated' | 'anonymous'

export interface AuthState {
  status: AuthStatus
  token: string | null
  user: User | null
}

const initialState: AuthState = {
  status: 'idle',
  token: null,
  user: null,
}

const authSlice = createSlice({
  name: 'auth',
  initialState,
  reducers: {
    sessionStart(state, action: PayloadAction<{ token: string; user: User }>) {
      state.status = 'authenticated'
      state.token = action.payload.token
      state.user = action.payload.user
    },
    sessionRestore(state, action: PayloadAction<User>) {
      // Restore the persisted token too, so refreshes/deep links keep the
      // session instead of bouncing to /login.
      state.status = 'authenticated'
      state.token = tokenStore.get()
      state.user = action.payload
    },
    sessionEnd(state) {
      state.status = 'anonymous'
      state.token = null
      state.user = null
    },
  },
})

export const { sessionStart, sessionRestore, sessionEnd } = authSlice.actions
export default authSlice.reducer
