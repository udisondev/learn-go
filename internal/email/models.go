package email

//go:generate go-enum --marshal --names --values --flag --nocase

import (
	"time"
)

// EmailType represents the type of email to send
// This enum is used to determine which template and configuration to use
// ENUM(verification, password_reset, notification)
type EmailType int

// Task represents an email task in the queue
// This is the main model that maps to the email_queue table
type Task struct {
	ID             int64
	EmailType      EmailType
	RecipientEmail string
	UserID         *int64 // nullable - some emails may not be user-specific
	Payload        []byte // JSONB - flexible data for different email types
	Attempts       int
	MaxAttempts    int
	Status         string
	Error          *string
	CreatedAt      time.Time
	ProcessedAt    *time.Time
	NextRetryAt    time.Time
}

// EmailConfig holds the configuration for a specific email type
// This includes the subject line and template name to use
type EmailConfig struct {
	Subject  string // Email subject line
	Template string // Template file name (without .html extension)
}

// emailConfigs maps each EmailType to its configuration
// This is the central place to configure all email types
// WHY: Using a map allows us to add new email types without changing code
// HOW: Worker looks up config by EmailType and uses it to send the email
var emailConfigs = map[EmailType]EmailConfig{
	EmailTypeVerification: {
		Subject:  "Подтвердите ваш email",
		Template: "verification",
	},
	EmailTypePasswordReset: {
		Subject:  "Сброс пароля",
		Template: "password_reset",
	},
	EmailTypeNotification: {
		Subject:  "Уведомление",
		Template: "notification",
	},
}

// GetConfig returns the configuration for a given email type
// Returns the config and a boolean indicating if the type is known
func GetConfig(emailType EmailType) (EmailConfig, bool) {
	config, ok := emailConfigs[emailType]
	return config, ok
}
