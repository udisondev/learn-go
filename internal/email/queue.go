package email

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Queue provides methods for working with the email_queue table
// This implements a simple task queue using PostgreSQL
type Queue struct {
	db *pgxpool.Pool
}

// NewQueue creates a new Queue instance
func NewQueue(db *pgxpool.Pool) *Queue {
	return &Queue{
		db: db,
	}
}

// Enqueue adds a new email task to the queue
// WHY: Web application calls this after user registration to schedule email sending
// HOW: Inserts a new row into email_queue with status='pending'
func (q *Queue) Enqueue(ctx context.Context, emailType EmailType, recipientEmail string, userID *int64, payload any) error {
	// Marshal payload to JSON
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}

	query, args, err := squirrel.Insert("email_queue").
		PlaceholderFormat(squirrel.Dollar).
		Columns(
			"email_type",
			"recipient_email",
			"user_id",
			"payload",
		).
		Values(
			emailType.String(),
			recipientEmail,
			userID,
			payloadBytes,
		).
		ToSql()

	if err != nil {
		return fmt.Errorf("build query: %w", err)
	}

	_, err = q.db.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("exec query: %w", err)
	}

	return nil
}

// Dequeue retrieves the next pending task from the queue
// WHY: Worker process calls this to get the next email to send
// HOW: Uses FOR UPDATE SKIP LOCKED to safely handle concurrent workers
//
// This query finds tasks that are:
// - status = 'pending'
// - next_retry_at <= NOW() (ready to be processed)
// - Orders by created_at (FIFO)
// - Locks the row (FOR UPDATE) so other workers can't take it
// - SKIP LOCKED means if another worker already locked a row, skip it
//
// After retrieving, the status is immediately set to 'processing'
func (q *Queue) Dequeue(ctx context.Context) (*Task, error) {
	tx, err := q.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Find and lock the next task using squirrel
	// Note: Squirrel doesn't support FOR UPDATE SKIP LOCKED, so we use Suffix
	query, args, err := squirrel.Select(
		"id",
		"email_type",
		"recipient_email",
		"user_id",
		"payload",
		"attempts",
		"max_attempts",
	).
		PlaceholderFormat(squirrel.Dollar).
		From("email_queue").
		Where(squirrel.Eq{"status": "pending"}).
		Where(squirrel.LtOrEq{"next_retry_at": time.Now().UTC()}).
		OrderBy("created_at ASC").
		Limit(1).
		Suffix("FOR UPDATE SKIP LOCKED").
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("build select query: %w", err)
	}

	var task Task
	var emailTypeStr string
	err = tx.QueryRow(ctx, query, args...).Scan(
		&task.ID,
		&emailTypeStr,
		&task.RecipientEmail,
		&task.UserID,
		&task.Payload,
		&task.Attempts,
		&task.MaxAttempts,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil // No tasks available
		}
		return nil, fmt.Errorf("query row: %w", err)
	}

	// Parse email type
	emailType, err := ParseEmailType(emailTypeStr)
	if err != nil {
		return nil, fmt.Errorf("parse email type: %w", err)
	}
	task.EmailType = emailType

	// Mark as processing
	updateQuery, updateArgs, err := squirrel.Update("email_queue").
		PlaceholderFormat(squirrel.Dollar).
		Set("status", "processing").
		Set("attempts", task.Attempts+1).
		Where(squirrel.Eq{"id": task.ID}).
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("build update query: %w", err)
	}

	_, err = tx.Exec(ctx, updateQuery, updateArgs...)
	if err != nil {
		return nil, fmt.Errorf("update status: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit transaction: %w", err)
	}

	task.Attempts++ // Increment for the current attempt
	return &task, nil
}

// MarkCompleted marks a task as successfully completed
// WHY: Worker calls this after successfully sending an email
// HOW: Sets status='completed' and processed_at=NOW()
func (q *Queue) MarkCompleted(ctx context.Context, taskID int64) error {
	query, args, err := squirrel.Update("email_queue").
		PlaceholderFormat(squirrel.Dollar).
		Set("status", "completed").
		Set("processed_at", time.Now()).
		Where(squirrel.Eq{"id": taskID}).
		ToSql()

	if err != nil {
		return fmt.Errorf("build query: %w", err)
	}

	_, err = q.db.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("exec query: %w", err)
	}

	return nil
}

// MarkFailed marks a task as failed and schedules retry if attempts remain
// WHY: Worker calls this when email sending fails (SMTP error, etc.)
// HOW: If attempts < max_attempts, sets status back to 'pending' with exponential backoff
//      If attempts >= max_attempts, sets status='failed' permanently
//
// Exponential backoff:
// - 1st retry: 1 minute
// - 2nd retry: 2 minutes
// - 3rd retry: 4 minutes
// - etc.
func (q *Queue) MarkFailed(ctx context.Context, taskID int64, attempts, maxAttempts int, errorMsg string) error {
	if attempts >= maxAttempts {
		// Exceeded max attempts, mark as permanently failed
		query, args, err := squirrel.Update("email_queue").
			PlaceholderFormat(squirrel.Dollar).
			Set("status", "failed").
			Set("error", errorMsg).
			Set("processed_at", time.Now()).
			Where(squirrel.Eq{"id": taskID}).
			ToSql()

		if err != nil {
			return fmt.Errorf("build query: %w", err)
		}

		_, err = q.db.Exec(ctx, query, args...)
		if err != nil {
			return fmt.Errorf("exec query: %w", err)
		}

		return nil
	}

	// Calculate next retry time with exponential backoff
	// 2^(attempts-1) minutes: 1m, 2m, 4m, 8m, etc.
	backoffMinutes := int(math.Pow(2, float64(attempts-1)))
	nextRetry := time.Now().Add(time.Duration(backoffMinutes) * time.Minute)

	query, args, err := squirrel.Update("email_queue").
		PlaceholderFormat(squirrel.Dollar).
		Set("status", "pending").
		Set("error", errorMsg).
		Set("next_retry_at", nextRetry).
		Where(squirrel.Eq{"id": taskID}).
		ToSql()

	if err != nil {
		return fmt.Errorf("build query: %w", err)
	}

	_, err = q.db.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("exec query: %w", err)
	}

	return nil
}
