package main

import (
	"api/internal/container"
	"api/internal/routes"
	logger "api/pkg/logging"
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func startServer(server *http.Server) {
	logger.Info("Started server on " + server.Addr)
	err := server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		logger.Fatalf("Server failed to start: %v", err)
	}
}

func main() {
	// Initialize logger
	logger.Init(logger.Config{
		Level: "debug",
	})

	// Create dependency container
	deps, err := container.NewContainer()
	if err != nil {
		logger.Fatalf("Failed to initialize dependencies: %v", err)
	}
	defer deps.Close()

	// Setup routes with dependency injection
	router := routes.SetupRoutes(deps)

	server := &http.Server{
		Addr:    deps.Config.GetPort(),
		Handler: router,
	}

	go startServer(server)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("Shutting down server...")

	// Give tasks time to finish cleanup
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Fatalf("Server forced to shutdown: %v", err)
	}

	logger.Info("Server exiting")
}
