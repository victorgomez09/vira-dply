// /cmd/consumer/main.go
package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/segmentio/kafka-go"
	"github.com/victorgomez09/vira-dply/pkg/domain/events"
	"github.com/victorgomez09/vira-dply/pkg/infrastructure/projectors"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	kafkaTopicMain  = "domain_events"
	kafkaTopicDLQ   = "domain_events_dlq"
	mongoDBName     = "readmodel_db"
	consumerGroupID = "product-read-model-group-1"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Advertencia: No se encontró archivo .env. Asumiendo variables de entorno ya configuradas.")
	}
	// Configuración
	mongoURL := os.Getenv("MONGO_URL")
	kafkaBroker := os.Getenv("KAFKA_BROKER")

	if kafkaBroker == "" || mongoURL == "" {
		log.Fatal("ERROR: Las variables de entorno KAFKA_BROKER y MONGO_URL deben estar configuradas.")
	}

	// 1. Inicializar Conexión a MongoDB
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(mongoURL))
	if err != nil {
		log.Fatalf("FATAL: Fallo al conectar a MongoDB: %v", err)
	}
	defer client.Disconnect(context.TODO())
	log.Println("Conexión exitosa a MongoDB.")

	// 2. Inicializar el Proyector
	projector := projectors.NewProductProjector(client, mongoDBName)

	// 3. Inicializar el Consumidor de Kafka (Reader)
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{kafkaBroker},
		Topic:    kafkaTopicMain,
		GroupID:  consumerGroupID,
		MinBytes: 10e3,
		MaxBytes: 10e6,
	})
	defer r.Close()

	// 4. Inicializar el Escritor para la DLQ (Dead Letter Queue)
	dlqWriter := &kafka.Writer{
		Addr:     kafka.TCP(kafkaBroker),
		Topic:    kafkaTopicDLQ,
		Balancer: &kafka.LeastBytes{},
	}
	defer dlqWriter.Close()

	log.Printf("Iniciando Consumer, escuchando en el tópico %s...", kafkaTopicMain)

	// Bucle principal de consumo
	for {
		// Leer mensaje con un timeout (opcional)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		m, err := r.ReadMessage(ctx)
		cancel()

		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				continue // Simplemente reintentar leer si hay timeout
			}
			log.Fatalf("FATAL: Error al leer mensaje de Kafka: %v", err)
		}

		// --- Paso A: Identificación del Evento ---
		var eventType string
		for _, h := range m.Headers {
			if h.Key == "EventType" {
				eventType = string(h.Value)
				break
			}
		}

		if eventType == "" {
			log.Printf("ERROR: Mensaje sin encabezado EventType. Enviando a DLQ.")
			sendToDLQ(dlqWriter, m)
			r.CommitMessages(context.Background(), m)
			continue
		}

		// --- Paso B: Deserialización y Proyección ---
		err = processMessage(projector, eventType, m.Value)

		if err != nil {
			// En producción, aquí se implementaría una estrategia de reintento.
			// Por simplicidad, un error de proyección (MongoDB) va directamente a DLQ.
			log.Printf("ERROR en proyección de %s (ID %s). Enviando a DLQ: %v", eventType, string(m.Key), err)
			sendToDLQ(dlqWriter, m)
		}

		// --- Paso C: Confirmación (Commit) ---
		// Solo confirmamos el mensaje si el proceso fue exitoso O si fue enviado a DLQ.
		r.CommitMessages(context.Background(), m)
	}
}

// processMessage realiza la deserialización y llama al método del Proyector
func processMessage(projector *projectors.ProductProjector, eventType string, payload []byte) error {
	switch eventType {
	case "ProductCreated":
		var event events.ProductCreated
		if err := json.Unmarshal(payload, &event); err != nil {
			// Esto es un error de dato (permanente)
			return fmt.Errorf("error de deserialización: %w", err)
		}
		// Llamar al proyector
		if err := projector.HandleProductCreated(&event); err != nil {
			// Esto es un error de infraestructura (ej. fallo de MongoDB)
			return fmt.Errorf("error de proyección: %w", err)
		}
		log.Printf("Proyectado ProductCreated para ID: %s", event.ProductID)
		return nil

	// Agrega aquí más casos para otros tipos de eventos (e.g., ProductRenamed)

	default:
		return fmt.Errorf("tipo de evento desconocido: %s", eventType)
	}
}

// sendToDLQ maneja el envío de mensajes que fallaron a la cola de errores
func sendToDLQ(writer *kafka.Writer, message kafka.Message) {
	// Clonar y modificar el mensaje para la DLQ
	dlqMessage := message
	dlqMessage.Topic = kafkaTopicDLQ

	// Opcional: Agregar un encabezado de error
	dlqMessage.Headers = append(dlqMessage.Headers, kafka.Header{
		Key:   "FailureTime",
		Value: []byte(time.Now().Format(time.RFC3339)),
	})

	if err := writer.WriteMessages(context.Background(), dlqMessage); err != nil {
		// ¡PELIGRO! Fallo al escribir en la DLQ. Esto requiere alerta operativa.
		log.Printf("FATAL: Fallo al enviar mensaje a la DLQ. El mensaje puede haberse perdido: %v", err)
	}
}
