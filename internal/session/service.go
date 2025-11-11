package session

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/udisondev/learn-go/internal/user"
)

// Service handles session business logic
type Service struct {
	repo *Repository
}

// NewService creates new session service
func NewService(db *pgxpool.Pool) *Service {
	return &Service{
		repo: NewRepository(db),
	}
}

// CreateSession creates new session for user
// WHY: Called after successful login or email verification
// HOW: Generate UUID session token and store in DB
//
// Returns session ID to set in cookie
func (s *Service) CreateSession(ctx context.Context, userID int64, ipAddress, userAgent string) (uuid.UUID, error) {
	return s.repo.Create(ctx, userID, ipAddress, userAgent)
}

// GetUserBySessionID retrieves user by session ID
// WHY: Authenticate user in middleware
// HOW: Query DB for session and return associated user
func (s *Service) GetUserBySessionID(ctx context.Context, sessionID uuid.UUID) (*user.User, error) {
	return s.repo.GetUserBySessionID(ctx, sessionID)
}

// DeleteSession deletes session (logout)
// WHY: Invalidate current session on logout
// HOW: Remove session from DB by ID
func (s *Service) DeleteSession(ctx context.Context, sessionID uuid.UUID) error {
	return s.repo.Delete(ctx, sessionID)
}
