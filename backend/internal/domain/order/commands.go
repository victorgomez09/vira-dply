package order

type CreateOrderCommand struct {
	OrderID string
}

type PayOrderCommand struct {
	OrderID string
}
