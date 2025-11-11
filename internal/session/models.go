package session

import (
	"time"

	"github.com/google/uuid"
)

// Session represents an active user session
// WHY: Store authentication state in database
// HOW: Session ID stored in HTTP cookie, used to identify user
type Session struct {
	ID        uuid.UUID // session token
	UserID    int64     // user who owns this session
	CreatedAt time.Time // when session was created
	IPAddress string    // IP address for security audit
	UserAgent string    // user agent for security audit
}
