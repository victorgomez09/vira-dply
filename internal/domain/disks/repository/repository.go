package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/mikrocloud/mikrocloud/internal/domain/disks"
)

type DiskRepository interface {
	Create(ctx context.Context, disk *disks.Disk) error
	GetByID(ctx context.Context, id disks.DiskID) (*disks.Disk, error)
	GetByProjectID(ctx context.Context, projectID uuid.UUID) ([]*disks.Disk, error)
	GetByServiceID(ctx context.Context, serviceID uuid.UUID) ([]*disks.Disk, error)
	Update(ctx context.Context, disk *disks.Disk) error
	Delete(ctx context.Context, id disks.DiskID) error
}

type DiskBackupRepository interface {
	Create(ctx context.Context, backup *disks.DiskBackup) error
	GetByID(ctx context.Context, id disks.DiskBackupID) (*disks.DiskBackup, error)
	GetByDiskID(ctx context.Context, diskID disks.DiskID) ([]*disks.DiskBackup, error)
	Update(ctx context.Context, backup *disks.DiskBackup) error
	Delete(ctx context.Context, id disks.DiskBackupID) error
}
