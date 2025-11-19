package handlers

import (
	"database/sql"
	"net/http"
	"runtime"
	"time"

	applicationsRepo "github.com/mikrocloud/mikrocloud/internal/domain/applications/repository"
	databasesRepo "github.com/mikrocloud/mikrocloud/internal/domain/databases/repository"
	projectsRepo "github.com/mikrocloud/mikrocloud/internal/domain/projects/repository"
	servicesRepo "github.com/mikrocloud/mikrocloud/internal/domain/services/repository"
	"github.com/mikrocloud/mikrocloud/internal/utils"
	"github.com/mikrocloud/mikrocloud/pkg/containers/manager"
)

type MaintenanceHandler struct {
	projectRepo     projectsRepo.Repository
	applicationRepo applicationsRepo.Repository
	databaseRepo    databasesRepo.DatabaseRepository
	serviceRepo     servicesRepo.TemplateRepository
	mainDB          *sql.DB
	containerMgr    manager.ContainerManager
}

func NewMaintenanceHandler(
	projectRepo projectsRepo.Repository,
	applicationRepo applicationsRepo.Repository,
	databaseRepo databasesRepo.DatabaseRepository,
	serviceRepo servicesRepo.TemplateRepository,
	mainDB *sql.DB,
	containerMgr manager.ContainerManager,
) *MaintenanceHandler {
	return &MaintenanceHandler{
		projectRepo:     projectRepo,
		applicationRepo: applicationRepo,
		databaseRepo:    databaseRepo,
		serviceRepo:     serviceRepo,
		mainDB:          mainDB,
		containerMgr:    containerMgr,
	}
}

type HealthCheckResponse struct {
	Status    string `json:"status"`
	Service   string `json:"service"`
	Version   string `json:"version"`
	Timestamp string `json:"timestamp"`
}

func (h *MaintenanceHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	resp := HealthCheckResponse{
		Status:    "ok",
		Service:   "mikrocloud",
		Version:   "0.1.0",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
	utils.SendJSON(w, http.StatusOK, resp)
}

type SystemStatusResponse struct {
	Body struct {
		Status     string `json:"status" example:"healthy"`
		Components struct {
			Database  string `json:"database" example:"ok"`
			Storage   string `json:"storage" example:"ok"`
			Container string `json:"container" example:"ok"`
		} `json:"components"`
		Timestamp string `json:"timestamp" example:"2024-01-01T00:00:00Z"`
	}
}

type SystemInfoResponse struct {
	Body struct {
		Version     string `json:"version" example:"0.1.0"`
		Platform    string `json:"platform" example:"linux/amd64"`
		GoVersion   string `json:"go_version" example:"1.21"`
		BuildTime   string `json:"build_time,omitempty" example:"2024-01-01T00:00:00Z"`
		Environment string `json:"environment" example:"production"`
	}
}

type ResourcesResponse struct {
	Body struct {
		Projects     int `json:"projects" example:"5"`
		Applications int `json:"applications" example:"10"`
		Databases    int `json:"databases" example:"3"`
		Services     int `json:"services" example:"8"`
	}
}

func (h *MaintenanceHandler) SystemStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	status := "healthy"
	components := struct {
		Database  string `json:"database"`
		Storage   string `json:"storage"`
		Container string `json:"container"`
	}{}

	if err := h.mainDB.PingContext(ctx); err != nil {
		components.Database = "error"
		status = "degraded"
	} else {
		components.Database = "ok"
	}

	components.Storage = "ok"

	if _, err := h.containerMgr.List(ctx); err != nil {
		components.Container = "error"
		status = "degraded"
	} else {
		components.Container = "ok"
	}

	resp := map[string]interface{}{
		"status":     status,
		"components": components,
		"timestamp":  time.Now().UTC().Format(time.RFC3339),
	}
	utils.SendJSON(w, http.StatusOK, resp)
}

func (h *MaintenanceHandler) GetResources(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	resp := map[string]int{
		"projects":     0,
		"applications": 0,
		"databases":    0,
		"services":     0,
	}

	projects, err := h.projectRepo.FindAll(ctx)
	if err == nil {
		resp["projects"] = len(projects)
	}

	applications, err := h.applicationRepo.FindAll(ctx)
	if err == nil {
		resp["applications"] = len(applications)
	}

	databases, err := h.databaseRepo.ListAllWithContainers()
	if err == nil {
		resp["databases"] = len(databases)
	}

	templates, err := h.serviceRepo.List()
	if err == nil {
		resp["services"] = len(templates)
	}

	utils.SendJSON(w, http.StatusOK, resp)
}

func (h *MaintenanceHandler) SystemInfo(w http.ResponseWriter, r *http.Request) {
	resp := map[string]string{
		"version":     "0.1.0",
		"platform":    runtime.GOOS + "/" + runtime.GOARCH,
		"go_version":  runtime.Version(),
		"environment": "production",
	}
	utils.SendJSON(w, http.StatusOK, resp)
}

type DomainListResponse struct {
	Domains []DomainInfo `json:"domains"`
	Total   int          `json:"total"`
}

type DomainInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Verified    bool   `json:"verified"`
	SSLEnabled  bool   `json:"ssl_enabled"`
	SSLExpiry   string `json:"ssl_expiry,omitempty"`
	ServiceID   string `json:"service_id,omitempty"`
	ServiceName string `json:"service_name,omitempty"`
	CreatedAt   string `json:"created_at"`
}

type AddDomainRequest struct {
	Name      string `json:"name"`
	ServiceID string `json:"service_id,omitempty"`
}

type EnableSSLRequest struct {
	Provider string `json:"provider"`
	Email    string `json:"email,omitempty"`
}

func (h *MaintenanceHandler) ListDomains(w http.ResponseWriter, r *http.Request) {
	resp := DomainListResponse{
		Domains: []DomainInfo{},
		Total:   0,
	}
	utils.SendJSON(w, http.StatusOK, resp)
}

func (h *MaintenanceHandler) AddDomain(w http.ResponseWriter, r *http.Request) {
	utils.SendError(w, http.StatusNotImplemented, "not_implemented", "Domain management infrastructure not yet available. Database schema and repositories need to be created first.")
}

func (h *MaintenanceHandler) RemoveDomain(w http.ResponseWriter, r *http.Request) {
	utils.SendError(w, http.StatusNotImplemented, "not_implemented", "Domain management infrastructure not yet available. Database schema and repositories need to be created first.")
}

func (h *MaintenanceHandler) EnableSSL(w http.ResponseWriter, r *http.Request) {
	utils.SendError(w, http.StatusNotImplemented, "not_implemented", "SSL management infrastructure not yet available. Domain tables and certificate management need to be implemented first.")
}
