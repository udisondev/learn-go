package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"github.com/udisondev/learn-go/internal/app"
	"github.com/udisondev/learn-go/pkg/config"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Setup context with graceful shutdown
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// Run application
	if err := app.Run(ctx, cfg); err != nil {
		log.Fatalf("Application error: %v", err)
	}
}
