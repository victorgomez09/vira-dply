package api

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/victorgomez09/vira-dply/pkg/application/commands"
	"github.com/victorgomez09/vira-dply/pkg/application/queries"
)

// Manejador para crear un producto (POST /products)
func CreateProductHandler(c echo.Context, handler *commands.CreateProductHandler) error {
	var cmd commands.CreateProductCommand
	if err := c.Bind(&cmd); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request")
	}

	if err := handler.Handle(cmd); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Command execution failed")
	}

	// 202 Accepted: El comando fue aceptado y será procesado (asíncronamente por Kafka)
	return c.NoContent(http.StatusAccepted)
}

// Manejador para obtener un producto (GET /products/:id)
func GetProductHandler(c echo.Context, handler *queries.GetProductHandler) error {
	query := queries.GetProductQuery{ID: c.Param("id")}

	dto, err := handler.Handle(query)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}

	return c.JSON(http.StatusOK, dto)
}

func RegisterProductRoutes(e *echo.Echo, createH *commands.CreateProductHandler, getH *queries.GetProductHandler) {
	e.POST("/products", func(c echo.Context) error { return CreateProductHandler(c, createH) })
	e.GET("/products/:id", func(c echo.Context) error { return GetProductHandler(c, getH) })
}
