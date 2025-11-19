package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/mikrocloud/mikrocloud/internal/domain/users"
)

type Repository interface {
	FindAll(ctx context.Context) ([]*users.Organization, error)
	FindByID(ctx context.Context, id users.OrganizationID) (*users.Organization, error)
	FindByUserID(ctx context.Context, userID users.UserID) ([]*users.Organization, error)
	Save(ctx context.Context, org *users.Organization) error
	Delete(ctx context.Context, id users.OrganizationID) error
}

type SQLiteOrganizationRepository struct {
	db *sql.DB
}

func NewSQLiteOrganizationRepository(db *sql.DB) Repository {
	return &SQLiteOrganizationRepository{db: db}
}

func (r *SQLiteOrganizationRepository) FindAll(ctx context.Context) ([]*users.Organization, error) {
	query := `
		SELECT id, name, slug, description, owner_id, billing_email, plan, status, created_at, updated_at
		FROM organizations
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query organizations: %w", err)
	}
	defer rows.Close()

	var organizations []*users.Organization
	for rows.Next() {
		var row organizationRow
		err := rows.Scan(
			&row.ID,
			&row.Name,
			&row.Slug,
			&row.Description,
			&row.OwnerID,
			&row.BillingEmail,
			&row.Plan,
			&row.Status,
			&row.CreatedAt,
			&row.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan organization row: %w", err)
		}

		org, err := r.mapRowToOrganization(row)
		if err != nil {
			return nil, fmt.Errorf("failed to map organization: %w", err)
		}

		organizations = append(organizations, org)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating organization rows: %w", err)
	}

	return organizations, nil
}

func (r *SQLiteOrganizationRepository) FindByID(ctx context.Context, id users.OrganizationID) (*users.Organization, error) {
	query := `
		SELECT id, name, slug, description, owner_id, billing_email, plan, status, created_at, updated_at
		FROM organizations
		WHERE id = ?
	`

	var row organizationRow
	err := r.db.QueryRowContext(ctx, query, id.String()).Scan(
		&row.ID,
		&row.Name,
		&row.Slug,
		&row.Description,
		&row.OwnerID,
		&row.BillingEmail,
		&row.Plan,
		&row.Status,
		&row.CreatedAt,
		&row.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("organization not found: %s", id.String())
		}
		return nil, fmt.Errorf("failed to find organization by ID: %w", err)
	}

	return r.mapRowToOrganization(row)
}

func (r *SQLiteOrganizationRepository) FindByUserID(ctx context.Context, userID users.UserID) ([]*users.Organization, error) {
	query := `
		SELECT o.id, o.name, o.slug, o.description, o.owner_id, o.billing_email, o.plan, o.status, o.created_at, o.updated_at
		FROM organizations o
		LEFT JOIN organization_members om ON o.id = om.organization_id
		WHERE o.owner_id = ? OR om.user_id = ?
		GROUP BY o.id
		ORDER BY o.created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID.String(), userID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to query organizations by user: %w", err)
	}
	defer rows.Close()

	var organizations []*users.Organization
	for rows.Next() {
		var row organizationRow
		err := rows.Scan(
			&row.ID,
			&row.Name,
			&row.Slug,
			&row.Description,
			&row.OwnerID,
			&row.BillingEmail,
			&row.Plan,
			&row.Status,
			&row.CreatedAt,
			&row.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan organization row: %w", err)
		}

		org, err := r.mapRowToOrganization(row)
		if err != nil {
			return nil, fmt.Errorf("failed to map organization: %w", err)
		}

		organizations = append(organizations, org)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating organization rows: %w", err)
	}

	return organizations, nil
}

func (r *SQLiteOrganizationRepository) Save(ctx context.Context, org *users.Organization) error {
	query := `
		INSERT OR REPLACE INTO organizations (id, name, slug, description, owner_id, billing_email, plan, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.ExecContext(ctx, query,
		org.ID().String(),
		org.Name(),
		org.Slug(),
		org.Description(),
		org.OwnerID().String(),
		org.BillingEmail(),
		org.Plan(),
		org.Status(),
		org.CreatedAt().Format(time.RFC3339),
		org.UpdatedAt().Format(time.RFC3339),
	)

	if err != nil {
		return fmt.Errorf("failed to save organization: %w", err)
	}

	return nil
}

func (r *SQLiteOrganizationRepository) Delete(ctx context.Context, id users.OrganizationID) error {
	query := `DELETE FROM organizations WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query, id.String())
	if err != nil {
		return fmt.Errorf("failed to delete organization: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("organization not found: %s", id.String())
	}

	return nil
}

type organizationRow struct {
	ID           string
	Name         string
	Slug         string
	Description  string
	OwnerID      string
	BillingEmail string
	Plan         string
	Status       string
	CreatedAt    string
	UpdatedAt    string
}

func (r *SQLiteOrganizationRepository) mapRowToOrganization(row organizationRow) (*users.Organization, error) {
	orgID, err := users.OrganizationIDFromString(row.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid organization ID: %w", err)
	}

	ownerID, err := users.UserIDFromString(row.OwnerID)
	if err != nil {
		return nil, fmt.Errorf("invalid owner ID: %w", err)
	}

	createdAt, err := time.Parse(time.RFC3339, row.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("invalid created_at timestamp: %w", err)
	}

	updatedAt, err := time.Parse(time.RFC3339, row.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("invalid updated_at timestamp: %w", err)
	}

	return users.ReconstructOrganization(
		orgID,
		row.Name,
		row.Slug,
		row.Description,
		ownerID,
		row.BillingEmail,
		users.OrganizationPlan(row.Plan),
		users.OrganizationStatus(row.Status),
		createdAt,
		updatedAt,
	), nil
}
