package app

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/udisondev/learn-go/internal/email"
	"github.com/udisondev/learn-go/internal/handler"
	"github.com/udisondev/learn-go/internal/router"
	"github.com/udisondev/learn-go/internal/session"
	"github.com/udisondev/learn-go/internal/templates"
	"github.com/udisondev/learn-go/internal/user"
	"github.com/udisondev/learn-go/pkg/config"
	"github.com/udisondev/learn-go/pkg/postgres"
)

// Run initializes and runs the application
func Run(ctx context.Context, cfg *config.Config) error {
	// 1. Setup logger
	var logLevel slog.Level
	if err := logLevel.UnmarshalText([]byte(cfg.App.LogLevel)); err != nil {
		logLevel = slog.LevelInfo // default to info if parsing fails
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	}))
	slog.SetDefault(logger)

	slog.Info("Starting Learn Go application",
		"env", cfg.App.Env,
		"port", cfg.App.Port,
	)

	// 2. Initialize database
	db, err := postgres.New(ctx, cfg.DB)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	// 3. Initialize services
	userService := user.NewService(db)
	sessionService := session.NewService(db)

	// 4. Initialize email queue
	emailQueue := email.NewQueue(db)

	// 5. Load templates
	tmpl, err := templates.Init()
	if err != nil {
		return fmt.Errorf("failed to load templates: %w", err)
	}

	// 6. Initialize handler
	h := handler.New(tmpl, userService, sessionService, emailQueue, cfg)

	// 7. Initialize router
	r := router.New(h, sessionService)

	// 8. Create HTTP server
	srv := &http.Server{
		Addr:         ":" + cfg.App.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// 9. Start server in goroutine
	serverErrors := make(chan error, 1)
	go func() {
		slog.Info("Server started", "address", srv.Addr)
		serverErrors <- srv.ListenAndServe()
	}()

	// 10. Wait for shutdown signal or server error
	select {
	case err := <-serverErrors:
		return fmt.Errorf("server error: %w", err)
	case <-ctx.Done():
		slog.Info("Shutdown signal received")
	}

	// 11. Graceful shutdown - give server time to finish requests
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		srv.Close() // force close if graceful shutdown failed
		return fmt.Errorf("graceful shutdown failed: %w", err)
	}

	slog.Info("Server stopped gracefully")
	return nil
}
