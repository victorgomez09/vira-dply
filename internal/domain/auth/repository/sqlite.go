package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/mikrocloud/mikrocloud/internal/domain/auth"
	"github.com/mikrocloud/mikrocloud/internal/domain/users"
	"github.com/stephenafamo/bob/dialect/sqlite"
	"github.com/stephenafamo/bob/dialect/sqlite/dm"
	"github.com/stephenafamo/bob/dialect/sqlite/im"
	"github.com/stephenafamo/bob/dialect/sqlite/sm"
	"github.com/stephenafamo/bob/dialect/sqlite/um"
)

// SQLiteSessionRepository implements SessionRepository for SQLite
type SQLiteSessionRepository struct {
	db *sql.DB
}

// SQLiteAuthRepository implements AuthRepository for SQLite
type SQLiteAuthRepository struct {
	db     *sql.DB
	userID users.UserID // for now, we'll use the existing users table
}

// NewSQLiteSessionRepository creates a new session repository
func NewSQLiteSessionRepository(db *sql.DB) SessionRepository {
	return &SQLiteSessionRepository{db: db}
}

// NewSQLiteAuthRepository creates a new auth repository
func NewSQLiteAuthRepository(db *sql.DB) AuthRepository {
	return &SQLiteAuthRepository{db: db}
}

// Session Repository Implementation

func (r *SQLiteSessionRepository) SaveSession(ctx context.Context, session *auth.Session) error {
	query := sqlite.Insert(
		im.Into("sessions", "id", "user_id", "token", "expires_at", "created_at", "is_revoked"),
		im.Values(
			sqlite.Arg(session.ID().String()),
			sqlite.Arg(session.UserID().String()),
			sqlite.Arg(session.Token()),
			sqlite.Arg(session.ExpiresAt().Format(time.RFC3339)),
			sqlite.Arg(session.CreatedAt().Format(time.RFC3339)),
			sqlite.Arg(session.IsRevoked()),
		),
		im.OnConflict("id").DoUpdate(
			im.SetCol("token").ToArg(session.Token()),
			im.SetCol("expires_at").ToArg(session.ExpiresAt().Format(time.RFC3339)),
			im.SetCol("is_revoked").ToArg(session.IsRevoked()),
		),
	)

	queryStr, args, err := query.Build(ctx)
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	_, err = r.db.ExecContext(ctx, queryStr, args...)
	if err != nil {
		return fmt.Errorf("failed to save session: %w", err)
	}

	return nil
}

func (r *SQLiteSessionRepository) GetSessionByToken(ctx context.Context, token string) (*auth.Session, error) {
	query := sqlite.Select(
		sm.Columns("id", "user_id", "token", "expires_at", "created_at", "is_revoked"),
		sm.From("sessions"),
		sm.Where(sqlite.Quote("token").EQ(sqlite.Arg(token))),
	)

	queryStr, args, err := query.Build(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	var row sessionRow
	err = r.db.QueryRowContext(ctx, queryStr, args...).Scan(
		&row.ID, &row.UserID, &row.Token, &row.ExpiresAt, &row.CreatedAt, &row.IsRevoked,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("session not found for token")
		}
		return nil, fmt.Errorf("failed to get session by token: %w", err)
	}

	return r.mapRowToSession(row)
}

func (r *SQLiteSessionRepository) GetSessionByID(ctx context.Context, sessionID auth.SessionID) (*auth.Session, error) {
	query := sqlite.Select(
		sm.Columns("id", "user_id", "token", "expires_at", "created_at", "is_revoked"),
		sm.From("sessions"),
		sm.Where(sqlite.Quote("id").EQ(sqlite.Arg(sessionID.String()))),
	)

	queryStr, args, err := query.Build(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	var row sessionRow
	err = r.db.QueryRowContext(ctx, queryStr, args...).Scan(
		&row.ID, &row.UserID, &row.Token, &row.ExpiresAt, &row.CreatedAt, &row.IsRevoked,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("session not found: %s", sessionID.String())
		}
		return nil, fmt.Errorf("failed to get session by ID: %w", err)
	}

	return r.mapRowToSession(row)
}

func (r *SQLiteSessionRepository) GetActiveSessionsByUserID(ctx context.Context, userID users.UserID) ([]*auth.Session, error) {
	query := sqlite.Select(
		sm.Columns("id", "user_id", "token", "expires_at", "created_at", "is_revoked"),
		sm.From("sessions"),
		sm.Where(
			sqlite.Quote("user_id").EQ(sqlite.Arg(userID.String())).
				And(sqlite.Quote("is_revoked").EQ(sqlite.Arg(false))).
				And(sqlite.Quote("expires_at").GT(sqlite.Arg(time.Now().Format(time.RFC3339)))),
		),
		sm.OrderBy("created_at").Desc(),
	)

	queryStr, args, err := query.Build(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	rows, err := r.db.QueryContext(ctx, queryStr, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query active sessions: %w", err)
	}
	defer rows.Close()

	var sessions []*auth.Session
	for rows.Next() {
		var row sessionRow
		err := rows.Scan(&row.ID, &row.UserID, &row.Token, &row.ExpiresAt, &row.CreatedAt, &row.IsRevoked)
		if err != nil {
			return nil, fmt.Errorf("failed to scan session row: %w", err)
		}

		session, err := r.mapRowToSession(row)
		if err != nil {
			return nil, fmt.Errorf("failed to map session: %w", err)
		}

		sessions = append(sessions, session)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over session rows: %w", err)
	}

	return sessions, nil
}

func (r *SQLiteSessionRepository) RevokeSession(ctx context.Context, sessionID auth.SessionID) error {
	query := sqlite.Update(
		um.Table("sessions"),
		um.SetCol("is_revoked").ToArg(true),
		um.Where(sqlite.Quote("id").EQ(sqlite.Arg(sessionID.String()))),
	)

	queryStr, args, err := query.Build(ctx)
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	result, err := r.db.ExecContext(ctx, queryStr, args...)
	if err != nil {
		return fmt.Errorf("failed to revoke session: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("session not found: %s", sessionID.String())
	}

	return nil
}

func (r *SQLiteSessionRepository) RevokeAllUserSessions(ctx context.Context, userID users.UserID) error {
	query := sqlite.Update(
		um.Table("sessions"),
		um.SetCol("is_revoked").ToArg(true),
		um.Where(sqlite.Quote("user_id").EQ(sqlite.Arg(userID.String()))),
	)

	queryStr, args, err := query.Build(ctx)
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	_, err = r.db.ExecContext(ctx, queryStr, args...)
	if err != nil {
		return fmt.Errorf("failed to revoke all user sessions: %w", err)
	}

	return nil
}

func (r *SQLiteSessionRepository) DeleteExpiredSessions(ctx context.Context) error {
	query := sqlite.Delete(
		dm.From("sessions"),
		dm.Where(sqlite.Quote("expires_at").LT(sqlite.Arg(time.Now().Format(time.RFC3339)))),
	)

	queryStr, args, err := query.Build(ctx)
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	_, err = r.db.ExecContext(ctx, queryStr, args...)
	if err != nil {
		return fmt.Errorf("failed to delete expired sessions: %w", err)
	}

	return nil
}

// Refresh Token Methods

func (r *SQLiteSessionRepository) SaveRefreshToken(ctx context.Context, refreshToken *auth.RefreshToken) error {
	query := sqlite.Insert(
		im.Into("refresh_tokens", "id", "user_id", "session_id", "token", "expires_at", "created_at", "is_used"),
		im.Values(
			sqlite.Arg(refreshToken.ID().String()),
			sqlite.Arg(refreshToken.UserID().String()),
			sqlite.Arg(refreshToken.SessionID().String()),
			sqlite.Arg(refreshToken.Token()),
			sqlite.Arg(refreshToken.ExpiresAt().Format(time.RFC3339)),
			sqlite.Arg(refreshToken.CreatedAt().Format(time.RFC3339)),
			sqlite.Arg(refreshToken.IsUsed()),
		),
		im.OnConflict("id").DoUpdate(
			im.SetCol("token").ToArg(refreshToken.Token()),
			im.SetCol("expires_at").ToArg(refreshToken.ExpiresAt().Format(time.RFC3339)),
			im.SetCol("is_used").ToArg(refreshToken.IsUsed()),
		),
	)

	queryStr, args, err := query.Build(ctx)
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	_, err = r.db.ExecContext(ctx, queryStr, args...)
	if err != nil {
		return fmt.Errorf("failed to save refresh token: %w", err)
	}

	return nil
}

func (r *SQLiteSessionRepository) GetRefreshTokenByToken(ctx context.Context, token string) (*auth.RefreshToken, error) {
	query := sqlite.Select(
		sm.Columns("id", "user_id", "session_id", "token", "expires_at", "created_at", "is_used"),
		sm.From("refresh_tokens"),
		sm.Where(sqlite.Quote("token").EQ(sqlite.Arg(token))),
	)

	queryStr, args, err := query.Build(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	var row refreshTokenRow
	err = r.db.QueryRowContext(ctx, queryStr, args...).Scan(
		&row.ID, &row.UserID, &row.SessionID, &row.Token, &row.ExpiresAt, &row.CreatedAt, &row.IsUsed,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("refresh token not found")
		}
		return nil, fmt.Errorf("failed to get refresh token by token: %w", err)
	}

	return r.mapRowToRefreshToken(row)
}

func (r *SQLiteSessionRepository) GetRefreshTokenByID(ctx context.Context, tokenID auth.RefreshTokenID) (*auth.RefreshToken, error) {
	query := sqlite.Select(
		sm.Columns("id", "user_id", "session_id", "token", "expires_at", "created_at", "is_used"),
		sm.From("refresh_tokens"),
		sm.Where(sqlite.Quote("id").EQ(sqlite.Arg(tokenID.String()))),
	)

	queryStr, args, err := query.Build(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	var row refreshTokenRow
	err = r.db.QueryRowContext(ctx, queryStr, args...).Scan(
		&row.ID, &row.UserID, &row.SessionID, &row.Token, &row.ExpiresAt, &row.CreatedAt, &row.IsUsed,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("refresh token not found: %s", tokenID.String())
		}
		return nil, fmt.Errorf("failed to get refresh token by ID: %w", err)
	}

	return r.mapRowToRefreshToken(row)
}

func (r *SQLiteSessionRepository) MarkRefreshTokenAsUsed(ctx context.Context, tokenID auth.RefreshTokenID) error {
	query := sqlite.Update(
		um.Table("refresh_tokens"),
		um.SetCol("is_used").ToArg(true),
		um.Where(sqlite.Quote("id").EQ(sqlite.Arg(tokenID.String()))),
	)

	queryStr, args, err := query.Build(ctx)
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	result, err := r.db.ExecContext(ctx, queryStr, args...)
	if err != nil {
		return fmt.Errorf("failed to mark refresh token as used: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("refresh token not found: %s", tokenID.String())
	}

	return nil
}

func (r *SQLiteSessionRepository) DeleteRefreshToken(ctx context.Context, tokenID auth.RefreshTokenID) error {
	query := sqlite.Delete(
		dm.From("refresh_tokens"),
		dm.Where(sqlite.Quote("id").EQ(sqlite.Arg(tokenID.String()))),
	)

	queryStr, args, err := query.Build(ctx)
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	result, err := r.db.ExecContext(ctx, queryStr, args...)
	if err != nil {
		return fmt.Errorf("failed to delete refresh token: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("refresh token not found: %s", tokenID.String())
	}

	return nil
}

func (r *SQLiteSessionRepository) DeleteExpiredRefreshTokens(ctx context.Context) error {
	query := sqlite.Delete(
		dm.From("refresh_tokens"),
		dm.Where(sqlite.Quote("expires_at").LT(sqlite.Arg(time.Now().Format(time.RFC3339)))),
	)

	queryStr, args, err := query.Build(ctx)
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	_, err = r.db.ExecContext(ctx, queryStr, args...)
	if err != nil {
		return fmt.Errorf("failed to delete expired refresh tokens: %w", err)
	}

	return nil
}

func (r *SQLiteSessionRepository) DeleteRefreshTokensBySessionID(ctx context.Context, sessionID auth.SessionID) error {
	query := sqlite.Delete(
		dm.From("refresh_tokens"),
		dm.Where(sqlite.Quote("session_id").EQ(sqlite.Arg(sessionID.String()))),
	)

	queryStr, args, err := query.Build(ctx)
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	_, err = r.db.ExecContext(ctx, queryStr, args...)
	if err != nil {
		return fmt.Errorf("failed to delete refresh tokens by session ID: %w", err)
	}

	return nil
}

// Auth Repository Implementation

func (r *SQLiteAuthRepository) GetUserByEmail(ctx context.Context, email users.Email) (*users.User, error) {
	query := sqlite.Select(
		sm.Columns("id", "email", "password_hash", "username", "name", "status", "email_verified_at", "last_login_at", "timezone", "created_at", "updated_at"),
		sm.From("users"),
		sm.Where(sqlite.Quote("email").EQ(sqlite.Arg(email.String()))),
	)

	queryStr, args, err := query.Build(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	var row userRow
	err = r.db.QueryRowContext(ctx, queryStr, args...).Scan(
		&row.ID, &row.Email, &row.PasswordHash, &row.Username, &row.Name, &row.Status,
		&row.EmailVerifiedAt, &row.LastLoginAt, &row.Timezone, &row.CreatedAt, &row.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found: %s", email.String())
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	return r.mapRowToUser(row)
}

func (r *SQLiteAuthRepository) GetUserByID(ctx context.Context, userID users.UserID) (*users.User, error) {
	query := sqlite.Select(
		sm.Columns("id", "email", "password_hash", "username", "name", "status", "email_verified_at", "last_login_at", "timezone", "created_at", "updated_at"),
		sm.From("users"),
		sm.Where(sqlite.Quote("id").EQ(sqlite.Arg(userID.String()))),
	)

	queryStr, args, err := query.Build(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	var row userRow
	err = r.db.QueryRowContext(ctx, queryStr, args...).Scan(
		&row.ID, &row.Email, &row.PasswordHash, &row.Username, &row.Name, &row.Status,
		&row.EmailVerifiedAt, &row.LastLoginAt, &row.Timezone, &row.CreatedAt, &row.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found: %s", userID.String())
		}
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}

	return r.mapRowToUser(row)
}

func (r *SQLiteAuthRepository) CreateUser(ctx context.Context, user *users.User) error {
	query := sqlite.Insert(
		im.Into("users",
			"id", "email", "password_hash", "username", "name", "status",
			"email_verified_at", "last_login_at", "timezone", "created_at", "updated_at",
		),
		im.Values(
			sqlite.Arg(user.ID().String()),
			sqlite.Arg(user.Email().String()),
			sqlite.Arg(user.PasswordHash()),
			sqlite.Arg(user.Username()),
			sqlite.Arg(user.Name()),
			sqlite.Arg(string(user.Status())),
			sqlite.Arg(user.EmailVerifiedAt()),
			sqlite.Arg(user.LastLoginAt()),
			sqlite.Arg(user.Timezone()),
			sqlite.Arg(user.CreatedAt().Format(time.RFC3339)),
			sqlite.Arg(user.UpdatedAt().Format(time.RFC3339)),
		),
	)

	queryStr, args, err := query.Build(ctx)
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	_, err = r.db.ExecContext(ctx, queryStr, args...)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

func (r *SQLiteAuthRepository) UpdateUserPassword(ctx context.Context, userID users.UserID, passwordHash string) error {
	query := sqlite.Update(
		um.Table("users"),
		um.SetCol("password_hash").ToArg(passwordHash),
		um.SetCol("updated_at").ToArg(time.Now().Format(time.RFC3339)),
		um.Where(sqlite.Quote("id").EQ(sqlite.Arg(userID.String()))),
	)

	queryStr, args, err := query.Build(ctx)
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	result, err := r.db.ExecContext(ctx, queryStr, args...)
	if err != nil {
		return fmt.Errorf("failed to update user password: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user not found: %s", userID.String())
	}

	return nil
}

func (r *SQLiteAuthRepository) UpdateLastLogin(ctx context.Context, userID users.UserID) error {
	query := sqlite.Update(
		um.Table("users"),
		um.SetCol("last_login_at").ToArg(time.Now().Format(time.RFC3339)),
		um.SetCol("updated_at").ToArg(time.Now().Format(time.RFC3339)),
		um.Where(sqlite.Quote("id").EQ(sqlite.Arg(userID.String()))),
	)

	queryStr, args, err := query.Build(ctx)
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	result, err := r.db.ExecContext(ctx, queryStr, args...)
	if err != nil {
		return fmt.Errorf("failed to update last login: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user not found: %s", userID.String())
	}

	return nil
}

func (r *SQLiteAuthRepository) UserExistsByEmail(ctx context.Context, email users.Email) (bool, error) {
	query := sqlite.Select(
		sm.Columns("COUNT(*)"),
		sm.From("users"),
		sm.Where(sqlite.Quote("email").EQ(sqlite.Arg(email.String()))),
	)

	queryStr, args, err := query.Build(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to build query: %w", err)
	}

	var count int
	err = r.db.QueryRowContext(ctx, queryStr, args...).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check if user exists: %w", err)
	}

	return count > 0, nil
}

func (r *SQLiteAuthRepository) HasAnyUsers(ctx context.Context) (bool, error) {
	query := sqlite.Select(
		sm.Columns("COUNT(*)"),
		sm.From("users"),
		sm.Where(sqlite.Quote("status").EQ(sqlite.Arg("active"))),
	)

	queryStr, args, err := query.Build(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to build query: %w", err)
	}

	var count int
	err = r.db.QueryRowContext(ctx, queryStr, args...).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check if any users exist: %w", err)
	}

	return count > 0, nil
}

// Helper types and functions

type sessionRow struct {
	ID        string
	UserID    string
	Token     string
	ExpiresAt string
	CreatedAt string
	IsRevoked bool
}

type refreshTokenRow struct {
	ID        string
	UserID    string
	SessionID string
	Token     string
	ExpiresAt string
	CreatedAt string
	IsUsed    bool
}

type userRow struct {
	ID              string
	Email           string
	PasswordHash    string
	Username        sql.NullString
	Name            string
	Status          string
	EmailVerifiedAt sql.NullTime
	LastLoginAt     sql.NullTime
	Timezone        string
	CreatedAt       string
	UpdatedAt       string
}

func (r *SQLiteSessionRepository) mapRowToSession(row sessionRow) (*auth.Session, error) {
	sessionID, err := auth.SessionIDFromString(row.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid session ID: %w", err)
	}

	userID, err := users.UserIDFromString(row.UserID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	expiresAt, err := time.Parse(time.RFC3339, row.ExpiresAt)
	if err != nil {
		return nil, fmt.Errorf("invalid expires_at timestamp: %w", err)
	}

	createdAt, err := time.Parse(time.RFC3339, row.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("invalid created_at timestamp: %w", err)
	}

	return auth.ReconstructSession(
		sessionID, userID, row.Token, expiresAt, createdAt, row.IsRevoked,
	), nil
}

func (r *SQLiteSessionRepository) mapRowToRefreshToken(row refreshTokenRow) (*auth.RefreshToken, error) {
	tokenID, err := auth.RefreshTokenIDFromString(row.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token ID: %w", err)
	}

	userID, err := users.UserIDFromString(row.UserID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	sessionID, err := auth.SessionIDFromString(row.SessionID)
	if err != nil {
		return nil, fmt.Errorf("invalid session ID: %w", err)
	}

	expiresAt, err := time.Parse(time.RFC3339, row.ExpiresAt)
	if err != nil {
		return nil, fmt.Errorf("invalid expires_at timestamp: %w", err)
	}

	createdAt, err := time.Parse(time.RFC3339, row.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("invalid created_at timestamp: %w", err)
	}

	return auth.ReconstructRefreshToken(
		tokenID, userID, sessionID, row.Token, expiresAt, createdAt, row.IsUsed,
	), nil
}

func (r *SQLiteAuthRepository) mapRowToUser(row userRow) (*users.User, error) {
	userID, err := users.UserIDFromString(row.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	email, err := users.NewEmail(row.Email)
	if err != nil {
		return nil, fmt.Errorf("invalid email: %w", err)
	}

	status := users.UserStatus(row.Status)

	createdAt, err := time.Parse(time.RFC3339, row.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("invalid created_at timestamp: %w", err)
	}

	updatedAt, err := time.Parse(time.RFC3339, row.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("invalid updated_at timestamp: %w", err)
	}

	var emailVerifiedAt *time.Time
	if row.EmailVerifiedAt.Valid {
		emailVerifiedAt = &row.EmailVerifiedAt.Time
	}

	var lastLoginAt *time.Time
	if row.LastLoginAt.Valid {
		lastLoginAt = &row.LastLoginAt.Time
	}

	var username *users.Username
	if row.Username.Valid && row.Username.String != "" {
		username, err = users.NewUsername(row.Username.String)
		if err != nil {
			return nil, fmt.Errorf("invalid username: %w", err)
		}
	}

	return users.ReconstructUser(
		userID, email, row.PasswordHash, row.Name, username, status,
		emailVerifiedAt, lastLoginAt, row.Timezone, createdAt, updatedAt,
	), nil
}
