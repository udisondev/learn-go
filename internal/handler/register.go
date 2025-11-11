package handler

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/udisondev/learn-go/internal/email"
	"github.com/udisondev/learn-go/internal/templates"
	"github.com/udisondev/learn-go/internal/user"
)

// HandleRegisterPage renders the registration page
func (h *Handler) HandleRegisterPage(w http.ResponseWriter, r *http.Request) {
	// Get authenticated user from context
	u, _ := user.FromCtx(r.Context())

	data := &templates.RegisterData{
		User: u,
	}

	if err := h.templates.RenderRegister(w, data); err != nil {
		slog.Error("Failed to render register page", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// HandleRegisterSubmit processes the registration form
func (h *Handler) HandleRegisterSubmit(w http.ResponseWriter, r *http.Request) {
	// Parse form data
	if err := r.ParseForm(); err != nil {
		slog.Error("Failed to parse form", "error", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	// Собираем данные из формы
	input := user.RegisterInput{
		Name:            r.FormValue("name"),
		Email:           r.FormValue("email"),
		Password:        r.FormValue("password"),
		PasswordConfirm: r.FormValue("password_confirm"),
		Phone:           r.FormValue("phone"),
	}

	// Вызываем service для регистрации
	result, err := h.userService.RegisterUser(r.Context(), input)
	if err != nil {
		// Проверяем тип ошибки - validation errors или системная ошибка
		var validationErrs user.ValidationErrors
		if errors.As(err, &validationErrs) {
			// Get authenticated user from context
			u, _ := user.FromCtx(r.Context())

			// Validation errors - отображаем в форме
			data := &templates.RegisterData{
				User:   u,
				Errors: make(map[string]string),
				Name:   input.Name,
				Email:  input.Email,
				Phone:  input.Phone,
			}

			// Преобразуем ValidationErrors в map для template
			for _, ve := range validationErrs {
				data.Errors[ve.Field] = ve.Message
			}

			// Возвращаем только форму с ошибками для HTMX
			if err := h.templates.RenderRegisterForm(w, data); err != nil {
				slog.Error("Failed to render register form with errors", "error", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
			return
		}

		// Системная ошибка
		slog.Error("Failed to register user", "error", err, "email", input.Email)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Регистрация успешна
	slog.Info("User registered successfully",
		"user_id", result.UserID,
		"email", input.Email,
		"has_verification_token", result.VerificationToken != "")

	// Отправляем задачу в очередь для email worker
	// WHY: Асинхронная отправка email не блокирует HTTP ответ
	// HOW: Добавляем задачу в email_queue, worker обработает её
	payload := map[string]string{
		"token":     result.VerificationToken,
		"user_name": input.Name,
	}

	if err := h.emailQueue.Enqueue(r.Context(), email.EmailTypeVerification, input.Email, &result.UserID, payload); err != nil {
		slog.Error("Failed to enqueue verification email", "error", err, "user_id", result.UserID)
		// Не возвращаем ошибку пользователю - регистрация прошла успешно
		// Email можно отправить позже вручную или через retry
	}

	// Возвращаем успешный ответ с триггером для модального окна
	w.Header().Set("HX-Trigger", "showSuccessModal")
	w.WriteHeader(http.StatusOK)
}
