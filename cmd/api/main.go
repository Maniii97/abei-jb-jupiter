package main

import (
	"api/pkg/logging"
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func startServer(server *http.Server) {
	logger.Info("Started server on :8080")
	err := server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		logger.Fatalf("Server failed to start: %v", err)
	}
}

func main() {
	// if err := config.Load(); err != nil {
	// 	log.Printf("Warning: Failed to load .env file: %v", err)
	// }

	// router := routes.SetupRoutes()

	logger.Init(logger.Config{
		Level: "debug",
	})

	server := &http.Server{
		Addr: ":8080",
		// Handler: router,
	}

	go startServer(server)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err := server.Shutdown(ctx)
	if err != nil {
		logger.Fatalf("Server forced to shutdown: %v", err)
	}

	logger.Info("Server exited")
}
