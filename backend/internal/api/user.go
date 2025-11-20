package api

import (
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/victorgomez09/vira-dply/internal/dto"
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

	return c.JSON(http.StatusOK, map[string]string{
		"token": t,
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
