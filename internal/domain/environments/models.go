package environments

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Environment represents an environment within a project (e.g., prod, dev, staging)
type Environment struct {
	id           EnvironmentID
	name         EnvironmentName
	isProduction bool
	projectID    uuid.UUID
	description  string
	variables    map[string]string
	createdAt    time.Time
	updatedAt    time.Time
}

// EnvironmentID is a value object for environment identification
type EnvironmentID struct {
	value uuid.UUID
}

func NewEnvironmentID() EnvironmentID {
	return EnvironmentID{value: uuid.New()}
}

func EnvironmentIDFromString(s string) (EnvironmentID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return EnvironmentID{}, fmt.Errorf("invalid environment ID: %w", err)
	}
	return EnvironmentID{value: id}, nil
}

func (id EnvironmentID) String() string {
	return id.value.String()
}

func (id EnvironmentID) UUID() uuid.UUID {
	return id.value
}

// EnvironmentName is a value object that enforces naming rules
type EnvironmentName struct {
	value string
}

func NewEnvironmentName(name string) (EnvironmentName, error) {
	if name == "" {
		return EnvironmentName{}, fmt.Errorf("environment name cannot be empty")
	}

	if len(name) > 50 {
		return EnvironmentName{}, fmt.Errorf("environment name cannot exceed 50 characters")
	}

	// Common environment names validation
	validNames := map[string]bool{
		"prod":        true,
		"production":  true,
		"dev":         true,
		"develop":     true,
		"development": true,
		"staging":     true,
		"test":        true,
		"testing":     true,
		"preview":     true,
	}

	// Allow custom names as well, but ensure they're lowercase and alphanumeric
	// This is just a suggestion - you might want different validation rules
	if !validNames[name] {
		// Custom validation for non-standard environment names
		// Could add regex validation here if needed
	}

	return EnvironmentName{value: name}, nil
}

func (n EnvironmentName) String() string {
	return n.value
}

// Predefined environment names
var (
	EnvironmentProduction  = EnvironmentName{value: "prod"}
	EnvironmentDevelopment = EnvironmentName{value: "dev"}
	EnvironmentStaging     = EnvironmentName{value: "staging"}
	EnvironmentTesting     = EnvironmentName{value: "test"}
)

// NewEnvironment creates a new environment with business rules enforcement
func NewEnvironment(name EnvironmentName, projectID uuid.UUID, description string, isProduction bool) *Environment {
	now := time.Now()

	return &Environment{
		id:           NewEnvironmentID(),
		name:         name,
		isProduction: isProduction,
		projectID:    projectID,
		description:  description,
		variables:    make(map[string]string),
		createdAt:    now,
		updatedAt:    now,
	}
}

// Getters
func (e *Environment) ID() EnvironmentID {
	return e.id
}

func (e *Environment) Name() EnvironmentName {
	return e.name
}

func (e *Environment) IsProduction() bool {
	return e.isProduction
}

func (e *Environment) ProjectID() uuid.UUID {
	return e.projectID
}

func (e *Environment) Description() string {
	return e.description
}

func (e *Environment) Variables() map[string]string {
	// Return a copy to maintain encapsulation
	vars := make(map[string]string)
	for k, v := range e.variables {
		vars[k] = v
	}
	return vars
}

func (e *Environment) CreatedAt() time.Time {
	return e.createdAt
}

func (e *Environment) UpdatedAt() time.Time {
	return e.updatedAt
}

// Business methods
func (e *Environment) UpdateDescription(description string) {
	e.description = description
	e.updatedAt = time.Now()
}

func (e *Environment) SetVariable(key, value string) error {
	if key == "" {
		return fmt.Errorf("environment variable key cannot be empty")
	}

	e.variables[key] = value
	e.updatedAt = time.Now()
	return nil
}

func (e *Environment) RemoveVariable(key string) {
	delete(e.variables, key)
	e.updatedAt = time.Now()
}

func (e *Environment) ChangeName(name EnvironmentName) error {
	e.name = name
	e.updatedAt = time.Now()
	return nil
}

func (e *Environment) SetProduction(isProduction bool) {
	e.isProduction = isProduction
	e.updatedAt = time.Now()
}

// GetVariable retrieves a specific environment variable
func (e *Environment) GetVariable(key string) (string, bool) {
	value, exists := e.variables[key]
	return value, exists
}

// ReconstructEnvironment recreates an environment from persistence data
func ReconstructEnvironment(
	id EnvironmentID,
	name EnvironmentName,
	isProduction bool,
	projectID uuid.UUID,
	description string,
	variables map[string]string,
	createdAt time.Time,
	updatedAt time.Time,
) *Environment {
	return &Environment{
		id:           id,
		name:         name,
		isProduction: isProduction,
		projectID:    projectID,
		description:  description,
		variables:    variables,
		createdAt:    createdAt,
		updatedAt:    updatedAt,
	}
}
