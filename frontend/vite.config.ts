import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

const backendUrl = process.env.VITE_BACKEND_URL ?? 'http://localhost:8080'
const apiBasePath = process.env.VITE_API_BASE_PATH ?? '/api/v1'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  server: {
    proxy: {
      [apiBasePath]: {
        target: backendUrl,
        changeOrigin: true,
      },
    },
  },
})
