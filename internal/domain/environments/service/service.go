package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/mikrocloud/mikrocloud/internal/domain/environments"
	"github.com/mikrocloud/mikrocloud/internal/domain/environments/repository"
)

// EnvironmentService handles environment business logic
type EnvironmentService struct {
	repo repository.Repository
}

// NewEnvironmentService creates a new environment service
func NewEnvironmentService(repo repository.Repository) *EnvironmentService {
	return &EnvironmentService{
		repo: repo,
	}
}

// Command types for service operations
type CreateEnvironmentCommand struct {
	Name         string
	Description  string
	ProjectID    uuid.UUID
	IsProduction bool
	Variables    map[string]string
}

type UpdateEnvironmentCommand struct {
	ID          environments.EnvironmentID
	Description string
	Variables   map[string]string
}

// CreateEnvironment creates a new environment
func (s *EnvironmentService) CreateEnvironment(ctx context.Context, cmd CreateEnvironmentCommand) (*environments.Environment, error) {
	// Validate environment name
	envName, err := environments.NewEnvironmentName(cmd.Name)
	if err != nil {
		return nil, fmt.Errorf("invalid environment name: %w", err)
	}

	// Check if environment with same name already exists in this project
	existingEnvs, err := s.repo.FindByProjectID(ctx, cmd.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing environments: %w", err)
	}

	for _, env := range existingEnvs {
		if env.Name().String() == cmd.Name {
			return nil, fmt.Errorf("environment with name '%s' already exists in this project", cmd.Name)
		}
	}

	// Create new environment
	env := environments.NewEnvironment(envName, cmd.ProjectID, cmd.Description, cmd.IsProduction)

	// Set environment variables
	for key, value := range cmd.Variables {
		if err := env.SetVariable(key, value); err != nil {
			return nil, fmt.Errorf("failed to set environment variable '%s': %w", key, err)
		}
	}

	// Save to repository
	if err := s.repo.Save(ctx, env); err != nil {
		return nil, fmt.Errorf("failed to save environment: %w", err)
	}

	return env, nil
}

// GetEnvironment retrieves an environment by ID
func (s *EnvironmentService) GetEnvironment(ctx context.Context, id environments.EnvironmentID) (*environments.Environment, error) {
	env, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to find environment: %w", err)
	}
	return env, nil
}

// ListEnvironmentsByProject retrieves all environments for a project
func (s *EnvironmentService) ListEnvironmentsByProject(ctx context.Context, projectID uuid.UUID) ([]*environments.Environment, error) {
	envs, err := s.repo.FindByProjectID(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list environments: %w", err)
	}
	return envs, nil
}

// ListAllEnvironments retrieves all environments
func (s *EnvironmentService) ListAllEnvironments(ctx context.Context) ([]*environments.Environment, error) {
	envs, err := s.repo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list all environments: %w", err)
	}
	return envs, nil
}

// UpdateEnvironment updates an existing environment
func (s *EnvironmentService) UpdateEnvironment(ctx context.Context, cmd UpdateEnvironmentCommand) (*environments.Environment, error) {
	// Get existing environment
	env, err := s.repo.FindByID(ctx, cmd.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to find environment: %w", err)
	}

	// Update description if provided
	if cmd.Description != "" {
		env.UpdateDescription(cmd.Description)
	}

	// Update variables if provided
	if cmd.Variables != nil {
		// Clear existing variables and set new ones
		for key := range env.Variables() {
			env.RemoveVariable(key)
		}

		for key, value := range cmd.Variables {
			if err := env.SetVariable(key, value); err != nil {
				return nil, fmt.Errorf("failed to set environment variable '%s': %w", key, err)
			}
		}
	}

	// Save updated environment
	if err := s.repo.Save(ctx, env); err != nil {
		return nil, fmt.Errorf("failed to update environment: %w", err)
	}

	return env, nil
}

// DeleteEnvironment removes an environment
func (s *EnvironmentService) DeleteEnvironment(ctx context.Context, id environments.EnvironmentID) error {
	// Check if environment exists
	_, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to find environment: %w", err)
	}

	// Delete environment
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete environment: %w", err)
	}

	return nil
}

// ToggleProduction toggles the production status of an environment
func (s *EnvironmentService) ToggleProduction(ctx context.Context, id environments.EnvironmentID, isProduction bool) (*environments.Environment, error) {
	env, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to find environment: %w", err)
	}

	env.SetProduction(isProduction)

	if err := s.repo.Save(ctx, env); err != nil {
		return nil, fmt.Errorf("failed to update environment production status: %w", err)
	}

	return env, nil
}
