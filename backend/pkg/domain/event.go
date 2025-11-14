package domain

import (
	"time"

	"github.com/victorgomez09/vira-dply/pkg/infrastructure/shared"
)

type Event interface {
	GetAggregateID() shared.ID
	GetType() string
	GetTimestamp() time.Time
}

// EventPublisher es la interfaz que permite publicar Eventos de Dominio
// a un Message Broker o sistema de cola.
type EventPublisher interface {
	// Publish toma un Evento de Dominio y lo envía al sistema de mensajería.
	Publish(event Event) error
}

// EventStore es la interfaz que define las operaciones para persistir y
// recuperar agregados basados en sus eventos.
type EventStore interface {
	// Save persiste los eventos no confirmados (uncommitted changes) de un AggregateRoot
	// y maneja el control de concurrencia optimista.
	Save(AggregateRoot) error

	// Load recupera un AggregateRoot dado su ID, reconstruyéndolo a partir
	// de su secuencia completa de eventos.
	Load(id shared.ID) (AggregateRoot, error)
}
