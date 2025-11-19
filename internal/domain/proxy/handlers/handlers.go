package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/mikrocloud/mikrocloud/internal/domain/proxy/service"
	"github.com/mikrocloud/mikrocloud/internal/utils"
)

type ProxyHandler struct {
	proxyService *service.ProxyService
	validator    *validator.Validate
}

func NewProxyHandler(proxyService *service.ProxyService) *ProxyHandler {
	return &ProxyHandler{
		proxyService: proxyService,
		validator:    validator.New(),
	}
}

// CreateProxyConfigRequest represents a request to create a proxy configuration
type CreateProxyConfigRequest struct {
	Name        string   `json:"name" validate:"required,min=1,max=100"`
	ServiceName string   `json:"service_name" validate:"required"`
	ContainerID string   `json:"container_id" validate:"required"`
	Hostnames   []string `json:"hostnames" validate:"required,min=1"`
	TargetURL   string   `json:"target_url" validate:"required,url"`
	Port        int      `json:"port" validate:"required,min=1,max=65535"`
	Protocol    string   `json:"protocol" validate:"required,oneof=http https tcp udp"`
	PathPrefix  string   `json:"path_prefix,omitempty"`
	StripPrefix bool     `json:"strip_prefix,omitempty"`
}

type UpdateProxyConfigRequest struct {
	Name        *string  `json:"name,omitempty" validate:"omitempty,min=1,max=100"`
	ServiceName *string  `json:"service_name,omitempty"`
	ContainerID *string  `json:"container_id,omitempty"`
	Hostnames   []string `json:"hostnames,omitempty" validate:"omitempty,min=1"`
	TargetURL   *string  `json:"target_url,omitempty" validate:"omitempty,url"`
	Port        *int     `json:"port,omitempty" validate:"omitempty,min=1,max=65535"`
	Protocol    *string  `json:"protocol,omitempty" validate:"omitempty,oneof=http https tcp udp"`
	PathPrefix  *string  `json:"path_prefix,omitempty"`
	StripPrefix *bool    `json:"strip_prefix,omitempty"`
}

type ListProxyConfigsResponse struct {
	ProxyConfigs []*service.ProxyConfigResponse `json:"proxy_configs"`
}

// CreateProxyConfig creates a new proxy configuration in a project
func (h *ProxyHandler) CreateProxyConfig(w http.ResponseWriter, r *http.Request) {
	var req CreateProxyConfigRequest

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

	// Create service request
	serviceReq := service.CreateProxyConfigRequest{
		Name:        req.Name,
		ProjectID:   projectID,
		ServiceName: req.ServiceName,
		ContainerID: req.ContainerID,
		Hostnames:   req.Hostnames,
		TargetURL:   req.TargetURL,
		Port:        req.Port,
		Protocol:    req.Protocol,
		PathPrefix:  req.PathPrefix,
		StripPrefix: req.StripPrefix,
	}

	// Create proxy config
	config, err := h.proxyService.CreateProxyConfig(r.Context(), serviceReq)
	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, "create_failed", "Failed to create proxy configuration: "+err.Error())
		return
	}

	utils.SendJSON(w, http.StatusCreated, config)
}

// ListProxyConfigs lists all proxy configurations for a project
func (h *ProxyHandler) ListProxyConfigs(w http.ResponseWriter, r *http.Request) {
	// Get project ID from URL
	projectIDStr := chi.URLParam(r, "project_id")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "invalid_project_id", "Invalid project ID")
		return
	}

	configs, err := h.proxyService.ListProxyConfigs(r.Context(), &projectID)
	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, "list_failed", "Failed to list proxy configurations")
		return
	}

	response := ListProxyConfigsResponse{
		ProxyConfigs: configs,
	}

	utils.SendJSON(w, http.StatusOK, response)
}

// GetProxyConfig retrieves a specific proxy configuration
func (h *ProxyHandler) GetProxyConfig(w http.ResponseWriter, r *http.Request) {
	// Get config ID from URL
	configID := chi.URLParam(r, "config_id")

	config, err := h.proxyService.GetProxyConfig(r.Context(), configID)
	if err != nil {
		utils.SendError(w, http.StatusNotFound, "config_not_found", "Proxy configuration not found")
		return
	}

	utils.SendJSON(w, http.StatusOK, config)
}

// UpdateProxyConfig updates an existing proxy configuration
func (h *ProxyHandler) UpdateProxyConfig(w http.ResponseWriter, r *http.Request) {
	var req UpdateProxyConfigRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.SendError(w, http.StatusBadRequest, "invalid_json", "Invalid JSON format")
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		utils.SendError(w, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	// Get config ID from URL
	configID := chi.URLParam(r, "config_id")

	// Get existing config to build update request
	existingConfig, err := h.proxyService.GetProxyConfig(r.Context(), configID)
	if err != nil {
		utils.SendError(w, http.StatusNotFound, "config_not_found", "Proxy configuration not found")
		return
	}

	// Build update request from existing config and new values
	updateReq := service.CreateProxyConfigRequest{
		Name:        existingConfig.Name,
		ProjectID:   existingConfig.ProjectID,
		ServiceName: existingConfig.ServiceName,
		ContainerID: existingConfig.ContainerID,
		Hostnames:   existingConfig.Hostnames,
		TargetURL:   existingConfig.TargetURL,
		Port:        existingConfig.Port,
		Protocol:    existingConfig.Protocol,
		PathPrefix:  existingConfig.PathPrefix,
		StripPrefix: existingConfig.StripPrefix,
	}

	// Apply updates
	if req.Name != nil {
		updateReq.Name = *req.Name
	}
	if req.ServiceName != nil {
		updateReq.ServiceName = *req.ServiceName
	}
	if req.ContainerID != nil {
		updateReq.ContainerID = *req.ContainerID
	}
	if len(req.Hostnames) > 0 {
		updateReq.Hostnames = req.Hostnames
	}
	if req.TargetURL != nil {
		updateReq.TargetURL = *req.TargetURL
	}
	if req.Port != nil {
		updateReq.Port = *req.Port
	}
	if req.Protocol != nil {
		updateReq.Protocol = *req.Protocol
	}
	if req.PathPrefix != nil {
		updateReq.PathPrefix = *req.PathPrefix
	}
	if req.StripPrefix != nil {
		updateReq.StripPrefix = *req.StripPrefix
	}

	// Update the configuration
	config, err := h.proxyService.UpdateProxyConfig(r.Context(), configID, updateReq)
	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, "update_failed", "Failed to update proxy configuration: "+err.Error())
		return
	}

	utils.SendJSON(w, http.StatusOK, config)
}

// DeleteProxyConfig deletes a proxy configuration
func (h *ProxyHandler) DeleteProxyConfig(w http.ResponseWriter, r *http.Request) {
	// Get config ID from URL
	configID := chi.URLParam(r, "config_id")

	if err := h.proxyService.DeleteProxyConfig(r.Context(), configID); err != nil {
		utils.SendError(w, http.StatusInternalServerError, "delete_failed", "Failed to delete proxy configuration")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
