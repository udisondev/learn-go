package middleware

import (
	"html/template"
	"net/http"

	"github.com/gorilla/csrf"
	"github.com/udisondev/learn-go/pkg/config"
)

// CSRF returns configured CSRF protection middleware
func CSRF(cfg *config.CSRFConfig) func(http.Handler) http.Handler {
	csrfMiddleware := csrf.Protect(
		[]byte(cfg.Secret),
		csrf.Secure(cfg.Secure),
		csrf.SameSite(csrf.SameSiteStrictMode),
		csrf.Path("/"),
	)

	return csrfMiddleware
}

// CSRFToken returns the CSRF token for the current request
func CSRFToken(r *http.Request) string {
	return csrf.Token(r)
}

// CSRFTemplateTag returns the HTML input field for CSRF token
func CSRFTemplateTag(r *http.Request) template.HTML {
	return csrf.TemplateField(r)
}
