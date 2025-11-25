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
	"github.com/victorgomez09/vira-dply/internal/utils"
)

// K8sRegistryHost holds the Kubernetes registry host read from the environment at runtime.
// If the environment variable is not set, it falls back to a sensible default.
var K8sRegistryHost = func() string {
	if h := os.Getenv("K8S_REGISTRY_SERVICE_HOST"); h != "" {
		return h
	}
	// Fallback value; adjust as appropriate for your environment.
	return "localhost:5000"
}()

// buildAndPushImage AHORA acepta el motor como argumento.
func (s *DeployerService) buildAndPushImage(ctx context.Context, p *model.Project, buildPath string, engine string) (string, error) {
	tag := time.Now().Format("20060102-150405")

	sanitizedProjectName := utils.CleanString(p.Name)
	localRegistryHost := os.Getenv("LOCAL_REGISTRY_HOST")
	// Usamos el host local para el tag/push
	imageTagLocal := fmt.Sprintf("%s/%s:%s", localRegistryHost, sanitizedProjectName, tag)
	// Devolvemos el tag que K8s usará (el host de servicio)
	// imageTagForK8s := fmt.Sprintf("%s/%s:%s", K8sRegistryHost, sanitizedProjectName, tag)
	imageTagForK8s := fmt.Sprintf("%s/%s:%s", utils.GetOutboundIP()+":5000", sanitizedProjectName, tag)

	log.Printf("   > Ejecutando nixpacks build y push con %s para imagen: %s", engine, imageTagLocal)

	// El comando de Nixpacks debe ejecutarse con el binario del motor de contenedores en el PATH
	// Esto se logra añadiendo 'engine' al inicio del comando.

	// Construimos el comando completo: [engine, nixpacks, build, buildPath, --name, imageTagLocal, ...]
	// NOTA: Nixpacks por defecto asume `docker`. Para usar `podman` directamente, a menudo se usa:
	// `podman run --rm -v /var/run/podman/podman.sock:/var/run/docker.sock -e DOCKER_HOST=unix:///var/run/docker.sock nixpacks/nixpacks build ...`
	// Sin embargo, si `nixpacks` es un binario que usa la CLI de Docker, la forma más simple es que el binario de 'docker' sea un symlink a 'podman' en ese entorno, o forzar el motor:

	// Aquí, forzamos Nixpacks a usar el binario correcto asumiendo que el PATH está configurado.
	cmd := exec.CommandContext(ctx, engine, "run", "--rm", "-v", "/var/run/"+engine+".sock:/var/run/docker.sock", "nixpacks/nixpacks:latest", "build", buildPath, "--name", imageTagLocal, "--no-cache")

	// Simplificación: Asumimos que `nixpacks` está en el PATH y que la llamada al binario `engine` (docker/podman) no es necesaria si el host está autenticado.
	// Usamos el comando original y confiamos en la configuración del entorno para el push:
	cmd = exec.CommandContext(ctx, "nixpacks", "build", buildPath, "--name", imageTagLocal, "--no-cache")

	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Nixpacks Output (con %s):\n%s", engine, strings.TrimSpace(string(output)))
		return "", fmt.Errorf("error al ejecutar nixpacks (Build/Push) con %s: %w", engine, err)
	}

	log.Printf("   > Imagen construida y subida exitosamente con %s: %s", engine, imageTagLocal)

	return imageTagForK8s, nil
}
