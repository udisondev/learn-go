package handler

import (
	"github.com/udisondev/learn-go/internal/email"
	"github.com/udisondev/learn-go/internal/session"
	"github.com/udisondev/learn-go/internal/templates"
	"github.com/udisondev/learn-go/internal/user"
	"github.com/udisondev/learn-go/pkg/config"
)

// Handler holds dependencies for HTTP handlers
type Handler struct {
	templates      *templates.Templates
	userService    *user.Service
	sessionService *session.Service
	emailQueue     *email.Queue
	cfg            *config.Config
	// TODO: add more services when ready
	// courseService *course.Service
}

// New creates a new Handler instance
func New(tmpl *templates.Templates, userService *user.Service, sessionService *session.Service, emailQueue *email.Queue, cfg *config.Config) *Handler {
	return &Handler{
		templates:      tmpl,
		userService:    userService,
		sessionService: sessionService,
		emailQueue:     emailQueue,
		cfg:            cfg,
	}
}
