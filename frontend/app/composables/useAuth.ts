// Define el estado global (accesible en toda la app)
export const useAuth = () => {
    // Usamos useState para la reactividad.
    const user = useState<any | null>('user', () => null)
    const loggedIn = computed(() => !!user.value)
    const authReady = useState<boolean>('authReady', () => false)
    
    // Función para obtener el token del almacenamiento (ej. Cookies)
    function getToken() {
        return useCookie('auth_token').value
    }

    // Función para establecer el token en el almacenamiento
    function setToken(token: string | null) {
        const cookie = useCookie('auth_token', {
            maxAge: 60 * 60 * 24 * 7,
            sameSite: 'strict',
            // secure: process.env.NODE_ENV === 'production'
        })
        cookie.value = token
    }

    // Función de Login
    const login = async (credentials: any) => {
        // const config = useRuntimeConfig()
        // const url = `${config.public.backendUrl}/api/login`

        try {
            // 1. Realizar la petición POST
            const response: { token: string, user_data: any } = await $fetch(`/api/auth/login`, {
                method: 'POST',
                body: credentials
            })
            console.log("response", response)

            // 2. Almacenar el token en la cookie
            setToken(response.token)

            // 3. Actualizar el estado global del usuario
            // (Aquí podrías decodificar el JWT o hacer una petición /user/me para obtener los datos)
            user.value = response.user_data || { email: credentials.email, role: 'user' }
            navigateTo('/private')

            return true
        } catch (error) {
            console.error('Login failed:', error)
            setToken(null)
            user.value = null
            throw error
        }
    }

    // Función de Logout
    const logout = () => {
        // setToken(null)
        user.value = null
        // Opcional: Redirigir al login
        navigateTo('/login')
    }

    const fetchProtected = async (url: string, options: any = {}) => {
    const token = getToken();
    if (!token) {
        throw new Error('No authentication token available.');
    }
    
    // Asegurar que los headers existen
    options.headers = options.headers || {};
    
    // 1. Inyectar el token en el encabezado
    options.headers.Authorization = `Bearer ${token}`; 
    
    // 2. Usar $fetch de Nuxt con el encabezado modificado
    return $fetch(url, options);
}

    // Función para verificar si el usuario ya está logueado (útil al iniciar la app)
    const checkAuth = async () => {
        if (loggedIn.value) return // Ya cargado

        const token = getToken()
        console.log("token", token)
        if (token) {
            try {
                user.value = await fetchProtected('/api/users/me')
                console.log("user.value", user.value)
            } catch (e) {
                // Token inválido o expirado
                // logout()
            }
        }
    }

    return {
        user,
        loggedIn,
        authReady,
        login,
        logout,
        getToken,
        checkAuth,
    }
}