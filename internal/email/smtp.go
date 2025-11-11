package email

import (
	"bytes"
	"fmt"
	"net/smtp"
	"text/template"

	"github.com/udisondev/learn-go/pkg/config"
)

// SMTPClient handles SMTP email sending
// WHY: Abstracts SMTP connection and authentication logic
// HOW: Uses standard library net/smtp package
type SMTPClient struct {
	host     string
	port     int
	username string
	password string
	from     string
	msgTmpl  *template.Template
}

// NewSMTPClient creates a new SMTP client from configuration
func NewSMTPClient(cfg *config.EmailConfig) (*SMTPClient, error) {
	// Parse email message template with headers
	// WHY: Using template allows for clean separation and easy modification
	// HOW: Template includes MIME headers and body placeholder
	msgTmpl, err := template.New("email").Parse(`From: {{.From}}
To: {{.To}}
Subject: {{.Subject}}
MIME-Version: 1.0
Content-Type: text/html; charset=UTF-8

{{.Body}}
`)
	if err != nil {
		return nil, fmt.Errorf("parse message template: %w", err)
	}

	return &SMTPClient{
		host:     cfg.Host,
		port:     cfg.Port,
		username: cfg.Username,
		password: cfg.Password,
		from:     cfg.From,
		msgTmpl:  msgTmpl,
	}, nil
}

// Send sends an email via SMTP
// WHY: Core method to actually send emails
// HOW: Connects to SMTP server, authenticates (if credentials provided), sends email
//
// Parameters:
// - to: recipient email address
// - subject: email subject line
// - body: HTML email body
//
// For Mailhog (development): no authentication required
// For production SMTP (Gmail, SendGrid, etc): requires username/password
func (c *SMTPClient) Send(to, subject, body string) error {
	// Build email message using template
	var buf bytes.Buffer
	err := c.msgTmpl.Execute(&buf, map[string]string{
		"From":    c.from,
		"To":      to,
		"Subject": subject,
		"Body":    body,
	})
	if err != nil {
		return fmt.Errorf("execute message template: %w", err)
	}

	// SMTP server address
	addr := fmt.Sprintf("%s:%d", c.host, c.port)

	// Setup authentication if credentials are provided
	// Mailhog doesn't require auth, but production SMTP does
	var auth smtp.Auth
	if c.username != "" && c.password != "" {
		auth = smtp.PlainAuth("", c.username, c.password, c.host)
	}

	// Send email
	// If auth is nil (Mailhog case), smtp.SendMail will skip authentication
	err = smtp.SendMail(addr, auth, c.from, []string{to}, buf.Bytes())
	if err != nil {
		return fmt.Errorf("send email: %w", err)
	}

	return nil
}
