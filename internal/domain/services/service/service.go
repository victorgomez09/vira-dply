package services

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/mikrocloud/mikrocloud/internal/domain/applications"
	"github.com/mikrocloud/mikrocloud/internal/domain/services"
	"github.com/mikrocloud/mikrocloud/internal/domain/services/repository"
)

// TemplateService handles service template business logic (marketplace)
type TemplateService struct {
	templateRepo       repository.TemplateRepository
	quickDeployService *repository.QuickDeployService
}

func NewTemplateService(
	templateRepo repository.TemplateRepository,
	quickDeployService *repository.QuickDeployService,
) *TemplateService {
	return &TemplateService{
		templateRepo:       templateRepo,
		quickDeployService: quickDeployService,
	}
}

// Template management operations

// CreateTemplate creates a new service template
func (s *TemplateService) CreateTemplate(ctx context.Context, template *services.ServiceTemplate) error {
	if template == nil {
		return fmt.Errorf("template cannot be nil")
	}

	// Check if template name already exists
	existing, err := s.templateRepo.FindByName(template.Name())
	if err == nil && existing != nil {
		return fmt.Errorf("template with name '%s' already exists", template.Name().String())
	}

	// Save the template
	if err := s.templateRepo.Save(template); err != nil {
		return fmt.Errorf("failed to save template: %w", err)
	}

	return nil
}

// GetTemplate retrieves a template by ID
func (s *TemplateService) GetTemplate(ctx context.Context, id services.TemplateID) (*services.ServiceTemplate, error) {
	template, err := s.templateRepo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to find template: %w", err)
	}
	return template, nil
}

// GetTemplateByName retrieves a template by name
func (s *TemplateService) GetTemplateByName(ctx context.Context, name services.TemplateName) (*services.ServiceTemplate, error) {
	template, err := s.templateRepo.FindByName(name)
	if err != nil {
		return nil, fmt.Errorf("failed to find template: %w", err)
	}
	return template, nil
}

// UpdateTemplate updates a template
func (s *TemplateService) UpdateTemplate(ctx context.Context, template *services.ServiceTemplate) error {
	if err := s.templateRepo.Update(template); err != nil {
		return fmt.Errorf("failed to update template: %w", err)
	}
	return nil
}

// DeleteTemplate deletes a template
func (s *TemplateService) DeleteTemplate(ctx context.Context, id services.TemplateID) error {
	if err := s.templateRepo.Delete(id); err != nil {
		return fmt.Errorf("failed to delete template: %w", err)
	}
	return nil
}

// ListTemplates lists all templates
func (s *TemplateService) ListTemplates(ctx context.Context) ([]*services.ServiceTemplate, error) {
	templates, err := s.templateRepo.List()
	if err != nil {
		return nil, fmt.Errorf("failed to list templates: %w", err)
	}
	return templates, nil
}

// ListTemplatesByCategory lists templates by category
func (s *TemplateService) ListTemplatesByCategory(ctx context.Context, category services.TemplateCategory) ([]*services.ServiceTemplate, error) {
	templates, err := s.templateRepo.FindByCategory(category)
	if err != nil {
		return nil, fmt.Errorf("failed to list templates by category: %w", err)
	}
	return templates, nil
}

// ListOfficialTemplates lists all official templates
func (s *TemplateService) ListOfficialTemplates(ctx context.Context) ([]*services.ServiceTemplate, error) {
	templates, err := s.templateRepo.ListOfficial()
	if err != nil {
		return nil, fmt.Errorf("failed to list official templates: %w", err)
	}
	return templates, nil
}

// Quick deployment operations

// DeployTemplate deploys a template as an application
func (s *TemplateService) DeployTemplate(ctx context.Context, templateID services.TemplateID, req services.DeploymentRequest) (*applications.Application, error) {
	app, err := s.quickDeployService.DeployTemplate(ctx, templateID, req)
	if err != nil {
		return nil, fmt.Errorf("failed to deploy template: %w", err)
	}
	return app, nil
}

// ValidateDeploymentRequest validates a deployment request for a template
func (s *TemplateService) ValidateDeploymentRequest(ctx context.Context, templateID services.TemplateID, req services.DeploymentRequest) error {
	// Validate that the request has all required fields
	if req.Name == "" {
		return fmt.Errorf("application name is required")
	}

	if req.ProjectID == uuid.Nil {
		return fmt.Errorf("project ID is required")
	}

	if req.EnvironmentID == uuid.Nil {
		return fmt.Errorf("environment ID is required")
	}

	// Templates can be either Git-based or Docker registry-based
	// No need to validate specific requirements here since CreateApplication handles both cases

	return nil
}

// PreviewDeployment shows what an application would look like when created from a template
func (s *TemplateService) PreviewDeployment(ctx context.Context, templateID services.TemplateID, req services.DeploymentRequest) (*services.DeploymentPreview, error) {
	// Get the template
	template, err := s.templateRepo.FindByID(templateID)
	if err != nil {
		return nil, fmt.Errorf("template not found: %w", err)
	}

	// Validate the request
	if err := s.ValidateDeploymentRequest(ctx, templateID, req); err != nil {
		return nil, err
	}

	// Create a preview without actually saving the application
	gitURL := template.GitURL()
	buildConfig := template.BuildConfig()

	environment := template.Environment()
	if req.Environment != nil {
		// Merge environments, with request overriding template
		merged := make(map[string]string)
		for k, v := range environment {
			merged[k] = v
		}
		for k, v := range req.Environment {
			merged[k] = v
		}
		environment = merged
	}

	preview := &services.DeploymentPreview{
		TemplateName:    template.Name(),
		ApplicationName: req.Name,
		ProjectID:       req.ProjectID,
		EnvironmentID:   req.EnvironmentID,
		GitURL:          gitURL,
		BuildConfig:     buildConfig,
		Environment:     environment,
		Ports:           template.Ports(),
		Volumes:         template.Volumes(),
	}

	return preview, nil
}
