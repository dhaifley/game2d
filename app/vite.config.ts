import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [react()],
  server: {
    proxy: {
      // Proxy API requests to the Go backend
      '/api': {
        target: 'http://localhost:8080', // Assuming the Go server runs on port 8080
        changeOrigin: true,
        secure: false,
      },
    },
  },
})
