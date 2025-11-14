package events

import (
	"time"

	"github.com/victorgomez09/vira-dply/pkg/infrastructure/shared"
)

// ProductCreated representa el evento de creacion
type ProductCreated struct {
	ProductID shared.ID `json:"product_id"`
	Name      string    `json:"name"`
	Timestamp time.Time `json:"timestamp"`
}

func NewProductCreated(id shared.ID, name string) *ProductCreated {
	return &ProductCreated{
		ProductID: id,
		Name:      name,
		Timestamp: time.Now(),
	}
}

func (e *ProductCreated) GetAggregateID() shared.ID {
	return e.ProductID
}

func (e *ProductCreated) GetType() string {
	return "ProductCreated"
}

func (e *ProductCreated) GetTimestamp() time.Time {
	return e.Timestamp
}
