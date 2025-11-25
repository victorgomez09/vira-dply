package service

import (
	"context"
	"encoding/base64" // Necesario para el secreto de Kubernetes
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// RegistryManager maneja la vida útil y autenticación del registro.
type RegistryManager struct {
	host     string
	port     int
	username string
	password string
	email    string
}

func NewRegistryManager() *RegistryManager {
	port, err := strconv.Atoi(os.Getenv("LOCAL_REGISTRY_PORT"))
	if err != nil {
		port = 5000
	}

	return &RegistryManager{
		host:     os.Getenv("LOCAL_REGISTRY_HOST"),
		port:     port,
		username: os.Getenv("REGISTRY_USERNAME"),
		password: os.Getenv("REGISTRY_PASSWORD"),
		email:    os.Getenv("REGISTRY_EMAIL"),
	}
}

// EnsureRegistryRunning verifica y levanta el contenedor 'registry:2' usando el motor especificado.
func (m *RegistryManager) EnsureRegistryRunning(ctx context.Context, engine string) error {
	registryName := "paas-local-registry"

	// 1. Detener/eliminar el antiguo por si acaso (ignorando errores)
	exec.Command(engine, "stop", registryName).Run()
	exec.Command(engine, "rm", registryName).Run()

	// 2. Ejecutar el registro (registry:2)
	fmt.Println(strconv.Itoa(m.port))
	cmd := exec.CommandContext(ctx, engine, "run", "-d", "-p", strconv.Itoa(m.port)+":5000", "--restart=always", "--name", registryName, "registry:2")

	log.Printf("Intentando levantar o asegurar el registro local (%d) con %s...", m.port, engine)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("fallo al levantar el registro local con %s: %s", engine, string(output))
	}

	log.Printf("Registro local asegurado y corriendo usando %s.", engine)
	return nil
}

// Authenticate localmente el cliente Docker/Podman usando el motor especificado.
func (m *RegistryManager) Authenticate(ctx context.Context, engine string) error {
	log.Printf("Autenticando cliente %s en %s...", engine, m.host)

	// Usamos --password-stdin para seguridad
	cmd := exec.CommandContext(ctx, engine, "login", m.host, "-u", m.username, "--password-stdin")

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("fallo al crear pipe para stdin: %w", err)
	}

	// Ejecutar el comando
	// Se usa el canal para leer el output después de escribir la contraseña
	var output strings.Builder
	cmd.Stdout = &output
	cmd.Stderr = &output

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("fallo al iniciar comando %s login: %w", engine, err)
	}

	// Escribir la contraseña
	if _, writeErr := stdin.Write([]byte(m.password + "\n")); writeErr != nil {
		stdin.Close()
		return fmt.Errorf("fallo al escribir contraseña: %w", writeErr)
	}
	stdin.Close()

	// Esperar a que el comando termine
	if err := cmd.Wait(); err != nil {
		log.Printf("Output de autenticación: %s", strings.TrimSpace(output.String()))
		return fmt.Errorf("fallo al autenticar %s en %s: %w", engine, m.host, err)
	}

	log.Println("Autenticación local exitosa.")
	return nil
}

// CreateK8sPullSecretData crea los datos necesarios para un ImagePullSecret en K8s.
func (m *RegistryManager) CreateK8sPullSecretData(k8sHost string) (map[string][]byte, error) {
	auth := base64.StdEncoding.EncodeToString([]byte(m.username + ":" + m.password))

	secretData := map[string]interface{}{
		"auths": map[string]interface{}{
			k8sHost: map[string]string{
				"username": m.username,
				"password": m.password,
				"email":    m.email,
				"auth":     auth,
			},
		},
	}

	dockerConfigJSON, err := json.Marshal(secretData)
	if err != nil {
		return nil, fmt.Errorf("error al serializar .dockerconfigjson: %w", err)
	}

	return map[string][]byte{
		".dockerconfigjson": dockerConfigJSON,
	}, nil
}
