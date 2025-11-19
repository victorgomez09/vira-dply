package main_db

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
	"golang.org/x/exp/slog"

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

// PostgreSQLDatabase implements MainDatabase interface for PostgreSQL
type PostgreSQLDatabase struct {
	db                      *sql.DB
	projectRepository       projectsRepo.Repository
	applicationRepository   applicationsRepo.Repository
	databaseRepository      databasesRepo.DatabaseRepository
	environmentRepository   environmentsRepo.Repository
	templateRepository      servicesRepo.TemplateRepository
	userRepository          usersRepo.Repository
	sessionRepository       authRepo.SessionRepository
	authRepository          authRepo.AuthRepository
	deploymentRepository    deploymentsRepo.DeploymentRepository
	proxyRepository         proxyRepo.ProxyRepository
	traefikConfigRepository proxyRepo.TraefikConfigRepository
	diskRepository          disksRepo.DiskRepository
	diskBackupRepository    disksRepo.DiskBackupRepository
	organizationRepository  organizationsRepo.Repository
}

// NewPostgreSQLDatabase creates a new PostgreSQL database instance
func NewPostgreSQLDatabase(connectionString string) (MainDatabase, error) {
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to open PostgreSQL database: %w", err)
	}

	// Configure PostgreSQL connection
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping PostgreSQL database: %w", err)
	}

	slog.Info("PostgreSQL database connection established", "connection", maskPassword(connectionString))

	// Initialize repositories
	// Note: These would need PostgreSQL-specific implementations
	// For now, using SQLite implementations which may need adaptation
	projectRepo := projectsRepo.NewSQLiteProjectRepository(db)
	applicationRepo := applicationsRepo.NewSQLiteApplicationRepository(db)
	databaseRepo := databasesRepo.NewSQLiteDatabaseRepository(db)
	environmentRepo := environmentsRepo.NewSQLiteEnvironmentRepository(db)
	templateRepo := servicesRepo.NewSQLiteTemplateRepository(db)
	userRepo := usersRepo.NewSQLiteUserRepository(db)
	sessionRepo := authRepo.NewSQLiteSessionRepository(db)
	authRepository := authRepo.NewSQLiteAuthRepository(db)
	deploymentRepo := deploymentsRepo.NewSQLiteDeploymentRepository(db)
	proxyRepository := proxyRepo.NewSQLiteProxyRepository(db)
	traefikConfigRepository := proxyRepo.NewSQLiteTraefikConfigRepository(db)
	diskRepo := disksRepo.NewSQLiteDiskRepository(db)
	diskBackupRepo := disksRepo.NewSQLiteDiskBackupRepository(db)
	organizationRepo := organizationsRepo.NewSQLiteOrganizationRepository(db)

	return &PostgreSQLDatabase{
		db:                      db,
		projectRepository:       projectRepo,
		applicationRepository:   applicationRepo,
		databaseRepository:      databaseRepo,
		environmentRepository:   environmentRepo,
		templateRepository:      templateRepo,
		userRepository:          userRepo,
		sessionRepository:       sessionRepo,
		authRepository:          authRepository,
		deploymentRepository:    deploymentRepo,
		proxyRepository:         proxyRepository,
		traefikConfigRepository: traefikConfigRepository,
		diskRepository:          diskRepo,
		diskBackupRepository:    diskBackupRepo,
		organizationRepository:  organizationRepo,
	}, nil
}

func (d *PostgreSQLDatabase) Close() error {
	if err := d.db.Close(); err != nil {
		slog.Error("Error closing PostgreSQL database", "error", err)
		return err
	}
	slog.Info("PostgreSQL database connection closed")
	return nil
}

func (d *PostgreSQLDatabase) DB() *sql.DB {
	return d.db
}

func (d *PostgreSQLDatabase) Ping(ctx context.Context) error {
	return d.db.PingContext(ctx)
}

// Repository getters
func (d *PostgreSQLDatabase) ProjectRepository() projectsRepo.Repository {
	return d.projectRepository
}

func (d *PostgreSQLDatabase) ApplicationRepository() applicationsRepo.Repository {
	return d.applicationRepository
}

func (d *PostgreSQLDatabase) DatabaseRepository() databasesRepo.DatabaseRepository {
	return d.databaseRepository
}

func (d *PostgreSQLDatabase) EnvironmentRepository() environmentsRepo.Repository {
	return d.environmentRepository
}

func (d *PostgreSQLDatabase) TemplateRepository() servicesRepo.TemplateRepository {
	return d.templateRepository
}

func (d *PostgreSQLDatabase) UserRepository() usersRepo.Repository {
	return d.userRepository
}

func (d *PostgreSQLDatabase) SessionRepository() authRepo.SessionRepository {
	return d.sessionRepository
}

func (d *PostgreSQLDatabase) AuthRepository() authRepo.AuthRepository {
	return d.authRepository
}

func (d *PostgreSQLDatabase) DeploymentRepository() deploymentsRepo.DeploymentRepository {
	return d.deploymentRepository
}

func (d *PostgreSQLDatabase) ProxyRepository() proxyRepo.ProxyRepository {
	return d.proxyRepository
}

func (d *PostgreSQLDatabase) TraefikConfigRepository() proxyRepo.TraefikConfigRepository {
	return d.traefikConfigRepository
}

func (d *PostgreSQLDatabase) DiskRepository() disksRepo.DiskRepository {
	return d.diskRepository
}

func (d *PostgreSQLDatabase) DiskBackupRepository() disksRepo.DiskBackupRepository {
	return d.diskBackupRepository
}

func (d *PostgreSQLDatabase) OrganizationRepository() organizationsRepo.Repository {
	return d.organizationRepository
}

// maskPassword masks the password in connection string for logging
func maskPassword(connectionString string) string {
	// Simple password masking for logging - in production, use proper URL parsing
	// This is a basic implementation
	return "postgresql://***:***@***"
}
