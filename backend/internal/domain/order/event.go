package order

import "time"

// Event base (opcional)
type Event interface {
	EventType() string
}

// OrderCreatedEvent
type OrderCreatedEvent struct {
	OrderID string
	Date    time.Time
}

func (e OrderCreatedEvent) EventType() string {
	return "OrderCreatedEvent"
}

// OrderPaidEvent
type OrderPaidEvent struct {
	OrderID string
	Date    time.Time
}

func (e OrderPaidEvent) EventType() string {
	return "OrderPaidEvent"
}
