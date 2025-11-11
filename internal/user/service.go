package user

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

// Service содержит бизнес-логику для работы с пользователями
// Отделяет валидацию и бизнес-правила от HTTP handlers
// Использует Repository для доступа к данным
type Service struct {
	repo *Repository
	db   *pgxpool.Pool
}

// NewService создает новый экземпляр сервиса
func NewService(db *pgxpool.Pool) *Service {
	return &Service{
		repo: NewRepository(db),
		db:   db,
	}
}

// RegisterInput содержит данные для регистрации пользователя
type RegisterInput struct {
	Name            string
	Email           string
	Password        string
	PasswordConfirm string
	Phone           string
}

// ValidationError представляет ошибки валидации полей формы
// Используется для отображения ошибок под каждым полем в UI
type ValidationError struct {
	Field   string // "name", "email", "password", "phone"
	Message string // Текст ошибки для пользователя
}

// ValidationErrors - список ошибок валидации
type ValidationErrors []ValidationError

func (ve ValidationErrors) Error() string {
	if len(ve) == 0 {
		return ""
	}
	var msgs []string
	for _, err := range ve {
		msgs = append(msgs, fmt.Sprintf("%s: %s", err.Field, err.Message))
	}
	return strings.Join(msgs, "; ")
}

// RegisterResult содержит результат регистрации
// Возвращает user_id и verification token для отправки email
type RegisterResult struct {
	UserID            int64
	VerificationToken string
}

// RegisterUser регистрирует нового пользователя в системе
// Выполняет полную валидацию входных данных и создает пользователя в БД
//
// Процесс регистрации:
// 1. Валидация всех полей (name, email, password, phone)
// 2. Проверка уникальности email
// 3. Хеширование пароля с bcrypt
// 4. Создание пользователя в БД (в транзакции)
// 5. Создание email verification token (в той же транзакции)
// 6. Возврат user_id и token для отправки email
//
// Почему используем транзакцию:
// - Пользователь и email_verification должны создаваться атомарно
// - Если не удалось создать verification - откатываем создание пользователя
// - Гарантирует консистентность данных
//
// Почему возвращаем token:
// - Service не занимается отправкой email (это делает отдельный worker)
// - Handler получает token и передает его в очередь для отправки
// - Разделение ответственности (12-factor app)
func (s *Service) RegisterUser(ctx context.Context, input RegisterInput) (*RegisterResult, error) {
	// Валидация входных данных
	if errs := s.validateRegisterInput(input); len(errs) > 0 {
		return nil, errs
	}

	// Проверяем что email еще не занят
	exists, err := s.repo.CheckEmailExists(ctx, input.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to check email: %w", err)
	}
	if exists {
		return nil, ValidationErrors{
			{Field: "email", Message: "Email уже зарегистрирован"},
		}
	}

	// Хешируем пароль с bcrypt
	// Почему bcrypt:
	// - Специально создан для хеширования паролей (медленный by design)
	// - Автоматически добавляет salt
	// - Параметр cost позволяет увеличивать время при росте мощности CPU
	// - Стандарт для Go (golang.org/x/crypto/bcrypt)
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	var result RegisterResult

	// Создаем пользователя и verification token в одной транзакции
	// Используем pgx.BeginTxFunc для автоматического commit/rollback
	err = pgx.BeginTxFunc(ctx, s.db, pgx.TxOptions{}, func(tx pgx.Tx) error {
		// Создаем пользователя
		userID, err := s.repo.CreateUser(ctx, tx, input.Name, input.Email, string(passwordHash), input.Phone)
		if err != nil {
			return fmt.Errorf("failed to create user: %w", err)
		}
		result.UserID = userID

		// Создаем email verification token
		token, err := s.repo.CreateEmailVerification(ctx, tx, userID)
		if err != nil {
			return fmt.Errorf("failed to create verification: %w", err)
		}
		result.VerificationToken = token

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &result, nil
}

// GetUserByEmail возвращает пользователя по email
func (s *Service) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	return s.repo.GetUserByEmail(ctx, email)
}

// VerifyEmail верифицирует email пользователя по токену
// Вызывается когда пользователь переходит по ссылке из письма
//
// Процесс верификации:
// 1. Хешируем токен из URL для поиска в БД
// 2. Ищем запись в email_verifications (проверяя expires_at)
// 3. Обновляем users.is_verified = true
// 4. Удаляем запись из email_verifications
// 5. Возвращаем user_id для создания сессии
//
// Почему все в транзакции:
// - Гарантирует атомарность (либо все, либо ничего)
// - Предотвращает race conditions при множественных кликах
//
// Возвращает user_id для автоматического логина после верификации
func (s *Service) VerifyEmail(ctx context.Context, emailToken string) (int64, error) {
	// Валидация токена
	if emailToken == "" {
		return 0, fmt.Errorf("token is required")
	}

	// Вызываем repository для верификации
	userID, err := s.repo.VerifyEmail(ctx, emailToken)
	if err != nil {
		return 0, fmt.Errorf("verification failed: %w", err)
	}

	return userID, nil
}

// validateRegisterInput валидирует все поля регистрации
// Возвращает список ошибок (может быть несколько ошибок одновременно)
//
// Правила валидации:
// - Name: обязательное, 2-100 символов
// - Email: обязательное, валидный email формат
// - Password: обязательное, минимум 8 символов, содержит буквы и цифры
// - Phone: опциональное, если указан - валидный формат телефона
func (s *Service) validateRegisterInput(input RegisterInput) ValidationErrors {
	var errors ValidationErrors

	// Валидация имени
	input.Name = strings.TrimSpace(input.Name)
	if input.Name == "" {
		errors = append(errors, ValidationError{
			Field:   "name",
			Message: "Имя обязательно для заполнения",
		})
	} else if len(input.Name) < 2 {
		errors = append(errors, ValidationError{
			Field:   "name",
			Message: "Имя должно содержать минимум 2 символа",
		})
	} else if len(input.Name) > 100 {
		errors = append(errors, ValidationError{
			Field:   "name",
			Message: "Имя не может быть длиннее 100 символов",
		})
	}

	// Валидация email
	input.Email = strings.TrimSpace(strings.ToLower(input.Email))
	if input.Email == "" {
		errors = append(errors, ValidationError{
			Field:   "email",
			Message: "Email обязателен для заполнения",
		})
	} else if !isValidEmail(input.Email) {
		errors = append(errors, ValidationError{
			Field:   "email",
			Message: "Некорректный формат email",
		})
	}

	// Валидация пароля
	if input.Password == "" {
		errors = append(errors, ValidationError{
			Field:   "password",
			Message: "Пароль обязателен для заполнения",
		})
	} else if len(input.Password) < 8 {
		errors = append(errors, ValidationError{
			Field:   "password",
			Message: "Пароль должен содержать минимум 8 символов",
		})
	} else if !isValidPassword(input.Password) {
		errors = append(errors, ValidationError{
			Field:   "password",
			Message: "Пароль должен содержать буквы и цифры",
		})
	}

	// Валидация подтверждения пароля
	if input.PasswordConfirm == "" {
		errors = append(errors, ValidationError{
			Field:   "password_confirm",
			Message: "Подтверждение пароля обязательно для заполнения",
		})
	} else if input.Password != input.PasswordConfirm {
		errors = append(errors, ValidationError{
			Field:   "password_confirm",
			Message: "Пароли не совпадают",
		})
	}

	// Валидация телефона (опциональное поле)
	input.Phone = strings.TrimSpace(input.Phone)
	if input.Phone != "" && !isValidPhone(input.Phone) {
		errors = append(errors, ValidationError{
			Field:   "phone",
			Message: "Некорректный формат телефона",
		})
	}

	return errors
}

// isValidEmail проверяет формат email
// Использует простое regex валидацию
// Почему простое regex:
// - Полная RFC 5322 валидация слишком сложная и не нужна
// - Реальная проверка email - это отправка письма с подтверждением
// - Простое regex защищает от очевидных опечаток
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

func isValidEmail(email string) bool {
	return emailRegex.MatchString(email)
}

// isValidPassword проверяет что пароль содержит буквы и цифры
// Почему такая проверка:
// - Баланс между безопасностью и UX
// - Требование спецсимволов часто раздражает пользователей
// - Длина (8+ символов) + буквы + цифры = достаточная энтропия
func isValidPassword(password string) bool {
	hasLetter := regexp.MustCompile(`[a-zA-Z]`).MatchString(password)
	hasDigit := regexp.MustCompile(`[0-9]`).MatchString(password)
	return hasLetter && hasDigit
}

// isValidPhone проверяет формат телефона
// Принимает форматы: +7XXXXXXXXXX, 8XXXXXXXXXX, 7XXXXXXXXXX
// Почему такая валидация:
// - Поддерживает российские номера (основная аудитория)
// - Можно расширить для международных номеров при необходимости
// - Убирает пробелы и дефисы для гибкости ввода
var phoneRegex = regexp.MustCompile(`^(\+7|7|8)\d{10}$`)

func isValidPhone(phone string) bool {
	// Убираем пробелы, дефисы, скобки для гибкости
	phone = strings.ReplaceAll(phone, " ", "")
	phone = strings.ReplaceAll(phone, "-", "")
	phone = strings.ReplaceAll(phone, "(", "")
	phone = strings.ReplaceAll(phone, ")", "")

	return phoneRegex.MatchString(phone)
}
