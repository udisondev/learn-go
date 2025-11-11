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
		"web/templates/layouts/base.html",
		"web/templates/components/header.html",
		"web/templates/components/register-form.html",
		"web/templates/pages/register.html",
	)
	if err != nil {
		return nil, err
	}

	return &Templates{
		landingTmpl:  landingTmpl,
		registerTmpl: registerTmpl,
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

// RenderLogin renders the login page
func (t *Templates) RenderLogin(w http.ResponseWriter, data *LoginData) error {
	// TODO: execute login template
	panic("not implemented")
}

// RenderRegister renders the register page
func (t *Templates) RenderRegister(w http.ResponseWriter, data *RegisterData) error {
	// Set content type
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// Execute base layout template with register content
	if err := t.registerTmpl.ExecuteTemplate(w, "base.html", data); err != nil {
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
	User  *user.User // Authenticated user (nil if anonymous)
	Error string
}

type RegisterData struct {
	User   *user.User        // Authenticated user (nil if anonymous)
	Errors map[string]string
	Name   string
	Email  string
	Phone  string
}
