package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	"github.com/radiophysiker/d56/internal/config"
	"github.com/radiophysiker/d56/internal/infrastructure/container"
)

func Run() error {
	// Initialize logger
	logger, err := zap.NewProduction()
	if err != nil {
		return fmt.Errorf("cannot create logger: %w", err)
	}
	zap.ReplaceGlobals(logger)
	defer func(logger *zap.Logger) {
		err := logger.Sync()
		if err != nil {
			logger.Error("cannot sync logger", zap.Error(err))
		}
	}(logger)

	// Load config
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("cannot load config: %w", err)
	}
	logger.Info("Loaded config", zap.Any("config", cfg))

	// Initialize dependency injection container
	// Здесь происходит подключение к БД в Infrastructure Layer
	diContainer, err := container.NewContainer(cfg, logger)
	if err != nil {
		return fmt.Errorf("cannot initialize DI container: %w", err)
	}
	defer diContainer.Close()
	logger.Info("Initialized application dependencies")

	// Get router from DI container
	httpHandler := diContainer.GetRouter().Setup()

	// Start server
	server := &http.Server{
		Addr:    cfg.RunAddress,
		Handler: httpHandler,
	}

	// Create context for graceful shutdown
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// Start accrual processor in background if configured
	if accrualProcessorService := diContainer.GetAccrualProcessorService(); accrualProcessorService != nil {
		go accrualProcessorService.Start(ctx)
		logger.Info("Started accrual processor service")
	}

	// Start server in a goroutine
	go func() {
		logger.Info("Starting server", zap.String("address", cfg.RunAddress))
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("HTTP server error", zap.Error(err))
			cancel()
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	<-ctx.Done()
	logger.Info("Shutting down server...")

	// The context is used to inform the server it has 5 seconds to finish
	// the request it is currently handling
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("Server forced to shutdown", zap.Error(err))
		return err
	}

	logger.Info("Server exited")
	return nil
}
