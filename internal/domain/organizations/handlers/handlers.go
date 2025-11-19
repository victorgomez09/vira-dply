package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/mikrocloud/mikrocloud/internal/domain/organizations/service"
	"github.com/mikrocloud/mikrocloud/internal/utils"
)

type OrganizationHandler struct {
	orgService *service.OrganizationService
}

func NewOrganizationHandler(orgService *service.OrganizationService) *OrganizationHandler {
	return &OrganizationHandler{
		orgService: orgService,
	}
}

type OrganizationResponse struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Slug         string `json:"slug"`
	Description  string `json:"description"`
	OwnerID      string `json:"owner_id"`
	BillingEmail string `json:"billing_email"`
	Plan         string `json:"plan"`
	Status       string `json:"status"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}

func (h *OrganizationHandler) ListOrganizations(w http.ResponseWriter, r *http.Request) {
	organizations, err := h.orgService.ListOrganizations(r.Context())
	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, "list_failed", "Failed to list organizations: "+err.Error())
		return
	}

	response := make([]OrganizationResponse, 0, len(organizations))
	for _, org := range organizations {
		response = append(response, OrganizationResponse{
			ID:           org.ID().String(),
			Name:         org.Name(),
			Slug:         org.Slug(),
			Description:  org.Description(),
			OwnerID:      org.OwnerID().String(),
			BillingEmail: org.BillingEmail(),
			Plan:         string(org.Plan()),
			Status:       string(org.Status()),
			CreatedAt:    org.CreatedAt().Format("2006-01-02T15:04:05Z"),
			UpdatedAt:    org.UpdatedAt().Format("2006-01-02T15:04:05Z"),
		})
	}

	utils.SendJSON(w, http.StatusOK, response)
}

func (h *OrganizationHandler) GetOrganization(w http.ResponseWriter, r *http.Request) {
	orgID := chi.URLParam(r, "organization_id")

	org, err := h.orgService.GetOrganization(r.Context(), orgID)
	if err != nil {
		utils.SendError(w, http.StatusNotFound, "not_found", "Organization not found: "+err.Error())
		return
	}

	response := OrganizationResponse{
		ID:           org.ID().String(),
		Name:         org.Name(),
		Slug:         org.Slug(),
		Description:  org.Description(),
		OwnerID:      org.OwnerID().String(),
		BillingEmail: org.BillingEmail(),
		Plan:         string(org.Plan()),
		Status:       string(org.Status()),
		CreatedAt:    org.CreatedAt().Format("2006-01-02T15:04:05Z"),
		UpdatedAt:    org.UpdatedAt().Format("2006-01-02T15:04:05Z"),
	}

	utils.SendJSON(w, http.StatusOK, response)
}
