package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/mikrocloud/mikrocloud/internal/domain/databases"
	"github.com/mikrocloud/mikrocloud/internal/domain/disks"
	"github.com/mikrocloud/mikrocloud/internal/domain/disks/service"
	"github.com/mikrocloud/mikrocloud/internal/utils"
)

type DatabaseService interface {
	GetDatabase(ctx context.Context, id databases.DatabaseID) (*databases.Database, error)
}

type DatabaseDeploymentService interface {
	Restart(ctx context.Context, database *databases.Database) error
}

type DiskHandler struct {
	diskService       *service.DiskService
	databaseService   DatabaseService
	deploymentService DatabaseDeploymentService
	validator         *validator.Validate
}

func NewDiskHandler(diskService *service.DiskService, databaseService DatabaseService, deploymentService DatabaseDeploymentService) *DiskHandler {
	return &DiskHandler{
		diskService:       diskService,
		databaseService:   databaseService,
		deploymentService: deploymentService,
		validator:         validator.New(),
	}
}

type DiskResponse struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	ProjectID     string `json:"project_id"`
	ServiceID     string `json:"service_id,omitempty"`
	Size          int64  `json:"size"`
	SizeGB        int64  `json:"size_gb"`
	MountPath     string `json:"mount_path"`
	Filesystem    string `json:"filesystem"`
	Status        string `json:"status"`
	Persistent    bool   `json:"persistent"`
	BackupEnabled bool   `json:"backup_enabled"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
}

type CreateDiskRequest struct {
	Name       string `json:"name" validate:"required,min=1,max=64"`
	SizeGB     int    `json:"size_gb" validate:"required,min=1"`
	MountPath  string `json:"mount_path" validate:"required"`
	Filesystem string `json:"filesystem" validate:"required,oneof=ext4 xfs btrfs zfs"`
	Persistent bool   `json:"persistent"`
}

type UpdateDiskRequest struct {
	SizeGB *int `json:"size_gb,omitempty" validate:"omitempty,min=1"`
}

type AttachDiskRequest struct {
	ServiceID string `json:"service_id" validate:"required,uuid"`
}

type ListDisksResponse struct {
	Disks []DiskResponse `json:"disks"`
}

func toDiskResponse(disk *disks.Disk) DiskResponse {
	resp := DiskResponse{
		ID:            disk.ID().String(),
		Name:          disk.Name().String(),
		ProjectID:     disk.ProjectID().String(),
		Size:          disk.Size().Bytes(),
		SizeGB:        disk.Size().GB(),
		MountPath:     disk.MountPath(),
		Filesystem:    string(disk.Filesystem()),
		Status:        string(disk.Status()),
		Persistent:    disk.Persistent(),
		BackupEnabled: disk.BackupEnabled(),
		CreatedAt:     disk.CreatedAt().Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:     disk.UpdatedAt().Format("2006-01-02T15:04:05Z07:00"),
	}

	if disk.ServiceID() != nil {
		resp.ServiceID = disk.ServiceID().String()
	}

	return resp
}

func (h *DiskHandler) ListDisks(w http.ResponseWriter, r *http.Request) {
	projectIDStr := chi.URLParam(r, "project_id")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "invalid_project_id", "Invalid project ID")
		return
	}

	disks, err := h.diskService.GetDisksByProject(r.Context(), projectID)
	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, "list_failed", "Failed to list disks")
		return
	}

	response := ListDisksResponse{
		Disks: make([]DiskResponse, len(disks)),
	}
	for i, disk := range disks {
		response.Disks[i] = toDiskResponse(disk)
	}

	utils.SendJSON(w, http.StatusOK, response)
}

func (h *DiskHandler) GetDisk(w http.ResponseWriter, r *http.Request) {
	projectIDStr := chi.URLParam(r, "project_id")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "invalid_project_id", "Invalid project ID")
		return
	}

	diskIDStr := chi.URLParam(r, "disk_id")
	diskID, err := disks.DiskIDFromString(diskIDStr)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "invalid_disk_id", "Invalid disk ID")
		return
	}

	disk, err := h.diskService.GetDisk(r.Context(), diskID)
	if err != nil {
		utils.SendError(w, http.StatusNotFound, "disk_not_found", "Disk not found")
		return
	}

	if disk.ProjectID() != projectID {
		utils.SendError(w, http.StatusNotFound, "disk_not_found", "Disk not found in project")
		return
	}

	utils.SendJSON(w, http.StatusOK, toDiskResponse(disk))
}

func (h *DiskHandler) CreateDisk(w http.ResponseWriter, r *http.Request) {
	projectIDStr := chi.URLParam(r, "project_id")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "invalid_project_id", "Invalid project ID")
		return
	}

	var req CreateDiskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.SendError(w, http.StatusBadRequest, "invalid_request", "Invalid request body")
		return
	}

	if err := h.validator.Struct(req); err != nil {
		utils.SendError(w, http.StatusBadRequest, "validation_failed", err.Error())
		return
	}

	diskName, err := disks.NewDiskName(req.Name)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "invalid_name", err.Error())
		return
	}

	diskSize, err := disks.NewDiskSizeFromGB(req.SizeGB)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "invalid_size", err.Error())
		return
	}

	disk, err := h.diskService.CreateDisk(
		r.Context(),
		diskName,
		projectID,
		diskSize,
		req.MountPath,
		disks.Filesystem(req.Filesystem),
		req.Persistent,
	)
	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, "creation_failed", "Failed to create disk")
		return
	}

	disk.ChangeStatus(disks.DiskStatusAvailable)

	utils.SendJSON(w, http.StatusCreated, toDiskResponse(disk))
}

func (h *DiskHandler) DeleteDisk(w http.ResponseWriter, r *http.Request) {
	projectIDStr := chi.URLParam(r, "project_id")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "invalid_project_id", "Invalid project ID")
		return
	}

	diskIDStr := chi.URLParam(r, "disk_id")
	diskID, err := disks.DiskIDFromString(diskIDStr)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "invalid_disk_id", "Invalid disk ID")
		return
	}

	disk, err := h.diskService.GetDisk(r.Context(), diskID)
	if err != nil {
		utils.SendError(w, http.StatusNotFound, "disk_not_found", "Disk not found")
		return
	}

	if disk.ProjectID() != projectID {
		utils.SendError(w, http.StatusNotFound, "disk_not_found", "Disk not found in project")
		return
	}

	if err := h.diskService.DeleteDisk(r.Context(), diskID); err != nil {
		utils.SendError(w, http.StatusBadRequest, "deletion_failed", err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *DiskHandler) AttachDisk(w http.ResponseWriter, r *http.Request) {
	projectIDStr := chi.URLParam(r, "project_id")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "invalid_project_id", "Invalid project ID")
		return
	}

	diskIDStr := chi.URLParam(r, "disk_id")
	diskID, err := disks.DiskIDFromString(diskIDStr)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "invalid_disk_id", "Invalid disk ID")
		return
	}

	var req AttachDiskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.SendError(w, http.StatusBadRequest, "invalid_request", "Invalid request body")
		return
	}

	if err := h.validator.Struct(req); err != nil {
		utils.SendError(w, http.StatusBadRequest, "validation_failed", err.Error())
		return
	}

	disk, err := h.diskService.GetDisk(r.Context(), diskID)
	if err != nil {
		utils.SendError(w, http.StatusNotFound, "disk_not_found", "Disk not found")
		return
	}

	if disk.ProjectID() != projectID {
		utils.SendError(w, http.StatusNotFound, "disk_not_found", "Disk not found in project")
		return
	}

	serviceID, err := uuid.Parse(req.ServiceID)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "invalid_service_id", "Invalid service ID")
		return
	}

	if err := h.diskService.AttachDisk(r.Context(), diskID, serviceID); err != nil {
		utils.SendError(w, http.StatusBadRequest, "attach_failed", err.Error())
		return
	}

	databaseID, err := databases.DatabaseIDFromString(serviceID.String())
	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, "restart_failed", "Failed to parse database ID")
		return
	}

	database, err := h.databaseService.GetDatabase(r.Context(), databaseID)
	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, "restart_failed", "Failed to get database for restart")
		return
	}

	if err := h.deploymentService.Restart(r.Context(), database); err != nil {
		utils.SendError(w, http.StatusInternalServerError, "restart_failed", "Failed to restart container with new volume")
		return
	}

	disk, _ = h.diskService.GetDisk(r.Context(), diskID)
	utils.SendJSON(w, http.StatusOK, toDiskResponse(disk))
}

func (h *DiskHandler) DetachDisk(w http.ResponseWriter, r *http.Request) {
	projectIDStr := chi.URLParam(r, "project_id")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "invalid_project_id", "Invalid project ID")
		return
	}

	diskIDStr := chi.URLParam(r, "disk_id")
	diskID, err := disks.DiskIDFromString(diskIDStr)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "invalid_disk_id", "Invalid disk ID")
		return
	}

	disk, err := h.diskService.GetDisk(r.Context(), diskID)
	if err != nil {
		utils.SendError(w, http.StatusNotFound, "disk_not_found", "Disk not found")
		return
	}

	if disk.ProjectID() != projectID {
		utils.SendError(w, http.StatusNotFound, "disk_not_found", "Disk not found in project")
		return
	}

	if disk.ServiceID() == nil {
		utils.SendError(w, http.StatusBadRequest, "not_attached", "Disk is not attached to any service")
		return
	}

	serviceID := *disk.ServiceID()

	if err := h.diskService.DetachDisk(r.Context(), diskID); err != nil {
		utils.SendError(w, http.StatusBadRequest, "detach_failed", err.Error())
		return
	}

	databaseID, err := databases.DatabaseIDFromString(serviceID.String())
	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, "restart_failed", "Failed to parse database ID")
		return
	}

	database, err := h.databaseService.GetDatabase(r.Context(), databaseID)
	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, "restart_failed", "Failed to get database for restart")
		return
	}

	if err := h.deploymentService.Restart(r.Context(), database); err != nil {
		utils.SendError(w, http.StatusInternalServerError, "restart_failed", "Failed to restart container without volume")
		return
	}

	disk, _ = h.diskService.GetDisk(r.Context(), diskID)
	utils.SendJSON(w, http.StatusOK, toDiskResponse(disk))
}

func (h *DiskHandler) ResizeDisk(w http.ResponseWriter, r *http.Request) {
	projectIDStr := chi.URLParam(r, "project_id")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "invalid_project_id", "Invalid project ID")
		return
	}

	diskIDStr := chi.URLParam(r, "disk_id")
	diskID, err := disks.DiskIDFromString(diskIDStr)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "invalid_disk_id", "Invalid disk ID")
		return
	}

	var req UpdateDiskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.SendError(w, http.StatusBadRequest, "invalid_request", "Invalid request body")
		return
	}

	if err := h.validator.Struct(req); err != nil {
		utils.SendError(w, http.StatusBadRequest, "validation_failed", err.Error())
		return
	}

	if req.SizeGB == nil {
		utils.SendError(w, http.StatusBadRequest, "missing_size", "Size is required")
		return
	}

	disk, err := h.diskService.GetDisk(r.Context(), diskID)
	if err != nil {
		utils.SendError(w, http.StatusNotFound, "disk_not_found", "Disk not found")
		return
	}

	if disk.ProjectID() != projectID {
		utils.SendError(w, http.StatusNotFound, "disk_not_found", "Disk not found in project")
		return
	}

	newSize, err := disks.NewDiskSizeFromGB(*req.SizeGB)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "invalid_size", err.Error())
		return
	}

	if err := h.diskService.ResizeDisk(r.Context(), diskID, newSize); err != nil {
		utils.SendError(w, http.StatusBadRequest, "resize_failed", err.Error())
		return
	}

	disk, _ = h.diskService.GetDisk(r.Context(), diskID)
	utils.SendJSON(w, http.StatusOK, toDiskResponse(disk))
}
