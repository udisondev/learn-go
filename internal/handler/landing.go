package handler

import (
	"log/slog"
	"net/http"

	"github.com/udisondev/learn-go/internal/templates"
	"github.com/udisondev/learn-go/internal/user"
)

// HandleLanding handles the landing page
func (h *Handler) HandleLanding(w http.ResponseWriter, r *http.Request) {
	// Get authenticated user from context (added by Auth middleware)
	u, _ := user.FromCtx(r.Context())

	data := &templates.LandingData{
		User: u, // nil if not authenticated
	}

	if err := h.templates.RenderLanding(w, data); err != nil {
		slog.Error("Failed to render landing page", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
