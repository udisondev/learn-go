package handler

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/udisondev/learn-go/internal/templates"
	"github.com/udisondev/learn-go/internal/user"
	"golang.org/x/crypto/bcrypt"
)

// GetLogin отображает страницу входа
func (h *Handler) GetLogin(w http.ResponseWriter, r *http.Request) {
	// Если пользователь уже авторизован - редирект на главную
	if _, ok := user.FromCtx(r.Context()); ok {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	data := templates.LoginData{
		Errors: make(map[string]string),
	}

	if err := h.templates.Render(w, "login.html", data); err != nil {
		slog.Error("Failed to render login page", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// PostLogin обрабатывает вход пользователя
func (h *Handler) PostLogin(w http.ResponseWriter, r *http.Request) {
	// Парсим форму
	if err := r.ParseForm(); err != nil {
		slog.Error("Failed to parse login form", "error", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	email := strings.TrimSpace(r.FormValue("email"))
	password := r.FormValue("password")

	// Валидация
	errors := make(map[string]string)

	if email == "" {
		errors["email"] = "Email обязателен для заполнения"
	}

	if password == "" {
		errors["password"] = "Пароль обязателен для заполнения"
	}

	// Если есть ошибки валидации - отправляем форму обратно
	if len(errors) > 0 {
		data := templates.LoginData{
			Email:  email,
			Errors: errors,
		}

		w.WriteHeader(http.StatusOK)
		if err := h.templates.RenderComponent(w, "login-form.html", data); err != nil {
			slog.Error("Failed to render login form", "error", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}

	// Ищем пользователя по email
	foundUser, err := h.userService.GetUserByEmail(r.Context(), email)
	if err != nil {
		slog.Error("Failed to find user", "error", err, "email", email)
		errors["email"] = "Неверный email или пароль"

		data := templates.LoginData{
			Email:  email,
			Errors: errors,
		}

		w.WriteHeader(http.StatusOK)
		if err := h.templates.RenderComponent(w, "login-form.html", data); err != nil {
			slog.Error("Failed to render login form", "error", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}

	// Проверяем пароль
	if err := bcrypt.CompareHashAndPassword([]byte(foundUser.PasswordHash), []byte(password)); err != nil {
		slog.Warn("Invalid password attempt", "email", email)
		errors["password"] = "Неверный email или пароль"

		data := templates.LoginData{
			Email:  email,
			Errors: errors,
		}

		w.WriteHeader(http.StatusOK)
		if err := h.templates.RenderComponent(w, "login-form.html", data); err != nil {
			slog.Error("Failed to render login form", "error", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}

	// Проверяем что email верифицирован
	if !foundUser.IsVerified {
		errors["email"] = "Email не подтвержден. Проверьте почту."

		data := templates.LoginData{
			Email:  email,
			Errors: errors,
		}

		w.WriteHeader(http.StatusOK)
		if err := h.templates.RenderComponent(w, "login-form.html", data); err != nil {
			slog.Error("Failed to render login form", "error", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}

	// Создаем сессию
	realIP := getRealIP(r)
	userAgent := r.UserAgent()

	sessionToken, err := h.sessionService.CreateSession(r.Context(), foundUser.ID, realIP, userAgent)
	if err != nil {
		slog.Error("Failed to create session", "error", err, "user_id", foundUser.ID)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Устанавливаем cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    sessionToken.String(),
		Path:     "/",
		HttpOnly: true,
		Secure:   h.cfg.Session.Secure,
		SameSite: http.SameSiteLaxMode,
	})

	slog.Info("User logged in successfully", "user_id", foundUser.ID, "email", foundUser.Email)

	// Редирект на главную
	w.Header().Set("HX-Redirect", "/")
	w.WriteHeader(http.StatusOK)
}
