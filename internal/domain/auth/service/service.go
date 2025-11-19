package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"golang.org/x/crypto/bcrypt"

	"github.com/mikrocloud/mikrocloud/internal/domain/auth"
	"github.com/mikrocloud/mikrocloud/internal/domain/auth/repository"
	"github.com/mikrocloud/mikrocloud/internal/domain/users"
	usersRepo "github.com/mikrocloud/mikrocloud/internal/domain/users/repository"
)

// Service errors
var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserNotFound       = errors.New("user not found")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrWeakPassword       = errors.New("password does not meet requirements")
	ErrInvalidEmail       = errors.New("invalid email format")
	ErrInvalidToken       = errors.New("invalid or expired token")
	ErrSessionExpired     = errors.New("session has expired")
)

// AuthService handles authentication business logic
type AuthService struct {
	sessionRepo repository.SessionRepository
	authRepo    repository.AuthRepository
	usersRepo   usersRepo.Repository

	// Configuration
	jwtSecret            string
	sessionDuration      time.Duration
	refreshTokenDuration time.Duration
}

// NewAuthService creates a new authentication service
func NewAuthService(
	sessionRepo repository.SessionRepository,
	authRepo repository.AuthRepository,
	usersRepo usersRepo.Repository,
	jwtSecret string,
) *AuthService {
	return &AuthService{
		sessionRepo:          sessionRepo,
		authRepo:             authRepo,
		usersRepo:            usersRepo,
		jwtSecret:            jwtSecret,
		sessionDuration:      24 * time.Hour,      // 24 hours for regular sessions
		refreshTokenDuration: 30 * 24 * time.Hour, // 30 days for refresh tokens
	}
}

// Command types for service operations
type LoginCommand struct {
	Email    string
	Password string
}

type RegisterCommand struct {
	Name     string
	Email    string
	Password string
}

// Result types for service operations
type LoginResult struct {
	User         *users.User
	Token        string
	RefreshToken string
}

type RegisterResult struct {
	User  *users.User
	Token string
}

type RefreshTokenResult struct {
	Token        string
	RefreshToken string
}

// Login authenticates a user and creates a new session
func (s *AuthService) Login(ctx context.Context, cmd LoginCommand) (*LoginResult, error) {
	// Validate input
	if cmd.Email == "" || cmd.Password == "" {
		return nil, ErrInvalidCredentials
	}

	// Create email value object
	email, err := users.NewEmail(cmd.Email)
	if err != nil {
		return nil, ErrInvalidEmail
	}

	// Get user by email
	user, err := s.authRepo.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, ErrUserNotFound
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash()), []byte(cmd.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	// Check if user is active
	if user.Status() != users.UserStatusActive {
		return nil, ErrInvalidCredentials
	}

	// Generate JWT token
	token, err := s.generateJWTToken(ctx, user.ID().String())
	if err != nil {
		return nil, fmt.Errorf("failed to generate JWT token: %w", err)
	}

	// Generate session token for database tracking
	sessionToken, err := s.generateSecureToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate session token: %w", err)
	}

	// Create session
	session := auth.NewSession(user.ID(), sessionToken, s.sessionDuration)
	if err := s.sessionRepo.SaveSession(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to save session: %w", err)
	}

	// Generate refresh token
	refreshTokenStr, err := s.generateSecureToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	refreshToken := auth.NewRefreshToken(user.ID(), session.ID(), refreshTokenStr, s.refreshTokenDuration)
	if err := s.sessionRepo.SaveRefreshToken(ctx, refreshToken); err != nil {
		return nil, fmt.Errorf("failed to save refresh token: %w", err)
	}

	// Update last login
	if err := s.authRepo.UpdateLastLogin(ctx, user.ID()); err != nil {
		// Log error but don't fail the login
		slog.Error("failed to update last login", "error", err, "user_id", user.ID())
	}

	return &LoginResult{
		User:         user,
		Token:        token,
		RefreshToken: refreshTokenStr,
	}, nil
}

// Register creates a new user account
func (s *AuthService) Register(ctx context.Context, cmd RegisterCommand) (*RegisterResult, error) {
	// Validate input
	if err := s.validateRegisterCommand(cmd); err != nil {
		return nil, err
	}

	// Create email value object
	email, err := users.NewEmail(cmd.Email)
	if err != nil {
		return nil, ErrInvalidEmail
	}

	// Check if user already exists
	exists, err := s.authRepo.UserExistsByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("failed to check user existence: %w", err)
	}
	if exists {
		return nil, ErrUserAlreadyExists
	}

	// Check if this is the first user (setup scenario)
	hasUsers, err := s.authRepo.HasAnyUsers(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to check if users exist: %w", err)
	}
	isFirstUser := !hasUsers

	// Hash password
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(cmd.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create new user
	user := users.NewUserWithName(email, string(passwordHash), cmd.Name)

	// If this is the first user, make them active (setup scenario)
	if isFirstUser {
		user.ChangeStatus(users.UserStatusActive)
	}

	// Save user
	if err := s.authRepo.CreateUser(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// If this is the first user, set up default organization and admin role
	if isFirstUser {
		if err := s.setupFirstUser(ctx, user); err != nil {
			return nil, fmt.Errorf("failed to set up first user: %w", err)
		}
	}

	// Generate JWT token
	token, err := s.generateJWTToken(ctx, user.ID().String())
	if err != nil {
		return nil, fmt.Errorf("failed to generate JWT token: %w", err)
	}

	// Generate session token for database tracking
	sessionToken, err := s.generateSecureToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate session token: %w", err)
	}

	// Create session
	session := auth.NewSession(user.ID(), sessionToken, s.sessionDuration)
	if err := s.sessionRepo.SaveSession(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to save session: %w", err)
	}

	return &RegisterResult{
		User:  user,
		Token: token,
	}, nil
}

// Logout invalidates a user session
func (s *AuthService) Logout(ctx context.Context, jwtToken string) error {
	if jwtToken == "" {
		return ErrInvalidToken
	}

	// Parse and verify JWT token
	token, err := jwt.Parse([]byte(jwtToken), jwt.WithKey(jwa.HS256, []byte(s.jwtSecret)))
	if err != nil {
		return ErrInvalidToken
	}

	// Extract user_id from claims
	userIDClaim, ok := token.Get("user_id")
	if !ok {
		return ErrInvalidToken
	}

	userIDStr, ok := userIDClaim.(string)
	if !ok || userIDStr == "" {
		return ErrInvalidToken
	}

	// Parse user ID
	userID, err := users.UserIDFromString(userIDStr)
	if err != nil {
		return ErrInvalidToken
	}

	// Revoke all sessions for this user
	if err := s.sessionRepo.RevokeAllUserSessions(ctx, userID); err != nil {
		return fmt.Errorf("failed to revoke user sessions: %w", err)
	}

	return nil
}

// RefreshToken creates a new session token using a refresh token
func (s *AuthService) RefreshToken(ctx context.Context, refreshTokenStr string) (*RefreshTokenResult, error) {
	if refreshTokenStr == "" {
		return nil, ErrInvalidToken
	}

	// Get refresh token
	refreshToken, err := s.sessionRepo.GetRefreshTokenByToken(ctx, refreshTokenStr)
	if err != nil {
		return nil, ErrInvalidToken
	}

	// Validate refresh token
	if !refreshToken.IsValid() {
		return nil, ErrInvalidToken
	}

	// Mark refresh token as used
	if err := s.sessionRepo.MarkRefreshTokenAsUsed(ctx, refreshToken.ID()); err != nil {
		return nil, fmt.Errorf("failed to mark refresh token as used: %w", err)
	}

	// Generate new session token
	newToken, err := s.generateSecureToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate session token: %w", err)
	}

	// Create new session
	newSession := auth.NewSession(refreshToken.UserID(), newToken, s.sessionDuration)
	if err := s.sessionRepo.SaveSession(ctx, newSession); err != nil {
		return nil, fmt.Errorf("failed to save session: %w", err)
	}

	// Generate new refresh token
	newRefreshTokenStr, err := s.generateSecureToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	newRefreshToken := auth.NewRefreshToken(refreshToken.UserID(), newSession.ID(), newRefreshTokenStr, s.refreshTokenDuration)
	if err := s.sessionRepo.SaveRefreshToken(ctx, newRefreshToken); err != nil {
		return nil, fmt.Errorf("failed to save refresh token: %w", err)
	}

	return &RefreshTokenResult{
		Token:        newToken,
		RefreshToken: newRefreshTokenStr,
	}, nil
}

// GetUserByID retrieves a user by their ID
func (s *AuthService) GetUserByID(ctx context.Context, userIDStr string) (*users.User, error) {
	userID, err := users.UserIDFromString(userIDStr)
	if err != nil {
		return nil, ErrUserNotFound
	}

	user, err := s.authRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, ErrUserNotFound
	}

	return user, nil
}

// ValidateSession validates a session token and returns the associated user
func (s *AuthService) ValidateSession(ctx context.Context, token string) (*users.User, error) {
	if token == "" {
		return nil, ErrInvalidToken
	}

	// Get session by token
	session, err := s.sessionRepo.GetSessionByToken(ctx, token)
	if err != nil {
		return nil, ErrInvalidToken
	}

	// Validate session
	if !session.IsValid() {
		return nil, ErrSessionExpired
	}

	// Get user
	user, err := s.authRepo.GetUserByID(ctx, session.UserID())
	if err != nil {
		return nil, ErrUserNotFound
	}

	return user, nil
}

// CleanupExpiredSessions removes expired sessions and refresh tokens
func (s *AuthService) CleanupExpiredSessions(ctx context.Context) error {
	if err := s.sessionRepo.DeleteExpiredSessions(ctx); err != nil {
		return fmt.Errorf("failed to delete expired sessions: %w", err)
	}

	if err := s.sessionRepo.DeleteExpiredRefreshTokens(ctx); err != nil {
		return fmt.Errorf("failed to delete expired refresh tokens: %w", err)
	}

	return nil
}

// SetupStatus represents the setup status of the application
type SetupStatus struct {
	HasUsers bool `json:"has_users"`
	IsSetup  bool `json:"is_setup"`
}

// GetSetupStatus checks if the application has been set up (has at least one user)
func (s *AuthService) GetSetupStatus(ctx context.Context) (*SetupStatus, error) {
	// Check if there are any users in the system
	hasUsers, err := s.authRepo.HasAnyUsers(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to check if users exist: %w", err)
	}

	return &SetupStatus{
		HasUsers: hasUsers,
		IsSetup:  hasUsers,
	}, nil
}

// HasAnyUsers checks if there are any users in the system
func (s *AuthService) HasAnyUsers(ctx context.Context) (bool, error) {
	return s.authRepo.HasAnyUsers(ctx)
}

// Helper methods

func (s *AuthService) validateRegisterCommand(cmd RegisterCommand) error {
	if cmd.Name == "" {
		return errors.New("name is required")
	}
	if cmd.Email == "" {
		return ErrInvalidEmail
	}
	if len(cmd.Password) < 8 {
		return ErrWeakPassword
	}
	// Additional password strength validation could be added here
	return nil
}

func (s *AuthService) generateSecureToken() (string, error) {
	// Generate 32 random bytes
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	// Encode to base64 URL-safe string
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// generateJWTToken generates a JWT token for the given user
func (s *AuthService) generateJWTToken(ctx context.Context, userID string) (string, error) {
	now := time.Now()

	// Parse user ID
	userIDObj, err := users.UserIDFromString(userID)
	if err != nil {
		return "", fmt.Errorf("invalid user ID: %w", err)
	}

	// Get user's organizations to include first one as org_id in JWT
	orgs, err := s.usersRepo.FindOrganizationsByUser(ctx, userIDObj)
	if err != nil {
		slog.Error("Failed to get user organizations", "error", err, "user_id", userID)
		return "", fmt.Errorf("failed to get user organizations: %w", err)
	}

	// Use the first organization ID, or empty string if no organizations
	var orgID string
	if len(orgs) > 0 {
		orgID = orgs[0].ID().String()
		slog.Info("Found organizations for user", "user_id", userID, "org_id", orgID, "org_count", len(orgs))
	} else {
		slog.Warn("No organizations found for user", "user_id", userID)
	}

	token, err := jwt.NewBuilder().
		Issuer("mikrocloud").
		IssuedAt(now).
		Expiration(now.Add(s.sessionDuration)).
		Claim("user_id", userID).
		Claim("org_id", orgID).
		Build()
	if err != nil {
		return "", fmt.Errorf("failed to build JWT: %w", err)
	}

	signed, err := jwt.Sign(token, jwt.WithKey(jwa.HS256, []byte(s.jwtSecret)))
	if err != nil {
		return "", fmt.Errorf("failed to sign JWT: %w", err)
	}

	return string(signed), nil
}

func (s *AuthService) setupFirstUser(ctx context.Context, user *users.User) error {
	org := users.NewOrganization(
		user.Name(),
		user.Name(),
		user.ID(),
	)

	if err := s.usersRepo.SaveOrganization(ctx, org); err != nil {
		return fmt.Errorf("failed to create default organization: %w", err)
	}

	if err := s.usersRepo.AddOrganizationMember(ctx, org.ID(), user.ID(), "owner", nil); err != nil {
		return fmt.Errorf("failed to add user to organization: %w", err)
	}

	adminRoleID, err := s.usersRepo.FindRoleByName(ctx, "admin")
	if err != nil {
		return fmt.Errorf("failed to find admin role: %w", err)
	}

	if err := s.usersRepo.AddUserRole(ctx, user.ID(), adminRoleID, nil); err != nil {
		return fmt.Errorf("failed to assign admin role: %w", err)
	}

	return nil
}
