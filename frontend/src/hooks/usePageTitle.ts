import { useEffect } from 'react'

/** Sets the document title for the current page; falls back to the brand default on unmount. */
export function usePageTitle(title?: string) {
  useEffect(() => {
    if (!title) return
    document.title = `${title} · SignFlow`
    return () => {
      document.title = 'SignFlow · Secure e-Signing Platform'
    }
  }, [title])
}
