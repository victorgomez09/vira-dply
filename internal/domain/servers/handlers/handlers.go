package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/mikrocloud/mikrocloud/internal/api/middleware"
	"github.com/mikrocloud/mikrocloud/internal/domain/servers"
	"github.com/mikrocloud/mikrocloud/internal/domain/servers/repository"
	"github.com/mikrocloud/mikrocloud/internal/domain/servers/service"
	"github.com/mikrocloud/mikrocloud/internal/utils"
)

type ServersHandler struct {
	service *service.ServersService
}

func NewServersHandler(service *service.ServersService) *ServersHandler {
	return &ServersHandler{service: service}
}

type CreateServerRequest struct {
	Name       string `json:"name"`
	Hostname   string `json:"hostname"`
	IPAddress  string `json:"ip_address"`
	Port       int    `json:"port"`
	ServerType string `json:"server_type"`
}

type UpdateServerRequest struct {
	Description string   `json:"description,omitempty"`
	Hostname    string   `json:"hostname,omitempty"`
	IPAddress   string   `json:"ip_address,omitempty"`
	Port        *int     `json:"port,omitempty"`
	Status      string   `json:"status,omitempty"`
	CPUCores    *int     `json:"cpu_cores,omitempty"`
	MemoryMB    *int     `json:"memory_mb,omitempty"`
	DiskGB      *int     `json:"disk_gb,omitempty"`
	OS          *string  `json:"os,omitempty"`
	OSVersion   *string  `json:"os_version,omitempty"`
	Tags        []string `json:"tags,omitempty"`
}

func (h *ServersHandler) ListServers(w http.ResponseWriter, r *http.Request) {
	orgIDStr := middleware.GetOrgID(r)
	if orgIDStr == "" {
		utils.SendError(w, http.StatusUnauthorized, "Organization ID not found", "")
		return
	}

	orgID, parseErr := uuid.Parse(orgIDStr)
	if parseErr != nil {
		utils.SendError(w, http.StatusBadRequest, "Invalid organization ID", parseErr.Error())
		return
	}

	serverType := r.URL.Query().Get("type")
	var serverList []*servers.Server
	var err error

	if serverType != "" {
		serverList, err = h.service.ListServersByType(orgID, servers.ServerType(serverType))
	} else {
		serverList, err = h.service.ListServersByOrganization(orgID)
	}

	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, "Failed to list servers", err.Error())
		return
	}

	dtos := make([]*repository.ServerDTO, 0, len(serverList))
	for _, server := range serverList {
		dto, err := repository.ToDTO(server)
		if err != nil {
			continue
		}
		dtos = append(dtos, dto)
	}

	utils.SendJSON(w, http.StatusOK, dtos)
}

func (h *ServersHandler) GetServer(w http.ResponseWriter, r *http.Request) {
	serverIDStr := chi.URLParam(r, "server_id")
	serverID, err := servers.ServerIDFromString(serverIDStr)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "Invalid server ID", err.Error())
		return
	}

	server, err := h.service.GetServer(serverID)
	if err != nil {
		utils.SendError(w, http.StatusNotFound, "Server not found", err.Error())
		return
	}

	dto, err := repository.ToDTO(server)
	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, "Failed to convert server", err.Error())
		return
	}

	utils.SendJSON(w, http.StatusOK, dto)
}

func (h *ServersHandler) CreateServer(w http.ResponseWriter, r *http.Request) {
	orgIDStr := middleware.GetOrgID(r)
	if orgIDStr == "" {
		utils.SendError(w, http.StatusUnauthorized, "Organization ID not found", "")
		return
	}

	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "Invalid organization ID", err.Error())
		return
	}

	var req CreateServerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.SendError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	server, err := h.service.CreateServer(req.Name, req.Hostname, req.IPAddress, req.Port, servers.ServerType(req.ServerType), orgID)
	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, "Failed to create server", err.Error())
		return
	}

	dto, err := repository.ToDTO(server)
	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, "Failed to convert server", err.Error())
		return
	}

	utils.SendJSON(w, http.StatusCreated, dto)
}

func (h *ServersHandler) UpdateServer(w http.ResponseWriter, r *http.Request) {
	serverIDStr := chi.URLParam(r, "server_id")
	serverID, err := servers.ServerIDFromString(serverIDStr)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "Invalid server ID", err.Error())
		return
	}

	server, err := h.service.GetServer(serverID)
	if err != nil {
		utils.SendError(w, http.StatusNotFound, "Server not found", err.Error())
		return
	}

	var req UpdateServerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.SendError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	if req.Description != "" {
		server.UpdateDescription(req.Description)
	}
	if req.Hostname != "" {
		server.UpdateHostname(req.Hostname)
	}
	if req.IPAddress != "" {
		server.UpdateIPAddress(req.IPAddress)
	}
	if req.Port != nil {
		server.UpdatePort(*req.Port)
	}
	if req.Status != "" {
		server.ChangeStatus(servers.ServerStatus(req.Status))
	}
	if req.CPUCores != nil || req.MemoryMB != nil || req.DiskGB != nil || req.OS != nil || req.OSVersion != nil {
		server.UpdateSpecs(req.CPUCores, req.MemoryMB, req.DiskGB, req.OS, req.OSVersion)
	}
	if req.Tags != nil {
		server.SetTags(req.Tags)
	}

	if err := h.service.UpdateServer(server); err != nil {
		utils.SendError(w, http.StatusInternalServerError, "Failed to update server", err.Error())
		return
	}

	dto, err := repository.ToDTO(server)
	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, "Failed to convert server", err.Error())
		return
	}

	utils.SendJSON(w, http.StatusOK, dto)
}

func (h *ServersHandler) DeleteServer(w http.ResponseWriter, r *http.Request) {
	serverIDStr := chi.URLParam(r, "server_id")
	serverID, err := servers.ServerIDFromString(serverIDStr)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "Invalid server ID", err.Error())
		return
	}

	if err := h.service.DeleteServer(serverID); err != nil {
		utils.SendError(w, http.StatusInternalServerError, "Failed to delete server", err.Error())
		return
	}

	utils.SendJSON(w, http.StatusOK, map[string]string{"message": "Server deleted successfully"})
}
