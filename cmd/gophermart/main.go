package main

import (
	"gophermart-service/internal/app"
	"gophermart-service/internal/config"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	logger, err := config.NewLogger(false)
	if err != nil {
		logger.Fatal(err)
		panic(err)
	}
	defer config.SyncLogger(logger)

	httpApp := app.NewHTTPApp(logger)

	httpApp.SetupCommonMiddleware()
	httpApp.SetupRoutes()

	err = httpApp.Start()
	if err != nil {
		logger.Fatal(err)
		panic(err)
	}

	<-sigChan

	logger.Info("Shutting down server...")
	err = httpApp.Stop()
	if err != nil {
		logger.Fatal(err)
		panic(err)
	}
	logger.Info("Server stopped")
}
