package handlers

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/mikrocloud/mikrocloud/internal/domain/activities/service"
	"github.com/mikrocloud/mikrocloud/internal/utils"
)

type ActivitiesHandlers struct {
	service *service.ActivitiesService
}

func NewActivitiesHandlers(service *service.ActivitiesService) *ActivitiesHandlers {
	return &ActivitiesHandlers{service: service}
}

func (h *ActivitiesHandlers) GetRecentActivities(w http.ResponseWriter, r *http.Request) {
	orgIDStr := chi.URLParam(r, "org_id")
	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "Invalid organization ID", err.Error())
		return
	}

	limit := 50
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	offset := 0
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	activities, err := h.service.GetRecentActivities(orgID, limit, offset)
	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, "Failed to fetch activities", err.Error())
		return
	}

	response := map[string]interface{}{
		"activities": activities,
		"total":      len(activities),
	}

	utils.SendJSON(w, http.StatusOK, response)
}

func (h *ActivitiesHandlers) GetResourceActivities(w http.ResponseWriter, r *http.Request) {
	resourceType := chi.URLParam(r, "resource_type")
	resourceIDStr := chi.URLParam(r, "resource_id")

	resourceID, err := uuid.Parse(resourceIDStr)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "Invalid resource ID", err.Error())
		return
	}

	limit := 20
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	activities, err := h.service.GetResourceActivities(resourceType, resourceID, limit)
	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, "Failed to fetch resource activities", err.Error())
		return
	}

	response := map[string]interface{}{
		"activities": activities,
		"total":      len(activities),
	}

	utils.SendJSON(w, http.StatusOK, response)
}
