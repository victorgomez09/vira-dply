package activities

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Activity struct {
	id             ActivityID
	activityType   ActivityType
	description    string
	initiatorID    *uuid.UUID
	initiatorName  string
	resourceType   *string
	resourceID     *uuid.UUID
	resourceName   *string
	metadata       string
	organizationID uuid.UUID
	createdAt      time.Time
}

type ActivityID struct {
	value uuid.UUID
}

func NewActivityID() ActivityID {
	return ActivityID{value: uuid.New()}
}

func ActivityIDFromUUID(id uuid.UUID) ActivityID {
	return ActivityID{value: id}
}

func ActivityIDFromString(id string) (ActivityID, error) {
	parsedID, err := uuid.Parse(id)
	if err != nil {
		return ActivityID{}, fmt.Errorf("invalid activity ID: %w", err)
	}
	return ActivityID{value: parsedID}, nil
}

func (id ActivityID) String() string {
	return id.value.String()
}

func (id ActivityID) UUID() uuid.UUID {
	return id.value
}

type ActivityType string

const (
	ActivityTypeProjectCreated     ActivityType = "project.created"
	ActivityTypeProjectUpdated     ActivityType = "project.updated"
	ActivityTypeProjectDeleted     ActivityType = "project.deleted"
	ActivityTypeAppCreated         ActivityType = "app.created"
	ActivityTypeAppUpdated         ActivityType = "app.updated"
	ActivityTypeAppDeleted         ActivityType = "app.deleted"
	ActivityTypeAppDeployed        ActivityType = "app.deployed"
	ActivityTypeAppStarted         ActivityType = "app.started"
	ActivityTypeAppStopped         ActivityType = "app.stopped"
	ActivityTypeAppRestarted       ActivityType = "app.restarted"
	ActivityTypeDatabaseCreated    ActivityType = "database.created"
	ActivityTypeDatabaseUpdated    ActivityType = "database.updated"
	ActivityTypeDatabaseDeleted    ActivityType = "database.deleted"
	ActivityTypeDatabaseStarted    ActivityType = "database.started"
	ActivityTypeDatabaseStopped    ActivityType = "database.stopped"
	ActivityTypeDatabaseRestarted  ActivityType = "database.restarted"
	ActivityTypeEnvironmentCreated ActivityType = "environment.created"
	ActivityTypeEnvironmentUpdated ActivityType = "environment.updated"
	ActivityTypeEnvironmentDeleted ActivityType = "environment.deleted"
	ActivityTypeDiskCreated        ActivityType = "disk.created"
	ActivityTypeDiskResized        ActivityType = "disk.resized"
	ActivityTypeDiskDeleted        ActivityType = "disk.deleted"
	ActivityTypeDiskAttached       ActivityType = "disk.attached"
	ActivityTypeDiskDetached       ActivityType = "disk.detached"
	ActivityTypeProxyCreated       ActivityType = "proxy.created"
	ActivityTypeProxyUpdated       ActivityType = "proxy.updated"
	ActivityTypeProxyDeleted       ActivityType = "proxy.deleted"
	ActivityTypeUserRegistered     ActivityType = "user.registered"
	ActivityTypeUserLogin          ActivityType = "user.login"
	ActivityTypeSettingsUpdated    ActivityType = "settings.updated"
	ActivityTypeBackupCreated      ActivityType = "backup.created"
	ActivityTypeBackupRestored     ActivityType = "backup.restored"
	ActivityTypeSystemStarted      ActivityType = "system.started"
	ActivityTypeSystemStopped      ActivityType = "system.stopped"
)

func NewActivity(
	activityType ActivityType,
	description string,
	initiatorID *uuid.UUID,
	initiatorName string,
	resourceType *string,
	resourceID *uuid.UUID,
	resourceName *string,
	metadata string,
	organizationID uuid.UUID,
) *Activity {
	return &Activity{
		id:             NewActivityID(),
		activityType:   activityType,
		description:    description,
		initiatorID:    initiatorID,
		initiatorName:  initiatorName,
		resourceType:   resourceType,
		resourceID:     resourceID,
		resourceName:   resourceName,
		metadata:       metadata,
		organizationID: organizationID,
		createdAt:      time.Now(),
	}
}

func (a *Activity) ID() ActivityID {
	return a.id
}

func (a *Activity) ActivityType() ActivityType {
	return a.activityType
}

func (a *Activity) Description() string {
	return a.description
}

func (a *Activity) InitiatorID() *uuid.UUID {
	return a.initiatorID
}

func (a *Activity) InitiatorName() string {
	return a.initiatorName
}

func (a *Activity) ResourceType() *string {
	return a.resourceType
}

func (a *Activity) ResourceID() *uuid.UUID {
	return a.resourceID
}

func (a *Activity) ResourceName() *string {
	return a.resourceName
}

func (a *Activity) Metadata() string {
	return a.metadata
}

func (a *Activity) OrganizationID() uuid.UUID {
	return a.organizationID
}

func (a *Activity) CreatedAt() time.Time {
	return a.createdAt
}

func ReconstructActivity(
	id ActivityID,
	activityType ActivityType,
	description string,
	initiatorID *uuid.UUID,
	initiatorName string,
	resourceType *string,
	resourceID *uuid.UUID,
	resourceName *string,
	metadata string,
	organizationID uuid.UUID,
	createdAt time.Time,
) *Activity {
	return &Activity{
		id:             id,
		activityType:   activityType,
		description:    description,
		initiatorID:    initiatorID,
		initiatorName:  initiatorName,
		resourceType:   resourceType,
		resourceID:     resourceID,
		resourceName:   resourceName,
		metadata:       metadata,
		organizationID: organizationID,
		createdAt:      createdAt,
	}
}
