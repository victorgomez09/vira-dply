package main_db

import (
	"context"
	"database/sql"

	applicationsRepo "github.com/mikrocloud/mikrocloud/internal/domain/applications/repository"
	authRepo "github.com/mikrocloud/mikrocloud/internal/domain/auth/repository"
	databasesRepo "github.com/mikrocloud/mikrocloud/internal/domain/databases/repository"
	deploymentsRepo "github.com/mikrocloud/mikrocloud/internal/domain/deployments/repository"
	disksRepo "github.com/mikrocloud/mikrocloud/internal/domain/disks/repository"
	environmentsRepo "github.com/mikrocloud/mikrocloud/internal/domain/environments/repository"
	organizationsRepo "github.com/mikrocloud/mikrocloud/internal/domain/organizations/repository"
	projectsRepo "github.com/mikrocloud/mikrocloud/internal/domain/projects/repository"
	proxyRepo "github.com/mikrocloud/mikrocloud/internal/domain/proxy/repository"
	servicesRepo "github.com/mikrocloud/mikrocloud/internal/domain/services/repository"
	usersRepo "github.com/mikrocloud/mikrocloud/internal/domain/users/repository"
)

// MainDatabase represents the main application database
type MainDatabase interface {
	// Core database operations
	Close() error
	DB() *sql.DB
	Ping(ctx context.Context) error

	// Repository access
	ProjectRepository() projectsRepo.Repository
	ApplicationRepository() applicationsRepo.Repository
	DatabaseRepository() databasesRepo.DatabaseRepository
	EnvironmentRepository() environmentsRepo.Repository
	TemplateRepository() servicesRepo.TemplateRepository
	UserRepository() usersRepo.Repository
	SessionRepository() authRepo.SessionRepository
	AuthRepository() authRepo.AuthRepository
	DeploymentRepository() deploymentsRepo.DeploymentRepository
	ProxyRepository() proxyRepo.ProxyRepository
	TraefikConfigRepository() proxyRepo.TraefikConfigRepository
	DiskRepository() disksRepo.DiskRepository
	DiskBackupRepository() disksRepo.DiskBackupRepository
	OrganizationRepository() organizationsRepo.Repository
}

// DatabaseType represents the type of main database
type DatabaseType string

const (
	SQLite     DatabaseType = "sqlite"
	PostgreSQL DatabaseType = "postgres"
)

// Factory creates a main database instance based on configuration
type Factory interface {
	Create(dbType DatabaseType, connectionString string) (MainDatabase, error)
}
