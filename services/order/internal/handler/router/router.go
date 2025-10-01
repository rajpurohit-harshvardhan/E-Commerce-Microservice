package router

import (
	"common/utils/http/middleware"
	"net/http"
	"order/internal/db"
	"order/internal/usecases/order"
)

func SetupRouter(db db.Db) *http.ServeMux {
	router := http.NewServeMux()

	router.HandleFunc("GET /", order.HealthCheck())
	router.HandleFunc("GET /health", order.HealthCheck())
	router.HandleFunc("GET /health-check", order.HealthCheck())

	authOnly := middleware.Authenticated
	router.Handle("POST /v1/order", authOnly(order.New(db)))
	router.Handle("DELETE /v1/order/{id}", authOnly(order.DeleteOrderById(db)))
	router.Handle("GET /v1/order/{id}", authOnly(order.GetOrderById(db)))
	router.Handle("GET /v1/order/list", authOnly(order.ListOrders(db)))
	router.Handle("PUT /v1/order/{id}", authOnly(order.UpdateOrderById(db)))
	return router
}
