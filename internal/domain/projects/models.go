// Package projects contains the Project aggregate following DDD principles
package projects

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/mikrocloud/mikrocloud/internal/domain/users"
)

// Project represents the core project entity
type Project struct {
	id             ProjectID
	name           ProjectName
	description    *string
	userID         users.UserID
	organizationID users.OrganizationID
	createdBy      users.UserID
	settings       string
	createdAt      time.Time
	updatedAt      time.Time
}

// ProjectID is a value object for project identification
type ProjectID struct {
	value uuid.UUID
}

func NewProjectID() ProjectID {
	return ProjectID{value: uuid.New()}
}

func ProjectIDFromUUID(id uuid.UUID) ProjectID {
	return ProjectID{value: id}
}

func ProjectIDFromString(id string) (ProjectID, error) {
	parsedID, err := uuid.Parse(id)
	if err != nil {
		return ProjectID{}, fmt.Errorf("invalid project ID: %w", err)
	}
	return ProjectID{value: parsedID}, nil
}

func (id ProjectID) String() string {
	return id.value.String()
}

func (id ProjectID) UUID() uuid.UUID {
	return id.value
}

// ProjectName is a value object that enforces naming rules
type ProjectName struct {
	value string
}

func NewProjectName(name string) (ProjectName, error) {
	if name == "" {
		return ProjectName{}, fmt.Errorf("project name cannot be empty")
	}

	if len(name) > 100 {
		return ProjectName{}, fmt.Errorf("project name cannot exceed 100 characters")
	}

	return ProjectName{value: name}, nil
}

func (n ProjectName) String() string {
	return n.value
}

// NewProject creates a new project with business rules enforcement
func NewProject(
	name ProjectName,
	description *string,
	userID users.UserID,
	organizationID users.OrganizationID,
) *Project {
	now := time.Now()

	return &Project{
		id:             NewProjectID(),
		name:           name,
		description:    description,
		userID:         userID,
		organizationID: organizationID,
		createdBy:      userID,
		settings:       "{}",
		createdAt:      now,
		updatedAt:      now,
	}
}

// Getters
func (p *Project) ID() ProjectID {
	return p.id
}

func (p *Project) Name() ProjectName {
	return p.name
}

func (p *Project) Description() *string {
	return p.description
}

func (p *Project) UserID() users.UserID {
	return p.userID
}

func (p *Project) OrganizationID() users.OrganizationID {
	return p.organizationID
}

func (p *Project) CreatedBy() users.UserID {
	return p.createdBy
}

func (p *Project) Settings() string {
	return p.settings
}

func (p *Project) CreatedAt() time.Time {
	return p.createdAt
}

func (p *Project) UpdatedAt() time.Time {
	return p.updatedAt
}

// Business methods
func (p *Project) UpdateDescription(description *string) {
	p.description = description
	p.updatedAt = time.Now()
}

func (p *Project) UpdateSettings(settings string) {
	p.settings = settings
	p.updatedAt = time.Now()
}

func (p *Project) ChangeName(name ProjectName) error {
	p.name = name
	p.updatedAt = time.Now()
	return nil
}

// ReconstructProject recreates a project from persistence data
// This is used by the infrastructure layer to reconstruct domain objects
func ReconstructProject(
	id ProjectID,
	name ProjectName,
	description *string,
	userID users.UserID,
	organizationID users.OrganizationID,
	createdBy users.UserID,
	settings string,
	createdAt time.Time,
	updatedAt time.Time,
) *Project {
	return &Project{
		id:             id,
		name:           name,
		description:    description,
		userID:         userID,
		organizationID: organizationID,
		createdBy:      createdBy,
		settings:       settings,
		createdAt:      createdAt,
		updatedAt:      updatedAt,
	}
}
