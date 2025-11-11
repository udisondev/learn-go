package middleware

import (
	"log/slog"
	"net/http"

	"github.com/google/uuid"
	"github.com/udisondev/learn-go/internal/session"
	"github.com/udisondev/learn-go/internal/user"
)

// Auth middleware checks for session cookie and loads user into context
// WHY: Make authenticated user available to all handlers
// HOW: Read session_id cookie, query DB, add user to context
//
// IMPORTANT: This middleware does NOT block unauthenticated requests
// It only adds user to context if session exists
// Use RequireAuth() middleware for protected routes
func Auth(sessionService *session.Service) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Try to get session cookie
			cookie, err := r.Cookie("session_id")
			if err != nil {
				// No cookie - continue as anonymous user
				next.ServeHTTP(w, r)
				return
			}

			// Parse UUID from cookie
			sessionID, err := uuid.Parse(cookie.Value)
			if err != nil {
				// Invalid UUID - clear cookie and continue as anonymous
				clearSessionCookie(w)
				next.ServeHTTP(w, r)
				return
			}

			// Get user by session ID
			u, err := sessionService.GetUserBySessionID(r.Context(), sessionID)
			if err != nil {
				// Invalid session - clear cookie and continue as anonymous
				slog.Debug("Invalid session", "session_id", sessionID, "error", err)
				clearSessionCookie(w)
				next.ServeHTTP(w, r)
				return
			}

			// Add user to context
			ctx := user.WithCtx(r.Context(), u)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// clearSessionCookie removes session cookie
func clearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})
}
