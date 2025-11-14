package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/victorgomez09/vira-dply/internal/application"
	"github.com/victorgomez09/vira-dply/internal/domain/order"
	"github.com/victorgomez09/vira-dply/internal/infrastructure/projection"
	"github.com/victorgomez09/vira-dply/internal/infrastructure/readmodel"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// Cargar .env
	if err := godotenv.Load(); err != nil {
		log.Println(".env file not found, using environment variables")
	}

	readDB := setupReadDB()
	retryStore := projection.NewRetryStore(readDB)
	readRepo := readmodel.NewReadModelRepo(readDB)

	// Crear handler
	handler := application.NewProjectionHandler(readRepo, retryStore, nil) // NATS opcional para DLQ

	maxRetries := 5
	if mr := os.Getenv("MAX_RETRIES"); mr != "" {
		fmt.Sscanf(mr, "%d", &maxRetries)
	}
	handler.MaxAttempts = maxRetries

	log.Println("Worker started, polling failed events...")

	for {
		time.Sleep(5 * time.Second)
		failedEvents, err := retryStore.GetDueRetries()
		if err != nil {
			log.Println("Error loading retries:", err)
			continue
		}

		for _, item := range failedEvents {
			var evtMap map[string]interface{}
			if err := json.Unmarshal(item.EventData, &evtMap); err != nil {
				log.Println("Failed to unmarshal event:", err)
				continue
			}

			var evt interface{}
			switch evtMap["EventType"] {
			case "OrderCreatedEvent":
				evt = order.OrderCreatedEvent{OrderID: evtMap["OrderID"].(string)}
			case "OrderPaidEvent":
				evt = order.OrderPaidEvent{OrderID: evtMap["OrderID"].(string)}
			default:
				log.Println("Unknown event type:", evtMap["EventType"])
				retryStore.Delete(item.ID)
				continue
			}

			err := handler.ProcessEvent(context.Background(), evt)
			if err != nil {
				log.Println("Failed to process event:", err)
				item.Attempts++
				if item.Attempts >= handler.MaxAttempts {
					log.Println("Max retries reached, discarding event:", evtMap["EventType"])
					retryStore.Delete(item.ID)
				} else {
					backoff := time.Duration(item.Attempts*item.Attempts) * 10 * time.Second
					item.NextRetry = time.Now().Add(backoff)
					readDB.Save(&item)
					log.Printf("Retry #%d scheduled in %v\n", item.Attempts, backoff)
				}
			} else {
				retryStore.Delete(item.ID)
			}
		}
	}
}

func setupReadDB() *gorm.DB {
	dsn := os.Getenv("READMODEL_DSN")
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to Read Model DB:", err)
	}
	return db
}
