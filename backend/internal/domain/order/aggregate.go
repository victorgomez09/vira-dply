package order

import "errors"

type OrderStatus string

const (
	OrderCreated OrderStatus = "CREATED"
	OrderPaid    OrderStatus = "PAID"
)

var ErrInvalidTransition = errors.New("invalid status transition")

type Order struct {
	ID      string
	Status  OrderStatus
	Version int
}

func NewOrder(id string) (*Order, []interface{}) {
	event := OrderCreatedEvent{OrderID: id}
	return &Order{ID: id, Status: OrderCreated}, []interface{}{event}
}

func (o *Order) Apply(event interface{}) {
	switch e := event.(type) {
	case OrderCreatedEvent:
		o.ID = e.OrderID
		o.Status = OrderCreated
	case OrderPaidEvent:
		o.Status = OrderPaid
	}
	o.Version++
}

func (o *Order) Pay() ([]interface{}, error) {
	if o.Status != OrderCreated {
		return nil, ErrInvalidTransition
	}
	return []interface{}{
		OrderPaidEvent{OrderID: o.ID},
	}, nil
}
