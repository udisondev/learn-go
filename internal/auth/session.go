package auth

import (
	"context"
	"fmt"
	"net/http"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	// SessionCookieName - имя cookie для хранения session ID
	// Используется для идентификации пользователя между запросами
	SessionCookieName = "session_id"

	// SessionDuration - длительность жизни сессии (30 дней)
	// После этого времени пользователю нужно будет залогиниться заново
	SessionDuration = 30 * 24 * time.Hour
)

var psql = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

// Session представляет сессию пользователя в системе
// Хранится в БД для:
// - Отзыва сессий при logout
// - Отслеживания активных сессий пользователя
// - Безопасности (можно удалить все сессии при смене пароля)
type Session struct {
	ID        string    // UUID v7 - time-ordered для лучшей производительности индексов
	UserID    int64     // ID пользователя из таблицы users
	CreatedAt time.Time // Когда создана сессия
	ExpiresAt time.Time // Когда истекает сессия
	IPAddress string    // IP адрес для аудита и безопасности
	UserAgent string    // User-Agent браузера для аудита
}

// CreateSession создает новую сессию для пользователя
// Вызывается после успешного логина или email verification
//
// Использует UUID v7 потому что:
// - Time-ordered - лучше для B-tree индексов в PostgreSQL
// - Криптографически стойкий - невозможно угадать
// - Можно извлечь timestamp создания
//
// Сохраняет IP и User-Agent для:
// - Обнаружения подозрительной активности
// - Показа пользователю списка активных сессий
func CreateSession(ctx context.Context, db *pgxpool.Pool, userID int64, r *http.Request) (*Session, error) {
	// Генерируем UUID v7 (time-ordered)
	sessionID, err := uuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("failed to generate session ID: %w", err)
	}

	now := time.Now().UTC()

	session := &Session{
		ID:        sessionID.String(),
		UserID:    userID,
		CreatedAt: now,
		ExpiresAt: now.Add(SessionDuration),
		IPAddress: getIP(r),
		UserAgent: r.UserAgent(),
	}

	// Используем squirrel для type-safe построения SQL
	query, args, err := psql.
		Insert("sessions").
		Columns("id", "user_id", "created_at", "expires_at", "ip_address", "user_agent").
		Values(session.ID, session.UserID, session.CreatedAt, session.ExpiresAt, session.IPAddress, session.UserAgent).
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("failed to build insert query: %w", err)
	}

	if _, err := db.Exec(ctx, query, args...); err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return session, nil
}

// GetSession получает сессию по ID и проверяет что она не истекла
// Возвращает ошибку если:
// - Сессия не найдена
// - Сессия истекла (expires_at < now)
//
// Проверка expires_at в SQL запросе для производительности:
// - Не нужно загружать истекшие сессии из БД
// - Индекс по expires_at ускоряет запрос
func GetSession(ctx context.Context, db *pgxpool.Pool, sessionID string) (*Session, error) {
	query, args, err := psql.
		Select("id", "user_id", "created_at", "expires_at", "ip_address", "user_agent").
		From("sessions").
		Where(sq.Eq{"id": sessionID}).
		Where(sq.Gt{"expires_at": time.Now().UTC()}). // Проверяем что не истекла
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("failed to build select query: %w", err)
	}

	session := &Session{}
	err = db.QueryRow(ctx, query, args...).Scan(
		&session.ID,
		&session.UserID,
		&session.CreatedAt,
		&session.ExpiresAt,
		&session.IPAddress,
		&session.UserAgent,
	)

	if err != nil {
		return nil, fmt.Errorf("session not found or expired: %w", err)
	}

	return session, nil
}

// DeleteSession удаляет сессию по ID
// Используется при logout - удаляет только текущую сессию,
// оставляя активными сессии на других устройствах
func DeleteSession(ctx context.Context, db *pgxpool.Pool, sessionID string) error {
	query, args, err := psql.
		Delete("sessions").
		Where(sq.Eq{"id": sessionID}).
		ToSql()

	if err != nil {
		return fmt.Errorf("failed to build delete query: %w", err)
	}

	if _, err := db.Exec(ctx, query, args...); err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	return nil
}

// DeleteUserSessions удаляет ВСЕ сессии пользователя
// Используется при:
// - Смене пароля (для безопасности)
// - Удалении аккаунта
// - Принудительном logout со всех устройств
func DeleteUserSessions(ctx context.Context, db *pgxpool.Pool, userID int64) error {
	query, args, err := psql.
		Delete("sessions").
		Where(sq.Eq{"user_id": userID}).
		ToSql()

	if err != nil {
		return fmt.Errorf("failed to build delete query: %w", err)
	}

	if _, err := db.Exec(ctx, query, args...); err != nil {
		return fmt.Errorf("failed to delete user sessions: %w", err)
	}

	return nil
}

// SetSessionCookie устанавливает cookie с session ID
//
// Параметры безопасности:
// HttpOnly: true - защита от XSS атак (JavaScript не может прочитать cookie)
// Secure: из конфига - true в production (только HTTPS), false в dev
// SameSite: Strict - защита от CSRF атак (cookie не отправляется на cross-site запросах)
// Path: / - cookie доступна на всех страницах сайта
// MaxAge: 30 дней - совпадает с SessionDuration
func SetSessionCookie(w http.ResponseWriter, sessionID string, secure bool) {
	cookie := &http.Cookie{
		Name:     SessionCookieName,
		Value:    sessionID,
		Path:     "/",
		MaxAge:   int(SessionDuration.Seconds()),
		HttpOnly: true,                    // Защита от XSS
		Secure:   secure,                  // Из конфига: true в production, false в dev
		SameSite: http.SameSiteStrictMode, // Защита от CSRF
	}
	http.SetCookie(w, cookie)
}

// ClearSessionCookie удаляет cookie сессии
// Используется при logout
// MaxAge: -1 означает "удалить немедленно"
func ClearSessionCookie(w http.ResponseWriter) {
	cookie := &http.Cookie{
		Name:     SessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	}
	http.SetCookie(w, cookie)
}

// GetSessionFromRequest извлекает session ID из cookie запроса
// Возвращает ошибку если cookie не найдена (пользователь не залогинен)
func GetSessionFromRequest(r *http.Request) (string, error) {
	cookie, err := r.Cookie(SessionCookieName)
	if err != nil {
		return "", fmt.Errorf("session cookie not found: %w", err)
	}
	return cookie.Value, nil
}

// getIP извлекает реальный IP адрес клиента
// Проверяет заголовки прокси (X-Forwarded-For, X-Real-IP)
// потому что приложение может быть за nginx/cloudflare
//
// Приоритет:
// 1. X-Forwarded-For - стандартный заголовок для цепочки прокси
// 2. X-Real-IP - альтернативный заголовок
// 3. RemoteAddr - прямое подключение без прокси
func getIP(r *http.Request) string {
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		return forwarded
	}

	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	return r.RemoteAddr
}
