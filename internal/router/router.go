package router

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/udisondev/learn-go/internal/handler"
	mw "github.com/udisondev/learn-go/internal/middleware"
	"github.com/udisondev/learn-go/internal/session"
)

// New creates and configures the HTTP router
func New(h *handler.Handler, sessionService *session.Service) *chi.Mux {
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))
	r.Use(mw.Auth(sessionService)) // Auth middleware - adds user to context if session exists

	// Static files
	fileServer := http.FileServer(http.Dir("web/static"))
	r.Handle("/static/*", http.StripPrefix("/static/", fileServer))

	// Public routes
	r.Get("/", h.HandleLanding)
	r.Get("/register", h.HandleRegisterPage)
	r.Post("/register", h.HandleRegisterSubmit)
	r.Get("/verify-email", h.HandleVerifyEmail)
	r.Post("/logout", h.HandleLogout)
	// TODO: r.Get("/login", h.HandleLogin)
	// TODO: r.Post("/login", h.HandleLoginSubmit)

	// Protected routes (require authentication)
	// TODO: r.Group(func(r chi.Router) {
	//   r.Use(middleware.AuthMiddleware)
	//   r.Get("/course", h.HandleCourse)
	//   r.Get("/profile", h.HandleProfile)
	//   r.Post("/logout", h.HandleLogout)
	//   r.Post("/submit", h.HandleSubmitCode)
	//   etc...
	// })

	return r
}
