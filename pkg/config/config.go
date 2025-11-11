package config

import (
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

// Config holds application configuration
type Config struct {
	App      AppConfig
	DB       DBConfig
	Session  SessionConfig
	CSRF     CSRFConfig
	Email    EmailConfig
	Executor ExecutorConfig
}

type AppConfig struct {
	Env      string `env:"APP_ENV" envDefault:"development"`
	Port     string `env:"APP_PORT" envDefault:"8080"`
	Host     string `env:"APP_HOST" envDefault:"localhost"`
	LogLevel string `env:"APP_LOG_LEVEL" envDefault:"info"`
}

type DBConfig struct {
	Host            string        `env:"DB_HOST" envDefault:"localhost"`
	Port            string        `env:"DB_PORT" envDefault:"5432"`
	User            string        `env:"DB_USER" envDefault:"postgres"`
	Password        string        `env:"DB_PASSWORD" envDefault:"postgres"`
	Name            string        `env:"DB_NAME" envDefault:"learn_go"`
	SSLMode         string        `env:"DB_SSL_MODE" envDefault:"disable"`
	MaxOpenConns    int           `env:"DB_MAX_OPEN_CONNS" envDefault:"25"`
	MaxIdleConns    int           `env:"DB_MAX_IDLE_CONNS" envDefault:"5"`
	ConnMaxLifetime time.Duration `env:"DB_CONN_MAX_LIFETIME" envDefault:"5m"`
	ConnMaxIdleTime time.Duration `env:"DB_CONN_MAX_IDLE_TIME" envDefault:"30m"`
	HealthCheckPeriod time.Duration `env:"DB_HEALTH_CHECK_PERIOD" envDefault:"1m"`
}

type SessionConfig struct {
	Secret string `env:"SESSION_SECRET" envDefault:"change-me"`
	MaxAge int    `env:"SESSION_MAX_AGE" envDefault:"86400"` // seconds
	Secure bool   `env:"SESSION_SECURE" envDefault:"false"`  // true in production for HTTPS
}

type CSRFConfig struct {
	Secret string `env:"CSRF_SECRET" envDefault:"32-byte-long-csrf-secret-key-change-in-production"`
	Secure bool   `env:"CSRF_SECURE" envDefault:"false"` // true in production
}

type EmailConfig struct {
	Host     string `env:"SMTP_HOST" envDefault:"localhost"`      // Mailhog default: localhost
	Port     int    `env:"SMTP_PORT" envDefault:"1025"`           // Mailhog default: 1025
	Username string `env:"SMTP_USERNAME" envDefault:""`           // Mailhog doesn't need auth
	Password string `env:"SMTP_PASSWORD" envDefault:""`           // Mailhog doesn't need auth
	From     string `env:"SMTP_FROM" envDefault:"noreply@learn-go.local"`
}

type ExecutorConfig struct {
	PoolSize       int           `env:"DOCKER_POOL_SIZE" envDefault:"10"`
	MaxContainers  int           `env:"DOCKER_MAX_CONTAINERS" envDefault:"20"`
	CPULimit       float64       `env:"DOCKER_CPU_LIMIT" envDefault:"0.5"`
	MemoryLimit    string        `env:"DOCKER_MEMORY_LIMIT" envDefault:"128m"`
	DefaultTimeout int           `env:"DOCKER_DEFAULT_TIMEOUT" envDefault:"10"`
	PollInterval   time.Duration `env:"EXECUTOR_POLL_INTERVAL" envDefault:"1s"`
	Workers        int           `env:"EXECUTOR_WORKERS" envDefault:"5"`
}

// Load loads configuration from .env file and environment variables
func Load() (*Config, error) {
	// Load .env file (ignore error if file doesn't exist)
	_ = godotenv.Load()

	cfg := &Config{}

	// Parse environment variables into config struct
	if err := env.Parse(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
