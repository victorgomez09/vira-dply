package broker

import (
	"context"
	"encoding/json"

	"github.com/nats-io/nats.go"
)

type NATSPublisher struct {
	nc *nats.Conn
}

func NewNATSPublisher(nc *nats.Conn) *NATSPublisher {
	return &NATSPublisher{nc}
}

func (p *NATSPublisher) Publish(ctx context.Context, evt interface{}) error {
	data, _ := json.Marshal(evt)
	return p.nc.Publish("events.order", data)
}
