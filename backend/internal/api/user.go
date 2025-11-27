package api

import (
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/victorgomez09/vira-dply/internal/dto"
	"github.com/victorgomez09/vira-dply/internal/middleware"
	"github.com/victorgomez09/vira-dply/internal/service"
)

type UserHandler struct {
	userSvc *service.UserService
}

func NewUserHandler(svc *service.UserService) *UserHandler {
	return &UserHandler{userSvc: svc}
}

type jwtCustomClaims struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Admin bool   `json:"admin"`
	jwt.RegisteredClaims
}

// Login maneja la autenticación de usuarios y genera un token JWT.
func (h *UserHandler) Login(c echo.Context) error {
	u := &dto.LoginRequest{}
	if err := c.Bind(u); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Solicitud inválida")
	}

	user, err := h.userSvc.GetUserByUsername(u.Username)
	if err != nil || h.userSvc.VerifyPassword(u.Password, user.Password) {
		return echo.NewHTTPError(http.StatusUnauthorized, "Credenciales inválidas")
	}

	claims := &jwtCustomClaims{
		Name:  user.Username,
		Admin: false,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(72 * time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	t, err := token.SignedString([]byte("secret")) // En producción, usa una clave segura
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "No se pudo generar el token")
	}

	return c.JSON(http.StatusOK, map[string]any{
		"token":     t,
		"user_data": map[string]string{"email": user.Email, "username": user.Username},
	})
}

// Register maneja el registro de nuevos usuarios.
func (h *UserHandler) Register(c echo.Context) error {
	u := &dto.CreateUserRequest{}
	if err := c.Bind(u); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Solicitud inválida")
	}

	_, err := h.userSvc.CreateUser(u.Username, u.Email, u.Password)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "No se pudo crear el usuario")
	}

	return c.JSON(http.StatusCreated, map[string]string{
		"message": "Usuario creado exitosamente",
	})
}

// MeHandler maneja la solicitud GET /users/me
func (h *UserHandler) MeHandler(c echo.Context) error {
	// 1. Obtener el ID de usuario del Contexto de Echo
	userIDValue := c.Get(middleware.ContextKeyUserID)
	if userIDValue == nil {
		// Esto solo ocurriría si el middleware no se ejecutó o falló.
		return c.String(http.StatusInternalServerError, "ID de usuario no encontrado en el contexto")
	}

	// Conversión del ID (ajustar el tipo según cómo lo guardó el middleware)
	userID, ok := userIDValue.(string) // Asumimos que el ID es un int
	if !ok {
		return c.String(http.StatusInternalServerError, "Tipo de ID de usuario no válido")
	}

	user, err := h.userSvc.GetUserByUsername(userID)
	if err != nil {
		// Usuario no encontrado en la DB
		return c.JSON(http.StatusNotFound, echo.Map{"error": "Usuario no encontrado"})
	}

	// 3. Respuesta Exitosa usando el método JSON de Echo
	// ¡IMPORTANTE! Asegúrate de que el struct 'user' NO incluya el hash de la contraseña.
	return c.JSON(http.StatusOK, user)
}
