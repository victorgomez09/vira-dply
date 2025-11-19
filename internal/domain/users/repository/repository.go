package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/mikrocloud/mikrocloud/internal/domain/users"
	"github.com/stephenafamo/bob/dialect/sqlite"
	"github.com/stephenafamo/bob/dialect/sqlite/dm"
	"github.com/stephenafamo/bob/dialect/sqlite/im"
	"github.com/stephenafamo/bob/dialect/sqlite/sm"
)

type Repository interface {
	Save(ctx context.Context, user *users.User) error
	SaveOrganization(ctx context.Context, org *users.Organization) error
	AddOrganizationMember(ctx context.Context, orgID users.OrganizationID, userID users.UserID, role string, invitedBy *users.UserID) error
	AddUserRole(ctx context.Context, userID users.UserID, roleID string, grantedBy *users.UserID) error
	FindRoleByName(ctx context.Context, roleName string) (string, error)
	FindByID(ctx context.Context, id users.UserID) (*users.User, error)
	FindByEmail(ctx context.Context, email users.Email) (*users.User, error)
	FindByUsername(ctx context.Context, username string) (*users.User, error)
	FindOrganizationByID(ctx context.Context, id users.OrganizationID) (*users.Organization, error)
	FindOrganizationBySlug(ctx context.Context, slug string) (*users.Organization, error)
	FindOrganizationsByUser(ctx context.Context, userID users.UserID) ([]*users.Organization, error)
	FindAll(ctx context.Context) ([]*users.User, error)
	Delete(ctx context.Context, id users.UserID) error
	DeleteOrganization(ctx context.Context, id users.OrganizationID) error
	Exists(ctx context.Context, email users.Email) (bool, error)
	OrganizationExists(ctx context.Context, slug string) (bool, error)
}

type SQLiteUserRepository struct {
	db *sql.DB
}

func NewSQLiteUserRepository(db *sql.DB) *SQLiteUserRepository {
	return &SQLiteUserRepository{db: db}
}

func (r *SQLiteUserRepository) Save(ctx context.Context, user *users.User) error {
	query := sqlite.Insert(
		im.Into("users"),
		im.Values(
			sqlite.Arg(user.ID().String()),
			sqlite.Arg(user.Email().String()),
			sqlite.Arg(user.PasswordHash()),
			sqlite.Arg(user.Username().String()),
			sqlite.Arg(string(user.Status())),
			sqlite.Arg(user.EmailVerifiedAt()),
			sqlite.Arg(user.LastLoginAt()),
			sqlite.Arg(user.Timezone()),
			sqlite.Arg(user.CreatedAt().Format(time.RFC3339)),
			sqlite.Arg(user.UpdatedAt().Format(time.RFC3339)),
		),
		im.OnConflict("id").DoUpdate(
			im.SetCol("email").ToArg(user.Email().String()),
			im.SetCol("password_hash").ToArg(user.PasswordHash()),
			im.SetCol("username").ToArg(user.Username().String()),
			im.SetCol("status").ToArg(string(user.Status())),
			im.SetCol("email_verified_at").ToArg(user.EmailVerifiedAt()),
			im.SetCol("last_login_at").ToArg(user.LastLoginAt()),
			im.SetCol("timezone").ToArg(user.Timezone()),
			im.SetCol("updated_at").ToArg(user.UpdatedAt().Format(time.RFC3339)),
		),
	)

	queryStr, args, err := query.Build(ctx)
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	_, err = r.db.ExecContext(ctx, queryStr, args...)
	if err != nil {
		return fmt.Errorf("failed to save user: %w", err)
	}

	return nil
}

func (r *SQLiteUserRepository) SaveOrganization(ctx context.Context, org *users.Organization) error {
	query := sqlite.Insert(
		im.Into("organizations"),
		im.Values(
			sqlite.Arg(org.ID().String()),
			sqlite.Arg(org.Name()),
			sqlite.Arg(org.Slug()),
			sqlite.Arg(org.Description()),
			sqlite.Arg(org.OwnerID().String()),
			sqlite.Arg(org.BillingEmail()),
			sqlite.Arg(string(org.Plan())),
			sqlite.Arg(string(org.Status())),
			sqlite.Arg(org.CreatedAt().Format(time.RFC3339)),
			sqlite.Arg(org.UpdatedAt().Format(time.RFC3339)),
		),
		im.OnConflict("id").DoUpdate(
			im.SetCol("name").ToArg(org.Name()),
			im.SetCol("slug").ToArg(org.Slug()),
			im.SetCol("description").ToArg(org.Description()),
			im.SetCol("billing_email").ToArg(org.BillingEmail()),
			im.SetCol("plan").ToArg(string(org.Plan())),
			im.SetCol("status").ToArg(string(org.Status())),
			im.SetCol("updated_at").ToArg(org.UpdatedAt().Format(time.RFC3339)),
		),
	)

	queryStr, args, err := query.Build(ctx)
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	_, err = r.db.ExecContext(ctx, queryStr, args...)
	if err != nil {
		return fmt.Errorf("failed to save organization: %w", err)
	}

	return nil
}

func (r *SQLiteUserRepository) FindByID(ctx context.Context, id users.UserID) (*users.User, error) {
	query := sqlite.Select(
		sm.Columns("id", "email", "password_hash", "username", "status", "email_verified_at", "last_login_at", "timezone", "created_at", "updated_at"),
		sm.From("users"),
		sm.Where(sqlite.Quote("id").EQ(sqlite.Arg(id.String()))),
	)

	queryStr, args, err := query.Build(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	var row userRow
	err = r.db.QueryRowContext(ctx, queryStr, args...).Scan(
		&row.ID, &row.Email, &row.PasswordHash, &row.Username, &row.Status,
		&row.EmailVerifiedAt, &row.LastLoginAt, &row.Timezone, &row.CreatedAt, &row.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found: %s", id.String())
		}
		return nil, fmt.Errorf("failed to find user by ID: %w", err)
	}

	return r.mapRowToUser(row)
}

func (r *SQLiteUserRepository) FindByEmail(ctx context.Context, email users.Email) (*users.User, error) {
	query := sqlite.Select(
		sm.Columns("id", "email", "password_hash", "username", "status", "email_verified_at", "last_login_at", "timezone", "created_at", "updated_at"),
		sm.From("users"),
		sm.Where(sqlite.Quote("email").EQ(sqlite.Arg(email.String()))),
	)

	queryStr, args, err := query.Build(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	var row userRow
	err = r.db.QueryRowContext(ctx, queryStr, args...).Scan(
		&row.ID, &row.Email, &row.PasswordHash, &row.Username, &row.Status,
		&row.EmailVerifiedAt, &row.LastLoginAt, &row.Timezone, &row.CreatedAt, &row.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found: %s", email.String())
		}
		return nil, fmt.Errorf("failed to find user by email: %w", err)
	}

	return r.mapRowToUser(row)
}

func (r *SQLiteUserRepository) FindByUsername(ctx context.Context, username string) (*users.User, error) {
	query := sqlite.Select(
		sm.Columns("id", "email", "password_hash", "username", "status", "email_verified_at", "last_login_at", "timezone", "created_at", "updated_at"),
		sm.From("users"),
		sm.Where(sqlite.Quote("username").EQ(sqlite.Arg(username))),
	)

	queryStr, args, err := query.Build(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	var row userRow
	err = r.db.QueryRowContext(ctx, queryStr, args...).Scan(
		&row.ID, &row.Email, &row.PasswordHash, &row.Username, &row.Status,
		&row.EmailVerifiedAt, &row.LastLoginAt, &row.Timezone, &row.CreatedAt, &row.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found: %s", username)
		}
		return nil, fmt.Errorf("failed to find user by username: %w", err)
	}

	return r.mapRowToUser(row)
}

func (r *SQLiteUserRepository) FindOrganizationByID(ctx context.Context, id users.OrganizationID) (*users.Organization, error) {
	query := sqlite.Select(
		sm.Columns("id", "name", "slug", "description", "owner_id", "billing_email", "plan", "status", "created_at", "updated_at"),
		sm.From("organizations"),
		sm.Where(sqlite.Quote("id").EQ(sqlite.Arg(id.String()))),
	)

	queryStr, args, err := query.Build(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	var row organizationRow
	err = r.db.QueryRowContext(ctx, queryStr, args...).Scan(
		&row.ID, &row.Name, &row.Slug, &row.Description, &row.OwnerID,
		&row.BillingEmail, &row.Plan, &row.Status, &row.CreatedAt, &row.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("organization not found: %s", id.String())
		}
		return nil, fmt.Errorf("failed to find organization by ID: %w", err)
	}

	return r.mapRowToOrganization(row)
}

func (r *SQLiteUserRepository) FindOrganizationBySlug(ctx context.Context, slug string) (*users.Organization, error) {
	query := sqlite.Select(
		sm.Columns("id", "name", "slug", "description", "owner_id", "billing_email", "plan", "status", "created_at", "updated_at"),
		sm.From("organizations"),
		sm.Where(sqlite.Quote("slug").EQ(sqlite.Arg(slug))),
	)

	queryStr, args, err := query.Build(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	var row organizationRow
	err = r.db.QueryRowContext(ctx, queryStr, args...).Scan(
		&row.ID, &row.Name, &row.Slug, &row.Description, &row.OwnerID,
		&row.BillingEmail, &row.Plan, &row.Status, &row.CreatedAt, &row.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("organization not found: %s", slug)
		}
		return nil, fmt.Errorf("failed to find organization by slug: %w", err)
	}

	return r.mapRowToOrganization(row)
}

func (r *SQLiteUserRepository) FindOrganizationsByUser(ctx context.Context, userID users.UserID) ([]*users.Organization, error) {
	// Use direct query since bob SQL builder had quoting issues
	queryStr := `
		SELECT o.id, o.name, o.slug, o.description, o.owner_id, o.billing_email, o.plan, o.status, o.created_at, o.updated_at
		FROM organizations o
		INNER JOIN organization_members om ON o.id = om.organization_id
		WHERE om.user_id = ?
		ORDER BY o.created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, queryStr, userID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to query organizations: %w", err)
	}
	defer rows.Close()

	var organizations []*users.Organization
	for rows.Next() {
		var row organizationRow
		err := rows.Scan(&row.ID, &row.Name, &row.Slug, &row.Description, &row.OwnerID,
			&row.BillingEmail, &row.Plan, &row.Status, &row.CreatedAt, &row.UpdatedAt)
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
		return nil, fmt.Errorf("error iterating over organization rows: %w", err)
	}

	return organizations, nil
}

func (r *SQLiteUserRepository) FindAll(ctx context.Context) ([]*users.User, error) {
	query := sqlite.Select(
		sm.Columns("id", "email", "password_hash", "username", "status", "email_verified_at", "last_login_at", "timezone", "created_at", "updated_at"),
		sm.From("users"),
		sm.OrderBy("created_at").Desc(),
	)

	queryStr, args, err := query.Build(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	rows, err := r.db.QueryContext(ctx, queryStr, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query all users: %w", err)
	}
	defer rows.Close()

	var userList []*users.User
	for rows.Next() {
		var row userRow
		err := rows.Scan(&row.ID, &row.Email, &row.PasswordHash, &row.Username, &row.Status,
			&row.EmailVerifiedAt, &row.LastLoginAt, &row.Timezone, &row.CreatedAt, &row.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user row: %w", err)
		}

		user, err := r.mapRowToUser(row)
		if err != nil {
			return nil, fmt.Errorf("failed to map user: %w", err)
		}

		userList = append(userList, user)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over user rows: %w", err)
	}

	return userList, nil
}

func (r *SQLiteUserRepository) Delete(ctx context.Context, id users.UserID) error {
	query := sqlite.Delete(
		dm.From("users"),
		dm.Where(sqlite.Quote("id").EQ(sqlite.Arg(id.String()))),
	)

	queryStr, args, err := query.Build(ctx)
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	result, err := r.db.ExecContext(ctx, queryStr, args...)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user not found: %s", id.String())
	}

	return nil
}

func (r *SQLiteUserRepository) DeleteOrganization(ctx context.Context, id users.OrganizationID) error {
	query := sqlite.Delete(
		dm.From("organizations"),
		dm.Where(sqlite.Quote("id").EQ(sqlite.Arg(id.String()))),
	)

	queryStr, args, err := query.Build(ctx)
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	result, err := r.db.ExecContext(ctx, queryStr, args...)
	if err != nil {
		return fmt.Errorf("failed to delete organization: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("organization not found: %s", id.String())
	}

	return nil
}

func (r *SQLiteUserRepository) Exists(ctx context.Context, email users.Email) (bool, error) {
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

func (r *SQLiteUserRepository) OrganizationExists(ctx context.Context, slug string) (bool, error) {
	query := sqlite.Select(
		sm.Columns("COUNT(*)"),
		sm.From("organizations"),
		sm.Where(sqlite.Quote("slug").EQ(sqlite.Arg(slug))),
	)

	queryStr, args, err := query.Build(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to build query: %w", err)
	}

	var count int
	err = r.db.QueryRowContext(ctx, queryStr, args...).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check if organization exists: %w", err)
	}

	return count > 0, nil
}

func (r *SQLiteUserRepository) AddOrganizationMember(ctx context.Context, orgID users.OrganizationID, userID users.UserID, role string, invitedBy *users.UserID) error {
	memberID := users.NewOrganizationID()

	var invitedByStr sql.NullString
	if invitedBy != nil {
		invitedByStr = sql.NullString{String: invitedBy.String(), Valid: true}
	}

	queryStr := `INSERT INTO organization_members (id, organization_id, user_id, role, invited_by, joined_at) VALUES (?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, queryStr, memberID.String(), orgID.String(), userID.String(), role, invitedByStr, time.Now().Format(time.RFC3339))
	if err != nil {
		return fmt.Errorf("failed to add organization member: %w", err)
	}

	return nil
}

func (r *SQLiteUserRepository) AddUserRole(ctx context.Context, userID users.UserID, roleID string, grantedBy *users.UserID) error {
	userRoleID := users.NewUserID()

	var grantedByStr sql.NullString
	if grantedBy != nil {
		grantedByStr = sql.NullString{String: grantedBy.String(), Valid: true}
	}

	queryStr := `INSERT INTO user_roles (id, user_id, role_id, granted_by, granted_at) VALUES (?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, queryStr, userRoleID.String(), userID.String(), roleID, grantedByStr, time.Now().Format(time.RFC3339))
	if err != nil {
		return fmt.Errorf("failed to add user role: %w", err)
	}

	return nil
}

func (r *SQLiteUserRepository) FindRoleByName(ctx context.Context, roleName string) (string, error) {
	var roleID string
	queryStr := `SELECT id FROM roles WHERE name = ?`
	err := r.db.QueryRowContext(ctx, queryStr, roleName).Scan(&roleID)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("role not found: %s", roleName)
		}
		return "", fmt.Errorf("failed to find role by name: %w", err)
	}

	return roleID, nil
}

type userRow struct {
	ID              string
	Email           string
	PasswordHash    string
	Username        sql.NullString
	Status          string
	EmailVerifiedAt sql.NullTime
	LastLoginAt     sql.NullTime
	Timezone        string
	CreatedAt       string
	UpdatedAt       string
}

type organizationRow struct {
	ID           string
	Name         string
	Slug         string
	Description  sql.NullString
	OwnerID      string
	BillingEmail sql.NullString
	Plan         string
	Status       string
	CreatedAt    string
	UpdatedAt    string
}

func (r *SQLiteUserRepository) mapRowToUser(row userRow) (*users.User, error) {
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
	if row.Username.Valid {
		username, err = users.NewUsername(row.Username.String)
		if err != nil {
			return nil, fmt.Errorf("invalid username: %w", err)
		}
	}

	// For now, use empty string for name since it's not in the database schema yet
	name := ""
	return users.ReconstructUser(
		userID, email, row.PasswordHash, name, username, status,
		emailVerifiedAt, lastLoginAt, row.Timezone, createdAt, updatedAt), nil
}

func (r *SQLiteUserRepository) mapRowToOrganization(row organizationRow) (*users.Organization, error) {
	orgID, err := users.OrganizationIDFromString(row.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid organization ID: %w", err)
	}

	ownerID, err := users.UserIDFromString(row.OwnerID)
	if err != nil {
		return nil, fmt.Errorf("invalid owner ID: %w", err)
	}

	plan := users.OrganizationPlan(row.Plan)
	status := users.OrganizationStatus(row.Status)

	createdAt, err := time.Parse(time.RFC3339, row.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("invalid created_at timestamp: %w", err)
	}

	updatedAt, err := time.Parse(time.RFC3339, row.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("invalid updated_at timestamp: %w", err)
	}

	description := ""
	if row.Description.Valid {
		description = row.Description.String
	}

	billingEmail := ""
	if row.BillingEmail.Valid {
		billingEmail = row.BillingEmail.String
	}

	return users.ReconstructOrganization(
		orgID, row.Name, row.Slug, description, ownerID,
		billingEmail, plan, status, createdAt, updatedAt), nil
}
