package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/mikrocloud/mikrocloud/internal/api/middleware"
	"github.com/mikrocloud/mikrocloud/internal/domain/projects/service"
	"github.com/mikrocloud/mikrocloud/internal/utils"
)

// ProjectHandler handles project-related HTTP requests
type ProjectHandler struct {
	projectService *service.ProjectService
	validator      *validator.Validate
}

// NewProjectHandler creates a new project handler
func NewProjectHandler(pgs *service.ProjectService) *ProjectHandler {
	return &ProjectHandler{
		projectService: pgs,
		validator:      validator.New(),
	}
}

// ProjectResponse represents a project in API responses
type ProjectResponse struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

type CreateProjectRequest struct {
	Name        string  `json:"name" validate:"required,min=1,max=100"`
	Description *string `json:"description,omitempty"`
}

type UpdateProjectRequest struct {
	Name        *string `json:"name,omitempty" validate:"omitempty,min=1,max=100"`
	Description *string `json:"description,omitempty"`
}

type ProjectListItem struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
	CreatedAt   string  `json:"created_at"`
}

type ListProjectsResponse struct {
	Projects []ProjectListItem `json:"projects"`
}

// CreateProject creates a new project
func (h *ProjectHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateProjectRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.SendError(w, http.StatusBadRequest, "invalid_json", "Invalid JSON format")
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		utils.SendError(w, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	userID := middleware.GetUserID(r)
	orgID := middleware.GetOrgID(r)

	cmd := service.CreateProjectCommand{
		Name:           req.Name,
		Description:    req.Description,
		UserID:         userID,
		OrganisationID: orgID,
	}

	proj, err := h.projectService.CreateProject(r.Context(), cmd)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "create_failed", "Failed to create project: "+err.Error())
		return
	}

	response := ProjectResponse{
		ID:          proj.ID().String(),
		Name:        proj.Name().String(),
		Description: proj.Description(),
		CreatedAt:   proj.CreatedAt().Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   proj.UpdatedAt().Format("2006-01-02T15:04:05Z"),
	}

	utils.SendJSON(w, http.StatusCreated, response)
}

// List lists all projects
func (h *ProjectHandler) List(w http.ResponseWriter, r *http.Request) {
	projects, err := h.projectService.ListProjects(r.Context())
	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, "list_failed", "Failed to list projects: "+err.Error())
		return
	}

	items := make([]ProjectListItem, len(projects))

	for i, proj := range projects {
		items[i] = ProjectListItem{
			ID:          proj.ID().String(),
			Name:        proj.Name().String(),
			Description: proj.Description(),
			CreatedAt:   proj.CreatedAt().Format("2006-01-02T15:04:05Z"),
		}
	}

	response := ListProjectsResponse{
		Projects: items,
	}

	utils.SendJSON(w, http.StatusOK, response)
}

// Get retrieves a specific project
func (h *ProjectHandler) Get(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "project_id")
	if projectID == "" {
		utils.SendError(w, http.StatusBadRequest, "missing_parameter", "Project ID is required")
		return
	}

	proj, err := h.projectService.GetProject(r.Context(), projectID)
	if err != nil {
		utils.SendError(w, http.StatusNotFound, "not_found", "Project not found: "+err.Error())
		return
	}

	response := ProjectResponse{
		ID:          proj.ID().String(),
		Name:        proj.Name().String(),
		Description: proj.Description(),
		CreatedAt:   proj.CreatedAt().Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   proj.UpdatedAt().Format("2006-01-02T15:04:05Z"),
	}

	utils.SendJSON(w, http.StatusOK, response)
}

// Update updates a specific project
func (h *ProjectHandler) Update(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "project_id")
	if projectID == "" {
		utils.SendError(w, http.StatusBadRequest, "missing_parameter", "Project ID is required")
		return
	}

	var req UpdateProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.SendError(w, http.StatusBadRequest, "invalid_json", "Invalid JSON format")
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		utils.SendError(w, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	proj, err := h.projectService.UpdateProject(r.Context(), projectID, req.Description)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "update_failed", "Failed to update project: "+err.Error())
		return
	}

	response := ProjectResponse{
		ID:          proj.ID().String(),
		Name:        proj.Name().String(),
		Description: proj.Description(),
		CreatedAt:   proj.CreatedAt().Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   proj.UpdatedAt().Format("2006-01-02T15:04:05Z"),
	}

	utils.SendJSON(w, http.StatusOK, response)
}

// Delete deletes a project and all its environments and services
func (h *ProjectHandler) Delete(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "project_id")
	if projectID == "" {
		utils.SendError(w, http.StatusBadRequest, "missing_parameter", "Project ID is required")
		return
	}

	err := h.projectService.DeleteProject(r.Context(), projectID)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "delete_failed", "Failed to delete project: "+err.Error())
		return
	}

	response := utils.SuccessResponse{
		Message: "Project deleted successfully",
	}

	utils.SendJSON(w, http.StatusOK, response)
}
