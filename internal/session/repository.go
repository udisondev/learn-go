package session

import (
	"context"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/udisondev/learn-go/internal/user"
)

var psql = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

// Repository handles session data access operations
type Repository struct {
	db *pgxpool.Pool
}

// NewRepository creates new session repository
func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// Create creates a new session in database
// WHY: Store session for authentication
// HOW: Generate UUID, insert with user_id, ip, user_agent
//
// Returns session ID to set in cookie
func (r *Repository) Create(ctx context.Context, userID int64, ipAddress, userAgent string) (uuid.UUID, error) {
	sessionID := uuid.New()

	query, args, err := psql.
		Insert("sessions").
		Columns("id", "user_id", "ip_address", "user_agent").
		Values(sessionID, userID, ipAddress, userAgent).
		ToSql()

	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to build insert query: %w", err)
	}

	_, err = r.db.Exec(ctx, query, args...)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to create session: %w", err)
	}

	return sessionID, nil
}

// GetUserBySessionID retrieves user by session ID
// WHY: Authenticate user from cookie
// HOW: JOIN sessions with users table
//
// Returns user if session valid, error if not found
func (r *Repository) GetUserBySessionID(ctx context.Context, sessionID uuid.UUID) (*user.User, error) {
	query, args, err := psql.
		Select(
			"u.id",
			"u.name",
			"u.email",
			"u.password_hash",
			"u.phone",
			"u.registered_at",
			"u.updated_at",
			"u.sub_plan",
			"u.score",
			"u.is_verified",
			"u.avatar_url",
		).
		From("sessions s").
		Join("users u ON u.id = s.user_id").
		Where(sq.Eq{"s.id": sessionID}).
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("failed to build select query: %w", err)
	}

	var u user.User
	err = r.db.QueryRow(ctx, query, args...).Scan(
		&u.ID,
		&u.Name,
		&u.Email,
		&u.PasswordHash,
		&u.Phone,
		&u.RegisteredAt,
		&u.UpdatedAt,
		&u.SubPlan,
		&u.Score,
		&u.IsVerified,
		&u.AvatarURL,
	)

	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("session not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user by session: %w", err)
	}

	return &u, nil
}

// Delete deletes session by ID (for logout)
// WHY: Invalidate session on logout
// HOW: DELETE FROM sessions WHERE id = ?
func (r *Repository) Delete(ctx context.Context, sessionID uuid.UUID) error {
	query, args, err := psql.
		Delete("sessions").
		Where(sq.Eq{"id": sessionID}).
		ToSql()

	if err != nil {
		return fmt.Errorf("failed to build delete query: %w", err)
	}

	_, err = r.db.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	return nil
}
