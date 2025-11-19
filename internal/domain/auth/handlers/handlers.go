package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/jwtauth/v5"
	"github.com/go-playground/validator/v10"
	"github.com/mikrocloud/mikrocloud/internal/domain/auth/service"
	"github.com/mikrocloud/mikrocloud/internal/utils"
)

// AuthHandler handles authentication-related HTTP requests
type AuthHandler struct {
	authService *service.AuthService
	validator   *validator.Validate
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(as *service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: as,
		validator:   validator.New(),
	}
}

// LoginRequest represents the login request payload
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

// RegisterRequest represents the registration request payload
type RegisterRequest struct {
	Name     string `json:"name" validate:"required,min=2,max=100"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

// AuthResponse represents the authentication response
type AuthResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

// User represents user data in responses
type User struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Email    string  `json:"email"`
	Username *string `json:"username,omitempty"`
}

// Login authenticates a user and returns a token
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.SendError(w, http.StatusBadRequest, "invalid_json", "Invalid JSON format")
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		utils.SendError(w, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	cmd := service.LoginCommand{
		Email:    req.Email,
		Password: req.Password,
	}

	result, err := h.authService.Login(r.Context(), cmd)
	if err != nil {
		// Check if it's an authentication error
		switch err {
		case service.ErrInvalidCredentials:
			utils.SendError(w, http.StatusUnauthorized, "invalid_credentials", "Invalid email or password")
		case service.ErrUserNotFound:
			utils.SendError(w, http.StatusUnauthorized, "invalid_credentials", "Invalid email or password")
		default:
			utils.SendError(w, http.StatusInternalServerError, "login_failed", "Login failed: "+err.Error())
		}
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    result.RefreshToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   60 * 60 * 24 * 7,
	})

	var username *string
	if result.User.Username() != nil {
		usernameStr := result.User.Username().String()
		username = &usernameStr
	}

	response := AuthResponse{
		Token: result.Token,
		User: User{
			ID:       result.User.ID().String(),
			Name:     result.User.Name(),
			Email:    result.User.Email().String(),
			Username: username,
		},
	}

	utils.SendJSON(w, http.StatusOK, response)
}

// Register creates a new user account
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.SendError(w, http.StatusBadRequest, "invalid_json", "Invalid JSON format")
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		utils.SendError(w, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	cmd := service.RegisterCommand{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
	}

	result, err := h.authService.Register(r.Context(), cmd)
	if err != nil {
		// Check if it's a user already exists error
		switch err {
		case service.ErrUserAlreadyExists:
			utils.SendError(w, http.StatusConflict, "user_exists", "A user with this email already exists")
		case service.ErrWeakPassword:
			utils.SendError(w, http.StatusBadRequest, "weak_password", "Password does not meet security requirements")
		case service.ErrInvalidEmail:
			utils.SendError(w, http.StatusBadRequest, "invalid_email", "Invalid email format")
		default:
			utils.SendError(w, http.StatusInternalServerError, "registration_failed", "Registration failed: "+err.Error())
		}
		return
	}

	var username *string
	if result.User.Username() != nil {
		usernameStr := result.User.Username().String()
		username = &usernameStr
	}

	response := AuthResponse{
		Token: result.Token,
		User: User{
			ID:       result.User.ID().String(),
			Name:     result.User.Name(),
			Email:    result.User.Email().String(),
			Username: username,
		},
	}

	utils.SendJSON(w, http.StatusCreated, response)
}

// Logout invalidates the user's token
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// Extract token from Authorization header
	authHeader := r.Header.Get("Authorization")

	if authHeader == "" {
		utils.SendError(w, http.StatusUnauthorized, "missing_token", "Authorization header is required")
		return
	}

	// Expect "Bearer <token>" format
	const bearerPrefix = "Bearer "
	if len(authHeader) <= len(bearerPrefix) || authHeader[:len(bearerPrefix)] != bearerPrefix {
		utils.SendError(w, http.StatusUnauthorized, "invalid_token_format", "Authorization header must be in 'Bearer <token>' format")
		return
	}

	token := authHeader[len(bearerPrefix):]
	if token == "" {
		utils.SendError(w, http.StatusUnauthorized, "missing_token", "Token is required")
		return
	}

	err := h.authService.Logout(r.Context(), token)
	if err != nil {
		switch err {
		case service.ErrInvalidToken:
			utils.SendError(w, http.StatusUnauthorized, "invalid_token", "Invalid or expired token")
		default:
			utils.SendError(w, http.StatusInternalServerError, "logout_failed", "Logout failed: "+err.Error())
		}
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   -1,
	})

	response := utils.SuccessResponse{
		Message: "Successfully logged out",
	}

	utils.SendJSON(w, http.StatusOK, response)
}

// RefreshToken refreshes an expired token
func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		utils.SendError(w, http.StatusUnauthorized, "missing_refresh_token", "Refresh token cookie not found")
		return
	}

	refreshToken := cookie.Value
	if refreshToken == "" {
		utils.SendError(w, http.StatusUnauthorized, "invalid_refresh_token", "Refresh token is empty")
		return
	}

	result, err := h.authService.RefreshToken(r.Context(), refreshToken)
	if err != nil {
		switch err {
		case service.ErrInvalidToken:
			utils.SendError(w, http.StatusUnauthorized, "invalid_refresh_token", "Invalid or expired refresh token")
		default:
			utils.SendError(w, http.StatusInternalServerError, "refresh_failed", "Token refresh failed: "+err.Error())
		}
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    result.RefreshToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   60 * 60 * 24 * 7,
	})

	type RefreshTokenResponse struct {
		Token string `json:"token"`
	}

	response := RefreshTokenResponse{
		Token: result.Token,
	}

	utils.SendJSON(w, http.StatusOK, response)
}

// GetProfile returns the current user's profile
func (h *AuthHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from JWT claims
	_, claims, err := jwtauth.FromContext(r.Context())
	if err != nil {
		utils.SendError(w, http.StatusUnauthorized, "unauthorized", "Invalid token")
		return
	}

	userID, ok := claims["user_id"].(string)
	if !ok || userID == "" {
		utils.SendError(w, http.StatusUnauthorized, "unauthorized", "User not authenticated")
		return
	}

	user, err := h.authService.GetUserByID(r.Context(), userID)
	if err != nil {
		switch err {
		case service.ErrUserNotFound:
			utils.SendError(w, http.StatusNotFound, "user_not_found", "User not found")
		default:
			utils.SendError(w, http.StatusInternalServerError, "profile_failed", "Failed to get profile: "+err.Error())
		}
		return
	}

	var username *string
	if user.Username() != nil {
		usernameStr := user.Username().String()
		username = &usernameStr
	}

	response := User{
		ID:       user.ID().String(),
		Name:     user.Name(),
		Email:    user.Email().String(),
		Username: username,
	}

	utils.SendJSON(w, http.StatusOK, response)
}

// SetupStatus represents the setup status response
type SetupStatus struct {
	IsSetup bool `json:"is_setup"`
}

// GetSetupStatus checks if the system has been set up (has any users)
func (h *AuthHandler) GetSetupStatus(w http.ResponseWriter, r *http.Request) {
	hasUsers, err := h.authService.HasAnyUsers(r.Context())
	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, "setup_check_failed", "Failed to check setup status: "+err.Error())
		return
	}

	response := SetupStatus{
		IsSetup: hasUsers,
	}

	utils.SendJSON(w, http.StatusOK, response)
}
