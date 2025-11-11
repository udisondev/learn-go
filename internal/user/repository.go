package user

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var psql = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

// Repository handles user data access operations
// Изолирует бизнес-логику от деталей работы с БД
// Использует squirrel для type-safe построения SQL запросов
type Repository struct {
	db *pgxpool.Pool
}

// NewRepository создает новый экземпляр репозитория
func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// CreateUser создает нового пользователя в БД внутри транзакции
// Вызывается после валидации в service layer
//
// Почему такие значения по умолчанию:
// - sub_plan: SubPlanFree - все новые пользователи начинают с бесплатного плана
// - score: 0 - начальный счет до прохождения заданий
// - is_verified: false - email еще не подтвержден
// - registered_at/updated_at: now - фиксируем время создания
//
// Почему принимает tx:
// - Должен выполняться в одной транзакции с CreateEmailVerification
// - Если создание verification не удалось - откатываем создание пользователя
// - Гарантирует атомарность операции регистрации
//
// Почему принимает ctx:
// - Для отмены операций и таймаутов
// - Передача метаданных (trace ID, user ID)
// - Стандартная практика Go - первый параметр любой I/O функции
//
// Возвращает ID созданного пользователя для последующего создания email_verification
func (r *Repository) CreateUser(ctx context.Context, tx pgx.Tx, name, email, passwordHash, phone string) (int64, error) {
	now := time.Now().UTC()

	query, args, err := psql.
		Insert("users").
		Columns("name", "email", "password_hash", "phone", "registered_at", "updated_at", "sub_plan", "score", "is_verified").
		Values(name, email, passwordHash, phone, now, now, SubPlanFree.String(), 0, false).
		Suffix("RETURNING id").
		ToSql()

	if err != nil {
		return 0, fmt.Errorf("failed to build insert query: %w", err)
	}

	var userID int64
	err = tx.QueryRow(ctx, query, args...).Scan(&userID)
	if err != nil {
		return 0, fmt.Errorf("failed to create user: %w", err)
	}

	return userID, nil
}

// CheckEmailExists проверяет существует ли пользователь с данным email
// Используется в service layer для валидации уникальности email
//
// Почему отдельная функция:
// - Валидация уникальности должна быть до попытки создания пользователя
// - Позволяет вернуть понятную ошибку "email уже занят"
// - Избегаем обработки constraint violation ошибок из PostgreSQL
func (r *Repository) CheckEmailExists(ctx context.Context, email string) (bool, error) {
	query, args, err := psql.
		Select("1").
		From("users").
		Where(sq.Eq{"email": email}).
		Limit(1).
		ToSql()

	if err != nil {
		return false, fmt.Errorf("failed to build select query: %w", err)
	}

	var exists int
	err = r.db.QueryRow(ctx, query, args...).Scan(&exists)
	if err == pgx.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to check email existence: %w", err)
	}

	return true, nil
}

// CreateEmailVerification создает запись для верификации email
// Вызывается сразу после CreateUser в той же транзакции
//
// Почему двойное хеширование:
// - rand.Text() генерирует base32 строку, которая выглядит специфично
// - Первое хеширование (generateEmailToken) маскирует rand.Text() -> обычный hex hash
// - Второе хеширование (hashEmailToken) защищает БД от rainbow table атак
// - Даже при компрометации БД атакующий не получит токен для отправки в URL
//
// Почему 48 часов:
// - Достаточно времени для проверки email
// - Не слишком долго для безопасности
// - После истечения пользователь может запросить новое письмо
//
// Возвращает email token (SHA256 hex) для отправки в письме
func (r *Repository) CreateEmailVerification(ctx context.Context, tx pgx.Tx, userID int64) (string, error) {
	// Генерируем токен для отправки в email (уже захеширован от rand.Text())
	emailToken := generateEmailToken()

	// Хешируем еще раз для хранения в БД
	dbTokenHash := hashEmailToken(emailToken)

	now := time.Now().UTC()
	expiresAt := now.Add(48 * time.Hour)

	query, args, err := psql.
		Insert("email_verifications").
		Columns("user_id", "token_hash", "created_at", "expires_at").
		Values(userID, dbTokenHash, now, expiresAt).
		ToSql()

	if err != nil {
		return "", fmt.Errorf("failed to build insert query: %w", err)
	}

	_, err = tx.Exec(ctx, query, args...)
	if err != nil {
		return "", fmt.Errorf("failed to create email verification: %w", err)
	}

	// Возвращаем email token для отправки в письме
	// В БД сохранен hash от этого токена
	return emailToken, nil
}

// GetUserByEmail получает пользователя по email
// Используется при email verification для обновления is_verified
//
// Почему нужна отдельная функция:
// - При верификации у нас есть только token, нужно получить user_id
// - При логине нужно получить пользователя для проверки пароля
// - Централизованное место для загрузки пользователя
func (r *Repository) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	query, args, err := psql.
		Select("id", "name", "email", "password_hash", "phone", "registered_at", "updated_at", "sub_plan", "score", "is_verified", "avatar_url").
		From("users").
		Where(sq.Eq{"email": email}).
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("failed to build select query: %w", err)
	}

	user := &User{}
	err = r.db.QueryRow(ctx, query, args...).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.PasswordHash,
		&user.Phone,
		&user.RegisteredAt,
		&user.UpdatedAt,
		&user.SubPlan,
		&user.Score,
		&user.IsVerified,
		&user.AvatarURL,
	)

	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

// VerifyEmail проверяет токен верификации и активирует пользователя
// Вызывается когда пользователь переходит по ссылке из email
//
// Почему используем pgx.BeginTxFunc:
// - Автоматический commit при успехе
// - Автоматический rollback при ошибке или панике
// - Чище и безопаснее чем ручное управление транзакцией
// - Гарантирует атомарность операций
//
// Почему транзакция:
// - Обновляем users.is_verified = true
// - Удаляем запись из email_verifications
// - Обе операции должны выполниться атомарно
// - Если одна не удалась - откатываем обе
//
// Почему проверяем expires_at в SQL:
// - Фильтрация на уровне БД эффективнее
// - Индекс по expires_at ускоряет запрос
// - Не загружаем истекшие токены из БД
//
// Возвращает user_id для создания сессии
func (r *Repository) VerifyEmail(ctx context.Context, emailToken string) (int64, error) {
	// Хешируем email token для поиска в БД
	dbTokenHash := hashEmailToken(emailToken)

	var userID int64

	// BeginTxFunc автоматически делает commit/rollback
	err := pgx.BeginTxFunc(ctx, r.db, pgx.TxOptions{}, func(tx pgx.Tx) error {
		// Находим verification по token_hash и проверяем что не истек
		query, args, err := psql.
			Select("user_id").
			From("email_verifications").
			Where(sq.Eq{"token_hash": dbTokenHash}).
			Where(sq.Gt{"expires_at": time.Now().UTC()}).
			ToSql()

		if err != nil {
			return fmt.Errorf("failed to build select query: %w", err)
		}

		err = tx.QueryRow(ctx, query, args...).Scan(&userID)
		if err == pgx.ErrNoRows {
			return fmt.Errorf("invalid or expired token")
		}
		if err != nil {
			return fmt.Errorf("failed to find verification: %w", err)
		}

		// Обновляем пользователя - устанавливаем is_verified = true
		updateQuery, updateArgs, err := psql.
			Update("users").
			Set("is_verified", true).
			Set("updated_at", time.Now().UTC()).
			Where(sq.Eq{"id": userID}).
			ToSql()

		if err != nil {
			return fmt.Errorf("failed to build update query: %w", err)
		}

		_, err = tx.Exec(ctx, updateQuery, updateArgs...)
		if err != nil {
			return fmt.Errorf("failed to update user: %w", err)
		}

		// Удаляем verification запись (больше не нужна)
		deleteQuery, deleteArgs, err := psql.
			Delete("email_verifications").
			Where(sq.Eq{"token_hash": dbTokenHash}).
			ToSql()

		if err != nil {
			return fmt.Errorf("failed to build delete query: %w", err)
		}

		_, err = tx.Exec(ctx, deleteQuery, deleteArgs...)
		if err != nil {
			return fmt.Errorf("failed to delete verification: %w", err)
		}

		return nil
	})

	if err != nil {
		return 0, err
	}

	return userID, nil
}

// generateEmailToken генерирует токен для отправки в email
// WHY: Маскирует rand.Text() чтобы токен выглядел как обычный hex hash
// HOW: rand.Text() → SHA256 → hex string
//
// Возвращает hex строку для использования в URL письма
func generateEmailToken() string {
	// Генерируем криптографически стойкий токен
	plainToken := rand.Text()

	// Хешируем чтобы замаскировать base32 формат rand.Text()
	hash := sha256.Sum256([]byte(plainToken))
	return hex.EncodeToString(hash[:])
}

// hashEmailToken хеширует email token для хранения в БД
// WHY: Защита от rainbow table атак и компрометации БД
// HOW: emailToken (hex) → SHA256 → hex string
//
// Второй слой хеширования для защиты БД
func hashEmailToken(emailToken string) string {
	hash := sha256.Sum256([]byte(emailToken))
	return hex.EncodeToString(hash[:])
}
