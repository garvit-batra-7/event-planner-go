package config

import (
	"os"
	"strconv"
	"time"
	"strings"

	"github.com/joho/godotenv"
)

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

type EventService struct {
	BaseURL string
	Timeout time.Duration
}

type UserService struct {
	BaseURL string
	Timeout time.Duration
}

type Config struct {
	Env          string
	Logging      LoggingConfig
	Database     DatabaseConfig
	HTTP         HTTPConfig
	EventService EventService
	UserService  UserService
	JWTSecret    string
}

// Load reads environment variables (optionally from .env) and builds Config.
func Load() (*Config, error) {
	// Try to load .env from repo root; ignore error if not found
	_ = godotenv.Load()

	cfg := &Config{
		Env: getEnv("APP_ENV", "development"),

		Logging: LoggingConfig{
			Level:  getEnv("ATTENDEE_LOG_LEVEL", "DEBUG"),
			Format: getEnv("ATTENDEE_LOG_FORMAT", "text"),
			Output: getEnv("ATTENDEE_LOG_OUTPUT", "stdout"),
		},

		Database: DatabaseConfig{
			DSN:             getEnv("ATTENDEE_DB_DSN", "./data/attendee.db"),
			MaxOpenConns:    getEnvInt("ATTENDEE_DB_MAX_OPEN_CONNS", 10),
			MaxIdleConns:    getEnvInt("ATTENDEE_DB_MAX_IDLE_CONNS", 5),
			ConnMaxLifetime: getEnvDuration("ATTENDEE_DB_CONN_MAX_LIFETIME", time.Hour),
			ConnMaxIdleTime: getEnvDuration("ATTENDEE_DB_CONN_MAX_IDLE_TIME", 15*time.Minute),
		},

		HTTP: HTTPConfig{
			// change default port here if you want (e.g. ":8080")
			Addr: getEnv("ATTENDEE_HTTP_ADDR", ":8083"),
		},

		EventService: EventService{
			BaseURL: getEnv("EVENT_SERVICE_URL", "http://localhost:8082"),
			Timeout: getEnvDuration("EVENT_SERVICE_TIMEOUT", 5*time.Second),
		},

		UserService: UserService{
			BaseURL: getEnv("USER_SERVICE_URL", "http://localhost:8081"),
			Timeout: getEnvDuration("USER_SERVICE_TIMEOUT", 5*time.Second),
		},

		JWTSecret: getEnv("JWT_SECRET", "super-secret-key"),
	}

	return cfg, nil
}

func (c *Config) IsProduction() bool {
	env := strings.ToLower(c.Env)
	return env == "prod" || env == "production"
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
