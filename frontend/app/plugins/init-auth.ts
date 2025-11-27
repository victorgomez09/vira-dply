// plugins/auth-init.ts
export default defineNuxtPlugin(async (_) => {
    const { user, checkAuth, getToken, authReady } = useAuth(); 

    if (!user.value && !authReady.value) {
        const token = getToken(); 
        if (token) {
            await checkAuth(); 
        } else {
            authReady.value = true;
        }
    }
});