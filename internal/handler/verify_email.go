package handler

import (
	"log/slog"
	"net/http"
)

// HandleVerifyEmail обрабатывает верификацию email по токену из ссылки
func (h *Handler) HandleVerifyEmail(w http.ResponseWriter, r *http.Request) {
	// Получаем токен из query параметра
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "Missing verification token", http.StatusBadRequest)
		return
	}

	// Вызываем service для верификации
	userID, err := h.userService.VerifyEmail(r.Context(), token)
	if err != nil {
		slog.Error("Failed to verify email", "error", err, "token", token)
		// TODO: Render error page with message "Invalid or expired token"
		http.Error(w, "Invalid or expired verification link", http.StatusBadRequest)
		return
	}

	// Верификация успешна
	slog.Info("Email verified successfully", "user_id", userID)

	// Создаем сессию для автологина после верификации
	sessionID, err := h.sessionService.CreateSession(
		r.Context(),
		userID,
		getRealIP(r),
		r.Header.Get("User-Agent"),
	)
	if err != nil {
		slog.Error("Failed to create session after verification", "error", err, "user_id", userID)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Устанавливаем session cookie
	// MaxAge = 0 означает session cookie (удалится при закрытии браузера)
	// Но сессия в БД безграничная, пользователь может вернуться позже
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    sessionID.String(),
		Path:     "/",
		MaxAge:   0,                     // Session cookie (browser lifetime)
		HttpOnly: true,                  // Защита от XSS
		Secure:   h.cfg.Session.Secure, // HTTPS only в production
		SameSite: http.SameSiteLaxMode,  // CSRF защита
	})

	slog.Info("Session created after verification", "user_id", userID, "session_id", sessionID)

	// Редирект на главную (пользователь уже залогинен)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
