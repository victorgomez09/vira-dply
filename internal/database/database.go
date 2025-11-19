package database

import (
	"context"
	"database/sql"
	"fmt"

	"golang.org/x/exp/slog"

	activitiesRepo "github.com/mikrocloud/mikrocloud/internal/domain/activities/repository"
	analyticsRepo "github.com/mikrocloud/mikrocloud/internal/domain/analytics/repository"
	applicationsRepo "github.com/mikrocloud/mikrocloud/internal/domain/applications/repository"
	authRepo "github.com/mikrocloud/mikrocloud/internal/domain/auth/repository"
	databasesRepo "github.com/mikrocloud/mikrocloud/internal/domain/databases/repository"
	deploymentsRepo "github.com/mikrocloud/mikrocloud/internal/domain/deployments/repository"
	disksRepo "github.com/mikrocloud/mikrocloud/internal/domain/disks/repository"
	environmentsRepo "github.com/mikrocloud/mikrocloud/internal/domain/environments/repository"
	logsRepo "github.com/mikrocloud/mikrocloud/internal/domain/logs/repository"
	organizationsRepo "github.com/mikrocloud/mikrocloud/internal/domain/organizations/repository"
	projectsRepo "github.com/mikrocloud/mikrocloud/internal/domain/projects/repository"
	proxyRepo "github.com/mikrocloud/mikrocloud/internal/domain/proxy/repository"
	serversRepo "github.com/mikrocloud/mikrocloud/internal/domain/servers/repository"
	servicesRepo "github.com/mikrocloud/mikrocloud/internal/domain/services/repository"
	settingsRepo "github.com/mikrocloud/mikrocloud/internal/domain/settings/repository"
	usersRepo "github.com/mikrocloud/mikrocloud/internal/domain/users/repository"

	"github.com/mikrocloud/mikrocloud/internal/config"
	analyticsdb "github.com/mikrocloud/mikrocloud/internal/database/analytics"
	maindb "github.com/mikrocloud/mikrocloud/internal/database/main"
	queuedb "github.com/mikrocloud/mikrocloud/internal/database/queue"
)

type Database struct {
	mainDB      maindb.MainDatabase
	analyticsDB analyticsdb.AnalyticsDatabase
	queueDB     queuedb.QueueDatabase

	ProjectRepository       projectsRepo.Repository
	ApplicationRepository   applicationsRepo.Repository
	DatabaseRepository      databasesRepo.DatabaseRepository
	EnvironmentRepository   environmentsRepo.Repository
	TemplateRepository      servicesRepo.TemplateRepository
	UserRepository          usersRepo.Repository
	SessionRepository       authRepo.SessionRepository
	AuthRepository          authRepo.AuthRepository
	DeploymentRepository    deploymentsRepo.DeploymentRepository
	ProxyRepository         proxyRepo.ProxyRepository
	TraefikConfigRepository proxyRepo.TraefikConfigRepository
	DiskRepository          disksRepo.DiskRepository
	DiskBackupRepository    disksRepo.DiskBackupRepository
	OrganizationRepository  organizationsRepo.Repository
	MetricRepository        analyticsRepo.MetricRepository
	LogRepository           logsRepo.LogRepository
	SettingsRepository      *settingsRepo.SettingsRepository
	ActivitiesRepository    *activitiesRepo.ActivitiesRepository
	ServersRepository       *serversRepo.ServersRepository
}

func New(cfg *config.Config) (*Database, error) {
	// Initialize main database factory and create database
	mainFactory := maindb.NewDatabaseFactory()
	mainDB, err := mainFactory.Create(maindb.DatabaseType(cfg.Database.Type), cfg.Database.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize main database: %w", err)
	}

	// Initialize analytics database factory and create database
	analyticsFactory := analyticsdb.NewAnalyticsFactory()
	analyticsDB, err := analyticsFactory.Create(analyticsdb.DatabaseType(cfg.Analytics.Type), cfg.Analytics.URL)
	if err != nil {
		slog.Warn("Failed to initialize analytics database, continuing without analytics", "error", err)
		// Return error for now - analytics DB is required
		// TODO: Make analytics optional once DuckDB extension loading is fixed
		return nil, fmt.Errorf("failed to initialize analytics database: %w", err)
	}

	// Initialize queue database factory and create database
	queueFactory := queuedb.NewQueueFactory()
	queueDB, err := queueFactory.Create(queuedb.DatabaseType(cfg.Queue.Type), cfg.Queue.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize queue database: %w", err)
	}

	slog.Info("Multi-database system initialized",
		"main_db", cfg.Database.Type,
		"analytics_db", cfg.Analytics.Type,
		"queue_db", cfg.Queue.Type)

	// Create analytics metric repository
	var metricRepo analyticsRepo.MetricRepository
	var logRepo logsRepo.LogRepository
	if sqlDB, ok := analyticsDB.DB().(*sql.DB); ok {
		metricRepo = analyticsRepo.NewSQLiteMetricRepository(sqlDB)
		logRepo = logsRepo.NewAnalyticsLogRepository(sqlDB)
	} else {
		return nil, fmt.Errorf("analytics database does not provide SQL DB interface")
	}

	return &Database{
		mainDB:                  mainDB,
		analyticsDB:             analyticsDB,
		queueDB:                 queueDB,
		ProjectRepository:       mainDB.ProjectRepository(),
		ApplicationRepository:   mainDB.ApplicationRepository(),
		DatabaseRepository:      mainDB.DatabaseRepository(),
		EnvironmentRepository:   mainDB.EnvironmentRepository(),
		TemplateRepository:      mainDB.TemplateRepository(),
		UserRepository:          mainDB.UserRepository(),
		SessionRepository:       mainDB.SessionRepository(),
		AuthRepository:          mainDB.AuthRepository(),
		DeploymentRepository:    mainDB.DeploymentRepository(),
		ProxyRepository:         mainDB.ProxyRepository(),
		TraefikConfigRepository: mainDB.TraefikConfigRepository(),
		DiskRepository:          mainDB.DiskRepository(),
		DiskBackupRepository:    mainDB.DiskBackupRepository(),
		OrganizationRepository:  mainDB.OrganizationRepository(),
		MetricRepository:        metricRepo,
		LogRepository:           logRepo,
		SettingsRepository:      settingsRepo.NewSettingsRepository(mainDB.DB()),
		ActivitiesRepository:    activitiesRepo.NewActivitiesRepository(mainDB.DB()),
		ServersRepository:       serversRepo.NewServersRepository(mainDB.DB()),
	}, nil
}

func (db *Database) Close() {
	if err := db.mainDB.Close(); err != nil {
		slog.Error("Error closing main database", "error", err)
	}
	if err := db.analyticsDB.Close(); err != nil {
		slog.Error("Error closing analytics database", "error", err)
	}
	if err := db.queueDB.Close(); err != nil {
		slog.Error("Error closing queue database", "error", err)
	}
	slog.Info("All database connections closed")
}

func (db *Database) DB() *sql.DB {
	return db.mainDB.DB()
}

// Health check method
func (db *Database) Ping(ctx context.Context) error {
	return db.mainDB.Ping(ctx)
}

// Access to specialized databases
func (db *Database) MainDB() maindb.MainDatabase {
	return db.mainDB
}

func (db *Database) AnalyticsDB() analyticsdb.AnalyticsDatabase {
	return db.analyticsDB
}

func (db *Database) QueueDB() queuedb.QueueDatabase {
	return db.queueDB
}
