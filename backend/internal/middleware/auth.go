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

// JWTMiddleware es un middleware de Echo que verifica el token de autorización.
func JWTMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// 1. Obtener el valor del encabezado Authorization
		authHeader := c.Request().Header.Get("Authorization")
		fmt.Println("AUTH HEADER", authHeader)

		// La lógica de Echo para devolver un error es retornar c.JSON o c.String
		if authHeader == "" {
			return c.JSON(http.StatusUnauthorized, echo.Map{
				"error": "Token de autorización requerido",
			})
		}

		// 2. Extraer el token (eliminar el prefijo "Bearer ")
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || strings.ToLower(tokenParts[0]) != "bearer" {
			return c.JSON(http.StatusUnauthorized, echo.Map{
				"error": "Formato de token no válido. Debe ser Bearer <token>",
			})
		}
		tokenString := tokenParts[1]

		// 3. Lógica para verificar y decodificar el token
		// La función verifyToken debe devolver los claims (incluyendo el ID) o un error.
		claims, err := verifyToken(tokenString) // <--- Usa tu función JWT
		fmt.Println("claims", claims)

		if err != nil {
			// Token inválido o expirado
			return c.JSON(http.StatusUnauthorized, echo.Map{
				"error": "Token inválido o expirado",
			})
		}

		// 4. Si es válido, almacena el ID de usuario en el contexto
		// Asegúrate de que claims.UserID sea el tipo de dato correcto (ej. int)
		c.Set(ContextKeyUserID, claims.Name)

		// 5. Continuar con el siguiente handler
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
			return "secret", nil
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
