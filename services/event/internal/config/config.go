package config

import (
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// ---- sub-configs ----

type LoggingConfig struct {
	Level  string
	Format string
	Output string
}

type DatabaseConfig struct {
	DSN             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

type HTTPConfig struct {
	Addr string
}

type AttendeeServiceConfig struct {
	BaseURL string
	Timeout time.Duration
}

// ---- main config ----

type Config struct {
	Env string

	Logging  LoggingConfig
	Database DatabaseConfig
	HTTP     HTTPConfig

	// JWT used by this service to validate tokens from auth service
	JWTSecret     string
	JWTExpiration time.Duration

	AttendeeService AttendeeServiceConfig
}

// Load reads from env (and .env) and builds Config.
func Load() (*Config, error) {
	// Load .env if present (ignore error if missing)
	_ = godotenv.Load()

	cfg := &Config{
		Env: strings.ToLower(getEnv("APP_ENV", "development")),

		Logging: LoggingConfig{
			Level:  getEnv("EVENT_LOG_LEVEL", "DEBUG"),
			Format: getEnv("EVENT_LOG_FORMAT", "text"),
			Output: getEnv("EVENT_LOG_OUTPUT", "stdout"),
		},

		Database: DatabaseConfig{
			DSN:             getEnv("EVENT_DB_DSN", "./data/event.db"),
			MaxOpenConns:    getEnvInt("EVENT_DB_MAX_OPEN_CONNS", 10),
			MaxIdleConns:    getEnvInt("EVENT_DB_MAX_IDLE_CONNS", 5),
			ConnMaxLifetime: getEnvDuration("EVENT_DB_CONN_MAX_LIFETIME", time.Hour),
			ConnMaxIdleTime: getEnvDuration("EVENT_DB_CONN_MAX_IDLE_TIME", 15*time.Minute),
		},

		HTTP: HTTPConfig{
			Addr: getEnv("EVENT_HTTP_ADDR", ":8082"),
		},

		JWTSecret:     getEnv("JWT_SECRET", "dev-secret-change-me"),             // must match auth service secret
		JWTExpiration: getEnvDuration("JWT_EXPIRATION", 24*time.Hour),          // usually same as auth

		AttendeeService: AttendeeServiceConfig{
			BaseURL: getEnv("ATTENDEE_SERVICE_URL", "http://localhost:8083"),
			Timeout: getEnvDuration("ATTENDEE_SERVICE_TIMEOUT", 5*time.Second),
		},
	}

	return cfg, nil
}

func (c *Config) IsProduction() bool {
	return c.Env == "prod" || c.Env == "production"
}

// ---- helpers ----

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getEnvInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return def
}

func getEnvDuration(key string, def time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return def
}
