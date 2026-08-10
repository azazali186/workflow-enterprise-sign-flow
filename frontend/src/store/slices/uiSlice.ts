import { createSlice, type PayloadAction } from '@reduxjs/toolkit'

export interface Toast {
  id: string
  kind: 'success' | 'error' | 'info'
  title: string
  description?: string
}

export interface UIState {
  sidebarCollapsed: boolean
  mobileNavOpen: boolean
  toasts: Toast[]
}

const initialState: UIState = {
  sidebarCollapsed: false,
  mobileNavOpen: false,
  toasts: [],
}

let toastSeq = 0

const uiSlice = createSlice({
  name: 'ui',
  initialState,
  reducers: {
    toggleSidebar(state) {
      state.sidebarCollapsed = !state.sidebarCollapsed
    },
    setMobileNav(state, action: PayloadAction<boolean>) {
      state.mobileNavOpen = action.payload
    },
    pushToast(state, action: PayloadAction<Omit<Toast, 'id'>>) {
      const id = `toast-${++toastSeq}`
      state.toasts.push({ ...action.payload, id })
    },
    dismissToast(state, action: PayloadAction<string>) {
      state.toasts = state.toasts.filter((t) => t.id !== action.payload)
    },
  },
})

export const { toggleSidebar, setMobileNav, pushToast, dismissToast } = uiSlice.actions
export default uiSlice.reducer
