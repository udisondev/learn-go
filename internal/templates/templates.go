package templates

import (
	"html/template"
	"net/http"

	"github.com/Masterminds/sprig/v3"
	"github.com/udisondev/learn-go/internal/user"
)

type Templates struct {
	landingTmpl  *template.Template
	registerTmpl *template.Template
	loginTmpl    *template.Template
}

// Init parses and loads all templates
func Init() (*Templates, error) {
	funcMap := sprig.FuncMap()

	// Add custom template functions
	funcMap["firstRune"] = func(s string) string {
		runes := []rune(s)
		if len(runes) == 0 {
			return ""
		}
		return string(runes[0])
	}

	// Parse landing page templates
	landingTmpl, err := template.New("").Funcs(funcMap).ParseFiles(
		"web/templates/layouts/base.html",
		"web/templates/components/header.html",
		"web/templates/components/feature-card.html",
		"web/templates/pages/landing.html",
	)
	if err != nil {
		return nil, err
	}

	// Parse register page templates
	registerTmpl, err := template.New("").Funcs(funcMap).ParseFiles(
		"web/templates/layouts/auth.html",
		"web/templates/components/register-form.html",
		"web/templates/pages/register.html",
	)
	if err != nil {
		return nil, err
	}

	// Parse login page templates
	loginTmpl, err := template.New("").Funcs(funcMap).ParseFiles(
		"web/templates/layouts/auth.html",
		"web/templates/components/login-form.html",
		"web/templates/pages/login.html",
	)
	if err != nil {
		return nil, err
	}

	return &Templates{
		landingTmpl:  landingTmpl,
		registerTmpl: registerTmpl,
		loginTmpl:    loginTmpl,
	}, nil
}

// RenderLanding renders the landing page
func (t *Templates) RenderLanding(w http.ResponseWriter, data *LandingData) error {
	// Set content type
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// Execute base layout template with landing content
	if err := t.landingTmpl.ExecuteTemplate(w, "base.html", data); err != nil {
		return err
	}

	return nil
}

// Render renders a full page
func (t *Templates) Render(w http.ResponseWriter, page string, data interface{}) error {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	var tmpl *template.Template
	switch page {
	case "landing.html":
		tmpl = t.landingTmpl
	case "register.html":
		tmpl = t.registerTmpl
	case "login.html":
		tmpl = t.loginTmpl
	default:
		return nil
	}

	// Use auth layout for login/register, base layout for others
	layoutName := "base.html"
	if page == "login.html" || page == "register.html" {
		layoutName = "auth.html"
	}
	return tmpl.ExecuteTemplate(w, layoutName, data)
}

// RenderComponent renders a component (for HTMX partial updates)
func (t *Templates) RenderComponent(w http.ResponseWriter, component string, data interface{}) error {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	var tmpl *template.Template
	var componentName string

	switch component {
	case "register-form.html":
		tmpl = t.registerTmpl
		componentName = "register-form"
	case "login-form.html":
		tmpl = t.loginTmpl
		componentName = "login-form"
	default:
		return nil
	}

	return tmpl.ExecuteTemplate(w, componentName, data)
}

// RenderRegister renders the register page
func (t *Templates) RenderRegister(w http.ResponseWriter, data *RegisterData) error {
	// Set content type
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// Execute auth layout template with register content
	if err := t.registerTmpl.ExecuteTemplate(w, "auth.html", data); err != nil {
		return err
	}

	return nil
}

// RenderRegisterForm renders only the register form (for HTMX partial updates)
func (t *Templates) RenderRegisterForm(w http.ResponseWriter, data *RegisterData) error {
	// Set content type
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// Execute only the form template
	if err := t.registerTmpl.ExecuteTemplate(w, "register-form", data); err != nil {
		return err
	}

	return nil
}

// Data structures

type LandingData struct {
	User *user.User // Authenticated user (nil if anonymous)
}

type LoginData struct {
	Email  string            // Preserved email on validation error
	Errors map[string]string // Field-specific errors
}

type RegisterData struct {
	Errors map[string]string
	Name   string
	Email  string
	Phone  string
}
