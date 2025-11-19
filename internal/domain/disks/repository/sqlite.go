package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/mikrocloud/mikrocloud/internal/domain/disks"
)

type SQLiteDiskRepository struct {
	db *sql.DB
}

func NewSQLiteDiskRepository(db *sql.DB) DiskRepository {
	return &SQLiteDiskRepository{db: db}
}

func (r *SQLiteDiskRepository) Create(ctx context.Context, disk *disks.Disk) error {
	query := `
		INSERT INTO disks (id, name, project_id, service_id, size_bytes, mount_path, filesystem, status, persistent, backup_enabled, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	var serviceID interface{}
	if disk.ServiceID() != nil {
		serviceID = disk.ServiceID().String()
	}

	_, err := r.db.ExecContext(ctx, query,
		disk.ID().String(),
		disk.Name().String(),
		disk.ProjectID().String(),
		serviceID,
		disk.Size().Bytes(),
		disk.MountPath(),
		string(disk.Filesystem()),
		string(disk.Status()),
		disk.Persistent(),
		disk.BackupEnabled(),
		disk.CreatedAt().Format(time.RFC3339),
		disk.UpdatedAt().Format(time.RFC3339),
	)

	if err != nil {
		return fmt.Errorf("failed to create disk: %w", err)
	}

	return nil
}

func (r *SQLiteDiskRepository) GetByID(ctx context.Context, id disks.DiskID) (*disks.Disk, error) {
	query := `
		SELECT id, name, project_id, service_id, size_bytes, mount_path, filesystem, status, persistent, backup_enabled, created_at, updated_at
		FROM disks
		WHERE id = ?
	`

	var row diskRow
	err := r.db.QueryRowContext(ctx, query, id.String()).Scan(
		&row.ID,
		&row.Name,
		&row.ProjectID,
		&row.ServiceID,
		&row.SizeBytes,
		&row.MountPath,
		&row.Filesystem,
		&row.Status,
		&row.Persistent,
		&row.BackupEnabled,
		&row.CreatedAt,
		&row.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("disk not found: %s", id.String())
		}
		return nil, fmt.Errorf("failed to get disk by ID: %w", err)
	}

	return r.mapRowToDisk(row)
}

func (r *SQLiteDiskRepository) GetByProjectID(ctx context.Context, projectID uuid.UUID) ([]*disks.Disk, error) {
	query := `
		SELECT id, name, project_id, service_id, size_bytes, mount_path, filesystem, status, persistent, backup_enabled, created_at, updated_at
		FROM disks
		WHERE project_id = ?
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, projectID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to get disks by project ID: %w", err)
	}
	defer rows.Close()

	var result []*disks.Disk
	for rows.Next() {
		var row diskRow
		err := rows.Scan(
			&row.ID,
			&row.Name,
			&row.ProjectID,
			&row.ServiceID,
			&row.SizeBytes,
			&row.MountPath,
			&row.Filesystem,
			&row.Status,
			&row.Persistent,
			&row.BackupEnabled,
			&row.CreatedAt,
			&row.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan disk row: %w", err)
		}

		disk, err := r.mapRowToDisk(row)
		if err != nil {
			return nil, fmt.Errorf("failed to map disk: %w", err)
		}

		result = append(result, disk)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating disk rows: %w", err)
	}

	return result, nil
}

func (r *SQLiteDiskRepository) GetByServiceID(ctx context.Context, serviceID uuid.UUID) ([]*disks.Disk, error) {
	query := `
		SELECT id, name, project_id, service_id, size_bytes, mount_path, filesystem, status, persistent, backup_enabled, created_at, updated_at
		FROM disks
		WHERE service_id = ?
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, serviceID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to get disks by service ID: %w", err)
	}
	defer rows.Close()

	var result []*disks.Disk
	for rows.Next() {
		var row diskRow
		err := rows.Scan(
			&row.ID,
			&row.Name,
			&row.ProjectID,
			&row.ServiceID,
			&row.SizeBytes,
			&row.MountPath,
			&row.Filesystem,
			&row.Status,
			&row.Persistent,
			&row.BackupEnabled,
			&row.CreatedAt,
			&row.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan disk row: %w", err)
		}

		disk, err := r.mapRowToDisk(row)
		if err != nil {
			return nil, fmt.Errorf("failed to map disk: %w", err)
		}

		result = append(result, disk)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating disk rows: %w", err)
	}

	return result, nil
}

func (r *SQLiteDiskRepository) Update(ctx context.Context, disk *disks.Disk) error {
	query := `
		UPDATE disks
		SET name = ?, service_id = ?, size_bytes = ?, mount_path = ?, filesystem = ?, status = ?, persistent = ?, backup_enabled = ?, updated_at = ?
		WHERE id = ?
	`

	var serviceID interface{}
	if disk.ServiceID() != nil {
		serviceID = disk.ServiceID().String()
	}

	result, err := r.db.ExecContext(ctx, query,
		disk.Name().String(),
		serviceID,
		disk.Size().Bytes(),
		disk.MountPath(),
		string(disk.Filesystem()),
		string(disk.Status()),
		disk.Persistent(),
		disk.BackupEnabled(),
		disk.UpdatedAt().Format(time.RFC3339),
		disk.ID().String(),
	)

	if err != nil {
		return fmt.Errorf("failed to update disk: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("disk not found: %s", disk.ID().String())
	}

	return nil
}

func (r *SQLiteDiskRepository) Delete(ctx context.Context, id disks.DiskID) error {
	query := `DELETE FROM disks WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query, id.String())
	if err != nil {
		return fmt.Errorf("failed to delete disk: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("disk not found: %s", id.String())
	}

	return nil
}

type diskRow struct {
	ID            string
	Name          string
	ProjectID     string
	ServiceID     sql.NullString
	SizeBytes     int64
	MountPath     string
	Filesystem    string
	Status        string
	Persistent    bool
	BackupEnabled bool
	CreatedAt     string
	UpdatedAt     string
}

func (r *SQLiteDiskRepository) mapRowToDisk(row diskRow) (*disks.Disk, error) {
	diskID, err := disks.DiskIDFromString(row.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid disk ID: %w", err)
	}

	diskName, err := disks.NewDiskName(row.Name)
	if err != nil {
		return nil, fmt.Errorf("invalid disk name: %w", err)
	}

	projectID, err := uuid.Parse(row.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("invalid project ID: %w", err)
	}

	var serviceID *uuid.UUID
	if row.ServiceID.Valid {
		parsedServiceID, err := uuid.Parse(row.ServiceID.String)
		if err != nil {
			return nil, fmt.Errorf("invalid service ID: %w", err)
		}
		serviceID = &parsedServiceID
	}

	diskSize, err := disks.NewDiskSize(row.SizeBytes)
	if err != nil {
		return nil, fmt.Errorf("invalid disk size: %w", err)
	}

	createdAt, err := time.Parse(time.RFC3339, row.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("invalid created_at timestamp: %w", err)
	}

	updatedAt, err := time.Parse(time.RFC3339, row.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("invalid updated_at timestamp: %w", err)
	}

	return disks.ReconstructDisk(
		diskID,
		diskName,
		projectID,
		serviceID,
		diskSize,
		row.MountPath,
		disks.Filesystem(row.Filesystem),
		disks.DiskStatus(row.Status),
		row.Persistent,
		row.BackupEnabled,
		createdAt,
		updatedAt,
	), nil
}

type SQLiteDiskBackupRepository struct {
	db *sql.DB
}

func NewSQLiteDiskBackupRepository(db *sql.DB) DiskBackupRepository {
	return &SQLiteDiskBackupRepository{db: db}
}

func (r *SQLiteDiskBackupRepository) Create(ctx context.Context, backup *disks.DiskBackup) error {
	query := `
		INSERT INTO disk_backups (id, disk_id, name, size_bytes, status, created_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.ExecContext(ctx, query,
		backup.ID().String(),
		backup.DiskID().String(),
		backup.Name(),
		backup.Size().Bytes(),
		string(backup.Status()),
		backup.CreatedAt().Format(time.RFC3339),
	)

	if err != nil {
		return fmt.Errorf("failed to create disk backup: %w", err)
	}

	return nil
}

func (r *SQLiteDiskBackupRepository) GetByID(ctx context.Context, id disks.DiskBackupID) (*disks.DiskBackup, error) {
	query := `
		SELECT id, disk_id, name, size_bytes, status, created_at
		FROM disk_backups
		WHERE id = ?
	`

	var row diskBackupRow
	err := r.db.QueryRowContext(ctx, query, id.String()).Scan(
		&row.ID,
		&row.DiskID,
		&row.Name,
		&row.SizeBytes,
		&row.Status,
		&row.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("disk backup not found: %s", id.String())
		}
		return nil, fmt.Errorf("failed to get disk backup by ID: %w", err)
	}

	return r.mapRowToDiskBackup(row)
}

func (r *SQLiteDiskBackupRepository) GetByDiskID(ctx context.Context, diskID disks.DiskID) ([]*disks.DiskBackup, error) {
	query := `
		SELECT id, disk_id, name, size_bytes, status, created_at
		FROM disk_backups
		WHERE disk_id = ?
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, diskID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to get disk backups by disk ID: %w", err)
	}
	defer rows.Close()

	var result []*disks.DiskBackup
	for rows.Next() {
		var row diskBackupRow
		err := rows.Scan(
			&row.ID,
			&row.DiskID,
			&row.Name,
			&row.SizeBytes,
			&row.Status,
			&row.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan disk backup row: %w", err)
		}

		backup, err := r.mapRowToDiskBackup(row)
		if err != nil {
			return nil, fmt.Errorf("failed to map disk backup: %w", err)
		}

		result = append(result, backup)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating disk backup rows: %w", err)
	}

	return result, nil
}

func (r *SQLiteDiskBackupRepository) Update(ctx context.Context, backup *disks.DiskBackup) error {
	query := `
		UPDATE disk_backups
		SET name = ?, size_bytes = ?, status = ?
		WHERE id = ?
	`

	result, err := r.db.ExecContext(ctx, query,
		backup.Name(),
		backup.Size().Bytes(),
		string(backup.Status()),
		backup.ID().String(),
	)

	if err != nil {
		return fmt.Errorf("failed to update disk backup: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("disk backup not found: %s", backup.ID().String())
	}

	return nil
}

func (r *SQLiteDiskBackupRepository) Delete(ctx context.Context, id disks.DiskBackupID) error {
	query := `DELETE FROM disk_backups WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query, id.String())
	if err != nil {
		return fmt.Errorf("failed to delete disk backup: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("disk backup not found: %s", id.String())
	}

	return nil
}

type diskBackupRow struct {
	ID        string
	DiskID    string
	Name      string
	SizeBytes int64
	Status    string
	CreatedAt string
}

func (r *SQLiteDiskBackupRepository) mapRowToDiskBackup(row diskBackupRow) (*disks.DiskBackup, error) {
	backupID, err := disks.DiskBackupIDFromString(row.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid backup ID: %w", err)
	}

	diskID, err := disks.DiskIDFromString(row.DiskID)
	if err != nil {
		return nil, fmt.Errorf("invalid disk ID: %w", err)
	}

	size, err := disks.NewDiskSize(row.SizeBytes)
	if err != nil {
		return nil, fmt.Errorf("invalid size: %w", err)
	}

	createdAt, err := time.Parse(time.RFC3339, row.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("invalid created_at timestamp: %w", err)
	}

	return disks.ReconstructDiskBackup(
		backupID,
		diskID,
		row.Name,
		size,
		disks.BackupStatus(row.Status),
		createdAt,
	), nil
}
