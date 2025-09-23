package main

import (
	"context"
	"gophermart-service/internal/app"
	"gophermart-service/internal/config"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	logger, err := config.NewLogger(false)
	if err != nil {
		// Если не можем создать логгер, выводим в stderr и завершаемся
		panic("failed to create logger: " + err.Error())
	}
	defer config.SyncLogger(logger)

	httpApp, err := app.NewHTTPApp(logger)
	if err != nil {
		logger.Error("Failed to create HTTP app", "error", err)
		os.Exit(1)
	}

	httpApp.SetupCommonMiddleware()
	httpApp.SetupRoutes()

	err = httpApp.Start()
	if err != nil {
		logger.Error("Failed to start server", "error", err)
		os.Exit(1)
	}

	<-ctx.Done()

	logger.Info("Shutting down server...")
	err = httpApp.Stop()
	if err != nil {
		logger.Error("Failed to stop server gracefully", "error", err)
		os.Exit(1)
	}
	logger.Info("Server stopped")
}
