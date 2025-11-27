import tailwindcss from "@tailwindcss/vite";

// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({
  compatibilityDate: '2025-07-15',
  devtools: { enabled: true },
  modules: ['@nuxt/ui', '@nuxt/hints'],
  css: ['./app/assets/css/main.css'],
  // runtimeConfig: {
  //   public: {
  //     backendUrl: process.env.NODE_ENV === 'production'
  //       ? process.env.BACKEND_API_URL
  //       : 'https://verbose-fiesta-g5wg45vqjpphv5rw-1323.app.github.dev',
  //   }
  // },
  devServer: {
    host: '0.0.0.0',
    port: 3000,
  },
  vite: {
    plugins: [
      tailwindcss(),
    ],
    server: {
      proxy: {
        // 1. Cuando el navegador pida /api/...
        '/api': {
          // 2. Redirige la petición a: http://localhost:8080
          target: 'https://verbose-fiesta-g5wg45vqjpphv5rw-1323.app.github.dev',
          // 3. Importante: Cambia el encabezado 'Host' a http://localhost:8080
          changeOrigin: true,
          // Opcional: si tu backend espera la ruta sin el prefijo /api
          // pathRewrite: { '^/api/': '/' },
        }
      }
    }
  }
})