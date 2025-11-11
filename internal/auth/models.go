package auth

import "time"

// EmailVerification represents an email verification token
type EmailVerification struct {
	ID        int64
	UserID    int64
	Token     string
	ExpiresAt time.Time
	CreatedAt time.Time
}

// PasswordReset represents a password reset token
type PasswordReset struct {
	ID        int64
	UserID    int64
	Token     string
	ExpiresAt time.Time
	UsedAt    *time.Time
	CreatedAt time.Time
}
