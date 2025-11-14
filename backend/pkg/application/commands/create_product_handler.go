package commands

import (
	"github.com/victorgomez09/vira-dply/pkg/domain"
	"github.com/victorgomez09/vira-dply/pkg/domain/model"
	"github.com/victorgomez09/vira-dply/pkg/infrastructure/shared"
)

type CreateProductCommand struct {
	Name string
}

type CreateProductHandler struct {
	eventStore domain.EventStore
	publisher  domain.EventPublisher
}

func NewCreateProductHandler(es domain.EventStore, p domain.EventPublisher) *CreateProductHandler {
	return &CreateProductHandler{eventStore: es, publisher: p}
}

func (h *CreateProductHandler) Handle(cmd CreateProductCommand) error {
	productID := shared.NewID()
	// 1. Crear el agregado (genera evento ProductCreated)
	p := model.NewProduct(productID, cmd.Name)

	// 2. Persistir los eventos en el Event Store (GORM/PostgreSQL)
	if err := h.eventStore.Save(p); err != nil {
		return err
	}

	// 3. Publicar el evento al Message Broker (Kafka)
	return h.publisher.Publish(p.GetUncommittedChanges()[0])
}
