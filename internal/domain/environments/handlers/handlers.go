package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/mikrocloud/mikrocloud/internal/domain/environments"
	"github.com/mikrocloud/mikrocloud/internal/domain/environments/service"
	"github.com/mikrocloud/mikrocloud/internal/utils"
)

type EnvironmentHandler struct {
	envService *service.EnvironmentService
	validator  *validator.Validate
}

func NewEnvironmentHandler(envService *service.EnvironmentService) *EnvironmentHandler {
	return &EnvironmentHandler{
		envService: envService,
		validator:  validator.New(),
	}
}

// EnvironmentResponse represents an environment in API responses
type EnvironmentResponse struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	ProjectID    string            `json:"project_id"`
	Description  string            `json:"description"`
	IsProduction bool              `json:"is_production"`
	Variables    map[string]string `json:"variables"`
	CreatedAt    string            `json:"created_at"`
	UpdatedAt    string            `json:"updated_at"`
}

type CreateEnvironmentRequest struct {
	Name         string            `json:"name" validate:"required,min=1,max=50"`
	Description  string            `json:"description,omitempty"`
	IsProduction bool              `json:"is_production,omitempty"`
	Variables    map[string]string `json:"variables,omitempty"`
}

type UpdateEnvironmentRequest struct {
	Name         *string           `json:"name,omitempty" validate:"omitempty,min=1,max=50"`
	Description  *string           `json:"description,omitempty"`
	IsProduction *bool             `json:"is_production,omitempty"`
	Variables    map[string]string `json:"variables,omitempty"`
}

type EnvironmentListItem struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	ProjectID    string `json:"project_id"`
	Description  string `json:"description"`
	IsProduction bool   `json:"is_production"`
	CreatedAt    string `json:"created_at"`
}

type ListEnvironmentsResponse struct {
	Environments []EnvironmentListItem `json:"environments"`
}

// CreateEnvironment creates a new environment in a project
func (h *EnvironmentHandler) CreateEnvironment(w http.ResponseWriter, r *http.Request) {
	var req CreateEnvironmentRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.SendError(w, http.StatusBadRequest, "invalid_json", "Invalid JSON format")
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		utils.SendError(w, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	// Get project ID from URL
	projectIDStr := chi.URLParam(r, "project_id")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "invalid_project_id", "Invalid project ID")
		return
	}

	cmd := service.CreateEnvironmentCommand{
		Name:         req.Name,
		Description:  req.Description,
		IsProduction: req.IsProduction,
		Variables:    req.Variables,
		ProjectID:    projectID,
	}

	env, err := h.envService.CreateEnvironment(r.Context(), cmd)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "create_failed", "Failed to create environment: "+err.Error())
		return
	}

	response := EnvironmentResponse{
		ID:           env.ID().String(),
		Name:         env.Name().String(),
		ProjectID:    env.ProjectID().String(),
		Description:  env.Description(),
		IsProduction: env.IsProduction(),
		Variables:    env.Variables(),
		CreatedAt:    env.CreatedAt().Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:    env.UpdatedAt().Format("2006-01-02T15:04:05Z07:00"),
	}

	utils.SendJSON(w, http.StatusCreated, response)
}

// GetEnvironment retrieves a specific environment
func (h *EnvironmentHandler) GetEnvironment(w http.ResponseWriter, r *http.Request) {
	// Get environment ID from URL
	envIDStr := chi.URLParam(r, "environment_id")
	envID, err := environments.EnvironmentIDFromString(envIDStr)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "invalid_environment_id", "Invalid environment ID")
		return
	}

	// Get project ID from URL
	projectIDStr := chi.URLParam(r, "project_id")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "invalid_project_id", "Invalid project ID")
		return
	}

	env, err := h.envService.GetEnvironment(r.Context(), envID)
	if err != nil {
		utils.SendError(w, http.StatusNotFound, "environment_not_found", "Environment not found")
		return
	}

	// Verify environment belongs to the specified project
	if env.ProjectID() != projectID {
		utils.SendError(w, http.StatusNotFound, "environment_not_found", "Environment not found in project")
		return
	}

	response := EnvironmentResponse{
		ID:           env.ID().String(),
		Name:         env.Name().String(),
		ProjectID:    env.ProjectID().String(),
		Description:  env.Description(),
		IsProduction: env.IsProduction(),
		Variables:    env.Variables(),
		CreatedAt:    env.CreatedAt().Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:    env.UpdatedAt().Format("2006-01-02T15:04:05Z07:00"),
	}

	utils.SendJSON(w, http.StatusOK, response)
}

// ListEnvironments lists all environments for a project
func (h *EnvironmentHandler) ListEnvironments(w http.ResponseWriter, r *http.Request) {
	// Get project ID from URL
	projectIDStr := chi.URLParam(r, "project_id")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "invalid_project_id", "Invalid project ID")
		return
	}

	envs, err := h.envService.ListEnvironmentsByProject(r.Context(), projectID)
	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, "list_failed", "Failed to list environments: "+err.Error())
		return
	}

	items := make([]EnvironmentListItem, len(envs))
	for i, env := range envs {
		items[i] = EnvironmentListItem{
			ID:           env.ID().String(),
			Name:         env.Name().String(),
			ProjectID:    env.ProjectID().String(),
			Description:  env.Description(),
			IsProduction: env.IsProduction(),
			CreatedAt:    env.CreatedAt().Format("2006-01-02T15:04:05Z07:00"),
		}
	}

	response := ListEnvironmentsResponse{
		Environments: items,
	}

	utils.SendJSON(w, http.StatusOK, response)
}

// UpdateEnvironment updates an existing environment
func (h *EnvironmentHandler) UpdateEnvironment(w http.ResponseWriter, r *http.Request) {
	var req UpdateEnvironmentRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.SendError(w, http.StatusBadRequest, "invalid_json", "Invalid JSON format")
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		utils.SendError(w, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	// Get environment ID from URL
	envIDStr := chi.URLParam(r, "environment_id")
	envID, err := environments.EnvironmentIDFromString(envIDStr)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "invalid_environment_id", "Invalid environment ID")
		return
	}

	// Get project ID from URL
	projectIDStr := chi.URLParam(r, "project_id")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "invalid_project_id", "Invalid project ID")
		return
	}

	// First verify the environment exists and belongs to the project
	env, err := h.envService.GetEnvironment(r.Context(), envID)
	if err != nil {
		utils.SendError(w, http.StatusNotFound, "environment_not_found", "Environment not found")
		return
	}

	if env.ProjectID() != projectID {
		utils.SendError(w, http.StatusNotFound, "environment_not_found", "Environment not found in project")
		return
	}

	// Handle production status toggle separately if provided
	if req.IsProduction != nil {
		env, err = h.envService.ToggleProduction(r.Context(), envID, *req.IsProduction)
		if err != nil {
			utils.SendError(w, http.StatusBadRequest, "toggle_production_failed", "Failed to toggle production status: "+err.Error())
			return
		}
	}

	// Prepare description for update command
	description := env.Description()
	if req.Description != nil {
		description = *req.Description
	}

	// Handle other updates (description, variables) if any are provided
	if req.Description != nil || req.Variables != nil {
		cmd := service.UpdateEnvironmentCommand{
			ID:          envID,
			Description: description,
			Variables:   req.Variables,
		}

		env, err = h.envService.UpdateEnvironment(r.Context(), cmd)
		if err != nil {
			utils.SendError(w, http.StatusBadRequest, "update_failed", "Failed to update environment: "+err.Error())
			return
		}
	}

	response := EnvironmentResponse{
		ID:           env.ID().String(),
		Name:         env.Name().String(),
		ProjectID:    env.ProjectID().String(),
		Description:  env.Description(),
		IsProduction: env.IsProduction(),
		Variables:    env.Variables(),
		CreatedAt:    env.CreatedAt().Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:    env.UpdatedAt().Format("2006-01-02T15:04:05Z07:00"),
	}

	utils.SendJSON(w, http.StatusOK, response)
}

// DeleteEnvironment deletes an environment
func (h *EnvironmentHandler) DeleteEnvironment(w http.ResponseWriter, r *http.Request) {
	// Get environment ID from URL
	envIDStr := chi.URLParam(r, "environment_id")
	envID, err := environments.EnvironmentIDFromString(envIDStr)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "invalid_environment_id", "Invalid environment ID")
		return
	}

	// Get project ID from URL
	projectIDStr := chi.URLParam(r, "project_id")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "invalid_project_id", "Invalid project ID")
		return
	}

	// First verify the environment exists and belongs to the project
	env, err := h.envService.GetEnvironment(r.Context(), envID)
	if err != nil {
		utils.SendError(w, http.StatusNotFound, "environment_not_found", "Environment not found")
		return
	}

	if env.ProjectID() != projectID {
		utils.SendError(w, http.StatusNotFound, "environment_not_found", "Environment not found in project")
		return
	}

	err = h.envService.DeleteEnvironment(r.Context(), envID)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "delete_failed", "Failed to delete environment: "+err.Error())
		return
	}

	response := map[string]string{
		"message": "Environment deleted successfully",
	}

	utils.SendJSON(w, http.StatusOK, response)
}
