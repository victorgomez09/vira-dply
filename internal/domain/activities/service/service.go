package service

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/mikrocloud/mikrocloud/internal/domain/activities"
	"github.com/mikrocloud/mikrocloud/internal/domain/activities/repository"
)

type ActivitiesService struct {
	repo *repository.ActivitiesRepository
}

func NewActivitiesService(repo *repository.ActivitiesRepository) *ActivitiesService {
	return &ActivitiesService{repo: repo}
}

func (s *ActivitiesService) LogActivity(
	activityType activities.ActivityType,
	description string,
	initiatorID *uuid.UUID,
	initiatorName string,
	resourceType *string,
	resourceID *uuid.UUID,
	resourceName *string,
	metadata map[string]interface{},
	organizationID uuid.UUID,
) error {
	metadataJSON := "{}"
	if metadata != nil {
		jsonBytes, err := json.Marshal(metadata)
		if err != nil {
			return fmt.Errorf("failed to marshal metadata: %w", err)
		}
		metadataJSON = string(jsonBytes)
	}

	activity := activities.NewActivity(
		activityType,
		description,
		initiatorID,
		initiatorName,
		resourceType,
		resourceID,
		resourceName,
		metadataJSON,
		organizationID,
	)

	return s.repo.Create(activity)
}

func (s *ActivitiesService) GetRecentActivities(organizationID uuid.UUID, limit int, offset int) ([]*repository.ActivityDTO, error) {
	activityList, err := s.repo.ListByOrganization(organizationID, limit, offset)
	if err != nil {
		return nil, err
	}

	dtos := make([]*repository.ActivityDTO, 0, len(activityList))
	for _, activity := range activityList {
		dto, err := repository.ToDTO(activity)
		if err != nil {
			continue
		}
		dtos = append(dtos, dto)
	}

	return dtos, nil
}

func (s *ActivitiesService) GetResourceActivities(resourceType string, resourceID uuid.UUID, limit int) ([]*repository.ActivityDTO, error) {
	activityList, err := s.repo.ListByResource(resourceType, resourceID, limit)
	if err != nil {
		return nil, err
	}

	dtos := make([]*repository.ActivityDTO, 0, len(activityList))
	for _, activity := range activityList {
		dto, err := repository.ToDTO(activity)
		if err != nil {
			continue
		}
		dtos = append(dtos, dto)
	}

	return dtos, nil
}

func (s *ActivitiesService) CleanupOldActivities(days int) error {
	return s.repo.DeleteOlderThan(days)
}
