package router

import (
	"auth/internal/db"
	"auth/internal/usecases/auth"
	"auth/internal/usecases/user"
	"common/utils/http/middleware"
	"net/http"
)

func SetupRouter(db db.Db) *http.ServeMux {
	router := http.NewServeMux()

	router.HandleFunc("GET /", user.HealthCheck())
	router.HandleFunc("GET /health", user.HealthCheck())
	router.HandleFunc("GET /health-check", user.HealthCheck())
	authOnly := middleware.Authenticated

	//router.HandleFunc("POST /v1/user", user.New(db))
	router.Handle("DELETE /api/v1/user/{id}", authOnly(user.DeleteUserById(db)))
	router.Handle("GET /api/v1/user/{id}", authOnly(user.GetUserById(db)))
	router.Handle("PUT /api/v1/user/{id}", authOnly(user.UpdateUserById(db)))

	router.HandleFunc("POST /api/v1/user/register", user.New(db))
	router.HandleFunc("POST /api/v1/user/login", auth.Login(db))
	router.Handle("POST /api/v1/user/refresh", authOnly(auth.RefreshTokens(db)))
	router.HandleFunc("POST /api/v1/user/logout", auth.Logout(db))
	return router
}
