package router

import (
	"net/http"
	"order/internal/db"
	"order/internal/usecases/order"
)

func SetupRouter(db db.Db) *http.ServeMux {
	router := http.NewServeMux()

	router.HandleFunc("GET /", order.HealthCheck())
	router.HandleFunc("GET /health", order.HealthCheck())
	router.HandleFunc("GET /health-check", order.HealthCheck())

	router.HandleFunc("POST /v1/order", order.New(db))
	router.HandleFunc("DELETE /v1/order/{id}", order.DeleteOrderById(db))
	router.HandleFunc("GET /v1/order/{id}", order.GetOrderById(db))
	router.HandleFunc("GET /v1/order/list", order.ListOrders(db))
	router.HandleFunc("PUT /v1/order/{id}", order.UpdateOrderById(db))
	return router
}
