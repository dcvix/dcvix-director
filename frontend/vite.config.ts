import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [react()],
  build: {
    sourcemap: true,
  },
  server: {
    proxy: {
      '/v1': {
        target: 'https://127.0.0.1:8445',
        changeOrigin: true,
        secure: false,
      }
    }
  }
})