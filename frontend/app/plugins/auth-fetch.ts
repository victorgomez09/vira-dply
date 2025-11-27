import { useAuth } from "~/composables/useAuth"

export default defineNuxtPlugin(() => {
  const { getToken } = useAuth()
  const token = getToken()

  // Extiende $fetch para añadir un interceptor
  globalThis.$fetch = $fetch.create({
    // baseURL: config.public.backendUrl as string || "",
    
    onRequest({ options }) {
      console.log("dentro")
      // 1. Verificar si la llamada va a tu backend y si hay un token
      if (token) {
        // 2. Adjuntar el header Authorization
        options.headers = options.headers || {}
        const headers = options.headers as unknown as Record<string, string>;
        headers.Authorization = `Bearer ${token}`
      }
    },
    
    onResponseError({ response }) {
      // Manejar el error 401 (No autorizado) globalmente
      if (response.status === 401) {
        // useAuth().logout()
        // Redirigir al login si el token es inválido/expirado
      }
    }
  })
})