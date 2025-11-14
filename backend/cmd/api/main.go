package main

import (
	"context"
	"encoding/json"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/nats-io/nats.go"
	"github.com/victorgomez09/vira-dply/internal/application"
	"github.com/victorgomez09/vira-dply/internal/domain/order"
	"github.com/victorgomez09/vira-dply/internal/infrastructure/projection"
	"github.com/victorgomez09/vira-dply/internal/infrastructure/readmodel"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// Cargar variables de entorno
	if err := godotenv.Load(); err != nil {
		log.Println(".env file not found, using environment variables")
	}

	// Conectar base de datos de lectura
	readDB := setupReadDB()
	readRepo := readmodel.NewReadModelRepo(readDB)
	retryStore := projection.NewRetryStore(readDB)

	// Conectar a NATS
	nc, err := nats.Connect(os.Getenv("NATS_URL"))
	if err != nil {
		log.Fatal("Failed to connect to NATS:", err)
	}
	defer nc.Close()

	// Crear handler
	handler := application.NewProjectionHandler(readRepo, retryStore, nc)

	// Suscribirse a eventos
	_, err = nc.Subscribe("events.>", func(msg *nats.Msg) {
		var evtMap map[string]interface{}
		if err := json.Unmarshal(msg.Data, &evtMap); err != nil {
			log.Println("Failed to unmarshal event:", err)
			return
		}

		var evt interface{}
		switch evtMap["EventType"] {
		case "OrderCreatedEvent":
			evt = order.OrderCreatedEvent{OrderID: evtMap["OrderID"].(string)}
		case "OrderPaidEvent":
			evt = order.OrderPaidEvent{OrderID: evtMap["OrderID"].(string)}
		default:
			log.Println("Unknown event type:", evtMap["EventType"])
			return
		}

		if err := handler.ProcessEvent(context.Background(), evt); err != nil {
			log.Println("Failed to process event:", err)
			handler.HandleFailedEvent(evt, err, 0)
		}
	})
	if err != nil {
		log.Fatal("Failed to subscribe to events:", err)
	}

	log.Println("API listening for events...")
	select {}
}

func setupReadDB() *gorm.DB {
	dsn := os.Getenv("READMODEL_DSN")
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to Read Model DB:", err)
	}
	return db
}
