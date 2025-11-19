package handlers

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/mikrocloud/mikrocloud/internal/api/middleware"
	"github.com/mikrocloud/mikrocloud/internal/domain/applications"
	"github.com/mikrocloud/mikrocloud/internal/domain/applications/service"
	deploymentService "github.com/mikrocloud/mikrocloud/internal/domain/deployments/service"
	"github.com/mikrocloud/mikrocloud/internal/domain/users"
	"github.com/mikrocloud/mikrocloud/internal/utils"
	"github.com/mikrocloud/mikrocloud/pkg/containers/manager"
)

type ApplicationHandler struct {
	appService        *service.ApplicationService
	deploymentService *deploymentService.DeploymentService
	containerManager  manager.ContainerManager
	validator         *validator.Validate
}

func NewApplicationHandler(appService *service.ApplicationService, deploymentService *deploymentService.DeploymentService, containerManager manager.ContainerManager) *ApplicationHandler {
	return &ApplicationHandler{
		appService:        appService,
		deploymentService: deploymentService,
		containerManager:  containerManager,
		validator:         validator.New(),
	}
}

// ApplicationResponse represents an application in API responses
type ApplicationResponse struct {
	ID               string                         `json:"id"`
	Name             string                         `json:"name"`
	Description      string                         `json:"description"`
	ProjectID        string                         `json:"project_id"`
	EnvironmentID    string                         `json:"environment_id"`
	DeploymentSource applications.DeploymentSource  `json:"deployment_source"`
	Domain           string                         `json:"domain"`
	CustomDomain     string                         `json:"custom_domain"`
	GeneratedDomain  string                         `json:"generated_domain"`
	ExposedPorts     []int                          `json:"exposed_ports"`
	PortMappings     []applications.PortMapping     `json:"port_mappings"`
	Buildpack        applications.BuildpackConfig   `json:"buildpack"`
	EnvVars          map[string]string              `json:"env_vars"`
	AutoDeploy       bool                           `json:"auto_deploy"`
	Status           applications.ApplicationStatus `json:"status"`
	CreatedAt        string                         `json:"created_at"`
	UpdatedAt        string                         `json:"updated_at"`
}

type CreateApplicationRequest struct {
	Name             string                        `json:"name" validate:"required,min=1,max=100"`
	Description      string                        `json:"description,omitempty"`
	EnvironmentID    string                        `json:"environment_id" validate:"required,uuid"`
	DeploymentSource applications.DeploymentSource `json:"deployment_source" validate:"required"`
	Buildpack        applications.BuildpackConfig  `json:"buildpack" validate:"required"`
	EnvVars          map[string]string             `json:"env_vars,omitempty"`
}

type UpdateApplicationRequest struct {
	Description      *string                        `json:"description,omitempty"`
	DeploymentSource *applications.DeploymentSource `json:"deployment_source,omitempty"`
	Domain           *string                        `json:"domain,omitempty"`
	Buildpack        *applications.BuildpackConfig  `json:"buildpack,omitempty"`
	EnvVars          map[string]string              `json:"env_vars,omitempty"`
	AutoDeploy       *bool                          `json:"auto_deploy,omitempty"`
}

type ApplicationListItem struct {
	ID            string                         `json:"id"`
	Name          string                         `json:"name"`
	Description   string                         `json:"description"`
	ProjectID     string                         `json:"project_id"`
	EnvironmentID string                         `json:"environment_id"`
	Domain        string                         `json:"domain"`
	Status        applications.ApplicationStatus `json:"status"`
	CreatedAt     string                         `json:"created_at"`
}

type ListApplicationsResponse struct {
	Applications []ApplicationListItem `json:"applications"`
}

type DeployApplicationRequest struct {
	Action string `json:"action" validate:"required,oneof=deploy stop"`
}

type UpdateGeneralRequest struct {
	Name        *string `json:"name" validate:"omitempty,min=1,max=100"`
	Description *string `json:"description"`
}

type AssignDomainRequest struct {
	Domain string `json:"domain" validate:"required"`
}

type UpdatePortsRequest struct {
	ExposedPorts []int                      `json:"exposed_ports" validate:"required,min=1,dive,min=1,max=65535"`
	PortMappings []applications.PortMapping `json:"port_mappings" validate:"dive"`
}

// CreateApplication creates a new application in a project
func (h *ApplicationHandler) CreateApplication(w http.ResponseWriter, r *http.Request) {
	var req CreateApplicationRequest

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

	// Parse environment ID
	environmentID, err := uuid.Parse(req.EnvironmentID)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "invalid_environment_id", "Invalid environment ID")
		return
	}

	cmd := service.CreateApplicationCommand{
		Name:             req.Name,
		Description:      req.Description,
		ProjectID:        projectID,
		EnvironmentID:    environmentID,
		DeploymentSource: req.DeploymentSource,
		BuildpackConfig:  convertLegacyBuildpackConfig(req.Buildpack),
		EnvVars:          req.EnvVars,
	}

	app, err := h.appService.CreateApplication(r.Context(), cmd)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "create_failed", "Failed to create application: "+err.Error())
		return
	}

	response := ApplicationResponse{
		ID:               app.ID().String(),
		Name:             app.Name().String(),
		Description:      app.Description(),
		ProjectID:        app.ProjectID().String(),
		EnvironmentID:    app.EnvironmentID().String(),
		DeploymentSource: app.DeploymentSource(),
		Domain:           app.Domain(),
		CustomDomain:     app.Domain(),
		GeneratedDomain:  app.GeneratedDomain(),
		ExposedPorts:     app.ExposedPorts(),
		PortMappings:     app.PortMappings(),
		Buildpack:        convertToLegacyBuildpackConfig(app.Buildpack()),
		EnvVars:          app.EnvVars(),
		AutoDeploy:       app.AutoDeploy(),
		Status:           app.Status(),
		CreatedAt:        app.CreatedAt().Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:        app.UpdatedAt().Format("2006-01-02T15:04:05Z07:00"),
	}

	utils.SendJSON(w, http.StatusCreated, response)
}

// GetApplication retrieves a specific application
func (h *ApplicationHandler) GetApplication(w http.ResponseWriter, r *http.Request) {
	// Get application ID from URL
	appIDStr := chi.URLParam(r, "application_id")
	appID, err := applications.ApplicationIDFromString(appIDStr)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "invalid_application_id", "Invalid application ID")
		return
	}

	// Get project ID from URL
	projectIDStr := chi.URLParam(r, "project_id")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "invalid_project_id", "Invalid project ID")
		return
	}

	app, err := h.appService.GetApplication(r.Context(), appID)
	if err != nil {
		utils.SendError(w, http.StatusNotFound, "application_not_found", "Application not found")
		return
	}

	// Verify application belongs to the specified project
	if app.ProjectID() != projectID {
		utils.SendError(w, http.StatusNotFound, "application_not_found", "Application not found in project")
		return
	}

	response := ApplicationResponse{
		ID:               app.ID().String(),
		Name:             app.Name().String(),
		Description:      app.Description(),
		ProjectID:        app.ProjectID().String(),
		EnvironmentID:    app.EnvironmentID().String(),
		DeploymentSource: app.DeploymentSource(),
		Domain:           app.Domain(),
		CustomDomain:     app.Domain(),
		GeneratedDomain:  app.GeneratedDomain(),
		ExposedPorts:     app.ExposedPorts(),
		PortMappings:     app.PortMappings(),
		Buildpack:        convertToLegacyBuildpackConfig(app.Buildpack()),
		EnvVars:          app.EnvVars(),
		AutoDeploy:       app.AutoDeploy(),
		Status:           app.Status(),
		CreatedAt:        app.CreatedAt().Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:        app.UpdatedAt().Format("2006-01-02T15:04:05Z07:00"),
	}

	utils.SendJSON(w, http.StatusOK, response)
}

// ListApplications lists all applications for a project
func (h *ApplicationHandler) ListApplications(w http.ResponseWriter, r *http.Request) {
	// Get project ID from URL
	projectIDStr := chi.URLParam(r, "project_id")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "invalid_project_id", "Invalid project ID")
		return
	}

	apps, err := h.appService.ListApplicationsByProject(r.Context(), projectID)
	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, "list_failed", "Failed to list applications: "+err.Error())
		return
	}

	items := make([]ApplicationListItem, len(apps))
	for i, app := range apps {
		items[i] = ApplicationListItem{
			ID:            app.ID().String(),
			Name:          app.Name().String(),
			Description:   app.Description(),
			ProjectID:     app.ProjectID().String(),
			EnvironmentID: app.EnvironmentID().String(),
			Domain:        app.Domain(),
			Status:        app.Status(),
			CreatedAt:     app.CreatedAt().Format("2006-01-02T15:04:05Z07:00"),
		}
	}

	response := ListApplicationsResponse{
		Applications: items,
	}

	utils.SendJSON(w, http.StatusOK, response)
}

// UpdateApplication updates an existing application
func (h *ApplicationHandler) UpdateApplication(w http.ResponseWriter, r *http.Request) {
	var req UpdateApplicationRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.SendError(w, http.StatusBadRequest, "invalid_json", "Invalid JSON format")
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		utils.SendError(w, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	// Get application ID from URL
	appIDStr := chi.URLParam(r, "application_id")
	appID, err := applications.ApplicationIDFromString(appIDStr)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "invalid_application_id", "Invalid application ID")
		return
	}

	// Get project ID from URL
	projectIDStr := chi.URLParam(r, "project_id")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "invalid_project_id", "Invalid project ID")
		return
	}

	// First verify the application exists and belongs to the project
	app, err := h.appService.GetApplication(r.Context(), appID)
	if err != nil {
		utils.SendError(w, http.StatusNotFound, "application_not_found", "Application not found")
		return
	}

	if app.ProjectID() != projectID {
		utils.SendError(w, http.StatusNotFound, "application_not_found", "Application not found in project")
		return
	}

	cmd := service.UpdateApplicationCommand{
		ID:               appID,
		Description:      req.Description,
		DeploymentSource: req.DeploymentSource,
		Domain:           req.Domain,
		BuildpackConfig:  convertLegacyBuildpackConfigPtr(req.Buildpack),
		EnvVars:          req.EnvVars,
		AutoDeploy:       req.AutoDeploy,
	}

	updatedApp, err := h.appService.UpdateApplication(r.Context(), cmd)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "update_failed", "Failed to update application: "+err.Error())
		return
	}

	response := ApplicationResponse{
		ID:               updatedApp.ID().String(),
		Name:             updatedApp.Name().String(),
		Description:      updatedApp.Description(),
		ProjectID:        updatedApp.ProjectID().String(),
		EnvironmentID:    updatedApp.EnvironmentID().String(),
		DeploymentSource: updatedApp.DeploymentSource(),
		Domain:           updatedApp.Domain(),
		CustomDomain:     updatedApp.Domain(),
		GeneratedDomain:  updatedApp.GeneratedDomain(),
		ExposedPorts:     updatedApp.ExposedPorts(),
		PortMappings:     updatedApp.PortMappings(),
		Buildpack:        convertToLegacyBuildpackConfig(updatedApp.Buildpack()),
		EnvVars:          updatedApp.EnvVars(),
		AutoDeploy:       updatedApp.AutoDeploy(),
		Status:           updatedApp.Status(),
		CreatedAt:        updatedApp.CreatedAt().Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:        updatedApp.UpdatedAt().Format("2006-01-02T15:04:05Z07:00"),
	}

	utils.SendJSON(w, http.StatusOK, response)
}

// DeleteApplication deletes an application
func (h *ApplicationHandler) DeleteApplication(w http.ResponseWriter, r *http.Request) {
	// Get application ID from URL
	appIDStr := chi.URLParam(r, "application_id")
	appID, err := applications.ApplicationIDFromString(appIDStr)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "invalid_application_id", "Invalid application ID")
		return
	}

	// Get project ID from URL
	projectIDStr := chi.URLParam(r, "project_id")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "invalid_project_id", "Invalid project ID")
		return
	}

	// First verify the application exists and belongs to the project
	app, err := h.appService.GetApplication(r.Context(), appID)
	if err != nil {
		utils.SendError(w, http.StatusNotFound, "application_not_found", "Application not found")
		return
	}

	if app.ProjectID() != projectID {
		utils.SendError(w, http.StatusNotFound, "application_not_found", "Application not found in project")
		return
	}

	err = h.appService.DeleteApplication(r.Context(), appID)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "delete_failed", "Failed to delete application: "+err.Error())
		return
	}

	response := map[string]string{
		"message": "Application deleted successfully",
	}

	utils.SendJSON(w, http.StatusOK, response)
}

// DeployApplication handles application deployment actions
func (h *ApplicationHandler) DeployApplication(w http.ResponseWriter, r *http.Request) {
	var req DeployApplicationRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.SendError(w, http.StatusBadRequest, "invalid_json", "Invalid JSON format")
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		utils.SendError(w, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	// Get application ID from URL
	appIDStr := chi.URLParam(r, "application_id")
	appID, err := applications.ApplicationIDFromString(appIDStr)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "invalid_application_id", "Invalid application ID")
		return
	}

	// Get project ID from URL
	projectIDStr := chi.URLParam(r, "project_id")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "invalid_project_id", "Invalid project ID")
		return
	}

	// First verify the application exists and belongs to the project
	app, err := h.appService.GetApplication(r.Context(), appID)
	if err != nil {
		utils.SendError(w, http.StatusNotFound, "application_not_found", "Application not found")
		return
	}

	if app.ProjectID() != projectID {
		utils.SendError(w, http.StatusNotFound, "application_not_found", "Application not found in project")
		return
	}

	switch req.Action {
	case "deploy":
		err = h.appService.StartDeployment(r.Context(), appID)
		if err != nil {
			utils.SendError(w, http.StatusBadRequest, "deploy_failed", "Failed to start deployment: "+err.Error())
			return
		}
	case "stop":
		err = h.appService.StopApplication(r.Context(), appID)
		if err != nil {
			utils.SendError(w, http.StatusBadRequest, "stop_failed", "Failed to stop application: "+err.Error())
			return
		}
	default:
		utils.SendError(w, http.StatusBadRequest, "invalid_action", "Invalid deployment action")
		return
	}

	response := map[string]string{
		"message": "Deployment action completed successfully",
		"action":  req.Action,
	}

	utils.SendJSON(w, http.StatusOK, response)
}

func (h *ApplicationHandler) StartApplication(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userIDStr := middleware.GetUserID(r)
	if userIDStr == "" {
		utils.SendError(w, http.StatusUnauthorized, "unauthorized", "User not authenticated")
		return
	}

	userID, err := users.UserIDFromString(userIDStr)
	if err != nil {
		utils.SendError(w, http.StatusUnauthorized, "invalid_user", "Invalid user ID")
		return
	}

	appIDStr := chi.URLParam(r, "application_id")
	appID, err := applications.ApplicationIDFromString(appIDStr)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "invalid_application_id", "Invalid application ID")
		return
	}

	projectIDStr := chi.URLParam(r, "project_id")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "invalid_project_id", "Invalid project ID")
		return
	}

	app, err := h.appService.GetApplication(r.Context(), appID)
	if err != nil {
		utils.SendError(w, http.StatusNotFound, "application_not_found", "Application not found")
		return
	}

	if app.ProjectID() != projectID {
		utils.SendError(w, http.StatusNotFound, "application_not_found", "Application not found in project")
		return
	}

	// Create deployment command for manual start
	cmd := deploymentService.CreateDeploymentCommand{
		ApplicationID: appID,
		IsProduction:  false,
		TriggeredBy:   &userID,
		TriggerType:   "manual",
		ImageTag:      generateImageTag(app, ""),
	}

	// Create and execute deployment
	_, err = h.deploymentService.CreateAndExecuteDeployment(r.Context(), cmd, h.appService)
	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, "start_failed", "Failed to start application: "+err.Error())
		return
	}

	utils.SendJSON(w, http.StatusOK, map[string]string{"message": "Application deployment started successfully"})
}

func generateImageTag(app *applications.Application, gitCommitHash string) string {
	imageName := app.Name().String()
	if gitCommitHash != "" {
		if len(gitCommitHash) > 7 {
			return imageName + ":" + gitCommitHash[:7]
		}
		return imageName + ":" + gitCommitHash
	}
	return imageName + ":latest"
}

func (h *ApplicationHandler) StopApplication(w http.ResponseWriter, r *http.Request) {
	appIDStr := chi.URLParam(r, "application_id")
	appID, err := applications.ApplicationIDFromString(appIDStr)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "invalid_application_id", "Invalid application ID")
		return
	}

	projectIDStr := chi.URLParam(r, "project_id")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "invalid_project_id", "Invalid project ID")
		return
	}

	app, err := h.appService.GetApplication(r.Context(), appID)
	if err != nil {
		utils.SendError(w, http.StatusNotFound, "application_not_found", "Application not found")
		return
	}

	if app.ProjectID() != projectID {
		utils.SendError(w, http.StatusNotFound, "application_not_found", "Application not found in project")
		return
	}

	err = h.appService.StopApplication(r.Context(), appID)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "stop_failed", "Failed to stop application: "+err.Error())
		return
	}

	utils.SendJSON(w, http.StatusOK, map[string]string{"message": "Application stopped successfully"})
}

func (h *ApplicationHandler) RestartApplication(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userIDStr := middleware.GetUserID(r)
	if userIDStr == "" {
		utils.SendError(w, http.StatusUnauthorized, "unauthorized", "User not authenticated")
		return
	}

	userID, err := users.UserIDFromString(userIDStr)
	if err != nil {
		utils.SendError(w, http.StatusUnauthorized, "invalid_user", "Invalid user ID")
		return
	}

	appIDStr := chi.URLParam(r, "application_id")
	appID, err := applications.ApplicationIDFromString(appIDStr)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "invalid_application_id", "Invalid application ID")
		return
	}

	projectIDStr := chi.URLParam(r, "project_id")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "invalid_project_id", "Invalid project ID")
		return
	}

	app, err := h.appService.GetApplication(r.Context(), appID)
	if err != nil {
		utils.SendError(w, http.StatusNotFound, "application_not_found", "Application not found")
		return
	}

	if app.ProjectID() != projectID {
		utils.SendError(w, http.StatusNotFound, "application_not_found", "Application not found in project")
		return
	}

	// Stop the application first
	err = h.appService.StopApplication(r.Context(), appID)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "restart_failed", "Failed to stop application: "+err.Error())
		return
	}

	// Create deployment command for manual restart
	cmd := deploymentService.CreateDeploymentCommand{
		ApplicationID: appID,
		IsProduction:  false,
		TriggeredBy:   &userID,
		TriggerType:   "manual",
		ImageTag:      generateImageTag(app, ""),
	}

	// Create and execute new deployment
	_, err = h.deploymentService.CreateAndExecuteDeployment(r.Context(), cmd, h.appService)
	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, "restart_failed", "Failed to start application: "+err.Error())
		return
	}

	utils.SendJSON(w, http.StatusOK, map[string]string{"message": "Application redeployment started successfully"})
}

func (h *ApplicationHandler) GetApplicationLogs(w http.ResponseWriter, r *http.Request) {
	appIDStr := chi.URLParam(r, "application_id")
	appID, err := applications.ApplicationIDFromString(appIDStr)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "invalid_application_id", "Invalid application ID")
		return
	}

	projectIDStr := chi.URLParam(r, "project_id")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "invalid_project_id", "Invalid project ID")
		return
	}

	app, err := h.appService.GetApplication(r.Context(), appID)
	if err != nil {
		utils.SendError(w, http.StatusNotFound, "application_not_found", "Application not found")
		return
	}

	if app.ProjectID() != projectID {
		utils.SendError(w, http.StatusNotFound, "application_not_found", "Application not found in project")
		return
	}

	deployment, err := h.deploymentService.GetLatestDeploymentByApplication(r.Context(), appID)
	if err != nil {
		utils.SendError(w, http.StatusNotFound, "no_deployment_found", "No deployment found for application")
		return
	}

	if deployment.ContainerID() == "" {
		utils.SendError(w, http.StatusNotFound, "no_container_found", "No container found for application")
		return
	}

	follow := r.URL.Query().Get("follow") == "true"

	logStream, err := h.containerManager.StreamLogs(r.Context(), deployment.ContainerID(), follow)
	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, "logs_failed", "Failed to get container logs: "+err.Error())
		return
	}
	defer func() {
		_ = logStream.Close()
	}()

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	if follow {
		w.Header().Set("Transfer-Encoding", "chunked")
	}

	_, err = io.Copy(w, logStream)
	if err != nil {
		return
	}
}

// Helper functions to convert between legacy and new buildpack configs

// convertLegacyBuildpackConfig converts from legacy BuildpackConfig to new BuildConfig
func convertLegacyBuildpackConfig(legacy applications.BuildpackConfig) *applications.BuildConfig {
	return applications.NewLegacyBuildpackConfig(legacy.Type, legacy.Config)
}

// convertLegacyBuildpackConfigPtr converts from pointer to legacy BuildpackConfig to new BuildConfig
func convertLegacyBuildpackConfigPtr(legacy *applications.BuildpackConfig) *applications.BuildConfig {
	if legacy == nil {
		return nil
	}
	return applications.NewLegacyBuildpackConfig(legacy.Type, legacy.Config)
}

// convertToLegacyBuildpackConfig converts from new BuildConfig to legacy BuildpackConfig
func convertToLegacyBuildpackConfig(config *applications.BuildConfig) applications.BuildpackConfig {
	if config == nil {
		return applications.BuildpackConfig{
			Type:   applications.BuildpackTypeNixpacks,
			Config: nil,
		}
	}

	var configData interface{}
	switch config.BuildpackType() {
	case applications.BuildpackTypeNixpacks:
		configData = config.NixpacksConfig()
	case applications.BuildpackTypeStatic:
		configData = config.StaticConfig()
	case applications.BuildpackTypeDockerfile:
		configData = config.DockerfileConfig()
	case applications.BuildpackTypeDockerCompose:
		configData = config.ComposeConfig()
	}

	return applications.BuildpackConfig{
		Type:   config.BuildpackType(),
		Config: configData,
	}
}

func (h *ApplicationHandler) UpdateGeneral(w http.ResponseWriter, r *http.Request) {
	var req UpdateGeneralRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.SendError(w, http.StatusBadRequest, "invalid_json", "Invalid JSON format")
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		utils.SendError(w, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	appIDStr := chi.URLParam(r, "application_id")
	appID, err := applications.ApplicationIDFromString(appIDStr)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "invalid_application_id", "Invalid application ID")
		return
	}

	projectIDStr := chi.URLParam(r, "project_id")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "invalid_project_id", "Invalid project ID")
		return
	}

	app, err := h.appService.GetApplication(r.Context(), appID)
	if err != nil {
		utils.SendError(w, http.StatusNotFound, "application_not_found", "Application not found")
		return
	}

	if app.ProjectID() != projectID {
		utils.SendError(w, http.StatusNotFound, "application_not_found", "Application not found in project")
		return
	}

	cmd := service.UpdateGeneralCommand{
		ID:          appID,
		Name:        req.Name,
		Description: req.Description,
	}

	updatedApp, err := h.appService.UpdateGeneral(r.Context(), cmd)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "update_failed", "Failed to update application: "+err.Error())
		return
	}

	response := ApplicationResponse{
		ID:               updatedApp.ID().String(),
		Name:             updatedApp.Name().String(),
		Description:      updatedApp.Description(),
		ProjectID:        updatedApp.ProjectID().String(),
		EnvironmentID:    updatedApp.EnvironmentID().String(),
		DeploymentSource: updatedApp.DeploymentSource(),
		Domain:           updatedApp.Domain(),
		CustomDomain:     updatedApp.Domain(),
		GeneratedDomain:  updatedApp.GeneratedDomain(),
		ExposedPorts:     updatedApp.ExposedPorts(),
		PortMappings:     updatedApp.PortMappings(),
		Buildpack:        convertToLegacyBuildpackConfig(updatedApp.Buildpack()),
		EnvVars:          updatedApp.EnvVars(),
		AutoDeploy:       updatedApp.AutoDeploy(),
		Status:           updatedApp.Status(),
		CreatedAt:        updatedApp.CreatedAt().Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:        updatedApp.UpdatedAt().Format("2006-01-02T15:04:05Z07:00"),
	}

	utils.SendJSON(w, http.StatusOK, response)
}

func (h *ApplicationHandler) GenerateDomain(w http.ResponseWriter, r *http.Request) {
	appIDStr := chi.URLParam(r, "application_id")
	appID, err := applications.ApplicationIDFromString(appIDStr)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "invalid_application_id", "Invalid application ID")
		return
	}

	projectIDStr := chi.URLParam(r, "project_id")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "invalid_project_id", "Invalid project ID")
		return
	}

	app, err := h.appService.GetApplication(r.Context(), appID)
	if err != nil {
		utils.SendError(w, http.StatusNotFound, "application_not_found", "Application not found")
		return
	}

	if app.ProjectID() != projectID {
		utils.SendError(w, http.StatusNotFound, "application_not_found", "Application not found in project")
		return
	}

	domain, err := h.appService.GenerateDomain(r.Context(), appID)
	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, "generate_failed", "Failed to generate domain: "+err.Error())
		return
	}

	response := map[string]string{
		"domain": domain,
	}

	utils.SendJSON(w, http.StatusOK, response)
}

func (h *ApplicationHandler) AssignDomain(w http.ResponseWriter, r *http.Request) {
	var req AssignDomainRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.SendError(w, http.StatusBadRequest, "invalid_json", "Invalid JSON format")
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		utils.SendError(w, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	appIDStr := chi.URLParam(r, "application_id")
	appID, err := applications.ApplicationIDFromString(appIDStr)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "invalid_application_id", "Invalid application ID")
		return
	}

	projectIDStr := chi.URLParam(r, "project_id")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "invalid_project_id", "Invalid project ID")
		return
	}

	app, err := h.appService.GetApplication(r.Context(), appID)
	if err != nil {
		utils.SendError(w, http.StatusNotFound, "application_not_found", "Application not found")
		return
	}

	if app.ProjectID() != projectID {
		utils.SendError(w, http.StatusNotFound, "application_not_found", "Application not found in project")
		return
	}

	err = h.appService.AssignDomain(r.Context(), appID, req.Domain)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "assign_failed", "Failed to assign domain: "+err.Error())
		return
	}

	updatedApp, err := h.appService.GetApplication(r.Context(), appID)
	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, "fetch_failed", "Failed to fetch updated application")
		return
	}

	response := ApplicationResponse{
		ID:               updatedApp.ID().String(),
		Name:             updatedApp.Name().String(),
		Description:      updatedApp.Description(),
		ProjectID:        updatedApp.ProjectID().String(),
		EnvironmentID:    updatedApp.EnvironmentID().String(),
		DeploymentSource: updatedApp.DeploymentSource(),
		Domain:           updatedApp.Domain(),
		CustomDomain:     updatedApp.Domain(),
		GeneratedDomain:  updatedApp.GeneratedDomain(),
		ExposedPorts:     updatedApp.ExposedPorts(),
		PortMappings:     updatedApp.PortMappings(),
		Buildpack:        convertToLegacyBuildpackConfig(updatedApp.Buildpack()),
		EnvVars:          updatedApp.EnvVars(),
		AutoDeploy:       updatedApp.AutoDeploy(),
		Status:           updatedApp.Status(),
		CreatedAt:        updatedApp.CreatedAt().Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:        updatedApp.UpdatedAt().Format("2006-01-02T15:04:05Z07:00"),
	}

	utils.SendJSON(w, http.StatusOK, response)
}

func (h *ApplicationHandler) UpdatePorts(w http.ResponseWriter, r *http.Request) {
	var req UpdatePortsRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.SendError(w, http.StatusBadRequest, "invalid_json", "Invalid JSON format")
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		utils.SendError(w, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	appIDStr := chi.URLParam(r, "application_id")
	appID, err := applications.ApplicationIDFromString(appIDStr)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "invalid_application_id", "Invalid application ID")
		return
	}

	projectIDStr := chi.URLParam(r, "project_id")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "invalid_project_id", "Invalid project ID")
		return
	}

	app, err := h.appService.GetApplication(r.Context(), appID)
	if err != nil {
		utils.SendError(w, http.StatusNotFound, "application_not_found", "Application not found")
		return
	}

	if app.ProjectID() != projectID {
		utils.SendError(w, http.StatusNotFound, "application_not_found", "Application not found in project")
		return
	}

	err = h.appService.UpdatePorts(r.Context(), appID, req.ExposedPorts, req.PortMappings)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "update_failed", "Failed to update ports: "+err.Error())
		return
	}

	updatedApp, err := h.appService.GetApplication(r.Context(), appID)
	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, "fetch_failed", "Failed to fetch updated application")
		return
	}

	response := ApplicationResponse{
		ID:               updatedApp.ID().String(),
		Name:             updatedApp.Name().String(),
		Description:      updatedApp.Description(),
		ProjectID:        updatedApp.ProjectID().String(),
		EnvironmentID:    updatedApp.EnvironmentID().String(),
		DeploymentSource: updatedApp.DeploymentSource(),
		Domain:           updatedApp.Domain(),
		Buildpack:        convertToLegacyBuildpackConfig(updatedApp.Buildpack()),
		EnvVars:          updatedApp.EnvVars(),
		AutoDeploy:       updatedApp.AutoDeploy(),
		Status:           updatedApp.Status(),
		CreatedAt:        updatedApp.CreatedAt().Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:        updatedApp.UpdatedAt().Format("2006-01-02T15:04:05Z07:00"),
	}

	utils.SendJSON(w, http.StatusOK, response)
}
