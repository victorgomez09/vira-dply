package main

import (
	"context"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/victorgomez09/vira-dply/pkg/application/commands"
	"github.com/victorgomez09/vira-dply/pkg/application/queries"
	"github.com/victorgomez09/vira-dply/pkg/infrastructure/api"
	eventpublisher "github.com/victorgomez09/vira-dply/pkg/infrastructure/event_publisher"
	eventstore "github.com/victorgomez09/vira-dply/pkg/infrastructure/event_store"
	"github.com/victorgomez09/vira-dply/pkg/infrastructure/repository"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Advertencia: No se encontró archivo .env. Asumiendo variables de entorno ya configuradas.")
	}
	// --- 1. Inicializar Infraestructura (Conexiones) ---
	pgURL := os.Getenv("POSTGRES_URL")
	db, err := gorm.Open(postgres.Open(pgURL), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL (Event Store): %v", err)
	}
	eventStore := eventstore.NewGormEventStore(db)

	kafkaBroker := os.Getenv("KAFKA_BROKER")
	eventPublisher := eventpublisher.NewKafkaPublisher(kafkaBroker)

	mongoURL := os.Getenv("MONGO_URL")
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(mongoURL))
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB (Read Model): %v", err)
	}
	defer client.Disconnect(context.TODO())

	// --- 2. Inyección de Dependencias (Handlers) ---
	createProductHandler := commands.NewCreateProductHandler(eventStore, eventPublisher)

	productReadRepo := repository.NewMongoProductRepository(client, "readmodel_db")
	getProductHandler := queries.NewGetProductHandler(productReadRepo)

	// --- 3. Inicializar Echo API ---
	e := echo.New()
	api.RegisterProductRoutes(e, createProductHandler, getProductHandler)

	log.Println("Starting API server on :8080")
	e.Logger.Fatal(e.Start(":8080"))
}
