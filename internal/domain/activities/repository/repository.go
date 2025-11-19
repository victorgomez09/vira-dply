package repository

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/mikrocloud/mikrocloud/internal/domain/activities"
)

type ActivitiesRepository struct {
	db *sql.DB
}

func NewActivitiesRepository(db *sql.DB) *ActivitiesRepository {
	return &ActivitiesRepository{db: db}
}

func (r *ActivitiesRepository) Create(activity *activities.Activity) error {
	_, err := r.db.Exec(`
		INSERT INTO activities (id, activity_type, description, initiator_id, initiator_name, 
			resource_type, resource_id, resource_name, metadata, organization_id, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		activity.ID().String(),
		string(activity.ActivityType()),
		activity.Description(),
		uuidToNullString(activity.InitiatorID()),
		activity.InitiatorName(),
		strPtrToNullString(activity.ResourceType()),
		uuidToNullString(activity.ResourceID()),
		strPtrToNullString(activity.ResourceName()),
		activity.Metadata(),
		activity.OrganizationID().String(),
		activity.CreatedAt(),
	)
	return err
}

func (r *ActivitiesRepository) ListByOrganization(organizationID uuid.UUID, limit int, offset int) ([]*activities.Activity, error) {
	rows, err := r.db.Query(`
		SELECT id, activity_type, description, initiator_id, initiator_name,
			resource_type, resource_id, resource_name, metadata, organization_id, created_at
		FROM activities
		WHERE organization_id = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`, organizationID.String(), limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanActivities(rows)
}

func (r *ActivitiesRepository) ListByResource(resourceType string, resourceID uuid.UUID, limit int) ([]*activities.Activity, error) {
	rows, err := r.db.Query(`
		SELECT id, activity_type, description, initiator_id, initiator_name,
			resource_type, resource_id, resource_name, metadata, organization_id, created_at
		FROM activities
		WHERE resource_type = ? AND resource_id = ?
		ORDER BY created_at DESC
		LIMIT ?
	`, resourceType, resourceID.String(), limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanActivities(rows)
}

func (r *ActivitiesRepository) DeleteOlderThan(days int) error {
	_, err := r.db.Exec(`
		DELETE FROM activities
		WHERE created_at < datetime('now', '-' || ? || ' days')
	`, days)
	return err
}

func (r *ActivitiesRepository) scanActivities(rows *sql.Rows) ([]*activities.Activity, error) {
	var result []*activities.Activity

	for rows.Next() {
		var (
			id             string
			activityType   string
			description    string
			initiatorID    sql.NullString
			initiatorName  string
			resourceType   sql.NullString
			resourceID     sql.NullString
			resourceName   sql.NullString
			metadata       string
			organizationID string
			createdAt      time.Time
		)

		if err := rows.Scan(
			&id, &activityType, &description, &initiatorID, &initiatorName,
			&resourceType, &resourceID, &resourceName, &metadata, &organizationID, &createdAt,
		); err != nil {
			return nil, err
		}

		activityID, _ := activities.ActivityIDFromString(id)
		orgID, _ := uuid.Parse(organizationID)

		activity := activities.ReconstructActivity(
			activityID,
			activities.ActivityType(activityType),
			description,
			nullStringToUUIDPtr(initiatorID),
			initiatorName,
			nullStringToStrPtr(resourceType),
			nullStringToUUIDPtr(resourceID),
			nullStringToStrPtr(resourceName),
			metadata,
			orgID,
			createdAt,
		)

		result = append(result, activity)
	}

	return result, rows.Err()
}

func uuidToNullString(u *uuid.UUID) sql.NullString {
	if u == nil {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: u.String(), Valid: true}
}

func strPtrToNullString(s *string) sql.NullString {
	if s == nil {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: *s, Valid: true}
}

func nullStringToUUIDPtr(ns sql.NullString) *uuid.UUID {
	if !ns.Valid {
		return nil
	}
	u, err := uuid.Parse(ns.String)
	if err != nil {
		return nil
	}
	return &u
}

func nullStringToStrPtr(ns sql.NullString) *string {
	if !ns.Valid {
		return nil
	}
	return &ns.String
}

type ActivityDTO struct {
	ID             string                 `json:"id"`
	ActivityType   string                 `json:"activity_type"`
	Description    string                 `json:"description"`
	InitiatorID    *string                `json:"initiator_id"`
	InitiatorName  string                 `json:"initiator_name"`
	ResourceType   *string                `json:"resource_type"`
	ResourceID     *string                `json:"resource_id"`
	ResourceName   *string                `json:"resource_name"`
	Metadata       map[string]interface{} `json:"metadata"`
	OrganizationID string                 `json:"organization_id"`
	CreatedAt      time.Time              `json:"created_at"`
}

func ToDTO(activity *activities.Activity) (*ActivityDTO, error) {
	var metadata map[string]interface{}
	if err := json.Unmarshal([]byte(activity.Metadata()), &metadata); err != nil {
		metadata = make(map[string]interface{})
	}

	var initiatorID *string
	if activity.InitiatorID() != nil {
		id := activity.InitiatorID().String()
		initiatorID = &id
	}

	var resourceID *string
	if activity.ResourceID() != nil {
		id := activity.ResourceID().String()
		resourceID = &id
	}

	return &ActivityDTO{
		ID:             activity.ID().String(),
		ActivityType:   string(activity.ActivityType()),
		Description:    activity.Description(),
		InitiatorID:    initiatorID,
		InitiatorName:  activity.InitiatorName(),
		ResourceType:   activity.ResourceType(),
		ResourceID:     resourceID,
		ResourceName:   activity.ResourceName(),
		Metadata:       metadata,
		OrganizationID: activity.OrganizationID().String(),
		CreatedAt:      activity.CreatedAt(),
	}, nil
}
