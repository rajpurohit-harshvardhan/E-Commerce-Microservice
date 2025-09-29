package router

import (
	"auth/internal/db"
	"auth/internal/usecases/user"
	"net/http"
)

func SetupRouter(db db.Db) *http.ServeMux {
	router := http.NewServeMux()

	router.HandleFunc("GET /", user.HealthCheck())
	router.HandleFunc("GET /health", user.HealthCheck())
	router.HandleFunc("GET /health-check", user.HealthCheck())

	router.HandleFunc("POST /v1/user", user.New(db))
	router.HandleFunc("DELETE /v1/user/{id}", user.DeleteUserById(db))
	router.HandleFunc("GET /v1/user/{id}", user.GetUserById(db))
	router.HandleFunc("PUT /v1/user/{id}", user.UpdateUserById(db))
	return router
}
