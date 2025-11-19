package disks

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Disk struct {
	id            DiskID
	name          DiskName
	projectID     uuid.UUID
	serviceID     *uuid.UUID
	size          DiskSize
	mountPath     string
	filesystem    Filesystem
	status        DiskStatus
	persistent    bool
	backupEnabled bool
	createdAt     time.Time
	updatedAt     time.Time
}

type DiskID struct {
	value string
}

func NewDiskID() DiskID {
	return DiskID{value: uuid.New().String()}
}

func DiskIDFromString(s string) (DiskID, error) {
	if s == "" {
		return DiskID{}, fmt.Errorf("disk ID cannot be empty")
	}
	return DiskID{value: s}, nil
}

func (id DiskID) String() string {
	return id.value
}

type DiskName struct {
	value string
}

func NewDiskName(name string) (DiskName, error) {
	if name == "" {
		return DiskName{}, fmt.Errorf("disk name cannot be empty")
	}
	if len(name) > 64 {
		return DiskName{}, fmt.Errorf("disk name cannot exceed 64 characters")
	}
	return DiskName{value: name}, nil
}

func (n DiskName) String() string {
	return n.value
}

type DiskSize struct {
	bytes int64
}

func NewDiskSize(sizeInBytes int64) (DiskSize, error) {
	if sizeInBytes < 0 {
		return DiskSize{}, fmt.Errorf("disk size cannot be negative")
	}
	if sizeInBytes > 0 && sizeInBytes < 1024*1024 { // Minimum 1MB if not unlimited
		return DiskSize{}, fmt.Errorf("disk size must be at least 1MB or 0 for unlimited")
	}
	return DiskSize{bytes: sizeInBytes}, nil
}

func NewDiskSizeFromGB(sizeInGB int) (DiskSize, error) {
	if sizeInGB < 0 {
		return DiskSize{}, fmt.Errorf("disk size cannot be negative")
	}
	return DiskSize{bytes: int64(sizeInGB) * 1024 * 1024 * 1024}, nil
}

func (s DiskSize) Bytes() int64 {
	return s.bytes
}

func (s DiskSize) MB() int64 {
	return s.bytes / (1024 * 1024)
}

func (s DiskSize) GB() int64 {
	return s.bytes / (1024 * 1024 * 1024)
}

func (s DiskSize) String() string {
	if s.bytes == 0 {
		return "unlimited"
	}
	if s.bytes >= 1024*1024*1024 {
		return fmt.Sprintf("%.1fGB", float64(s.bytes)/(1024*1024*1024))
	}
	if s.bytes >= 1024*1024 {
		return fmt.Sprintf("%.1fMB", float64(s.bytes)/(1024*1024))
	}
	return fmt.Sprintf("%dB", s.bytes)
}

type Filesystem string

const (
	FilesystemExt4  Filesystem = "ext4"
	FilesystemXFS   Filesystem = "xfs"
	FilesystemBtrfs Filesystem = "btrfs"
	FilesystemZFS   Filesystem = "zfs"
)

type DiskStatus string

const (
	DiskStatusCreating  DiskStatus = "creating"
	DiskStatusAvailable DiskStatus = "available"
	DiskStatusAttached  DiskStatus = "attached"
	DiskStatusDeleting  DiskStatus = "deleting"
	DiskStatusError     DiskStatus = "error"
)

type DiskBackup struct {
	id        DiskBackupID
	diskID    DiskID
	name      string
	size      DiskSize
	status    BackupStatus
	createdAt time.Time
}

type DiskBackupID struct {
	value string
}

func NewDiskBackupID() DiskBackupID {
	return DiskBackupID{value: uuid.New().String()}
}

func DiskBackupIDFromString(s string) (DiskBackupID, error) {
	if s == "" {
		return DiskBackupID{}, fmt.Errorf("disk backup ID cannot be empty")
	}
	return DiskBackupID{value: s}, nil
}

func (id DiskBackupID) String() string {
	return id.value
}

type BackupStatus string

const (
	BackupStatusCreating  BackupStatus = "creating"
	BackupStatusCompleted BackupStatus = "completed"
	BackupStatusRestoring BackupStatus = "restoring"
	BackupStatusDeleting  BackupStatus = "deleting"
	BackupStatusError     BackupStatus = "error"
)

func NewDisk(
	name DiskName,
	projectID uuid.UUID,
	size DiskSize,
	mountPath string,
	filesystem Filesystem,
	persistent bool,
) (*Disk, error) {
	if mountPath == "" {
		return nil, fmt.Errorf("mount path cannot be empty")
	}
	if mountPath[0] != '/' {
		return nil, fmt.Errorf("mount path must be absolute")
	}

	now := time.Now()
	return &Disk{
		id:            NewDiskID(),
		name:          name,
		projectID:     projectID,
		size:          size,
		mountPath:     mountPath,
		filesystem:    filesystem,
		status:        DiskStatusCreating,
		persistent:    persistent,
		backupEnabled: false,
		createdAt:     now,
		updatedAt:     now,
	}, nil
}

func (d *Disk) ID() DiskID {
	return d.id
}

func (d *Disk) Name() DiskName {
	return d.name
}

func (d *Disk) ProjectID() uuid.UUID {
	return d.projectID
}

func (d *Disk) ServiceID() *uuid.UUID {
	return d.serviceID
}

func (d *Disk) Size() DiskSize {
	return d.size
}

func (d *Disk) MountPath() string {
	return d.mountPath
}

func (d *Disk) Filesystem() Filesystem {
	return d.filesystem
}

func (d *Disk) Status() DiskStatus {
	return d.status
}

func (d *Disk) Persistent() bool {
	return d.persistent
}

func (d *Disk) BackupEnabled() bool {
	return d.backupEnabled
}

func (d *Disk) CreatedAt() time.Time {
	return d.createdAt
}

func (d *Disk) UpdatedAt() time.Time {
	return d.updatedAt
}

func (d *Disk) AttachToService(serviceID uuid.UUID) error {
	if d.status != DiskStatusAvailable {
		return fmt.Errorf("disk is not available for attachment")
	}
	d.serviceID = &serviceID
	d.status = DiskStatusAttached
	d.updatedAt = time.Now()
	return nil
}

func (d *Disk) DetachFromService() error {
	if d.status != DiskStatusAttached {
		return fmt.Errorf("disk is not attached to any service")
	}
	d.serviceID = nil
	d.status = DiskStatusAvailable
	d.updatedAt = time.Now()
	return nil
}

func (d *Disk) Resize(newSize DiskSize) error {
	if newSize.bytes <= d.size.bytes {
		return fmt.Errorf("new size must be larger than current size")
	}
	d.size = newSize
	d.updatedAt = time.Now()
	return nil
}

func (d *Disk) EnableBackup() {
	d.backupEnabled = true
	d.updatedAt = time.Now()
}

func (d *Disk) DisableBackup() {
	d.backupEnabled = false
	d.updatedAt = time.Now()
}

func (d *Disk) ChangeStatus(status DiskStatus) {
	d.status = status
	d.updatedAt = time.Now()
}

func (d *Disk) CanDelete() error {
	if d.status == DiskStatusAttached {
		return fmt.Errorf("disk is attached to a service")
	}
	if d.status == DiskStatusDeleting {
		return fmt.Errorf("disk is already being deleted")
	}
	return nil
}

func NewDiskBackup(diskID DiskID, name string, size DiskSize) *DiskBackup {
	return &DiskBackup{
		id:        NewDiskBackupID(),
		diskID:    diskID,
		name:      name,
		size:      size,
		status:    BackupStatusCreating,
		createdAt: time.Now(),
	}
}

func (db *DiskBackup) ID() DiskBackupID {
	return db.id
}

func (db *DiskBackup) DiskID() DiskID {
	return db.diskID
}

func (db *DiskBackup) Name() string {
	return db.name
}

func (db *DiskBackup) Size() DiskSize {
	return db.size
}

func (db *DiskBackup) Status() BackupStatus {
	return db.status
}

func (db *DiskBackup) CreatedAt() time.Time {
	return db.createdAt
}

func (db *DiskBackup) ChangeStatus(status BackupStatus) {
	db.status = status
}

func ReconstructDisk(
	id DiskID,
	name DiskName,
	projectID uuid.UUID,
	serviceID *uuid.UUID,
	size DiskSize,
	mountPath string,
	filesystem Filesystem,
	status DiskStatus,
	persistent, backupEnabled bool,
	createdAt, updatedAt time.Time,
) *Disk {
	return &Disk{
		id:            id,
		name:          name,
		projectID:     projectID,
		serviceID:     serviceID,
		size:          size,
		mountPath:     mountPath,
		filesystem:    filesystem,
		status:        status,
		persistent:    persistent,
		backupEnabled: backupEnabled,
		createdAt:     createdAt,
		updatedAt:     updatedAt,
	}
}

func ReconstructDiskBackup(
	id DiskBackupID,
	diskID DiskID,
	name string,
	size DiskSize,
	status BackupStatus,
	createdAt time.Time,
) *DiskBackup {
	return &DiskBackup{
		id:        id,
		diskID:    diskID,
		name:      name,
		size:      size,
		status:    status,
		createdAt: createdAt,
	}
}
