# Learn Go - GoSpace: Путешествие к звездам

Интерактивная онлайн-платформа для обучения программированию на языке Go с геймификацией и практическими заданиями.

## Описание

Learn Go - это образовательная платформа, где студенты изучают Go через увлекательную сюжетную линию о космической станции "Гофер-1". Каждый урок содержит теоретическую часть и практические задания с автоматической проверкой кода.

### Особенности

- Server-side rendering (html/template + HTMX)
- Monaco Editor для написания кода
- Docker-based code execution с изоляцией
- Система прогресса и достижений
- Email верификация и уведомления
- Система подписок (Free, Basic, Standard, Premium)

## Архитектура

Проект построен как monorepo с двумя основными сервисами:

- **Web App** - основное веб-приложение с SSR
- **Executor** - сервис для выполнения пользовательского кода в изолированных Docker контейнерах

### Структура проекта

```
learn-go/
├── cmd/
│   ├── web/              # Веб-приложение
│   └── executor/         # Executor сервис
├── internal/
│   ├── domain/           # Доменные модели
│   ├── auth/             # Аутентификация
│   ├── user/             # Управление пользователями
│   ├── course/           # Модули и уроки
│   ├── exercise/         # Упражнения и проверка
│   ├── executor/         # Docker pool
│   ├── progress/         # Прогресс пользователей
│   ├── achievement/      # Система достижений
│   ├── email/            # Email уведомления
│   ├── middleware/       # HTTP middleware
│   └── validator/        # Валидация данных
├── web/
│   ├── templates/        # HTML шаблоны
│   └── static/           # Статические файлы
├── migrations/           # Goose миграции БД
├── tests/                # Тесты
└── pkg/                  # Переиспользуемые пакеты
```

## Технологический стек

- **Backend**: Go 1.23
- **Database**: PostgreSQL 16
- **Migrations**: Goose
- **Sessions**: gorilla/sessions
- **Templating**: html/template
- **Frontend**: HTMX + Tailwind CSS
- **Code Editor**: Monaco Editor
- **Container**: Docker

## Начало работы

### Требования

- Go 1.23+
- Docker & Docker Compose
- Make (опционально)

### Установка

1. Клонируйте репозиторий:
```bash
git clone https://github.com/udisondev/learn-go.git
cd learn-go
```

2. Скопируйте и настройте переменные окружения:
```bash
cp .env.example .env
# Отредактируйте .env файл
```

3. Запустите PostgreSQL:
```bash
make docker-up
```

4. Установите goose для миграций:
```bash
go install github.com/pressly/goose/v3/cmd/goose@latest
```

5. Примените миграции:
```bash
make db-migrate-up
```

6. Запустите приложение:
```bash
make run-web
```

Приложение будет доступно по адресу: http://localhost:8080

## Команды Makefile

- `make run-web` - Запустить веб-приложение
- `make run-executor` - Запустить executor сервис
- `make build-web` - Собрать веб-приложение
- `make build-executor` - Собрать executor
- `make test` - Запустить все тесты
- `make docker-up` - Запустить Docker контейнеры
- `make docker-down` - Остановить Docker контейнеры
- `make db-migrate-up` - Применить миграции
- `make db-migrate-down` - Откатить миграции
- `make db-migrate-create NAME=migration_name` - Создать новую миграцию
- `make db-reset` - Сбросить базу данных

## Планы подписок

- **Free**: Доступ к теоретическим урокам и базовым упражнениям
- **Basic**: Free + все практические задания
- **Standard**: Basic + mock-interview + HR консультация
- **Premium**: Standard + чат с ментором + 3 месяца поддержки

## Разработка

### Создание новой миграции

```bash
make db-migrate-create NAME=create_users_table
```

### Запуск тестов

```bash
# Все тесты
make test

# Только unit тесты
make test-unit

# Только integration тесты
make test-integration
```

## Лицензия

MIT

## Контакты

- Автор: [Your Name]
- Email: [your-email]
- GitHub: [@udisondev](https://github.com/udisondev)
# learn-go
