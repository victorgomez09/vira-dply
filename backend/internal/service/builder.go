package service

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/victorgomez09/vira-dply/internal/model"
)

func (s *DeployerService) buildAndPushImage(ctx context.Context, p *model.Project, buildPath string) (string, error) {
	// Usaremos un tag único basado en el tiempo
	tag := time.Now().Format("20060102-150405")
	imageName := fmt.Sprintf("%s/%s:%s", RegistryHost, p.Name, tag)

	log.Printf("   > Ejecutando nixpacks build y push para imagen: %s", imageName)

	// El comando nixpacks necesita `--name` para taggear la imagen, y se recomienda `--no-cache` en entornos CI/CD.
	cmd := exec.CommandContext(ctx, "nixpacks", "build", buildPath, "--name", imageName, "--no-cache")

	// Configurar PATH para encontrar el binario de Nixpacks y Docker/Podman
	cmd.Env = os.Environ()

	// Capturar la salida para logging y errores
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Nixpacks Output:\n%s", strings.TrimSpace(string(output)))
		return "", fmt.Errorf("error al ejecutar nixpacks (Build/Push): %w", err)
	}

	log.Printf("   > Imagen construida y subida exitosamente: %s", imageName)
	return imageName, nil
}
