// middleware/auth.ts

export default defineNuxtRouteMiddleware((to, from) => {
  // Obtenemos el estado
  const { loggedIn, authReady } = useAuth();
  
  // 🚨 CRÍTICO: Si la autenticación aún no está lista, no hagas nada.
  // En SSR, esto se resuelve inmediatamente gracias al plugin.
  if (!authReady.value) {
      // El middleware espera pasivamente a que el plugin resuelva el estado
      // antes de que la aplicación avance con la redirección.
      return
  }

  // Ahora, si la autenticación ya está lista:
  if (authReady.value && !loggedIn.value && to.meta.auth !== false) {
    // Si no está logueado y la ruta requiere autenticación
    return navigateTo('/login');
  }
});