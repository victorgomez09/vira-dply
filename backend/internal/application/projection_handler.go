package application

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/nats-io/nats.go"
	"github.com/victorgomez09/vira-dply/internal/domain/order"
	"github.com/victorgomez09/vira-dply/internal/infrastructure/projection"
	"github.com/victorgomez09/vira-dply/internal/infrastructure/readmodel"
)

type ProjectionHandler struct {
	readRepo    readmodel.ReadModelRepository
	retryStore  *projection.RetryStore
	nats        *nats.Conn
	MaxAttempts int
}

func NewProjectionHandler(readRepo readmodel.ReadModelRepository, retryStore *projection.RetryStore, nc *nats.Conn) *ProjectionHandler {
	return &ProjectionHandler{
		readRepo:    readRepo,
		retryStore:  retryStore,
		nats:        nc,
		MaxAttempts: 5,
	}
}

// Método público para que el worker u otros paquetes puedan usarlo
func (p *ProjectionHandler) ProcessEvent(ctx context.Context, evt interface{}) error {
	return p.processEvent(ctx, evt)
}

// Método privado con la lógica real de proyección
func (p *ProjectionHandler) processEvent(ctx context.Context, evt interface{}) error {
	switch e := evt.(type) {
	case order.OrderCreatedEvent:
		return p.readRepo.InsertOrder(ctx, e.OrderID)
	case order.OrderPaidEvent:
		return p.readRepo.UpdateOrderPaid(ctx, e.OrderID)
	default:
		return fmt.Errorf("unknown event type: %T", evt)
	}
}

// Manejo de fallo: guardar en retryStore y eventualmente en DLQ
func (p *ProjectionHandler) HandleFailedEvent(evt interface{}, err error, attempts int) error {
	data, _ := json.Marshal(evt)

	nextRetry := projection.NextBackoff(attempts)

	if attempts >= p.MaxAttempts {
		// Enviar a DLQ
		p.nats.Publish("events.dlq", data)
		return nil
	}

	// Guardar en retryStore para reintentos
	return p.retryStore.SaveFailed(
		getEventName(evt),
		data,
		attempts,
		p.MaxAttempts,
		nextRetry,
		err,
	)
}

func getEventName(evt interface{}) string {
	return fmt.Sprintf("%T", evt)
}
