package queries

import "errors"

type GetProductQuery struct {
	ID string
}

type ProductDTO struct { // Data Transfer Object para la respuesta
	ID   string `json:"id" bson:"_id"`
	Name string `json:"name" bson:"name"`
}

// Interfaz para el repositorio de lectura (MongoDB)
type ProductReadRepository interface {
	FindByID(id string) (*ProductDTO, error)
}
type GetProductHandler struct{ readRepo ProductReadRepository }

func NewGetProductHandler(repo ProductReadRepository) *GetProductHandler {
	return &GetProductHandler{readRepo: repo}
}

func (h *GetProductHandler) Handle(query GetProductQuery) (*ProductDTO, error) {
	// 1. Consultar directamente el modelo de lectura (MongoDB)
	dto, err := h.readRepo.FindByID(query.ID)
	if err != nil {
		return nil, errors.New("product not found")
	}
	return dto, nil
}
