package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/mikrocloud/mikrocloud/internal/api/middleware"
	"github.com/mikrocloud/mikrocloud/internal/domain/git"
	"github.com/mikrocloud/mikrocloud/internal/domain/git/service"
	"github.com/mikrocloud/mikrocloud/internal/utils"
)

type GitHandler struct {
	gitService *service.GitService
	validator  *validator.Validate
}

func NewGitHandler(gs *service.GitService) *GitHandler {
	return &GitHandler{
		gitService: gs,
		validator:  validator.New(),
	}
}

func (h *GitHandler) ValidateRepository(w http.ResponseWriter, r *http.Request) {
	var req git.ValidateRepositoryRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.SendError(w, http.StatusBadRequest, "invalid_json", "Invalid JSON format")
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		utils.SendError(w, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	fmt.Println(req)

	response, err := h.gitService.ValidateRepository(r.Context(), req)
	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, "validation_failed", err.Error())
		return
	}

	utils.SendJSON(w, http.StatusOK, response)
}

func (h *GitHandler) ListBranches(w http.ResponseWriter, r *http.Request) {
	var req git.ListBranchesRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.SendError(w, http.StatusBadRequest, "invalid_json", "Invalid JSON format")
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		utils.SendError(w, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	response, err := h.gitService.ListBranches(r.Context(), req)
	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, "list_failed", err.Error())
		return
	}

	utils.SendJSON(w, http.StatusOK, response)
}

func (h *GitHandler) DetectBuildMethod(w http.ResponseWriter, r *http.Request) {
	var req git.DetectBuildMethodRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.SendError(w, http.StatusBadRequest, "invalid_json", "Invalid JSON format")
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		utils.SendError(w, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	response, err := h.gitService.DetectBuildMethod(r.Context(), req)
	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, "detection_failed", err.Error())
		return
	}

	utils.SendJSON(w, http.StatusOK, response)
}

func (h *GitHandler) CreateGitSource(w http.ResponseWriter, r *http.Request) {
	var req git.CreateGitSourceRequest

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

	source, err := h.gitService.CreateGitSource(r.Context(), orgID, userID, req)
	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, "create_failed", err.Error())
		return
	}

	utils.SendJSON(w, http.StatusCreated, source)
}

func (h *GitHandler) GetGitSource(w http.ResponseWriter, r *http.Request) {
	sourceID := chi.URLParam(r, "source_id")
	if sourceID == "" {
		utils.SendError(w, http.StatusBadRequest, "missing_parameter", "Source ID is required")
		return
	}

	source, err := h.gitService.GetGitSource(r.Context(), sourceID)
	if err != nil {
		utils.SendError(w, http.StatusNotFound, "not_found", err.Error())
		return
	}

	utils.SendJSON(w, http.StatusOK, source)
}

func (h *GitHandler) ListGitSources(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)

	sources, err := h.gitService.ListGitSourcesByUser(r.Context(), userID)
	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, "list_failed", err.Error())
		return
	}

	utils.SendJSON(w, http.StatusOK, map[string]interface{}{
		"sources": sources,
	})
}

func (h *GitHandler) UpdateGitSource(w http.ResponseWriter, r *http.Request) {
	sourceID := chi.URLParam(r, "source_id")
	if sourceID == "" {
		utils.SendError(w, http.StatusBadRequest, "missing_parameter", "Source ID is required")
		return
	}

	var req git.UpdateGitSourceRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.SendError(w, http.StatusBadRequest, "invalid_json", "Invalid JSON format")
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		utils.SendError(w, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	source, err := h.gitService.UpdateGitSource(r.Context(), sourceID, req)
	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, "update_failed", err.Error())
		return
	}

	utils.SendJSON(w, http.StatusOK, source)
}

func (h *GitHandler) DeleteGitSource(w http.ResponseWriter, r *http.Request) {
	sourceID := chi.URLParam(r, "source_id")
	if sourceID == "" {
		utils.SendError(w, http.StatusBadRequest, "missing_parameter", "Source ID is required")
		return
	}

	if err := h.gitService.DeleteGitSource(r.Context(), sourceID); err != nil {
		utils.SendError(w, http.StatusInternalServerError, "delete_failed", err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
