package eventstore

import (
	"context"
	"encoding/json"

	"github.com/segmentio/kafka-go"
	"github.com/victorgomez09/vira-dply/pkg/domain"
)

type KafkaPublisher struct{ writer *kafka.Writer }

func NewKafkaPublisher(broker string) domain.EventPublisher {
	writer := &kafka.Writer{
		Addr:     kafka.TCP(broker),
		Topic:    "domain_events", // Tema único para todos los eventos
		Balancer: &kafka.LeastBytes{},
	}
	return &KafkaPublisher{writer: writer}
}
func (p *KafkaPublisher) Publish(event domain.Event) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}
	message := kafka.Message{
		Key:     []byte(event.GetAggregateID().String()),
		Value:   payload,
		Headers: []kafka.Header{{Key: "EventType", Value: []byte(event.GetType())}},
	}
	return p.writer.WriteMessages(context.Background(), message)
}
