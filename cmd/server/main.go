package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/minsau/dominator/internal/config"
	"github.com/minsau/dominator/internal/database"
	"github.com/minsau/dominator/internal/registrar"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	cfg := config.Load()
	logger.Info("starting dominator service", "port", cfg.Port)

	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		logger.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()
	logger.Info("connected to postgres database")

	initCtx, initCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer initCancel()

	if err := database.InitSchema(initCtx, db); err != nil {
		logger.Error("failed to initialize schema", "error", err)
		os.Exit(1)
	}
	logger.Info("database schema initialized")

	repo := registrar.NewPostgresRepository(db)
	controller := registrar.NewController(repo)
	handler := registrar.NewHandler(controller, logger)

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Port),
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		logger.Info("http server listening", "addr", server.Addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	<-stopChan
	logger.Info("shutting down server gracefully...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("error during server shutdown", "error", err)
	}

	logger.Info("server stopped")
}
