package email

import (
	"context"
	"fmt"
	"net/smtp"
	"net/url"

	"github.com/udisondev/learn-go/pkg/config"
)

type MailtrapSender struct {
	cfg *config.EmailConfig
}

func NewMailtrapSender(cfg *config.EmailConfig) *MailtrapSender {
	return &MailtrapSender{
		cfg: cfg,
	}
}

func (m *MailtrapSender) SendVerificationEmail(ctx context.Context, to, name, token string) error {
	// Build verification URL
	u := &url.URL{
		Scheme: "http",
		Host:   "localhost:8080", // TODO: get from config
		Path:   "/verify-email",
	}
	q := u.Query()
	q.Set("token", token)
	u.RawQuery = q.Encode()

	verificationURL := u.String()

	subject := "Подтвердите ваш email"
	body := fmt.Sprintf(`
Привет, %s!

Спасибо за регистрацию на learn-go!

Пожалуйста, подтвердите ваш email, перейдя по ссылке:
%s

Ссылка действительна в течение 24 часов.

С уважением,
Команда learn-go
`, name, verificationURL)

	return m.send(ctx, to, subject, body)
}

func (m *MailtrapSender) SendPasswordResetEmail(ctx context.Context, to, name, token string) error {
	// Build reset URL
	u := &url.URL{
		Scheme: "http",
		Host:   "localhost:8080", // TODO: get from config
		Path:   "/reset-password",
	}
	q := u.Query()
	q.Set("token", token)
	u.RawQuery = q.Encode()

	resetURL := u.String()

	subject := "Сброс пароля"
	body := fmt.Sprintf(`
Привет, %s!

Вы запросили сброс пароля.

Перейдите по ссылке для создания нового пароля:
%s

Ссылка действительна в течение 1 часа.

Если вы не запрашивали сброс пароля, проигнорируйте это письмо.

С уважением,
Команда learn-go
`, name, resetURL)

	return m.send(ctx, to, subject, body)
}

func (m *MailtrapSender) SendReminderEmail(ctx context.Context, to, userName string) error {
	subject := "Мы скучаем по вам!"
	body := fmt.Sprintf(`
Привет, %s!

Мы заметили, что вы давно не заходили на learn-go.

Продолжите обучение Go прямо сейчас:
http://localhost:8080/course

С уважением,
Команда learn-go
`, userName)

	return m.send(ctx, to, subject, body)
}

func (m *MailtrapSender) send(ctx context.Context, to, subject, body string) error {
	from := m.cfg.From

	// Compose message
	message := fmt.Sprintf("From: %s\r\n", from)
	message += fmt.Sprintf("To: %s\r\n", to)
	message += fmt.Sprintf("Subject: %s\r\n", subject)
	message += "\r\n"
	message += body

	// SMTP authentication
	auth := smtp.PlainAuth("", m.cfg.Username, m.cfg.Password, m.cfg.Host)

	// Send email
	addr := fmt.Sprintf("%s:%d", m.cfg.Host, m.cfg.Port)
	if err := smtp.SendMail(addr, auth, from, []string{to}, []byte(message)); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}
