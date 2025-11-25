package main

import (
	"context"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/victorgomez09/vira-dply/internal/api"
	"github.com/victorgomez09/vira-dply/internal/model"
	"github.com/victorgomez09/vira-dply/internal/service"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Advertencia: No se encontró archivo .env. Continuando con variables de entorno del sistema.")
	}

	log.Println("🚀 Iniciando PaaS Controller...")
	// Init k8s
	k8sClient, err := service.NewK8sClient()
	if err != nil {
		log.Fatalf("Fallo crítico al inicializar K8s Client: %v", err)
	}
	k8sClient.CreateNamespace(context.Background(), "production")

	// Init database
	db, err := gorm.Open(sqlite.Open("vira-dply.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	db.AutoMigrate(&model.Project{}, &model.User{})

	registryManager := service.NewRegistryManager()

	// Usamos un motor FIJO para la gestión inicial del registro local.
	// Esto asume que uno de los dos motores está disponible para iniciar el contenedor de registro.
	initialEngine := os.Getenv("CONTAINER_ENGINE")

	ctx := context.Background()

	// 2.1 Asegurar que el registro esté corriendo (usando el motor fijo)
	if err := registryManager.EnsureRegistryRunning(ctx, initialEngine); err != nil {
		log.Fatalf("Fallo crítico al iniciar el registro local: %v", err)
	}

	// 2.2 Autenticar la CLI local en el registro (usando el motor fijo)
	if err := registryManager.Authenticate(ctx, initialEngine); err != nil {
		log.Fatalf("Fallo crítico en la autenticación local del registro: %v", err)
	}

	// Init services
	deployerSvc := service.NewDeployerService(k8sClient, db)
	userSvc := service.NewUserService(db)

	e := echo.New()
	e.Use(middleware.CORS())
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	setupRoutes(e, deployerSvc, userSvc)
	e.Logger.Fatal(e.Start(":1323"))
}

func setupRoutes(e *echo.Echo, deployerSvc *service.DeployerService, userSvc *service.UserService) {
	ph := api.NewDeployerHandler(deployerSvc)
	uh := api.NewUserHandler(userSvc)

	a := e.Group("/auth")
	a.POST("/login", uh.Login)
	a.POST("/register", uh.Register)

	p := e.Group("/projects")
	p.POST("", ph.CreateProjectHandler)
	p.POST("/:id/deploy", ph.DeployProject)
	p.GET("", ph.GetProjectsHandler)
	p.GET("/:id", ph.GetProjectByIDHandler)
}
