import '@testing-library/jest-dom/vitest'

// jsdom lacks matchMedia, which some components query indirectly.
if (typeof window !== 'undefined' && !window.matchMedia) {
  window.matchMedia = (query: string) =>
    ({
      matches: false,
      media: query,
      onchange: null,
      addListener: () => undefined,
      removeListener: () => undefined,
      addEventListener: () => undefined,
      removeEventListener: () => undefined,
      dispatchEvent: () => false,
    }) as MediaQueryList
}

// Silence React Router/Query console noise in tests when needed.
globalThis.ResizeObserver = globalThis.ResizeObserver ?? class {
  observe() {}
  unobserve() {}
  disconnect() {}
}

// jsdom's Storage can be incomplete depending on the version — provide a
// reliable in-memory implementation for the token store used by the client.
const memoryStore = new Map<string, string>()
const fakeStorage: Storage = {
  get length() {
    return memoryStore.size
  },
  clear: () => memoryStore.clear(),
  getItem: (k: string) => memoryStore.get(k) ?? null,
  key: (i: number) => Array.from(memoryStore.keys())[i] ?? null,
  removeItem: (k: string) => void memoryStore.delete(k),
  setItem: (k: string, v: string) => void memoryStore.set(k, String(v)),
}
Object.defineProperty(globalThis, 'localStorage', { value: fakeStorage, configurable: true })
