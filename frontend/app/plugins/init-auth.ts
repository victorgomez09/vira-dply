// plugins/auth-init.ts

export default defineNuxtPlugin(async (nuxtApp) => {
    const { user, checkAuth, getToken, authReady } = useAuth(); 

    // 1. Lógica de Inicialización (SSR/CSR inicial)
    // Esto se ejecuta ANTES de que el router inicie en SSR.
    if (!user.value && !authReady.value) {
        const token = getToken(); 
        if (token) {
            await checkAuth(); 
        } else {
            authReady.value = true;
        }
    }

    // 2. 🚨 SOLUCIÓN: Usar nuxtApp.hook('app:beforeMount') para acceder al Router.
    // El hook 'app:beforeEach' NO existe directamente en nuxtApp.hook().
    
    // Si necesitas bloquear la navegación ANTES del middleware, usa nuxtApp.hook('app:mounted') 
    // y luego Vue Router.
    nuxtApp.hook('app:mounted', () => {
        const router = useRouter(); // Esto es una composable, seguro en app:mounted o setup
        router.beforeEach(async (to, from, next) => {
            // Esta lógica se ejecuta en el lado del cliente (CSR) antes de la navegación.
            
            // Si la autenticación aún NO está lista (sólo debería pasar en CSR inicial)
            if (!authReady.value) {
                // Si el token no ha sido verificado, espera o ejecuta checkAuth()
                const token = getToken();
                if (token) {
                    await checkAuth();
                    next(); // Continuar después de la verificación
                    return;
                }
            }
            
            // Si todo está listo (authReady=true), dejar que el middleware se encargue.
            next();
        });
    });
});