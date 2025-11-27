// plugins/auth-init.ts
export default defineNuxtPlugin(async (nuxtApp) => {
    // Obtenemos la composable de autenticación
    const { user, checkAuth, getToken } = useAuth(); 

    // Solo si el estado 'user' no ha sido cargado aún
    // y si la aplicación se ejecuta por primera vez.
    if (!user.value) {
        // Obtenemos el token almacenado en la cookie.
        const token = getToken(); 
        
        if (token) {
            // Si existe un token, llamamos a la lógica para verificarlo 
            // y cargar los datos del usuario.
            await checkAuth(); 
        }
    }
    
    // Si la llamada a checkAuth fallara (token expirado), el estado 'user' 
    // seguiría siendo null, y el middleware se encargaría de redirigir 
    // en el siguiente paso del ciclo de vida.
});