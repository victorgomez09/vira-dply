package api

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/victorgomez09/vira-dply/internal/application"
	"github.com/victorgomez09/vira-dply/internal/domain/order"
)

type OrderAPI struct {
	cmd *application.OrderCommandHandler
}

func NewOrderAPI(cmd *application.OrderCommandHandler) *OrderAPI {
	return &OrderAPI{cmd}
}

func (a *OrderAPI) Register(e *echo.Echo) {
	e.POST("/orders/:id", a.CreateOrder)
	e.POST("/orders/:id/pay", a.PayOrder)
}

func (a *OrderAPI) CreateOrder(c echo.Context) error {
	id := c.Param("id")
	err := a.cmd.HandleCreateOrder(c.Request().Context(), order.CreateOrderCommand{OrderID: id})
	if err != nil {
		return c.JSON(500, err.Error())
	}
	return c.JSON(http.StatusCreated, map[string]string{"id": id})
}

func (a *OrderAPI) PayOrder(c echo.Context) error {
	id := c.Param("id")
	err := a.cmd.HandlePayOrder(c.Request().Context(), order.PayOrderCommand{OrderID: id})
	if err != nil {
		return c.JSON(500, err.Error())
	}
	return c.JSON(200, "ok")
}
