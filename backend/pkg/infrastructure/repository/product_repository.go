package repository

import (
	"context"
	"errors"
	"time"

	"github.com/victorgomez09/vira-dply/pkg/application/queries"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type MongoProductRepository struct{ collection *mongo.Collection }

func NewMongoProductRepository(client *mongo.Client, dbName string) *MongoProductRepository {
	db := client.Database(dbName)
	collection := db.Collection("products_view")
	return &MongoProductRepository{collection: collection}
}

// FindByID implementa la Query (usado por GetProductHandler)
func (r *MongoProductRepository) FindByID(id string) (*queries.ProductDTO, error) {
	var dto queries.ProductDTO
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&dto)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("product not found")
		}
		return nil, err
	}
	return &dto, nil
}

// Nota: Aquí iría la lógica del Proyector para guardar en MongoDB (otro servicio/worker)
