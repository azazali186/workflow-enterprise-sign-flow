/// <reference types="vitest/config" />
import path from 'node:path'
import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'

// The backend (Hertz) serves the API at http://localhost:8080 by default.
// All /api requests are proxied in dev so the frontend can call relative
// paths and never worries about CORS.
const BACKEND_URL = process.env.VITE_API_PROXY || 'http://localhost:8080'

export default defineConfig({
  plugins: [react(), tailwindcss()],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, 'src'),
    },
  },
  server: {
    port: 5173,
    proxy: {
      '/api': {
        target: BACKEND_URL,
        changeOrigin: true,
      },
    },
  },
  test: {
    environment: 'jsdom',
    globals: true,
    setupFiles: ['./src/test/setup.ts'],
    include: ['src/**/*.test.{ts,tsx}'],
    coverage: {
      provider: 'v8',
      reporter: ['text', 'html'],
      include: ['src/utils/**', 'src/hooks/**', 'src/services/api/**', 'src/components/ui/**'],
    },
  },
  build: {
    sourcemap: false,
    chunkSizeWarningLimit: 900,
    rollupOptions: {
      output: {
        // Stable vendor chunks: libraries split from app code.
        manualChunks(id) {
          if (id.includes('node_modules')) {
            if (id.includes('react') || id.includes('react-router')) return 'vendor-react'
            if (id.includes('redux')) return 'vendor-redux'
            if (id.includes('@tanstack')) return 'vendor-query'
            if (id.includes('lucide')) return 'vendor-icons'
            return 'vendor'
          }
        },
      },
    },
  },
})
