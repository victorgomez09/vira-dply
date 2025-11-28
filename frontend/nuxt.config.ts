import tailwindcss from "@tailwindcss/vite";

// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({
  compatibilityDate: '2025-07-15',
  devtools: { enabled: true },
  modules: ['@nuxt/ui', '@nuxt/hints'],
  css: ['./app/assets/css/main.css'],
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
        // 🚨 CRÍTICO: Usamos la ruta /api como prefijo
        '/api': {
          target: 'http://localhost:1323',
          changeOrigin: true,
          secure: false,
          configure: (proxy, options) => {
            // El 'proxy' es una instancia de http-proxy-middleware
            
            // Este hook se ejecuta justo antes de enviar la solicitud al target (Echo)
            proxy.on('proxyReq', (proxyReq, req, res) => {
              // Copiamos TODOS los headers de la solicitud original (req) 
              // a la solicitud que va al backend (proxyReq).
              
              // Los headers ya están inyectados por el interceptor de $fetch 
              // en la solicitud original de Nuxt.

              // La solicitud original (req) incluye el Authorization inyectado
              // y el Cookie enviado por el navegador.
              
              // Este paso asegura que se copien, previniendo la eliminación.
            
              // No necesitas código aquí si changeOrigin: true es suficiente,
              // pero para depuración, puedes ver los headers:
              
              // console.log("--- PROXYING HEADERS ---");
              // console.log("Authorization:", req.headers['authorization']); 
              // console.log("------------------------");
              
              // Nota: Dado que changeOrigin: true a menudo es suficiente para
              // esto, si sigue fallando, la única forma es la manipulación manual
              // que ya hace changeOrigin.

              // Si el header 'Authorization' sigue faltando, deshabilitamos la caché de headers:
              if (proxyReq.getHeader('authorization')) {
                  // Si el header existe, lo dejamos pasar. 
              } else if (req.headers['authorization']) {
                  // Si no está en el proxyReq, lo añadimos de la solicitud original.
                  proxyReq.setHeader('authorization', req.headers['authorization']);
              }
              
              // También nos aseguramos de que el host sea el local, para evitar conflictos:
              proxyReq.setHeader('host', 'localhost:1323');
            });
          }
        }
      }
    }
  }
})