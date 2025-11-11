package handler

import (
	"log/slog"
	"net/http"

	"github.com/google/uuid"
)

// HandleLogout handles user logout
// WHY: Invalidate current session and clear cookie
// HOW: Get session ID from cookie, delete from DB, clear cookie, redirect
func (h *Handler) HandleLogout(w http.ResponseWriter, r *http.Request) {
	// Get session cookie
	cookie, err := r.Cookie("session_id")
	if err != nil {
		// No session cookie - just redirect to home
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// Parse session ID
	sessionID, err := uuid.Parse(cookie.Value)
	if err != nil {
		// Invalid session ID - clear cookie and redirect
		clearSessionCookie(w)
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// Delete session from database
	if err := h.sessionService.DeleteSession(r.Context(), sessionID); err != nil {
		slog.Error("Failed to delete session", "error", err, "session_id", sessionID)
		// Continue anyway - clear cookie even if DB delete failed
	}

	// Clear session cookie
	clearSessionCookie(w)

	slog.Info("User logged out", "session_id", sessionID)

	// Redirect to home page
	http.Redirect(w, r, "/", http.StatusSeeOther)
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
