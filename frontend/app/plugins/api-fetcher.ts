// plugins/api-fetcher.ts (Solución para SSR/CSR)
import { useAuth } from "~/composables/useAuth";

export default defineNuxtPlugin(() => {
    
    globalThis.$fetch = $fetch.create({
        baseURL: '', 

        onRequest({ options }) {
            const { getToken } = useAuth(); 
            const token = getToken();
            
            // 🚨 1. Obtener la Cookie del navegador (Solo en SSR)
            const headers = process.server ? useRequestHeaders(['cookie']) : {}; 
            
            options.headers = options.headers || {} as Record<string, string>;
            const fetchHeaders = options.headers as unknown as Record<string, string>;

            // 2. Inyectar la Cookie (CRÍTICO)
            if (headers.cookie) {
                fetchHeaders.Cookie = headers.cookie; 
            }
            
            // 3. Inyectar el Authorization
            if (token) {
                fetchHeaders.Authorization = `Bearer ${token}`;
            }
        },
    })
})