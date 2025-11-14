package model

import (
	"github.com/victorgomez09/vira-dply/pkg/domain"
	events "github.com/victorgomez09/vira-dply/pkg/domain/events"
	"github.com/victorgomez09/vira-dply/pkg/infrastructure/shared"
)

type Product struct {
	ID      shared.ID
	Name    string
	Version int
	changes []domain.Event
}

func (p *Product) GetID() shared.ID {
	return p.ID
}

func (p *Product) GetName() string {
	return p.Name
}

func (p *Product) GetVersion() int {
	return p.Version
}

// Este método identifica de manera única el tipo de agregado (entidad) en el sistema.
func (p *Product) GetType() string {
	// Debe devolver el nombre del agregado como un string constante.
	return "Product"
}

func (p *Product) GetUncommittedChanges() []domain.Event {
	return p.changes
}

func (p *Product) MarkChangesCommitted() {
	p.changes = nil
}

func NewEmptyProduct(id shared.ID) *Product {
	return &Product{
		ID:      id,
		Version: 0,
		changes: nil, // Sin cambios al inicio
	}
}

// Factory (creacion que genera un evento)
func NewProduct(id shared.ID, name string) *Product {
	event := events.NewProductCreated(id, name)
	p := &Product{Version: 0, changes: []domain.Event{event}}
	p.ApplyChange(event)

	return p
}

func (p *Product) ApplyChange(event domain.Event) {
	p.Version++
	switch e := event.(type) {
	case *events.ProductCreated:
		p.ID = e.ProductID
		p.Name = e.Name
		// ... más casos para otros eventos (e.g., ProductRenamed)
	}
}
