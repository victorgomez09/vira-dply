package application

import (
	"context"

	"github.com/victorgomez09/vira-dply/internal/domain/order"
)

type EventStore interface {
	Load(ctx context.Context, aggregateID string) ([]interface{}, int, error)
	Append(ctx context.Context, aggregateID string, expectedVersion int, events []interface{}) error
}

type BrokerPublisher interface {
	Publish(ctx context.Context, event interface{}) error
}

type OrderCommandHandler struct {
	store     EventStore
	publisher BrokerPublisher
}

func NewOrderCommandHandler(store EventStore, publisher BrokerPublisher) *OrderCommandHandler {
	return &OrderCommandHandler{store, publisher}
}

func (h *OrderCommandHandler) HandleCreateOrder(ctx context.Context, cmd order.CreateOrderCommand) error {
	agg, events := order.NewOrder(cmd.OrderID)

	if err := h.store.Append(ctx, agg.ID, agg.Version, events); err != nil {
		return err
	}

	for _, evt := range events {
		_ = h.publisher.Publish(ctx, evt)
	}

	return nil
}

func (h *OrderCommandHandler) HandlePayOrder(ctx context.Context, cmd order.PayOrderCommand) error {
	events, version, err := h.store.Load(ctx, cmd.OrderID)
	if err != nil {
		return err
	}

	agg := &order.Order{}
	for _, e := range events {
		agg.Apply(e)
	}

	newEvents, err := agg.Pay()
	if err != nil {
		return err
	}

	if err := h.store.Append(ctx, agg.ID, version, newEvents); err != nil {
		return err
	}

	for _, evt := range newEvents {
		_ = h.publisher.Publish(ctx, evt)
	}

	return nil
}
