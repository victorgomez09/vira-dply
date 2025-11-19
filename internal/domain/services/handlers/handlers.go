package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/mikrocloud/mikrocloud/internal/domain/applications"
	"github.com/mikrocloud/mikrocloud/internal/domain/services"
	servicesService "github.com/mikrocloud/mikrocloud/internal/domain/services/service"
	"github.com/mikrocloud/mikrocloud/internal/utils"
)

type TemplateHandler struct {
	templateService *servicesService.TemplateService
	validator       *validator.Validate
}

func NewTemplateHandler(templateService *servicesService.TemplateService) *TemplateHandler {
	return &TemplateHandler{
		templateService: templateService,
		validator:       validator.New(),
	}
}

// Template Management DTOs

type CreateTemplateRequest struct {
	Name        string            `json:"name" validate:"required,min=1,max=50"`
	Description string            `json:"description" validate:"required,min=1,max=500"`
	Category    string            `json:"category" validate:"required"`
	Version     string            `json:"version" validate:"required"`
	GitURL      string            `json:"git_url,omitempty"`
	GitBranch   string            `json:"git_branch,omitempty"`
	ContextRoot string            `json:"context_root,omitempty"`
	Environment map[string]string `json:"environment,omitempty"`
	Ports       []PortRequest     `json:"ports,omitempty"`
	Volumes     []VolumeRequest   `json:"volumes,omitempty"`
}

type PortRequest struct {
	ContainerPort int    `json:"container_port" validate:"required,min=1,max=65535"`
	HostPort      int    `json:"host_port,omitempty"`
	Protocol      string `json:"protocol,omitempty"`
	Name          string `json:"name,omitempty"`
}

type VolumeRequest struct {
	ContainerPath string `json:"container_path" validate:"required"`
	HostPath      string `json:"host_path,omitempty"`
	Type          string `json:"type,omitempty"`
	Name          string `json:"name,omitempty"`
}

type UpdateTemplateRequest struct {
	Version     string            `json:"version,omitempty"`
	Environment map[string]string `json:"environment,omitempty"`
	Ports       []PortRequest     `json:"ports,omitempty"`
	Volumes     []VolumeRequest   `json:"volumes,omitempty"`
}

// Deployment DTOs

type DeployTemplateRequest struct {
	Name          string                 `json:"name" validate:"required,min=1,max=50"`
	ProjectID     string                 `json:"project_id" validate:"required,uuid"`
	EnvironmentID string                 `json:"environment_id" validate:"required,uuid"`
	Environment   map[string]string      `json:"environment,omitempty"`
	CustomConfig  map[string]interface{} `json:"custom_config,omitempty"`
}

type PreviewDeploymentRequest struct {
	Name          string                 `json:"name" validate:"required,min=1,max=50"`
	ProjectID     string                 `json:"project_id" validate:"required,uuid"`
	EnvironmentID string                 `json:"environment_id" validate:"required,uuid"`
	Environment   map[string]string      `json:"environment,omitempty"`
	CustomConfig  map[string]interface{} `json:"custom_config,omitempty"`
}

// Response DTOs

type TemplateResponse struct {
	ID          string                    `json:"id"`
	Name        string                    `json:"name"`
	Description string                    `json:"description"`
	Category    services.TemplateCategory `json:"category"`
	Version     string                    `json:"version"`
	GitURL      *applications.GitURL      `json:"git_url"`
	Environment map[string]string         `json:"environment"`
	Ports       []services.Port           `json:"ports"`
	Volumes     []services.Volume         `json:"volumes"`
	Official    bool                      `json:"official"`
	CreatedAt   string                    `json:"created_at,omitempty"`
	UpdatedAt   string                    `json:"updated_at,omitempty"`
}

type TemplateListResponse struct {
	Templates []TemplateResponse `json:"templates"`
	Count     int                `json:"count"`
}

type DeploymentResponse struct {
	ApplicationID string `json:"application_id"`
	Status        string `json:"status"`
	Message       string `json:"message"`
}

type DeploymentPreviewResponse struct {
	Preview *services.DeploymentPreview `json:"preview"`
}

// Template Management Handlers

func (h *TemplateHandler) CreateTemplate(w http.ResponseWriter, r *http.Request) {
	var req CreateTemplateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.SendError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	if err := h.validator.Struct(req); err != nil {
		utils.SendError(w, http.StatusBadRequest, "Validation failed", err.Error())
		return
	}

	// Create template name
	templateName, err := services.NewTemplateName(req.Name)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "Invalid template name", err.Error())
		return
	}

	// Parse category
	category := services.TemplateCategory(req.Category)

	// Create Git URL if provided
	var gitURL *applications.GitURL
	if req.GitURL != "" {
		gitURL, err = applications.NewGitURL(req.GitURL, req.GitBranch, req.ContextRoot)
		if err != nil {
			utils.SendError(w, http.StatusBadRequest, "Invalid Git URL", err.Error())
			return
		}
	}

	// Convert ports
	ports := make([]services.Port, len(req.Ports))
	for i, p := range req.Ports {
		ports[i] = services.Port{
			Port:     p.ContainerPort,
			Protocol: p.Protocol,
			Public:   p.HostPort > 0,
		}
	}

	// Convert volumes
	volumes := make([]services.Volume, len(req.Volumes))
	for i, v := range req.Volumes {
		volumes[i] = services.Volume{
			Name:      v.Name,
			MountPath: v.ContainerPath,
			Size:      "",
			ReadOnly:  false,
		}
	}

	// Create template
	template := services.NewServiceTemplate(
		templateName,
		req.Description,
		category,
		req.Version,
		gitURL,
		nil,   // Build config would be set later or derived from Git URL
		false, // Not official by default
	)

	// Set environment variables
	if req.Environment != nil {
		for k, v := range req.Environment {
			if err := template.SetEnvironmentVariable(k, v); err != nil {
				utils.SendError(w, http.StatusBadRequest, "invalid_environment_variable", fmt.Sprintf("Invalid environment variable %s: %v", k, err))
				return
			}
		}
	}

	// Set ports and volumes
	for _, port := range ports {
		template.AddPort(port)
	}
	for _, volume := range volumes {
		template.AddVolume(volume)
	}

	// Save template
	if err := h.templateService.CreateTemplate(r.Context(), template); err != nil {
		utils.SendError(w, http.StatusInternalServerError, "Failed to create template", err.Error())
		return
	}

	response := h.convertToTemplateResponse(template)
	utils.SendJSON(w, http.StatusCreated, response)
}

func (h *TemplateHandler) GetTemplate(w http.ResponseWriter, r *http.Request) {
	templateIDStr := chi.URLParam(r, "id")
	templateID, err := services.TemplateIDFromString(templateIDStr)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "Invalid template ID", err.Error())
		return
	}

	template, err := h.templateService.GetTemplate(r.Context(), templateID)
	if err != nil {
		utils.SendError(w, http.StatusNotFound, "Template not found", err.Error())
		return
	}

	response := h.convertToTemplateResponse(template)
	utils.SendJSON(w, http.StatusOK, response)
}

func (h *TemplateHandler) UpdateTemplate(w http.ResponseWriter, r *http.Request) {
	templateIDStr := chi.URLParam(r, "id")
	templateID, err := services.TemplateIDFromString(templateIDStr)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "Invalid template ID", err.Error())
		return
	}

	var req UpdateTemplateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.SendError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	if err := h.validator.Struct(req); err != nil {
		utils.SendError(w, http.StatusBadRequest, "Validation failed", err.Error())
		return
	}

	// Get existing template
	template, err := h.templateService.GetTemplate(r.Context(), templateID)
	if err != nil {
		utils.SendError(w, http.StatusNotFound, "Template not found", err.Error())
		return
	}

	// Update version if provided
	if req.Version != "" {
		template.UpdateVersion(req.Version)
	}

	// Update environment variables
	if req.Environment != nil {
		for k, v := range req.Environment {
			if err := template.SetEnvironmentVariable(k, v); err != nil {
				utils.SendError(w, http.StatusBadRequest, "invalid_environment_variable", fmt.Sprintf("Invalid environment variable %s: %v", k, err))
				return
			}
		}
	}

	// Add new ports
	if req.Ports != nil {
		for _, p := range req.Ports {
			port := services.Port{
				Port:     p.ContainerPort,
				Protocol: p.Protocol,
				Public:   p.HostPort > 0,
			}
			template.AddPort(port)
		}
	}

	// Add new volumes
	if req.Volumes != nil {
		for _, v := range req.Volumes {
			volume := services.Volume{
				Name:      v.Name,
				MountPath: v.ContainerPath,
				Size:      "",
				ReadOnly:  false,
			}
			template.AddVolume(volume)
		}
	}

	// Save changes
	if err := h.templateService.UpdateTemplate(r.Context(), template); err != nil {
		utils.SendError(w, http.StatusInternalServerError, "Failed to update template", err.Error())
		return
	}

	response := h.convertToTemplateResponse(template)
	utils.SendJSON(w, http.StatusOK, response)
}

func (h *TemplateHandler) DeleteTemplate(w http.ResponseWriter, r *http.Request) {
	templateIDStr := chi.URLParam(r, "id")
	templateID, err := services.TemplateIDFromString(templateIDStr)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "Invalid template ID", err.Error())
		return
	}

	if err := h.templateService.DeleteTemplate(r.Context(), templateID); err != nil {
		utils.SendError(w, http.StatusInternalServerError, "Failed to delete template", err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *TemplateHandler) ListTemplates(w http.ResponseWriter, r *http.Request) {
	category := r.URL.Query().Get("category")

	var templates []*services.ServiceTemplate
	var err error

	if category != "" {
		templateCategory := services.TemplateCategory(category)
		templates, err = h.templateService.ListTemplatesByCategory(r.Context(), templateCategory)
	} else {
		templates, err = h.templateService.ListTemplates(r.Context())
	}

	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, "Failed to list templates", err.Error())
		return
	}

	// Convert to response format
	response := TemplateListResponse{
		Templates: make([]TemplateResponse, len(templates)),
		Count:     len(templates),
	}

	for i, template := range templates {
		response.Templates[i] = h.convertToTemplateResponse(template)
	}

	utils.SendJSON(w, http.StatusOK, response)
}

func (h *TemplateHandler) ListOfficialTemplates(w http.ResponseWriter, r *http.Request) {
	templates, err := h.templateService.ListOfficialTemplates(r.Context())
	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, "Failed to list official templates", err.Error())
		return
	}

	// Convert to response format
	response := TemplateListResponse{
		Templates: make([]TemplateResponse, len(templates)),
		Count:     len(templates),
	}

	for i, template := range templates {
		response.Templates[i] = h.convertToTemplateResponse(template)
	}

	utils.SendJSON(w, http.StatusOK, response)
}

// Deployment Handlers

func (h *TemplateHandler) DeployTemplate(w http.ResponseWriter, r *http.Request) {
	templateIDStr := chi.URLParam(r, "id")
	templateID, err := services.TemplateIDFromString(templateIDStr)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "Invalid template ID", err.Error())
		return
	}

	var req DeployTemplateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.SendError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	if err := h.validator.Struct(req); err != nil {
		utils.SendError(w, http.StatusBadRequest, "Validation failed", err.Error())
		return
	}

	projectID, err := uuid.Parse(req.ProjectID)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "Invalid project ID", err.Error())
		return
	}

	environmentID, err := uuid.Parse(req.EnvironmentID)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "Invalid environment ID", err.Error())
		return
	}

	// Create deployment request
	deployReq := services.DeploymentRequest{
		ProjectID:     projectID,
		EnvironmentID: environmentID,
		Name:          req.Name,
		Environment:   req.Environment,
		CustomConfig:  req.CustomConfig,
	}

	// Deploy the template
	app, err := h.templateService.DeployTemplate(r.Context(), templateID, deployReq)
	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, "Failed to deploy template", err.Error())
		return
	}

	response := DeploymentResponse{
		ApplicationID: app.ID().String(),
		Status:        "created",
		Message:       "Application created successfully from template",
	}

	utils.SendJSON(w, http.StatusCreated, response)
}

func (h *TemplateHandler) PreviewDeployment(w http.ResponseWriter, r *http.Request) {
	templateIDStr := chi.URLParam(r, "id")
	templateID, err := services.TemplateIDFromString(templateIDStr)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "Invalid template ID", err.Error())
		return
	}

	var req PreviewDeploymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.SendError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	if err := h.validator.Struct(req); err != nil {
		utils.SendError(w, http.StatusBadRequest, "Validation failed", err.Error())
		return
	}

	projectID, err := uuid.Parse(req.ProjectID)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "Invalid project ID", err.Error())
		return
	}

	environmentID, err := uuid.Parse(req.EnvironmentID)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "Invalid environment ID", err.Error())
		return
	}

	// Create deployment request
	deployReq := services.DeploymentRequest{
		ProjectID:     projectID,
		EnvironmentID: environmentID,
		Name:          req.Name,
		Environment:   req.Environment,
		CustomConfig:  req.CustomConfig,
	}

	// Preview the deployment
	preview, err := h.templateService.PreviewDeployment(r.Context(), templateID, deployReq)
	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, "Failed to preview deployment", err.Error())
		return
	}

	response := DeploymentPreviewResponse{Preview: preview}
	utils.SendJSON(w, http.StatusOK, response)
}

// Helper function to convert domain model to response DTO
func (h *TemplateHandler) convertToTemplateResponse(template *services.ServiceTemplate) TemplateResponse {
	return TemplateResponse{
		ID:          template.ID().String(),
		Name:        template.Name().String(),
		Description: template.Description(),
		Category:    template.Category(),
		Version:     template.Version(),
		GitURL:      template.GitURL(),
		Environment: template.Environment(),
		Ports:       template.Ports(),
		Volumes:     template.Volumes(),
		Official:    template.IsOfficial(),
	}
}
