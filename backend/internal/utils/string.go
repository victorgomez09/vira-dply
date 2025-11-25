package utils

import "strings"

// CleanString elimina caracteres no válidos para una etiqueta de contenedor.
func CleanString(name string) string {
	cleaned := strings.ReplaceAll(name, " ", "-")
	cleaned = strings.ToLower(cleaned)

	return cleaned
}
