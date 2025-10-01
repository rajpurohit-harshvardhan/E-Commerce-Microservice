package router

import (
	"common/utils/http/middleware"
	"net/http"
	"product/internal/db"
	"product/internal/usecases/product"
)

func SetupRouter(db db.Db) *http.ServeMux {
	router := http.NewServeMux()

	router.HandleFunc("GET /", product.HealthCheck())
	router.HandleFunc("GET /health", product.HealthCheck())
	router.HandleFunc("GET /health-check", product.HealthCheck())

	authOnly := middleware.Authenticated
	router.Handle("POST /v1/product", authOnly(product.New(db)))
	router.Handle("GET /v1/product/list", authOnly(product.ListProducts(db)))
	router.Handle("GET /v1/product/{id}", authOnly(product.GetProductById(db)))
	router.Handle("DELETE /v1/product/{id}", authOnly(product.DeleteProductById(db)))
	router.Handle("PUT /v1/product/{id}", authOnly(product.UpdateProductById(db)))

	return router
}
