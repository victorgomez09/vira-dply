package projection

import "time"

// NextBackoff calcula el próximo retry con backoff exponencial
func NextBackoff(attempts int) time.Time {
	// Base de 10 segundos, exponencial al cuadrado de intentos
	backoff := time.Duration(attempts*attempts) * 10 * time.Second
	return time.Now().Add(backoff)
}
