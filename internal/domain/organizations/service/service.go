package service

import (
	"context"
	"fmt"

	"github.com/mikrocloud/mikrocloud/internal/domain/organizations/repository"
	"github.com/mikrocloud/mikrocloud/internal/domain/users"
)

type OrganizationService struct {
	orgRepo repository.Repository
}

func NewOrganizationService(orgRepo repository.Repository) *OrganizationService {
	return &OrganizationService{
		orgRepo: orgRepo,
	}
}

func (s *OrganizationService) ListOrganizations(ctx context.Context) ([]*users.Organization, error) {
	return s.orgRepo.FindAll(ctx)
}

func (s *OrganizationService) GetOrganization(ctx context.Context, id string) (*users.Organization, error) {
	orgID, err := users.OrganizationIDFromString(id)
	if err != nil {
		return nil, fmt.Errorf("invalid organization ID: %w", err)
	}

	return s.orgRepo.FindByID(ctx, orgID)
}

func (s *OrganizationService) GetUserOrganizations(ctx context.Context, userID string) ([]*users.Organization, error) {
	uid, err := users.UserIDFromString(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	return s.orgRepo.FindByUserID(ctx, uid)
}
