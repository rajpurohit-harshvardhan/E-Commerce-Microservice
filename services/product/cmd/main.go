package main

import (
	"context"
	"errors"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"product/internal/config"
	"product/internal/db/postgres"
	"product/internal/handler/router"
	"syscall"
	"time"
)

func main() {
	// loading config
	cfg := config.MustLoad()

	// create db connection
	db, err := postgres.New(cfg)
	if err != nil {
		log.Fatal(err)
	}

	// router setup
	handler := router.SetupRouter(db)

	// server setup
	server := http.Server{
		Addr:    cfg.Addr,
		Handler: handler,
	}

	slog.Info("Starting server on", slog.String("address", cfg.Addr))

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		err := server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal("Failed to start server")
		}
	}()

	<-done
	slog.Info("Shutting down server")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		slog.Error("failed to shutdown server", slog.String("error", err.Error()))
	}
	slog.Info("Server gracefully Shutdown")
}
