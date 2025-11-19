package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/mikrocloud/mikrocloud/internal/domain/disks"
	"github.com/mikrocloud/mikrocloud/internal/domain/disks/repository"
)

type DiskService struct {
	diskRepo   repository.DiskRepository
	backupRepo repository.DiskBackupRepository
}

func NewDiskService(
	diskRepo repository.DiskRepository,
	backupRepo repository.DiskBackupRepository,
) *DiskService {
	return &DiskService{
		diskRepo:   diskRepo,
		backupRepo: backupRepo,
	}
}

// GetDiskMounts returns all disk mounts for a service formatted for container volumes
func (s *DiskService) GetDiskMounts(ctx context.Context, serviceID uuid.UUID) (map[string]string, error) {
	disks, err := s.diskRepo.GetByServiceID(ctx, serviceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get disks for service: %w", err)
	}

	volumes := make(map[string]string)
	for _, disk := range disks {
		// Format: /var/lib/mikrocloud/volumes/<disk-id>:/mount/path
		hostPath := fmt.Sprintf("/var/lib/mikrocloud/volumes/%s", disk.ID())
		volumes[hostPath] = disk.MountPath()
	}

	return volumes, nil
}

func (s *DiskService) CreateDisk(
	ctx context.Context,
	name disks.DiskName,
	projectID uuid.UUID,
	size disks.DiskSize,
	mountPath string,
	filesystem disks.Filesystem,
	persistent bool,
) (*disks.Disk, error) {
	disk, err := disks.NewDisk(name, projectID, size, mountPath, filesystem, persistent)
	if err != nil {
		return nil, fmt.Errorf("failed to create disk: %w", err)
	}

	if err := s.diskRepo.Create(ctx, disk); err != nil {
		return nil, fmt.Errorf("failed to save disk: %w", err)
	}

	return disk, nil
}

func (s *DiskService) GetDisk(ctx context.Context, id disks.DiskID) (*disks.Disk, error) {
	return s.diskRepo.GetByID(ctx, id)
}

func (s *DiskService) GetDisksByProject(ctx context.Context, projectID uuid.UUID) ([]*disks.Disk, error) {
	return s.diskRepo.GetByProjectID(ctx, projectID)
}

func (s *DiskService) GetDisksByService(ctx context.Context, serviceID uuid.UUID) ([]*disks.Disk, error) {
	return s.diskRepo.GetByServiceID(ctx, serviceID)
}

func (s *DiskService) AttachDisk(ctx context.Context, diskID disks.DiskID, serviceID uuid.UUID) error {
	disk, err := s.diskRepo.GetByID(ctx, diskID)
	if err != nil {
		return err
	}

	if err := disk.AttachToService(serviceID); err != nil {
		return err
	}

	return s.diskRepo.Update(ctx, disk)
}

func (s *DiskService) DetachDisk(ctx context.Context, diskID disks.DiskID) error {
	disk, err := s.diskRepo.GetByID(ctx, diskID)
	if err != nil {
		return err
	}

	if err := disk.DetachFromService(); err != nil {
		return err
	}

	return s.diskRepo.Update(ctx, disk)
}

func (s *DiskService) ResizeDisk(ctx context.Context, diskID disks.DiskID, newSize disks.DiskSize) error {
	disk, err := s.diskRepo.GetByID(ctx, diskID)
	if err != nil {
		return err
	}

	if err := disk.Resize(newSize); err != nil {
		return err
	}

	return s.diskRepo.Update(ctx, disk)
}

func (s *DiskService) EnableBackup(ctx context.Context, diskID disks.DiskID) error {
	disk, err := s.diskRepo.GetByID(ctx, diskID)
	if err != nil {
		return err
	}

	disk.EnableBackup()
	return s.diskRepo.Update(ctx, disk)
}

func (s *DiskService) DisableBackup(ctx context.Context, diskID disks.DiskID) error {
	disk, err := s.diskRepo.GetByID(ctx, diskID)
	if err != nil {
		return err
	}

	disk.DisableBackup()
	return s.diskRepo.Update(ctx, disk)
}

func (s *DiskService) DeleteDisk(ctx context.Context, diskID disks.DiskID) error {
	disk, err := s.diskRepo.GetByID(ctx, diskID)
	if err != nil {
		return err
	}

	if err := disk.CanDelete(); err != nil {
		return err
	}

	backups, err := s.backupRepo.GetByDiskID(ctx, diskID)
	if err != nil {
		return fmt.Errorf("failed to check disk backups: %w", err)
	}

	for _, backup := range backups {
		if err := s.backupRepo.Delete(ctx, backup.ID()); err != nil {
			return fmt.Errorf("failed to delete backup %s: %w", backup.ID(), err)
		}
	}

	return s.diskRepo.Delete(ctx, diskID)
}

func (s *DiskService) CreateBackup(ctx context.Context, diskID disks.DiskID, name string) (*disks.DiskBackup, error) {
	disk, err := s.diskRepo.GetByID(ctx, diskID)
	if err != nil {
		return nil, err
	}

	if !disk.BackupEnabled() {
		return nil, fmt.Errorf("backup is not enabled for disk %s", diskID)
	}

	backup := disks.NewDiskBackup(diskID, name, disk.Size())
	if err := s.backupRepo.Create(ctx, backup); err != nil {
		return nil, fmt.Errorf("failed to create backup: %w", err)
	}

	return backup, nil
}

func (s *DiskService) GetBackup(ctx context.Context, backupID disks.DiskBackupID) (*disks.DiskBackup, error) {
	return s.backupRepo.GetByID(ctx, backupID)
}

func (s *DiskService) GetBackupsByDisk(ctx context.Context, diskID disks.DiskID) ([]*disks.DiskBackup, error) {
	return s.backupRepo.GetByDiskID(ctx, diskID)
}

func (s *DiskService) DeleteBackup(ctx context.Context, backupID disks.DiskBackupID) error {
	return s.backupRepo.Delete(ctx, backupID)
}
