import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  server: {
    proxy: {
      // For mirrord demo: VITE_PROXY_TARGET=http://127.0.0.1:PORT (from minikube service)
      // getInventory uses VITE_INVENTORY_API=http://localhost:18082 for branch data
      '/api': {
        target: process.env.VITE_PROXY_TARGET || 'http://127.0.0.1:55587',
        changeOrigin: true,
      },
    },
  },
})
