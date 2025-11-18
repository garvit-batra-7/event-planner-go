package config

import (
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Env       string
	Auth      AuthConfig
	Logging   LoggingConfig
}

type AuthConfig struct {
	JWTSecret       string
	JWTExpiration   time.Duration
	PasswordMinLength int
}

type LoggingConfig struct {
	Level   string
	Format  string
	Output  string
}

func Load() *Config {
	_ = godotenv.Load() // loads .env if present

	return &Config{
		Env: os.Getenv("APP_ENV"),
		Auth: AuthConfig{
			JWTSecret:       os.Getenv("JWT_SECRET"),
			JWTExpiration:   time.Hour * 24,
			PasswordMinLength: 8,
		},
		Logging: LoggingConfig{
			Level:  os.Getenv("LOG_LEVEL"),
			Format: os.Getenv("LOG_FORMAT"),
			Output: os.Getenv("LOG_OUTPUT"),
		},
	}
}

// convenience helper
func (c *Config) IsProduction() bool {
	return c.Env == "production"
}
