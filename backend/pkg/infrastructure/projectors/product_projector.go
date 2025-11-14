package projectors

import (
	"context"
	"time"

	"github.com/victorgomez09/vira-dply/pkg/domain/events"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// ProductProjector se encarga de actualizar la vista de lectura (MongoDB).
type ProductProjector struct {
	collection *mongo.Collection
}

func NewProductProjector(client *mongo.Client, dbName string) *ProductProjector {
	db := client.Database(dbName)
	// Apunta a la colección que contendrá la vista desnormalizada de productos
	collection := db.Collection("products_view")
	return &ProductProjector{collection: collection}
}

// HandleProductCreated: Lógica para manejar la creación de un producto.
func (p *ProductProjector) HandleProductCreated(event *events.ProductCreated) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Estructura optimizada para la lectura en MongoDB
	projection := bson.M{
		"_id":   event.ProductID.String(), // Usamos el ID del agregado como clave primaria
		"name":  event.Name,
		"stock": 0, // Campos adicionales que puedan necesitar la vista
	}

	// Insertar el nuevo documento en la colección de MongoDB
	_, err := p.collection.InsertOne(ctx, projection)
	return err
}

// Nota: Aquí se agregarían otros métodos Handle para eventos de actualización
// (e.g., HandleProductRenamed, HandleProductStockAdjusted, etc.) que harían UpdateOne.
