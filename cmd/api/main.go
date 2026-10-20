package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"hexagonal-go-backend/internal/platform/config"
	"hexagonal-go-backend/internal/platform/database"

	"hexagonal-go-backend/internal/platform/logger"
	"hexagonal-go-backend/internal/platform/server"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("invalid configuration", "error", err)
		os.Exit(1)
	}

	log := logger.New(cfg.App.LogLevel, cfg.App.Env)
	autoMigrate := cfg.App.Env == "development"
	entClient, _, err := database.OpenPostgres(context.Background(), cfg.Database.URL, cfg.Database.MaxOpenConns, cfg.Database.MaxIdleConns, cfg.Database.ConnMaxLifetime, autoMigrate)
	if err != nil {
		log.Error("failed to connect to postgres", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := entClient.Close(); err != nil {
			log.Error("failed to close postgres", "error", err)
		}
	}()

	router, err := buildRouter(cfg, log, entClient)
	if err != nil {
		log.Error("failed to create development administrator", "error", err)
		os.Exit(1)
	}
	httpServer := server.New(router, cfg.App)

	shutdownSignal, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	serverErrors := make(chan error, 1)
	go func() {
		log.Info("http server started", "address", httpServer.Addr, "environment", cfg.App.Env)
		serverErrors <- httpServer.ListenAndServe()
	}()

	select {
	case err := <-serverErrors:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("http server failed", "error", err)
			os.Exit(1)
		}
	case <-shutdownSignal.Done():
		log.Info("shutdown signal received")
	}

	shutdownContext, cancel := context.WithTimeout(context.Background(), cfg.App.ShutdownTimeout)
	defer cancel()
	if err := httpServer.Shutdown(shutdownContext); err != nil {
		log.Error("graceful shutdown failed", "error", err)
		os.Exit(1)
	}
	log.Info("http server stopped")
}
