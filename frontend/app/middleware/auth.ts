import { useAuth } from "~/composables/useAuth"

export default defineNuxtRouteMiddleware((to, from) => {
  const { loggedIn } = useAuth()

  // Si el usuario NO está logueado y trata de acceder a una ruta protegida
  if (!loggedIn.value && to.meta.auth !== false) {
    // Redirigir a la página de login
    return navigateTo('/login')
  }
})