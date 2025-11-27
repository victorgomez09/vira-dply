package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v4"
	"github.com/labstack/echo/v4"
)

type jwtCustomClaims struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Admin bool   `json:"admin"`
	jwt.RegisteredClaims
}

// ContextKeyUserID es la clave para almacenar el ID de usuario en el contexto de Echo.
const ContextKeyUserID = "name"

// 🚨 Nombre de la cookie donde Nuxt almacena el token
const AuthCookieName = "auth_token"

// JWTMiddleware es un middleware de Echo que verifica el token de autorización,
// buscando primero en la cookie 'auth_token' y luego en el encabezado 'Authorization'.
func JWTMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		var tokenString string

		// 1. 🚨 INTENTAR OBTENER EL TOKEN DE LA COOKIE
		cookie, err := c.Cookie(AuthCookieName)
		if err == nil && cookie.Value != "" {
			// Token encontrado en la cookie
			tokenString = cookie.Value
		}

		// 2. INTENTAR OBTENER EL TOKEN DEL ENCABEZADO (MECANISMO DE RESPALDO)
		if tokenString == "" {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader != "" {
				// Extraer el token (eliminar el prefijo "Bearer ")
				tokenParts := strings.Split(authHeader, " ")
				if len(tokenParts) == 2 && strings.ToLower(tokenParts[0]) == "bearer" {
					tokenString = tokenParts[1]
				}
			}
		}

		// 3. VERIFICAR SI SE ENCONTRÓ ALGÚN TOKEN
		if tokenString == "" {
			return c.JSON(http.StatusUnauthorized, echo.Map{
				"error": "Token de autorización (Cookie o Header) requerido",
			})
		}

		// 4. Lógica para verificar y decodificar el token
		claims, err := verifyToken(tokenString) // <--- Usa tu función JWT

		if err != nil {
			// Token inválido o expirado
			return c.JSON(http.StatusUnauthorized, echo.Map{
				"error": "Token inválido o expirado: " + err.Error(),
			})
		}

		// 5. Si es válido, almacena el ID de usuario en el contexto
		c.Set(ContextKeyUserID, claims.Name)

		// 6. Continuar con el siguiente handler
		return next(c)
	}
}

// verifyToken verifica la validez de un token JWT y devuelve los claims.
func verifyToken(tokenString string) (*jwtCustomClaims, error) {
	// 1. Parsing del token
	token, err := jwt.ParseWithClaims(
		tokenString,
		&jwtCustomClaims{}, // Usamos la estructura definida
		func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("método de firma inesperado: %v", token.Method.Alg())
			}
			// 🚨 CRÍTICO: Reemplaza "secret" con tu clave secreta REAL
			return []byte("secret"), nil
		},
	)

	if err != nil {
		return nil, fmt.Errorf("fallo al parsear el token: %w", err)
	}

	// 2. Extracción y validación de los claims
	if claims, ok := token.Claims.(*jwtCustomClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("token no válido")
}
