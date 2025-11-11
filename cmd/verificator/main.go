package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/udisondev/learn-go/internal/email"
	"github.com/udisondev/learn-go/pkg/config"
	"github.com/udisondev/learn-go/pkg/postgres"
)

func main() {
	// Setup context with cancellation for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		slog.Error("Failed to load config", "error", err)
		os.Exit(1)
	}

	// Setup logger
	var logLevel slog.Level
	if err := logLevel.UnmarshalText([]byte(cfg.App.LogLevel)); err != nil {
		logLevel = slog.LevelInfo
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	}))
	slog.SetDefault(logger)

	slog.Info("Starting email worker (verificator)")

	// Initialize database connection
	db, err := postgres.New(ctx, cfg.DB)
	if err != nil {
		slog.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	slog.Info("Database connected", "host", cfg.DB.Host, "port", cfg.DB.Port)

	// Initialize email queue
	queue := email.NewQueue(db)

	// Initialize SMTP client
	smtpClient, err := email.NewSMTPClient(&cfg.Email)
	if err != nil {
		slog.Error("Failed to create SMTP client", "error", err)
		os.Exit(1)
	}

	// Initialize email sender with templates
	sender, err := email.NewSender(smtpClient, "web/templates/email")
	if err != nil {
		slog.Error("Failed to create email sender", "error", err)
		os.Exit(1)
	}

	slog.Info("Email worker initialized",
		"smtp_host", cfg.Email.Host,
		"smtp_port", cfg.Email.Port,
		"poll_interval", cfg.Executor.PollInterval,
	)

	// Setup graceful shutdown
	// WHY: Allow worker to finish current task before exiting
	// HOW: Listen for SIGINT/SIGTERM and cancel context
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		slog.Info("Shutdown signal received, finishing current task...")
		cancel()
	}()

	// Main processing loop
	// WHY: Continuously poll for new tasks and process them
	// HOW: Use ticker with configurable interval to check for tasks
	slog.Info("Starting task processing loop")

	ticker := time.NewTicker(cfg.Executor.PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			slog.Info("Worker stopped gracefully")
			return

		case <-ticker.C:
			// Process one task
			if err := processNextTask(ctx, queue, sender); err != nil {
				slog.Error("Error processing task", "error", err)
			}
		}
	}
}

// processNextTask dequeues and processes a single email task
// WHY: Separates task processing logic for better testability
// HOW: Dequeue → Send → Mark completed/failed
func processNextTask(ctx context.Context, queue *email.Queue, sender *email.Sender) error {
	// Dequeue next task
	task, err := queue.Dequeue(ctx)
	if err != nil {
		return err
	}

	// No tasks available
	if task == nil {
		return nil
	}

	slog.Info("Processing email task",
		"task_id", task.ID,
		"email_type", task.EmailType.String(),
		"recipient", task.RecipientEmail,
		"attempt", task.Attempts,
	)

	// Send email
	if err := sender.Send(ctx, task); err != nil {
		// Email sending failed - mark for retry
		slog.Error("Failed to send email",
			"task_id", task.ID,
			"error", err,
			"attempts", task.Attempts,
			"max_attempts", task.MaxAttempts,
		)

		if markErr := queue.MarkFailed(ctx, task.ID, task.Attempts, task.MaxAttempts, err.Error()); markErr != nil {
			slog.Error("Failed to mark task as failed", "task_id", task.ID, "error", markErr)
		}

		if task.Attempts >= task.MaxAttempts {
			slog.Warn("Task permanently failed after max attempts",
				"task_id", task.ID,
				"attempts", task.Attempts,
			)
		} else {
			slog.Info("Task will be retried",
				"task_id", task.ID,
				"next_attempt", task.Attempts+1,
			)
		}

		return err
	}

	// Email sent successfully - mark as completed
	if err := queue.MarkCompleted(ctx, task.ID); err != nil {
		slog.Error("Failed to mark task as completed", "task_id", task.ID, "error", err)
		return err
	}

	slog.Info("Email sent successfully",
		"task_id", task.ID,
		"email_type", task.EmailType.String(),
		"recipient", task.RecipientEmail,
	)

	return nil
}
