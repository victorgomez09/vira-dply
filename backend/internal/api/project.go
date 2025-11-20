package api

import (
	"log"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/victorgomez09/vira-dply/internal/dto"
	"github.com/victorgomez09/vira-dply/internal/service"
	"gorm.io/gorm"
)

type DeployerHandler struct {
	deployerSvc *service.DeployerService
}

func NewDeployerHandler(svc *service.DeployerService) *DeployerHandler {
	return &DeployerHandler{deployerSvc: svc}
}

// CreateProjectHandler maneja la petición POST para crear un proyecto.
func (h *DeployerHandler) CreateProjectHandler(c echo.Context) error {
	var req dto.CreateProjectRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Solicitud inválida. Asegúrate de enviar 'name', 'git_url', y 'branch'.")
	}

	// 1. Guardar el objeto Project inicial en la DB usando el servicio
	newProjectGORM, err := h.deployerSvc.SaveProject(&req)
	if err != nil {
		// Generalmente, un error 500 para fallos de DB o 400 si Name es duplicado (uniqueIndex)
		log.Printf("Error al guardar proyecto en DB: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Fallo al registrar el proyecto.")
	}

	log.Printf("Proyecto creado y registrado con ID %d. Iniciando proceso asíncrono...", newProjectGORM.ID)

	// 2. Iniciar el flujo de despliegue de forma ASÍNCRONA
	// Se pasa el objeto GORM recién creado que contiene el ID.
	h.deployerSvc.TriggerDeployment(newProjectGORM)

	// Devolver el objeto GORM (que ya tiene el ID y CreatedAt)
	return c.JSON(http.StatusAccepted, newProjectGORM)
}

// GetProjectsHandler lista todos los proyectos de la base de datos.
func (h *DeployerHandler) GetProjectsHandler(c echo.Context) error {
	projects, err := h.deployerSvc.GetAllProjects()
	if err != nil {
		log.Printf("Error al obtener proyectos de DB: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Fallo al obtener la lista de proyectos.")
	}

	// Si no hay proyectos, devuelve una lista vacía (StatusOK 200)
	return c.JSON(http.StatusOK, projects)
}

// GetProjectByIDHandler obtiene un proyecto específico por ID.
func (h *DeployerHandler) GetProjectByIDHandler(c echo.Context) error {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "ID de proyecto inválido.")
	}

	project, err := h.deployerSvc.GetProjectByID(uint(id))

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return echo.NewHTTPError(http.StatusNotFound, "Proyecto no encontrado.")
		}
		log.Printf("Error al buscar proyecto %d en DB: %v", id, err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Fallo interno del servidor.")
	}

	return c.JSON(http.StatusOK, project)
}
