package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/mikrocloud/mikrocloud/internal/domain/git"
)

type SQLiteGitRepository struct {
	db *sql.DB
}

func NewSQLiteGitRepository(db *sql.DB) GitRepository {
	return &SQLiteGitRepository{db: db}
}

func (r *SQLiteGitRepository) Create(ctx context.Context, source *git.GitSource) error {
	query := `
		INSERT INTO git_sources (id, org_id, user_id, provider, name, access_token, refresh_token, token_expires_at, custom_url, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	var tokenExpiresAt interface{}
	if source.TokenExpiresAt != nil {
		tokenExpiresAt = source.TokenExpiresAt.Format(time.RFC3339)
	}

	var customURL interface{}
	if source.CustomURL != nil {
		customURL = *source.CustomURL
	}

	_, err := r.db.ExecContext(ctx, query,
		source.ID,
		source.OrgID,
		source.UserID,
		source.Provider,
		source.Name,
		source.AccessToken,
		source.RefreshToken,
		tokenExpiresAt,
		customURL,
		source.CreatedAt.Format(time.RFC3339),
		source.UpdatedAt.Format(time.RFC3339),
	)

	if err != nil {
		return fmt.Errorf("failed to create git source: %w", err)
	}

	return nil
}

func (r *SQLiteGitRepository) GetByID(ctx context.Context, id string) (*git.GitSource, error) {
	query := `
		SELECT id, org_id, user_id, provider, name, access_token, refresh_token, token_expires_at, custom_url, created_at, updated_at
		FROM git_sources
		WHERE id = ?
	`

	var row gitSourceRow
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&row.ID,
		&row.OrgID,
		&row.UserID,
		&row.Provider,
		&row.Name,
		&row.AccessToken,
		&row.RefreshToken,
		&row.TokenExpiresAt,
		&row.CustomURL,
		&row.CreatedAt,
		&row.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("git source not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get git source: %w", err)
	}

	return r.mapRowToGitSource(row)
}

func (r *SQLiteGitRepository) GetByUserID(ctx context.Context, userID string) ([]*git.GitSource, error) {
	query := `
		SELECT id, org_id, user_id, provider, name, access_token, refresh_token, token_expires_at, custom_url, created_at, updated_at
		FROM git_sources
		WHERE user_id = ?
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query git sources by user: %w", err)
	}
	defer rows.Close()

	var sources []*git.GitSource
	for rows.Next() {
		var row gitSourceRow
		err := rows.Scan(
			&row.ID,
			&row.OrgID,
			&row.UserID,
			&row.Provider,
			&row.Name,
			&row.AccessToken,
			&row.RefreshToken,
			&row.TokenExpiresAt,
			&row.CustomURL,
			&row.CreatedAt,
			&row.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan git source row: %w", err)
		}

		source, err := r.mapRowToGitSource(row)
		if err != nil {
			return nil, fmt.Errorf("failed to map git source: %w", err)
		}

		sources = append(sources, source)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating git source rows: %w", err)
	}

	return sources, nil
}

func (r *SQLiteGitRepository) GetByOrgID(ctx context.Context, orgID string) ([]*git.GitSource, error) {
	query := `
		SELECT id, org_id, user_id, provider, name, access_token, refresh_token, token_expires_at, custom_url, created_at, updated_at
		FROM git_sources
		WHERE org_id = ?
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to query git sources by org: %w", err)
	}
	defer rows.Close()

	var sources []*git.GitSource
	for rows.Next() {
		var row gitSourceRow
		err := rows.Scan(
			&row.ID,
			&row.OrgID,
			&row.UserID,
			&row.Provider,
			&row.Name,
			&row.AccessToken,
			&row.RefreshToken,
			&row.TokenExpiresAt,
			&row.CustomURL,
			&row.CreatedAt,
			&row.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan git source row: %w", err)
		}

		source, err := r.mapRowToGitSource(row)
		if err != nil {
			return nil, fmt.Errorf("failed to map git source: %w", err)
		}

		sources = append(sources, source)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating git source rows: %w", err)
	}

	return sources, nil
}

func (r *SQLiteGitRepository) Update(ctx context.Context, id string, source *git.GitSource) error {
	query := `
		UPDATE git_sources
		SET name = ?, access_token = ?, refresh_token = ?, token_expires_at = ?, updated_at = ?
		WHERE id = ?
	`

	var tokenExpiresAt interface{}
	if source.TokenExpiresAt != nil {
		tokenExpiresAt = source.TokenExpiresAt.Format(time.RFC3339)
	}

	result, err := r.db.ExecContext(ctx, query,
		source.Name,
		source.AccessToken,
		source.RefreshToken,
		tokenExpiresAt,
		source.UpdatedAt.Format(time.RFC3339),
		id,
	)

	if err != nil {
		return fmt.Errorf("failed to update git source: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("git source not found: %s", id)
	}

	return nil
}

func (r *SQLiteGitRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM git_sources WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete git source: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("git source not found: %s", id)
	}

	return nil
}

type gitSourceRow struct {
	ID             string
	OrgID          string
	UserID         string
	Provider       string
	Name           string
	AccessToken    string
	RefreshToken   sql.NullString
	TokenExpiresAt sql.NullString
	CustomURL      sql.NullString
	CreatedAt      string
	UpdatedAt      string
}

func (r *SQLiteGitRepository) mapRowToGitSource(row gitSourceRow) (*git.GitSource, error) {
	createdAt, err := time.Parse(time.RFC3339, row.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("invalid created_at timestamp: %w", err)
	}

	updatedAt, err := time.Parse(time.RFC3339, row.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("invalid updated_at timestamp: %w", err)
	}

	var tokenExpiresAt *time.Time
	if row.TokenExpiresAt.Valid {
		t, err := time.Parse(time.RFC3339, row.TokenExpiresAt.String)
		if err != nil {
			return nil, fmt.Errorf("invalid token_expires_at timestamp: %w", err)
		}
		tokenExpiresAt = &t
	}

	var customURL *string
	if row.CustomURL.Valid {
		customURL = &row.CustomURL.String
	}

	return &git.GitSource{
		ID:             row.ID,
		OrgID:          row.OrgID,
		UserID:         row.UserID,
		Provider:       git.GitProvider(row.Provider),
		Name:           row.Name,
		AccessToken:    row.AccessToken,
		RefreshToken:   row.RefreshToken.String,
		TokenExpiresAt: tokenExpiresAt,
		CustomURL:      customURL,
		CreatedAt:      createdAt,
		UpdatedAt:      updatedAt,
	}, nil
}
