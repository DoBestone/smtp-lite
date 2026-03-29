import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { fileURLToPath, URL } from 'node:url'

export default defineConfig({
  plugins: [vue()],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url))
    }
  },
  server: {
    port: 8091,
    proxy: {
      '/api': 'http://localhost:8090',
      '/track': 'http://localhost:8090'
    }
  },
  build: {
    outDir: '../web/dist'
  }
})