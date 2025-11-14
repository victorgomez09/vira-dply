package eventstore

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/victorgomez09/vira-dply/pkg/domain"
	"github.com/victorgomez09/vira-dply/pkg/domain/events"
	"github.com/victorgomez09/vira-dply/pkg/domain/model"
	"github.com/victorgomez09/vira-dply/pkg/infrastructure/shared"
	"gorm.io/gorm"
)

type GormOutbox struct {
	ID          uint `gorm:"primaryKey;autoIncrement"`
	AggregateID string
	EventType   string
	Payload     []byte `gorm:"type:jsonb"`
	Timestamp   time.Time
	IsPublished bool `gorm:"default:false"` // Marcador de estado
}

func (GormOutbox) TableName() string { return "outbox" }

type GormEvent struct {
	ID            uint   `gorm:"primaryKey;autoIncrement"`
	AggregateID   string `gorm:"index:idx_aggregate_version,unique"`
	AggregateType string
	Version       int `gorm:"index:idx_aggregate_version,unique"` // Único para Concurrencia Optimista
	EventType     string
	Payload       []byte `gorm:"type:jsonb"`
	Timestamp     time.Time
}

func (GormEvent) TableName() string { return "events" }

type GormEventStore struct {
	db *gorm.DB
}

func NewGormEventStore(db *gorm.DB) domain.EventStore {
	db.AutoMigrate(&GormEvent{}) // Migración al inicio
	db.AutoMigrate(&GormOutbox{})

	return &GormEventStore{db: db}
}

func (s *GormEventStore) Save(aggregate domain.AggregateRoot) error {

	return s.db.Transaction(func(tx *gorm.DB) error {

		newEvents := aggregate.GetUncommittedChanges()

		for i, event := range newEvents {
			eventVersion := aggregate.GetVersion() - len(newEvents) + i + 1
			payload, err := json.Marshal(event)
			if err != nil {
				return err
			}

			gormEvent := GormEvent{
				AggregateID:   aggregate.GetID().String(),
				AggregateType: aggregate.GetType(),
				Version:       eventVersion,
				EventType:     event.GetType(),
				Payload:       payload,
				Timestamp:     event.GetTimestamp(),
			}

			// --- 1. Persistencia del Evento de Dominio (Event Sourcing) ---
			// Código de persistencia del GormEvent aquí...
			if result := tx.Create(&gormEvent); result.Error != nil {
				return errors.New("optimistic lock failed: version mismatch")
			}

			// --- 2. Persistencia del Mensaje de Outbox (Misma Transacción) ---
			outboxEntry := GormOutbox{
				AggregateID: aggregate.GetID().String(),
				EventType:   event.GetType(),
				Payload:     payload,
				Timestamp:   time.Now(),
				IsPublished: false,
			}
			if result := tx.Create(&outboxEntry); result.Error != nil {
				return errors.New("failed to save to outbox")
			}
		}
		aggregate.MarkChangesCommitted()
		return nil
	})
}

func (s *GormEventStore) Load(id shared.ID) (domain.AggregateRoot, error) {
	var gormEvents []GormEvent

	// 1. Consultar todos los eventos para el AggregateID, ordenados por versión.
	result := s.db.Where("aggregate_id = ?", id.String()).Order("version asc").Find(&gormEvents)
	if result.Error != nil {
		return nil, result.Error
	}

	if len(gormEvents) == 0 {
		return nil, errors.New("aggregate not found")
	}

	// 2. Determinar el tipo de agregado e inicializarlo.
	// En un sistema real, usarías un registro (map[string]func() AggregateRoot).
	if gormEvents[0].AggregateType != "Product" {
		return nil, fmt.Errorf("unsupported aggregate type: %s", gormEvents[0].AggregateType)
	}

	// Crear el agregado inicial con la versión 0.
	productAggregate := model.NewEmptyProduct(id)

	// 3. Aplicar los eventos para reconstruir el estado.
	for _, ge := range gormEvents {
		// Des-serializar el payload (JSONB) al tipo de evento Go concreto.
		event, err := s.deserializeEvent(ge.EventType, ge.Payload, id)
		if err != nil {
			return nil, fmt.Errorf("failed to deserialize event %s: %w", ge.EventType, err)
		}

		// Aplicar el evento, mutando el estado interno del agregado.
		productAggregate.ApplyChange(event)
	}

	// 4. Marcar los cambios como confirmados (ya están en la DB).
	productAggregate.MarkChangesCommitted()

	return productAggregate, nil
}

func (s *GormEventStore) deserializeEvent(eventType string, payload []byte, aggregateID shared.ID) (domain.Event, error) {
	switch eventType {
	case "ProductCreated":
		var e events.ProductCreated
		if err := json.Unmarshal(payload, &e); err != nil {
			return nil, err
		}
		return &e, nil

	// Agrega más casos aquí para otros eventos del producto (e.g., "ProductRenamed", etc.)

	default:
		return nil, fmt.Errorf("unknown event type for deserialization: %s", eventType)
	}
}
