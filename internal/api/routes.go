package api

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
	"github.com/mikrocloud/mikrocloud/internal/api/middleware"
	"github.com/mikrocloud/mikrocloud/internal/config"
	"github.com/mikrocloud/mikrocloud/internal/database"
	activitiesHandlers "github.com/mikrocloud/mikrocloud/internal/domain/activities/handlers"
	activitiesService "github.com/mikrocloud/mikrocloud/internal/domain/activities/service"
	appHandlers "github.com/mikrocloud/mikrocloud/internal/domain/applications/handlers"
	applicationsService "github.com/mikrocloud/mikrocloud/internal/domain/applications/service"
	authHandlers "github.com/mikrocloud/mikrocloud/internal/domain/auth/handlers"
	authService "github.com/mikrocloud/mikrocloud/internal/domain/auth/service"
	databaseHandlers "github.com/mikrocloud/mikrocloud/internal/domain/databases/handlers"
	databaseService "github.com/mikrocloud/mikrocloud/internal/domain/databases/service"
	deploymentHandlers "github.com/mikrocloud/mikrocloud/internal/domain/deployments/handlers"
	deploymentService "github.com/mikrocloud/mikrocloud/internal/domain/deployments/service"
	diskHandlers "github.com/mikrocloud/mikrocloud/internal/domain/disks/handlers"
	diskService "github.com/mikrocloud/mikrocloud/internal/domain/disks/service"
	"github.com/mikrocloud/mikrocloud/internal/domain/domains"
	envHandlers "github.com/mikrocloud/mikrocloud/internal/domain/environments/handlers"
	environmentService "github.com/mikrocloud/mikrocloud/internal/domain/environments/service"
	gitHandlers "github.com/mikrocloud/mikrocloud/internal/domain/git/handlers"
	gitRepo "github.com/mikrocloud/mikrocloud/internal/domain/git/repository"
	gitService "github.com/mikrocloud/mikrocloud/internal/domain/git/service"
	maintenanceHandlers "github.com/mikrocloud/mikrocloud/internal/domain/maintenance/handlers"
	organizationsHandlers "github.com/mikrocloud/mikrocloud/internal/domain/organizations/handlers"
	organizationsService "github.com/mikrocloud/mikrocloud/internal/domain/organizations/service"
	projectHandlers "github.com/mikrocloud/mikrocloud/internal/domain/projects/handlers"
	projectService "github.com/mikrocloud/mikrocloud/internal/domain/projects/service"
	proxyHandlers "github.com/mikrocloud/mikrocloud/internal/domain/proxy/handlers"
	proxyService "github.com/mikrocloud/mikrocloud/internal/domain/proxy/service"
	serversHandlers "github.com/mikrocloud/mikrocloud/internal/domain/servers/handlers"
	serversService "github.com/mikrocloud/mikrocloud/internal/domain/servers/service"
	serviceHandlers "github.com/mikrocloud/mikrocloud/internal/domain/services/handlers"
	"github.com/mikrocloud/mikrocloud/internal/domain/services/repository"
	servicesService "github.com/mikrocloud/mikrocloud/internal/domain/services/service"
	settingsHandlers "github.com/mikrocloud/mikrocloud/internal/domain/settings/handlers"
	settingsService "github.com/mikrocloud/mikrocloud/internal/domain/settings/service"
	buildService "github.com/mikrocloud/mikrocloud/pkg/containers/build"
	databaseContainers "github.com/mikrocloud/mikrocloud/pkg/containers/database"
	"github.com/mikrocloud/mikrocloud/pkg/containers/manager"
	proxyContainers "github.com/mikrocloud/mikrocloud/pkg/containers/proxy"
)

func SetupRoutes(api chi.Router, db *database.Database, cfg *config.Config, tokenAuth *jwtauth.JWTAuth, ctx context.Context) (*proxyContainers.TraefikService, *databaseService.StatusSyncService, error) {
	// Apply CORS middleware
	api.Use(middleware.CORS(cfg.Server.AllowedOrigins))

	// Create container manager
	containerManager, err := createContainerManager(cfg)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create container manager: %w", err)
	}

	// Create service instances
	envSvc := environmentService.NewEnvironmentService(db.EnvironmentRepository)
	projSvc := projectService.NewProjectService(db.ProjectRepository, db.EnvironmentRepository)
	authSvc := authService.NewAuthService(db.SessionRepository, db.AuthRepository, db.UserRepository, cfg.Auth.JWTSecret)

	// Create disk service (needed by database deployment service)
	diskSvc := diskService.NewDiskService(db.DiskRepository, db.DiskBackupRepository)

	// Create database container deployment service
	dbImageResolver := databaseContainers.NewDefaultImageResolver()
	dbConfigBuilder := databaseContainers.NewDefaultContainerConfigBuilder(dbImageResolver)
	dbDeploymentSvc := databaseContainers.NewDatabaseDeploymentService(containerManager, dbImageResolver, dbConfigBuilder, diskSvc)
	dbSvc := databaseService.NewDatabaseService(db.DatabaseRepository, dbDeploymentSvc, diskSvc)

	// Create database status sync service (will be started by server with proper context)
	logger := slog.Default()
	statusSyncSvc := databaseService.NewStatusSyncService(dbSvc, containerManager, logger, 30*time.Second)
	go statusSyncSvc.Start(ctx)

	// Create build service
	buildSvc := buildService.NewBuildService(containerManager, cfg.Docker.SocketPath)

	// Create deployment service
	deploymentSvc := deploymentService.NewDeploymentService(
		db.DeploymentRepository,
		buildSvc,
		containerManager,
	)

	// Create application service with deployment service as container recreator
	domainGenerator := domains.NewDomainGenerator(cfg.Server.PublicIP)
	appSvc := applicationsService.NewApplicationService(db.ApplicationRepository, domainGenerator, deploymentSvc)

	// Create QuickDeployService wrapper for ApplicationService
	quickDeployService := repository.NewQuickDeployService(db.TemplateRepository, appSvc)
	templateSvc := servicesService.NewTemplateService(db.TemplateRepository, quickDeployService)

	// Create proxy services
	proxySvc := proxyService.New(db.ProxyRepository, db.TraefikConfigRepository)
	traefikConfigDir := filepath.Join(cfg.Server.DataDir, "traefik")
	traefikSvc := proxyContainers.NewTraefikService(containerManager, traefikConfigDir, cfg.Docker.NetworkMode)

	// Create handler dependencies
	authHandler := authHandlers.NewAuthHandler(authSvc)
	projectHandler := projectHandlers.NewProjectHandler(projSvc)
	environmentHandler := envHandlers.NewEnvironmentHandler(envSvc)
	applicationHandler := appHandlers.NewApplicationHandler(appSvc, deploymentSvc, containerManager)
	databaseHandler := databaseHandlers.NewDatabaseHandler(dbSvc, containerManager)
	studioHandler := databaseHandlers.NewStudioHandler(dbSvc)
	deploymentHandler := deploymentHandlers.NewDeploymentHandler(deploymentSvc, appSvc)
	templateHandler := serviceHandlers.NewTemplateHandler(templateSvc)
	proxyHandler := proxyHandlers.NewProxyHandler(proxySvc)
	diskHandler := diskHandlers.NewDiskHandler(diskSvc, dbSvc, dbDeploymentSvc)
	maintenanceHandler := maintenanceHandlers.NewMaintenanceHandler(
		db.ProjectRepository,
		db.ApplicationRepository,
		db.DatabaseRepository,
		db.TemplateRepository,
		db.DB(),
		containerManager,
	)

	gitRepository := gitRepo.NewSQLiteGitRepository(db.DB())
	gitSvc := gitService.NewGitService(gitRepository)
	gitHandler := gitHandlers.NewGitHandler(gitSvc)

	settingsSvc := settingsService.NewSettingsService(db.SettingsRepository)
	settingsHandler := settingsHandlers.NewSettingsHandler(settingsSvc)

	activitiesSvc := activitiesService.NewActivitiesService(db.ActivitiesRepository)
	activitiesHandler := activitiesHandlers.NewActivitiesHandlers(activitiesSvc)

	serversSvc := serversService.NewServersService(db.ServersRepository)
	serversHandler := serversHandlers.NewServersHandler(serversSvc)

	organizationsSvc := organizationsService.NewOrganizationService(db.OrganizationRepository)
	organizationsHandler := organizationsHandlers.NewOrganizationHandler(organizationsSvc)

	// Protected routes that require authentication
	api.Group(func(r chi.Router) {
		r.Use(middleware.WebSocketTokenInjector())
		r.Use(jwtauth.Verifier(tokenAuth))
		r.Use(jwtauth.Authenticator(tokenAuth))
		r.Use(middleware.ExtractUserOrg())

		// Project routes
		r.Route("/projects", func(r chi.Router) {
			r.Get("/", projectHandler.List)
			r.Post("/", projectHandler.Create)
			r.Route("/{project_id}", func(r chi.Router) {
				r.Get("/", projectHandler.Get)
				r.Put("/", projectHandler.Update)
				r.Delete("/", projectHandler.Delete)

				// Environment routes within project
				r.Route("/environments", func(r chi.Router) {
					r.Get("/", environmentHandler.ListEnvironments)
					r.Post("/", environmentHandler.CreateEnvironment)
					r.Route("/{environment_id}", func(r chi.Router) {
						r.Get("/", environmentHandler.GetEnvironment)
						r.Put("/", environmentHandler.UpdateEnvironment)
						r.Delete("/", environmentHandler.DeleteEnvironment)
					})
				})

				// Application routes within project
				r.Route("/applications", func(r chi.Router) {
					r.Get("/", applicationHandler.ListApplications)
					r.Post("/", applicationHandler.CreateApplication)
					r.Route("/{application_id}", func(r chi.Router) {
						r.Get("/", applicationHandler.GetApplication)
						r.Put("/", applicationHandler.UpdateApplication)
						r.Delete("/", applicationHandler.DeleteApplication)
						r.Post("/deploy", applicationHandler.DeployApplication)
						r.Post("/start", applicationHandler.StartApplication)
						r.Post("/stop", applicationHandler.StopApplication)
						r.Post("/restart", applicationHandler.RestartApplication)
						r.Get("/logs", applicationHandler.GetApplicationLogs)
						r.Patch("/general", applicationHandler.UpdateGeneral)
						r.Post("/domain/generate", applicationHandler.GenerateDomain)
						r.Put("/domain", applicationHandler.AssignDomain)
						r.Put("/ports", applicationHandler.UpdatePorts)

						// Deployment routes within application
						r.Route("/deployments", func(r chi.Router) {
							r.Get("/", deploymentHandler.ListDeployments)
							r.Post("/", deploymentHandler.CreateDeployment)
							r.Route("/{deployment_id}", func(r chi.Router) {
								r.Get("/", deploymentHandler.GetDeployment)
								r.Post("/stop", deploymentHandler.StopDeployment)
								r.Post("/cancel", deploymentHandler.CancelDeployment)
								r.Get("/logs", deploymentHandler.GetDeploymentLogs)
								r.Get("/logs/stream", deploymentHandler.StreamDeploymentLogs)
							})
						})
					})
				})

				// Database routes within project
				r.Route("/databases", func(r chi.Router) {
					r.Get("/", databaseHandler.ListDatabases)
					r.Post("/", databaseHandler.CreateDatabase)
					r.Get("/types", databaseHandler.GetDatabaseTypes)
					r.Get("/types/{type}/config", databaseHandler.GetDefaultDatabaseConfig)
					r.Route("/{database_id}", func(r chi.Router) {
						r.Get("/", databaseHandler.GetDatabase)
						r.Put("/", databaseHandler.UpdateDatabase)
						r.Delete("/", databaseHandler.DeleteDatabase)
						r.Post("/action", databaseHandler.DatabaseAction)
						r.Get("/logs", databaseHandler.GetDatabaseLogs)
						r.Get("/terminal", databaseHandler.HandleTerminal)

						// Database studio routes
						r.Route("/studio", func(r chi.Router) {
							r.Get("/info", studioHandler.GetDatabaseInfo)
							r.Get("/schemas", studioHandler.ListSchemas)
							r.Get("/tables", studioHandler.ListTables)
							r.Get("/tables/{table_name}/schema", studioHandler.GetTableSchema)
							r.Get("/tables/{table_name}/data", studioHandler.GetTableData)
							r.Post("/tables/{table_name}/data", studioHandler.GetTableData)
							r.Post("/query", studioHandler.ExecuteQuery)
							r.Post("/tables/{table_name}/rows", studioHandler.InsertRow)
							r.Put("/tables/{table_name}/rows", studioHandler.UpdateRow)
							r.Delete("/tables/{table_name}/rows", studioHandler.DeleteRow)
						})
					})
				})

				// Proxy routes within project
				r.Route("/proxy", func(r chi.Router) {
					r.Get("/", proxyHandler.ListProxyConfigs)
					r.Post("/", proxyHandler.CreateProxyConfig)
					r.Route("/{config_id}", func(r chi.Router) {
						r.Get("/", proxyHandler.GetProxyConfig)
						r.Put("/", proxyHandler.UpdateProxyConfig)
						r.Delete("/", proxyHandler.DeleteProxyConfig)
					})
				})

				// Disk routes within project
				r.Route("/disks", func(r chi.Router) {
					r.Get("/", diskHandler.ListDisks)
					r.Post("/", diskHandler.CreateDisk)
					r.Route("/{disk_id}", func(r chi.Router) {
						r.Get("/", diskHandler.GetDisk)
						r.Put("/resize", diskHandler.ResizeDisk)
						r.Delete("/", diskHandler.DeleteDisk)
						r.Post("/attach", diskHandler.AttachDisk)
						r.Post("/detach", diskHandler.DetachDisk)
					})
				})
			})
		})

		// Template routes
		r.Route("/templates", func(r chi.Router) {
			r.Get("/", templateHandler.ListTemplates)
			r.Post("/", templateHandler.CreateTemplate)
			r.Get("/official", templateHandler.ListOfficialTemplates)
			r.Route("/{id}", func(r chi.Router) {
				r.Get("/", templateHandler.GetTemplate)
				r.Put("/", templateHandler.UpdateTemplate)
				r.Delete("/", templateHandler.DeleteTemplate)
				r.Post("/deploy", templateHandler.DeployTemplate)
				r.Post("/preview", templateHandler.PreviewDeployment)
			})
		})

		// Git routes
		r.Route("/git", func(r chi.Router) {
			r.Post("/validate", gitHandler.ValidateRepository)
			r.Post("/branches", gitHandler.ListBranches)
			r.Post("/detect-build", gitHandler.DetectBuildMethod)
			r.Post("/sources", gitHandler.CreateGitSource)
			r.Get("/sources", gitHandler.ListGitSources)
			r.Route("/sources/{source_id}", func(r chi.Router) {
				r.Get("/", gitHandler.GetGitSource)
				r.Put("/", gitHandler.UpdateGitSource)
				r.Delete("/", gitHandler.DeleteGitSource)
			})
		})
	})

	// Public routes (no authentication required)
	api.Route("/auth", func(r chi.Router) {
		r.Get("/setup", authHandler.GetSetupStatus)
		r.Post("/login", authHandler.Login)
		r.Post("/register", authHandler.Register)
		r.Post("/refresh", authHandler.RefreshToken)

		// Protected auth routes
		r.Group(func(r chi.Router) {
			r.Use(jwtauth.Verifier(tokenAuth))
			r.Use(jwtauth.Authenticator(tokenAuth))
			r.Post("/logout", authHandler.Logout)
			r.Get("/profile", authHandler.GetProfile)
		})
	})

	// Maintenance routes (protected)
	api.Route("/maintenance", func(r chi.Router) {
		r.Use(jwtauth.Verifier(tokenAuth))
		r.Use(jwtauth.Authenticator(tokenAuth))

		r.Get("/health", maintenanceHandler.HealthCheck)
		r.Get("/status", maintenanceHandler.SystemStatus)
		r.Get("/resources", maintenanceHandler.GetResources)
		r.Get("/info", maintenanceHandler.SystemInfo)

		r.Route("/domains", func(r chi.Router) {
			r.Get("/", maintenanceHandler.ListDomains)
			r.Post("/", maintenanceHandler.AddDomain)
			r.Delete("/{domain_id}", maintenanceHandler.RemoveDomain)
			r.Post("/{domain_id}/ssl", maintenanceHandler.EnableSSL)
		})
	})

	// Settings routes (protected)
	api.Route("/settings", func(r chi.Router) {
		r.Use(jwtauth.Verifier(tokenAuth))
		r.Use(jwtauth.Authenticator(tokenAuth))

		r.Get("/general", settingsHandler.GetGeneralSettings)
		r.Post("/general", settingsHandler.SaveGeneralSettings)
		r.Get("/advanced", settingsHandler.GetAdvancedSettings)
		r.Post("/advanced", settingsHandler.SaveAdvancedSettings)
		r.Get("/updates", settingsHandler.GetUpdateSettings)
		r.Post("/updates", settingsHandler.SaveUpdateSettings)
		r.Post("/backup", settingsHandler.CreateBackup)
		r.Post("/restore", settingsHandler.RestoreBackup)
	})

	// Activities routes (protected)
	api.Route("/activities", func(r chi.Router) {
		r.Use(jwtauth.Verifier(tokenAuth))
		r.Use(jwtauth.Authenticator(tokenAuth))

		r.Get("/{org_id}", activitiesHandler.GetRecentActivities)
		r.Get("/{resource_type}/{resource_id}", activitiesHandler.GetResourceActivities)
	})

	// Servers routes (protected)
	api.Route("/servers", func(r chi.Router) {
		r.Use(jwtauth.Verifier(tokenAuth))
		r.Use(jwtauth.Authenticator(tokenAuth))

		r.Get("/", serversHandler.ListServers)
		r.Post("/", serversHandler.CreateServer)
		r.Route("/{server_id}", func(r chi.Router) {
			r.Get("/", serversHandler.GetServer)
			r.Put("/", serversHandler.UpdateServer)
			r.Delete("/", serversHandler.DeleteServer)
		})
	})

	// Organizations routes (protected)
	api.Route("/organizations", func(r chi.Router) {
		r.Use(jwtauth.Verifier(tokenAuth))
		r.Use(jwtauth.Authenticator(tokenAuth))

		r.Get("/", organizationsHandler.ListOrganizations)
		r.Get("/{organization_id}", organizationsHandler.GetOrganization)
	})

	return traefikSvc, statusSyncSvc, nil
}

func createContainerManager(cfg *config.Config) (manager.ContainerManager, error) {
	switch cfg.Docker.Runtime {
	case "docker":
		return manager.NewDockerManager()
	case "podman":
		return manager.NewPodmanManager()
	default:
		return manager.NewDockerManager() // Default to Docker
	}
}
