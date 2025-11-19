package main_db

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
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

// SQLiteDatabase implements MainDatabase interface for SQLite
type SQLiteDatabase struct {
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

// NewSQLiteDatabase creates a new SQLite database instance
func NewSQLiteDatabase(databaseURL string) (MainDatabase, error) {
	// Ensure data directory exists
	if err := ensureDataDir(databaseURL); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	db, err := sql.Open("sqlite3", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open SQLite database: %w", err)
	}

	// Configure SQLite connection
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping SQLite database: %w", err)
	}

	// Enable foreign keys and WAL mode for better performance
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
	}
	if _, err := db.Exec("PRAGMA journal_mode = WAL"); err != nil {
		return nil, fmt.Errorf("failed to set WAL mode: %w", err)
	}

	slog.Info("SQLite database connection established", "path", databaseURL)

	// Initialize repositories
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

	return &SQLiteDatabase{
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

func (d *SQLiteDatabase) Close() error {
	if err := d.db.Close(); err != nil {
		slog.Error("Error closing SQLite database", "error", err)
		return err
	}
	slog.Info("SQLite database connection closed")
	return nil
}

func (d *SQLiteDatabase) DB() *sql.DB {
	return d.db
}

func (d *SQLiteDatabase) Ping(ctx context.Context) error {
	return d.db.PingContext(ctx)
}

// Repository getters
func (d *SQLiteDatabase) ProjectRepository() projectsRepo.Repository {
	return d.projectRepository
}

func (d *SQLiteDatabase) ApplicationRepository() applicationsRepo.Repository {
	return d.applicationRepository
}

func (d *SQLiteDatabase) DatabaseRepository() databasesRepo.DatabaseRepository {
	return d.databaseRepository
}

func (d *SQLiteDatabase) EnvironmentRepository() environmentsRepo.Repository {
	return d.environmentRepository
}

func (d *SQLiteDatabase) TemplateRepository() servicesRepo.TemplateRepository {
	return d.templateRepository
}

func (d *SQLiteDatabase) UserRepository() usersRepo.Repository {
	return d.userRepository
}

func (d *SQLiteDatabase) SessionRepository() authRepo.SessionRepository {
	return d.sessionRepository
}

func (d *SQLiteDatabase) AuthRepository() authRepo.AuthRepository {
	return d.authRepository
}

func (d *SQLiteDatabase) DeploymentRepository() deploymentsRepo.DeploymentRepository {
	return d.deploymentRepository
}

func (d *SQLiteDatabase) ProxyRepository() proxyRepo.ProxyRepository {
	return d.proxyRepository
}

func (d *SQLiteDatabase) TraefikConfigRepository() proxyRepo.TraefikConfigRepository {
	return d.traefikConfigRepository
}

func (d *SQLiteDatabase) DiskRepository() disksRepo.DiskRepository {
	return d.diskRepository
}

func (d *SQLiteDatabase) DiskBackupRepository() disksRepo.DiskBackupRepository {
	return d.diskBackupRepository
}

func (d *SQLiteDatabase) OrganizationRepository() organizationsRepo.Repository {
	return d.organizationRepository
}

// ensureDataDir creates the directory for the SQLite database if it doesn't exist
func ensureDataDir(dbPath string) error {
	dir := filepath.Dir(dbPath)
	if dir == "." || dir == "/" {
		return nil // Current directory or root, no need to create
	}
	return os.MkdirAll(dir, 0755)
}
