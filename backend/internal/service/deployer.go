package service

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/victorgomez09/vira-dply/internal/dto"
	"github.com/victorgomez09/vira-dply/internal/model"
	"gorm.io/gorm"
)

const RegistryHost = "myregistry.com" // Reemplazar con tu registro de contenedores
const DefaultDomain = "localhost"     // Dominio base para Ingress

// DeployerService maneja la lógica de negocio para los despliegues.
type DeployerService struct {
	k8sClient *K8sClient
	db        *gorm.DB
}

func NewDeployerService(k8sClient *K8sClient, db *gorm.DB) *DeployerService {
	return &DeployerService{k8sClient: k8sClient, db: db}
}

// SaveProject crea un nuevo registro en la DB a partir de la solicitud inicial.
func (s *DeployerService) SaveProject(p *dto.CreateProjectRequest) (*model.Project, error) {
	newProject := &model.Project{
		Name:          p.Name,
		GitUrl:        p.GitURL,
		GitBranch:     p.Branch,
		GitSourcePath: p.SourcePath,
		Status:        "Pending", // Estado inicial
		K8sNamespace:  "proj-" + p.Name,
		PublicUrl:     "",
	}

	// db.DBClient es el cliente GORM inicializado globalmente
	result := s.db.Create(newProject)
	if result.Error != nil {
		return nil, fmt.Errorf("fallo al crear proyecto en DB: %w", result.Error)
	}
	return newProject, nil
}

// UpdateProjectStatus actualiza el estado y opcionalmente la URL después del despliegue.
func (s *DeployerService) UpdateProjectStatus(project *model.Project, status, url string) error {
	updates := map[string]interface{}{
		"Status":    status,
		"PublicURL": url,
	}
	// GORM usa la clave primaria del struct `project` para saber qué fila actualizar.
	result := s.db.Model(project).Updates(updates)
	if result.Error != nil {
		return fmt.Errorf("fallo al actualizar proyecto en DB: %w", result.Error)
	}
	return nil
}

// GetProjectByID recupera un proyecto por su ID. Usado por el API Handler.
func (s *DeployerService) GetProjectByID(id uint) (*model.Project, error) {
	var project model.Project
	result := s.db.First(&project, id)
	return &project, result.Error
}

// GetAllProjects recupera todos los proyectos de la base de datos. Usado por el API Handler.
func (s *DeployerService) GetAllProjects() ([]model.Project, error) {
	var projects []model.Project
	result := s.db.Find(&projects)
	return projects, result.Error
}

// TriggerDeployment inicia el flujo de Build/Deploy de manera asíncrona.
func (s *DeployerService) TriggerDeployment(p *model.Project) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancel()

	go func() {
		log.Printf("🤖 Iniciando despliegue para el proyecto: %s (NS: %s)", p.Name, p.K8sNamespace)

		// 1. Crear directorio temporal
		tempDir, err := os.MkdirTemp("", "vira-dkply")
		if err != nil {
			s.handleDeploymentError(p, "Fallo al crear dir temporal", err)
			return
		}
		defer os.RemoveAll(tempDir) // Limpiar el directorio al finalizar/fallar

		// 2. Clonar el repositorio
		if err := s.cloneRepository(ctx, p, tempDir); err != nil {
			s.handleDeploymentError(p, "Fallo al clonar repo", err)
			return
		}

		// 3. Determinar la ruta de construcción REAL
		// Aquí se combina el directorio temporal con la SourcePath
		buildPath := filepath.Join(tempDir, p.GitSourcePath)

		// Verificar si la ruta existe. Es crucial.
		if _, err := os.Stat(buildPath); os.IsNotExist(err) {
			s.handleDeploymentError(p, fmt.Sprintf("La subcarpeta '%s' no existe en el repositorio clonado.", p.GitSourcePath), err)
			return
		}

		// 4. Construir la imagen con Nixpacks, usando la ruta específica
		imageName, err := s.buildAndPushImage(ctx, p, buildPath)
		if err != nil {
			s.handleDeploymentError(p, "Fallo al construir/push imagen", err)
			return
		}

		// 5. Crear recursos en Kubernetes
		if err := s.k8sClient.CreateDeployment(ctx, p.K8sNamespace, p.Name, imageName); err != nil {
			s.handleDeploymentError(p, "Fallo al crear recursos de K8s", err)
			return
		}

		// 6. Finalizar y actualizar el estado
		p.Status = "Running"
		p.PublicUrl = fmt.Sprintf("https://%s.%s", p.Name, DefaultDomain)
		// **Actualizar estado en la base de datos** (Lógica de DB omitida aquí)
		log.Printf("✅ Despliegue completado para %s. URL: %s", p.Name, p.PublicUrl)

	}()
}

func (s *DeployerService) handleDeploymentError(p *model.Project, msg string, err error) {
	p.Status = "Failed"
	// **Actualizar estado en la base de datos**
	log.Printf("❌ ERROR en despliegue de %s: %s. Detalle: %v", p.Name, msg, err)
}
