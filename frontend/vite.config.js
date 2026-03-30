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
    port: 6700,
    proxy: {
      '/api': 'http://localhost:6710',
      '/track': 'http://localhost:6710'
    }
  },
  build: {
    outDir: '../web/dist'
  }
})