package repository

import (
	"context"

	"github.com/mikrocloud/mikrocloud/internal/domain/auth"
	"github.com/mikrocloud/mikrocloud/internal/domain/users"
)

// SessionRepository defines the interface for session storage operations
type SessionRepository interface {
	// Session operations
	SaveSession(ctx context.Context, session *auth.Session) error
	GetSessionByToken(ctx context.Context, token string) (*auth.Session, error)
	GetSessionByID(ctx context.Context, sessionID auth.SessionID) (*auth.Session, error)
	GetActiveSessionsByUserID(ctx context.Context, userID users.UserID) ([]*auth.Session, error)
	RevokeSession(ctx context.Context, sessionID auth.SessionID) error
	RevokeAllUserSessions(ctx context.Context, userID users.UserID) error
	DeleteExpiredSessions(ctx context.Context) error

	// Refresh token operations
	SaveRefreshToken(ctx context.Context, refreshToken *auth.RefreshToken) error
	GetRefreshTokenByToken(ctx context.Context, token string) (*auth.RefreshToken, error)
	GetRefreshTokenByID(ctx context.Context, tokenID auth.RefreshTokenID) (*auth.RefreshToken, error)
	MarkRefreshTokenAsUsed(ctx context.Context, tokenID auth.RefreshTokenID) error
	DeleteRefreshToken(ctx context.Context, tokenID auth.RefreshTokenID) error
	DeleteExpiredRefreshTokens(ctx context.Context) error
	DeleteRefreshTokensBySessionID(ctx context.Context, sessionID auth.SessionID) error
}

// AuthRepository defines the interface for authentication-related user operations
type AuthRepository interface {
	// User authentication operations
	GetUserByEmail(ctx context.Context, email users.Email) (*users.User, error)
	GetUserByID(ctx context.Context, userID users.UserID) (*users.User, error)
	CreateUser(ctx context.Context, user *users.User) error
	UpdateUserPassword(ctx context.Context, userID users.UserID, passwordHash string) error
	UpdateLastLogin(ctx context.Context, userID users.UserID) error
	UserExistsByEmail(ctx context.Context, email users.Email) (bool, error)
	HasAnyUsers(ctx context.Context) (bool, error)
}
