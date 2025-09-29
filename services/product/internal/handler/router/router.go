package router

import (
	"net/http"
	"product/internal/db"
	"product/internal/usecases/product"
)

func SetupRouter(db db.Db) *http.ServeMux {
	router := http.NewServeMux()

	router.HandleFunc("GET /", product.HealthCheck())
	router.HandleFunc("GET /health", product.HealthCheck())
	router.HandleFunc("GET /health-check", product.HealthCheck())

	router.HandleFunc("POST /v1/product", product.New(db))
	router.HandleFunc("GET /v1/product/list", product.ListProducts(db))
	router.HandleFunc("GET /v1/product/{id}", product.GetProductById(db))
	router.HandleFunc("DELETE /v1/product/{id}", product.DeleteProductById(db))
	router.HandleFunc("PUT /v1/product/{id}", product.UpdateProductById(db))

	return router
}
