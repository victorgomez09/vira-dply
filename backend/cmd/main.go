package main

import (
	"log"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/victorgomez09/vira-dply/internal/api"
	"github.com/victorgomez09/vira-dply/internal/model"
	"github.com/victorgomez09/vira-dply/internal/service"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	// Init k8s
	k8sClient, err := service.NewK8sClient()
	if err != nil {
		log.Fatalf("Fallo crítico al inicializar K8s Client: %v", err)
	}

	// Init database
	db, err := gorm.Open(sqlite.Open("vira-dply.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	db.AutoMigrate(&model.Project{}, &model.User{})

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
	p.GET("", ph.GetProjectsHandler)
	p.GET("/:id", ph.GetProjectByIDHandler)
}
