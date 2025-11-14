package broker

import (
	"context"
	"encoding/json"

	"github.com/nats-io/nats.go"
)

type EventHandler func(ctx context.Context, evt interface{})

func SubscribeEvents(nc *nats.Conn, handler EventHandler) error {
	_, err := nc.Subscribe("events.order", func(msg *nats.Msg) {
		var raw map[string]interface{}
		json.Unmarshal(msg.Data, &raw)
		handler(context.Background(), raw)
	})
	return err
}
