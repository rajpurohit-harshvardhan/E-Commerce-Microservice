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
	router.Handle("POST /api/v1/products", authOnly(product.New(db)))
	router.Handle("GET /api/v1/products/list", authOnly(product.ListProducts(db)))
	router.Handle("GET /api/v1/products/{id}", authOnly(product.GetProductById(db)))
	router.Handle("DELETE /api/v1/products/{id}", authOnly(product.DeleteProductById(db)))
	router.Handle("PUT /api/v1/products/{id}", authOnly(product.UpdateProductById(db)))

	return router
}
