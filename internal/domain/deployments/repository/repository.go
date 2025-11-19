package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/mikrocloud/mikrocloud/internal/domain/applications"
	"github.com/mikrocloud/mikrocloud/internal/domain/deployments"
	"github.com/mikrocloud/mikrocloud/internal/domain/users"
	"github.com/stephenafamo/bob/dialect/sqlite"
	"github.com/stephenafamo/bob/dialect/sqlite/im"
	"github.com/stephenafamo/bob/dialect/sqlite/sm"
)

type DeploymentRepository interface {
	Create(ctx context.Context, deployment *deployments.Deployment) error
	GetByID(ctx context.Context, id deployments.DeploymentID) (*deployments.Deployment, error)
	Update(ctx context.Context, deployment *deployments.Deployment) error
	Delete(ctx context.Context, id deployments.DeploymentID) error
	List(ctx context.Context) ([]*deployments.Deployment, error)
	ListByApplication(ctx context.Context, applicationID applications.ApplicationID) ([]*deployments.Deployment, error)
	GetLatestByApplication(ctx context.Context, applicationID applications.ApplicationID) (*deployments.Deployment, error)
	ListByStatus(ctx context.Context, status deployments.DeploymentStatus) ([]*deployments.Deployment, error)
}

type sqliteDeploymentRepository struct {
	db *sql.DB
}

func NewSQLiteDeploymentRepository(db *sql.DB) DeploymentRepository {
	return &sqliteDeploymentRepository{db: db}
}

func (r *sqliteDeploymentRepository) Create(ctx context.Context, deployment *deployments.Deployment) error {
	var triggeredBy *string
	if deployment.TriggeredBy() != nil {
		userID := deployment.TriggeredBy().String()
		triggeredBy = &userID
	}

	query := sqlite.Insert(
		im.Into("deployments"),
		im.Values(
			sqlite.Arg(deployment.ID().String()),
			sqlite.Arg(deployment.ApplicationID().String()),
			sqlite.Arg(deployment.DeploymentNumber()),
			sqlite.Arg(boolToInt(deployment.IsProduction())),
			sqlite.Arg(triggeredBy),
			sqlite.Arg(string(deployment.TriggerType())),
			sqlite.Arg(string(deployment.Status())),
			sqlite.Arg(deployment.ContainerID()),
			sqlite.Arg(deployment.ImageTag()),
			sqlite.Arg(deployment.ImageDigest()),
			sqlite.Arg(deployment.GitCommitHash()),
			sqlite.Arg(deployment.GitCommitMessage()),
			sqlite.Arg(deployment.GitBranch()),
			sqlite.Arg(deployment.GitAuthorName()),
			sqlite.Arg(deployment.BuildLogs()),
			sqlite.Arg(deployment.DeployLogs()),
			sqlite.Arg(deployment.ErrorMessage()),
			sqlite.Arg(deployment.StartedAt().Format(time.RFC3339)),
			sqlite.Arg(formatTimePtr(deployment.BuildStartedAt())),
			sqlite.Arg(formatTimePtr(deployment.BuildCompletedAt())),
			sqlite.Arg(formatTimePtr(deployment.DeployStartedAt())),
			sqlite.Arg(formatTimePtr(deployment.DeployCompletedAt())),
			sqlite.Arg(formatTimePtr(deployment.StoppedAt())),
			sqlite.Arg(deployment.BuildDurationSeconds()),
			sqlite.Arg(deployment.DeployDurationSeconds()),
			sqlite.Arg(deployment.UpdatedAt().Format(time.RFC3339)),
		),
	)

	queryStr, args, err := query.Build(ctx)
	if err != nil {
		return fmt.Errorf("failed to build insert query: %w", err)
	}

	_, err = r.db.ExecContext(ctx, queryStr, args...)
	if err != nil {
		return fmt.Errorf("failed to create deployment: %w", err)
	}

	return nil
}

func (r *sqliteDeploymentRepository) GetByID(ctx context.Context, id deployments.DeploymentID) (*deployments.Deployment, error) {
	query := sqlite.Select(
		sm.Columns(
			"id", "application_id", "deployment_number", "is_production", "triggered_by",
			"trigger_type", "status", "container_id", "image_tag", "image_digest",
			"git_commit_hash", "git_commit_message", "git_branch", "git_author_name",
			"build_logs", "deploy_logs", "error_message", "started_at",
			"build_started_at", "build_completed_at", "deploy_started_at",
			"deploy_completed_at", "stopped_at", "build_duration_seconds",
			"deploy_duration_seconds", "updated_at",
		),
		sm.From("deployments"),
		sm.Where(sqlite.Quote("id").EQ(sqlite.Arg(id.String()))),
	)

	queryStr, args, err := query.Build(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	row := deploymentRow{}
	err = r.db.QueryRowContext(ctx, queryStr, args...).Scan(
		&row.ID, &row.ApplicationID, &row.DeploymentNumber, &row.IsProduction, &row.TriggeredBy,
		&row.TriggerType, &row.Status, &row.ContainerID, &row.ImageTag, &row.ImageDigest,
		&row.GitCommitHash, &row.GitCommitMessage, &row.GitBranch, &row.GitAuthorName,
		&row.BuildLogs, &row.DeployLogs, &row.ErrorMessage, &row.StartedAt,
		&row.BuildStartedAt, &row.BuildCompletedAt, &row.DeployStartedAt,
		&row.DeployCompletedAt, &row.StoppedAt, &row.BuildDurationSeconds,
		&row.DeployDurationSeconds, &row.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("deployment not found: %s", id.String())
		}
		return nil, fmt.Errorf("failed to get deployment: %w", err)
	}

	return r.mapRowToDeployment(row)
}

func (r *sqliteDeploymentRepository) Update(ctx context.Context, deployment *deployments.Deployment) error {
	var triggeredBy *string
	if deployment.TriggeredBy() != nil {
		userID := deployment.TriggeredBy().String()
		triggeredBy = &userID
	}

	// Use direct SQL for update since Bob's update is complex
	query := `UPDATE deployments SET 
		deployment_number = ?, is_production = ?, triggered_by = ?, trigger_type = ?, 
		status = ?, container_id = ?, image_tag = ?, image_digest = ?,
		git_commit_hash = ?, git_commit_message = ?, git_branch = ?, git_author_name = ?,
		build_logs = ?, deploy_logs = ?, error_message = ?, started_at = ?,
		build_started_at = ?, build_completed_at = ?, deploy_started_at = ?, 
		deploy_completed_at = ?, stopped_at = ?, build_duration_seconds = ?,
		deploy_duration_seconds = ?, updated_at = ?
		WHERE id = ?`

	_, err := r.db.ExecContext(ctx, query,
		deployment.DeploymentNumber(),
		boolToInt(deployment.IsProduction()),
		triggeredBy,
		string(deployment.TriggerType()),
		string(deployment.Status()),
		deployment.ContainerID(),
		deployment.ImageTag(),
		deployment.ImageDigest(),
		deployment.GitCommitHash(),
		deployment.GitCommitMessage(),
		deployment.GitBranch(),
		deployment.GitAuthorName(),
		deployment.BuildLogs(),
		deployment.DeployLogs(),
		deployment.ErrorMessage(),
		deployment.StartedAt().Format(time.RFC3339),
		formatTimePtr(deployment.BuildStartedAt()),
		formatTimePtr(deployment.BuildCompletedAt()),
		formatTimePtr(deployment.DeployStartedAt()),
		formatTimePtr(deployment.DeployCompletedAt()),
		formatTimePtr(deployment.StoppedAt()),
		deployment.BuildDurationSeconds(),
		deployment.DeployDurationSeconds(),
		deployment.UpdatedAt().Format(time.RFC3339),
		deployment.ID().String(),
	)
	if err != nil {
		return fmt.Errorf("failed to update deployment: %w", err)
	}

	return nil
}

func (r *sqliteDeploymentRepository) Delete(ctx context.Context, id deployments.DeploymentID) error {
	query := `DELETE FROM deployments WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, id.String())
	if err != nil {
		return fmt.Errorf("failed to delete deployment: %w", err)
	}
	return nil
}

func (r *sqliteDeploymentRepository) List(ctx context.Context) ([]*deployments.Deployment, error) {
	query := sqlite.Select(
		sm.Columns(
			"id", "application_id", "deployment_number", "is_production", "triggered_by",
			"trigger_type", "status", "container_id", "image_tag", "image_digest",
			"git_commit_hash", "git_commit_message", "git_branch", "git_author_name",
			"build_logs", "deploy_logs", "error_message", "started_at",
			"build_started_at", "build_completed_at", "deploy_started_at",
			"deploy_completed_at", "stopped_at", "build_duration_seconds",
			"deploy_duration_seconds", "updated_at",
		),
		sm.From("deployments"),
		sm.OrderBy("started_at").Desc(),
	)

	queryStr, args, err := query.Build(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	rows, err := r.db.QueryContext(ctx, queryStr, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list deployments: %w", err)
	}
	defer rows.Close()

	var result []*deployments.Deployment
	for rows.Next() {
		row := deploymentRow{}
		err := rows.Scan(
			&row.ID, &row.ApplicationID, &row.DeploymentNumber, &row.IsProduction, &row.TriggeredBy,
			&row.TriggerType, &row.Status, &row.ContainerID, &row.ImageTag, &row.ImageDigest,
			&row.GitCommitHash, &row.GitCommitMessage, &row.GitBranch, &row.GitAuthorName,
			&row.BuildLogs, &row.DeployLogs, &row.ErrorMessage, &row.StartedAt,
			&row.BuildStartedAt, &row.BuildCompletedAt, &row.DeployStartedAt,
			&row.DeployCompletedAt, &row.StoppedAt, &row.BuildDurationSeconds,
			&row.DeployDurationSeconds, &row.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan deployment row: %w", err)
		}

		deployment, err := r.mapRowToDeployment(row)
		if err != nil {
			return nil, err
		}
		result = append(result, deployment)
	}

	return result, nil
}

func (r *sqliteDeploymentRepository) ListByApplication(ctx context.Context, applicationID applications.ApplicationID) ([]*deployments.Deployment, error) {
	query := sqlite.Select(
		sm.Columns(
			"id", "application_id", "deployment_number", "is_production", "triggered_by",
			"trigger_type", "status", "container_id", "image_tag", "image_digest",
			"git_commit_hash", "git_commit_message", "git_branch", "git_author_name",
			"build_logs", "deploy_logs", "error_message", "started_at",
			"build_started_at", "build_completed_at", "deploy_started_at",
			"deploy_completed_at", "stopped_at", "build_duration_seconds",
			"deploy_duration_seconds", "updated_at",
		),
		sm.From("deployments"),
		sm.Where(sqlite.Quote("application_id").EQ(sqlite.Arg(applicationID.String()))),
		sm.OrderBy("deployment_number").Desc(),
	)

	queryStr, args, err := query.Build(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	rows, err := r.db.QueryContext(ctx, queryStr, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list deployments by application: %w", err)
	}
	defer rows.Close()

	var result []*deployments.Deployment
	for rows.Next() {
		row := deploymentRow{}
		err := rows.Scan(
			&row.ID, &row.ApplicationID, &row.DeploymentNumber, &row.IsProduction, &row.TriggeredBy,
			&row.TriggerType, &row.Status, &row.ContainerID, &row.ImageTag, &row.ImageDigest,
			&row.GitCommitHash, &row.GitCommitMessage, &row.GitBranch, &row.GitAuthorName,
			&row.BuildLogs, &row.DeployLogs, &row.ErrorMessage, &row.StartedAt,
			&row.BuildStartedAt, &row.BuildCompletedAt, &row.DeployStartedAt,
			&row.DeployCompletedAt, &row.StoppedAt, &row.BuildDurationSeconds,
			&row.DeployDurationSeconds, &row.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan deployment row: %w", err)
		}

		deployment, err := r.mapRowToDeployment(row)
		if err != nil {
			return nil, err
		}
		result = append(result, deployment)
	}

	return result, nil
}

func (r *sqliteDeploymentRepository) GetLatestByApplication(ctx context.Context, applicationID applications.ApplicationID) (*deployments.Deployment, error) {
	query := sqlite.Select(
		sm.Columns(
			"id", "application_id", "deployment_number", "is_production", "triggered_by",
			"trigger_type", "status", "container_id", "image_tag", "image_digest",
			"git_commit_hash", "git_commit_message", "git_branch", "git_author_name",
			"build_logs", "deploy_logs", "error_message", "started_at",
			"build_started_at", "build_completed_at", "deploy_started_at",
			"deploy_completed_at", "stopped_at", "build_duration_seconds",
			"deploy_duration_seconds", "updated_at",
		),
		sm.From("deployments"),
		sm.Where(sqlite.Quote("application_id").EQ(sqlite.Arg(applicationID.String()))),
		sm.OrderBy("deployment_number").Desc(),
		sm.Limit(1),
	)

	queryStr, args, err := query.Build(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	row := deploymentRow{}
	err = r.db.QueryRowContext(ctx, queryStr, args...).Scan(
		&row.ID, &row.ApplicationID, &row.DeploymentNumber, &row.IsProduction, &row.TriggeredBy,
		&row.TriggerType, &row.Status, &row.ContainerID, &row.ImageTag, &row.ImageDigest,
		&row.GitCommitHash, &row.GitCommitMessage, &row.GitBranch, &row.GitAuthorName,
		&row.BuildLogs, &row.DeployLogs, &row.ErrorMessage, &row.StartedAt,
		&row.BuildStartedAt, &row.BuildCompletedAt, &row.DeployStartedAt,
		&row.DeployCompletedAt, &row.StoppedAt, &row.BuildDurationSeconds,
		&row.DeployDurationSeconds, &row.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no deployments found for application: %s", applicationID.String())
		}
		return nil, fmt.Errorf("failed to get latest deployment: %w", err)
	}

	return r.mapRowToDeployment(row)
}

func (r *sqliteDeploymentRepository) ListByStatus(ctx context.Context, status deployments.DeploymentStatus) ([]*deployments.Deployment, error) {
	query := sqlite.Select(
		sm.Columns(
			"id", "application_id", "deployment_number", "is_production", "triggered_by",
			"trigger_type", "status", "container_id", "image_tag", "image_digest",
			"git_commit_hash", "git_commit_message", "git_branch", "git_author_name",
			"build_logs", "deploy_logs", "error_message", "started_at",
			"build_started_at", "build_completed_at", "deploy_started_at",
			"deploy_completed_at", "stopped_at", "build_duration_seconds",
			"deploy_duration_seconds", "updated_at",
		),
		sm.From("deployments"),
		sm.Where(sqlite.Quote("status").EQ(sqlite.Arg(string(status)))),
		sm.OrderBy("started_at").Desc(),
	)

	queryStr, args, err := query.Build(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	rows, err := r.db.QueryContext(ctx, queryStr, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list deployments by status: %w", err)
	}
	defer rows.Close()

	var result []*deployments.Deployment
	for rows.Next() {
		row := deploymentRow{}
		err := rows.Scan(
			&row.ID, &row.ApplicationID, &row.DeploymentNumber, &row.IsProduction, &row.TriggeredBy,
			&row.TriggerType, &row.Status, &row.ContainerID, &row.ImageTag, &row.ImageDigest,
			&row.GitCommitHash, &row.GitCommitMessage, &row.GitBranch, &row.GitAuthorName,
			&row.BuildLogs, &row.DeployLogs, &row.ErrorMessage, &row.StartedAt,
			&row.BuildStartedAt, &row.BuildCompletedAt, &row.DeployStartedAt,
			&row.DeployCompletedAt, &row.StoppedAt, &row.BuildDurationSeconds,
			&row.DeployDurationSeconds, &row.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan deployment row: %w", err)
		}

		deployment, err := r.mapRowToDeployment(row)
		if err != nil {
			return nil, err
		}
		result = append(result, deployment)
	}

	return result, nil
}

type deploymentRow struct {
	ID                    string
	ApplicationID         string
	DeploymentNumber      int
	IsProduction          int
	TriggeredBy           *string
	TriggerType           string
	Status                string
	ContainerID           string
	ImageTag              string
	ImageDigest           string
	GitCommitHash         string
	GitCommitMessage      string
	GitBranch             string
	GitAuthorName         string
	BuildLogs             string
	DeployLogs            string
	ErrorMessage          string
	StartedAt             string
	BuildStartedAt        *string
	BuildCompletedAt      *string
	DeployStartedAt       *string
	DeployCompletedAt     *string
	StoppedAt             *string
	BuildDurationSeconds  *int
	DeployDurationSeconds *int
	UpdatedAt             string
}

func (r *sqliteDeploymentRepository) mapRowToDeployment(row deploymentRow) (*deployments.Deployment, error) {
	deploymentID, err := deployments.DeploymentIDFromString(row.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid deployment ID: %w", err)
	}

	appID, err := applications.ApplicationIDFromString(row.ApplicationID)
	if err != nil {
		return nil, fmt.Errorf("invalid application ID: %w", err)
	}

	var triggeredBy *users.UserID
	if row.TriggeredBy != nil {
		userID, err := users.UserIDFromString(*row.TriggeredBy)
		if err != nil {
			return nil, fmt.Errorf("invalid triggered by user ID: %w", err)
		}
		triggeredBy = &userID
	}

	startedAt, err := time.Parse(time.RFC3339, row.StartedAt)
	if err != nil {
		return nil, fmt.Errorf("invalid started at time: %w", err)
	}

	updatedAt, err := time.Parse(time.RFC3339, row.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("invalid updated at time: %w", err)
	}

	var buildStartedAt, buildCompletedAt, deployStartedAt, deployCompletedAt, stoppedAt *time.Time

	if row.BuildStartedAt != nil {
		t, err := time.Parse(time.RFC3339, *row.BuildStartedAt)
		if err != nil {
			return nil, fmt.Errorf("invalid build started at time: %w", err)
		}
		buildStartedAt = &t
	}

	if row.BuildCompletedAt != nil {
		t, err := time.Parse(time.RFC3339, *row.BuildCompletedAt)
		if err != nil {
			return nil, fmt.Errorf("invalid build completed at time: %w", err)
		}
		buildCompletedAt = &t
	}

	if row.DeployStartedAt != nil {
		t, err := time.Parse(time.RFC3339, *row.DeployStartedAt)
		if err != nil {
			return nil, fmt.Errorf("invalid deploy started at time: %w", err)
		}
		deployStartedAt = &t
	}

	if row.DeployCompletedAt != nil {
		t, err := time.Parse(time.RFC3339, *row.DeployCompletedAt)
		if err != nil {
			return nil, fmt.Errorf("invalid deploy completed at time: %w", err)
		}
		deployCompletedAt = &t
	}

	if row.StoppedAt != nil {
		t, err := time.Parse(time.RFC3339, *row.StoppedAt)
		if err != nil {
			return nil, fmt.Errorf("invalid stopped at time: %w", err)
		}
		stoppedAt = &t
	}

	return deployments.ReconstructDeployment(
		deploymentID,
		appID,
		row.DeploymentNumber,
		intToBool(row.IsProduction),
		triggeredBy,
		deployments.TriggerType(row.TriggerType),
		deployments.DeploymentStatus(row.Status),
		row.ContainerID,
		row.ImageTag,
		row.ImageDigest,
		row.GitCommitHash,
		row.GitCommitMessage,
		row.GitBranch,
		row.GitAuthorName,
		row.BuildLogs,
		row.DeployLogs,
		row.ErrorMessage,
		startedAt,
		buildStartedAt,
		buildCompletedAt,
		deployStartedAt,
		deployCompletedAt,
		stoppedAt,
		row.BuildDurationSeconds,
		row.DeployDurationSeconds,
		updatedAt,
	), nil
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func intToBool(i int) bool {
	return i != 0
}

func formatTimePtr(t *time.Time) *string {
	if t == nil {
		return nil
	}
	formatted := t.Format(time.RFC3339)
	return &formatted
}
