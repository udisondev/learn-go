package email

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"path/filepath"
)

// Sender handles email sending with template rendering
// WHY: Provides high-level API for sending emails with type-specific templates
// HOW: Uses SMTPClient for actual sending, renders HTML templates with data
type Sender struct {
	smtp      *SMTPClient
	templates map[string]*template.Template
}

// NewSender creates a new Sender instance
// WHY: Initializes sender with SMTP client and loads all email templates
// HOW: Parses templates from web/templates/email/ directory
func NewSender(smtp *SMTPClient, templatesDir string) (*Sender, error) {
	sender := &Sender{
		smtp:      smtp,
		templates: make(map[string]*template.Template),
	}

	// Load all email templates
	// WHY: Pre-parse templates at startup for better performance
	// HOW: Each email type has its own HTML template file
	templateFiles := []string{
		"verification",
		"password_reset",
		"notification",
	}

	for _, name := range templateFiles {
		tmplPath := filepath.Join(templatesDir, name+".html")
		tmpl, err := template.ParseFiles(tmplPath)
		if err != nil {
			return nil, fmt.Errorf("parse template %s: %w", name, err)
		}
		sender.templates[name] = tmpl
	}

	return sender, nil
}

// Send sends an email based on task configuration
// WHY: Single method to send any type of email
// HOW: Looks up config by EmailType, renders template, sends via SMTP
//
// This is the main method called by the worker:
// 1. Get email config (subject + template name) by type
// 2. Parse JSON payload into map
// 3. Render HTML template with payload data
// 4. Send via SMTP
func (s *Sender) Send(ctx context.Context, task *Task) error {
	// Get configuration for this email type
	config, ok := GetConfig(task.EmailType)
	if !ok {
		return fmt.Errorf("unknown email type: %s", task.EmailType)
	}

	// Parse payload into map for template rendering
	var data map[string]any
	if err := json.Unmarshal(task.Payload, &data); err != nil {
		return fmt.Errorf("unmarshal payload: %w", err)
	}

	// Render HTML template with data
	body, err := s.renderTemplate(config.Template, data)
	if err != nil {
		return fmt.Errorf("render template: %w", err)
	}

	// Send email via SMTP
	if err := s.smtp.Send(task.RecipientEmail, config.Subject, body); err != nil {
		return fmt.Errorf("smtp send: %w", err)
	}

	return nil
}

// renderTemplate renders an HTML template with the given data
// WHY: Centralizes template rendering logic
// HOW: Looks up pre-parsed template and executes it with data
func (s *Sender) renderTemplate(templateName string, data any) (string, error) {
	tmpl, ok := s.templates[templateName]
	if !ok {
		return "", fmt.Errorf("template not found: %s", templateName)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("execute template: %w", err)
	}

	return buf.String(), nil
}
