import { defineConfig } from 'vite';

export default defineConfig({
  // Output directory - this will build to your Go server's static directory
  build: {
    outDir: '../backend/static',
    emptyOutDir: true
  },
  // Base path for asset URLs
  base: '/',
  // Enable SPA routing support
  server: {
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true
      }
    }
  }
});