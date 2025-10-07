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
	router.Handle("POST /api/v1/orders", authOnly(order.New(db)))
	router.Handle("DELETE /api/v1/orders/{id}", authOnly(order.DeleteOrderById(db)))
	router.Handle("GET /api/v1/orders/{id}", authOnly(order.GetOrderById(db)))
	router.Handle("GET /api/v1/orders/list", authOnly(order.ListOrders(db)))
	router.Handle("PUT /api/v1/orders/{id}", authOnly(order.UpdateOrderById(db)))
	return router
}
