package shared

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Environment represents the application environment
type Environment string

const (
	EnvDevelopment Environment = "development"
	EnvStaging     Environment = "staging"
	EnvProduction  Environment = "production"
	EnvTest        Environment = "test"
)

// Config holds all configuration for the event-planner application
type Config struct {
	// Application settings
	App AppConfig `json:"app"`
	
	// Server settings
	Server ServerConfig `json:"server"`
	
	// Database settings
	Database DatabaseConfig `json:"database"`
	
	// Redis/Cache settings
	Cache CacheConfig `json:"cache"`
	
	// Authentication settings
	Auth AuthConfig `json:"auth"`
	
	// Email settings
	Email EmailConfig `json:"email"`
	
	// File storage settings
	Storage StorageConfig `json:"storage"`
	
	// Logging settings
	Logging LoggingConfig `json:"logging"`
	
	// External API settings
	ExternalAPIs ExternalAPIConfig `json:"external_apis"`
	
	// Feature flags
	Features FeatureFlags `json:"features"`
}

// AppConfig contains general application settings
type AppConfig struct {
	Name        string      `json:"name"`
	Version     string      `json:"version"`
	Environment Environment `json:"environment"`
	Debug       bool        `json:"debug"`
	TimeZone    string      `json:"timezone"`
	BaseURL     string      `json:"base_url"`
}

// ServerConfig contains HTTP server settings
type ServerConfig struct {
	Host            string        `json:"host"`
	Port            int           `json:"port"`
	ReadTimeout     time.Duration `json:"read_timeout"`
	WriteTimeout    time.Duration `json:"write_timeout"`
	IdleTimeout     time.Duration `json:"idle_timeout"`
	ShutdownTimeout time.Duration `json:"shutdown_timeout"`
	TLS             TLSConfig     `json:"tls"`
}

// TLSConfig contains TLS/SSL settings
type TLSConfig struct {
	Enabled  bool   `json:"enabled"`
	CertFile string `json:"cert_file"`
	KeyFile  string `json:"key_file"`
}

// DatabaseConfig contains database connection settings
type DatabaseConfig struct {
	Driver          string        `json:"driver"`
	Host            string        `json:"host"`
	Port            int           `json:"port"`
	Name            string        `json:"name"`
	Username        string        `json:"username"`
	Password        string        `json:"password"`
	SSLMode         string        `json:"ssl_mode"`
	MaxOpenConns    int           `json:"max_open_conns"`
	MaxIdleConns    int           `json:"max_idle_conns"`
	ConnMaxLifetime time.Duration `json:"conn_max_lifetime"`
	ConnMaxIdleTime time.Duration `json:"conn_max_idle_time"`
}

// CacheConfig contains Redis/cache settings
type CacheConfig struct {
	Enabled      bool          `json:"enabled"`
	Host         string        `json:"host"`
	Port         int           `json:"port"`
	Password     string        `json:"password"`
	Database     int           `json:"database"`
	PoolSize     int           `json:"pool_size"`
	MinIdleConns int           `json:"min_idle_conns"`
	MaxRetries   int           `json:"max_retries"`
	TTL          time.Duration `json:"ttl"`
}

// AuthConfig contains authentication and authorization settings
type AuthConfig struct {
	JWTSecret           string        `json:"jwt_secret"`
	JWTExpiration       time.Duration `json:"jwt_expiration"`
	RefreshTokenExpiry  time.Duration `json:"refresh_token_expiry"`
	PasswordMinLength   int           `json:"password_min_length"`
	MaxLoginAttempts    int           `json:"max_login_attempts"`
	LoginAttemptWindow  time.Duration `json:"login_attempt_window"`
	RequireEmailVerify  bool          `json:"require_email_verify"`
	EnableTwoFactor     bool          `json:"enable_two_factor"`
}

// EmailConfig contains email service settings
type EmailConfig struct {
	Provider    string `json:"provider"`
	SMTPHost    string `json:"smtp_host"`
	SMTPPort    int    `json:"smtp_port"`
	SMTPUser    string `json:"smtp_user"`
	SMTPPass    string `json:"smtp_pass"`
	FromEmail   string `json:"from_email"`
	FromName    string `json:"from_name"`
	ReplyToEmail string `json:"reply_to_email"`
	
	// Email templates
	Templates EmailTemplatesConfig `json:"templates"`
}

// EmailTemplatesConfig contains email template settings
type EmailTemplatesConfig struct {
	WelcomeTemplate      string `json:"welcome_template"`
	ResetPasswordTemplate string `json:"reset_password_template"`
	EventInviteTemplate   string `json:"event_invite_template"`
	EventReminderTemplate string `json:"event_reminder_template"`
}

// StorageConfig contains file storage settings
type StorageConfig struct {
	Provider        string `json:"provider"` // "local", "s3", "gcs"
	LocalPath       string `json:"local_path"`
	MaxFileSize     int64  `json:"max_file_size"`
	AllowedTypes    []string `json:"allowed_types"`
	
	// S3 Configuration
	S3Config S3Config `json:"s3"`
}

// S3Config contains AWS S3 settings
type S3Config struct {
	AccessKeyID     string `json:"access_key_id"`
	SecretAccessKey string `json:"secret_access_key"`
	Region          string `json:"region"`
	Bucket          string `json:"bucket"`
	Endpoint        string `json:"endpoint"` // For S3-compatible services
}

// LoggingConfig contains logging settings
type LoggingConfig struct {
	Level     string `json:"level"`
	Format    string `json:"format"`
	Output    string `json:"output"`
	Component string `json:"component"`
}

// ExternalAPIConfig contains settings for external API integrations
type ExternalAPIConfig struct {
	GoogleMaps    GoogleMapsConfig    `json:"google_maps"`
	PaymentGateway PaymentGatewayConfig `json:"payment_gateway"`
	CalendarAPI   CalendarAPIConfig   `json:"calendar_api"`
}

// GoogleMapsConfig contains Google Maps API settings
type GoogleMapsConfig struct {
	APIKey  string `json:"api_key"`
	Enabled bool   `json:"enabled"`
}

// PaymentGatewayConfig contains payment processing settings
type PaymentGatewayConfig struct {
	Provider    string `json:"provider"` // "stripe", "paypal", etc.
	APIKey      string `json:"api_key"`
	SecretKey   string `json:"secret_key"`
	WebhookSecret string `json:"webhook_secret"`
	Enabled     bool   `json:"enabled"`
}

// CalendarAPIConfig contains calendar integration settings
type CalendarAPIConfig struct {
	Provider     string `json:"provider"` // "google", "outlook", etc.
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	RedirectURL  string `json:"redirect_url"`
	Enabled      bool   `json:"enabled"`
}

// FeatureFlags contains feature toggle settings
type FeatureFlags struct {
	EnableRegistration    bool `json:"enable_registration"`
	EnableSocialLogin     bool `json:"enable_social_login"`
	EnablePayments        bool `json:"enable_payments"`
	EnableNotifications   bool `json:"enable_notifications"`
	EnableEventSharing    bool `json:"enable_event_sharing"`
	EnableAdvancedSearch  bool `json:"enable_advanced_search"`
	EnableEventTemplates  bool `json:"enable_event_templates"`
	EnableMultiLanguage   bool `json:"enable_multi_language"`
}

var (
	appConfig *Config
)

// LoadConfig loads configuration from environment variables with defaults
func LoadConfig() (*Config, error) {
	config := &Config{
		App: AppConfig{
			Name:        getEnvString("APP_NAME", "Event Planner"),
			Version:     getEnvString("APP_VERSION", "1.0.0"),
			Environment: Environment(getEnvString("APP_ENV", string(EnvDevelopment))),
			Debug:       getEnvBool("APP_DEBUG", true),
			TimeZone:    getEnvString("APP_TIMEZONE", "UTC"),
			BaseURL:     getEnvString("APP_BASE_URL", "http://localhost:8080"),
		},
		Server: ServerConfig{
			Host:            getEnvString("SERVER_HOST", "localhost"),
			Port:            getEnvInt("SERVER_PORT", 8080),
			ReadTimeout:     getEnvDuration("SERVER_READ_TIMEOUT", 15*time.Second),
			WriteTimeout:    getEnvDuration("SERVER_WRITE_TIMEOUT", 15*time.Second),
			IdleTimeout:     getEnvDuration("SERVER_IDLE_TIMEOUT", 60*time.Second),
			ShutdownTimeout: getEnvDuration("SERVER_SHUTDOWN_TIMEOUT", 30*time.Second),
			TLS: TLSConfig{
				Enabled:  getEnvBool("SERVER_TLS_ENABLED", false),
				CertFile: getEnvString("SERVER_TLS_CERT_FILE", ""),
				KeyFile:  getEnvString("SERVER_TLS_KEY_FILE", ""),
			},
		},
		Database: DatabaseConfig{
			Driver:          getEnvString("DB_DRIVER", "postgres"),
			Host:            getEnvString("DB_HOST", "localhost"),
			Port:            getEnvInt("DB_PORT", 5432),
			Name:            getEnvString("DB_NAME", "event_planner"),
			Username:        getEnvString("DB_USERNAME", "postgres"),
			Password:        getEnvString("DB_PASSWORD", ""),
			SSLMode:         getEnvString("DB_SSL_MODE", "disable"),
			MaxOpenConns:    getEnvInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getEnvInt("DB_MAX_IDLE_CONNS", 10),
			ConnMaxLifetime: getEnvDuration("DB_CONN_MAX_LIFETIME", 5*time.Minute),
			ConnMaxIdleTime: getEnvDuration("DB_CONN_MAX_IDLE_TIME", 5*time.Minute),
		},
		Cache: CacheConfig{
			Enabled:      getEnvBool("CACHE_ENABLED", false),
			Host:         getEnvString("CACHE_HOST", "localhost"),
			Port:         getEnvInt("CACHE_PORT", 6379),
			Password:     getEnvString("CACHE_PASSWORD", ""),
			Database:     getEnvInt("CACHE_DATABASE", 0),
			PoolSize:     getEnvInt("CACHE_POOL_SIZE", 10),
			MinIdleConns: getEnvInt("CACHE_MIN_IDLE_CONNS", 5),
			MaxRetries:   getEnvInt("CACHE_MAX_RETRIES", 3),
			TTL:          getEnvDuration("CACHE_TTL", 1*time.Hour),
		},
		Auth: AuthConfig{
			JWTSecret:          getEnvString("JWT_SECRET", "your-super-secret-jwt-key"),
			JWTExpiration:      getEnvDuration("JWT_EXPIRATION", 24*time.Hour),
			RefreshTokenExpiry: getEnvDuration("REFRESH_TOKEN_EXPIRY", 7*24*time.Hour),
			PasswordMinLength:  getEnvInt("PASSWORD_MIN_LENGTH", 8),
			MaxLoginAttempts:   getEnvInt("MAX_LOGIN_ATTEMPTS", 5),
			LoginAttemptWindow: getEnvDuration("LOGIN_ATTEMPT_WINDOW", 15*time.Minute),
			RequireEmailVerify: getEnvBool("REQUIRE_EMAIL_VERIFY", true),
			EnableTwoFactor:    getEnvBool("ENABLE_TWO_FACTOR", false),
		},
		Email: EmailConfig{
			Provider:     getEnvString("EMAIL_PROVIDER", "smtp"),
			SMTPHost:     getEnvString("SMTP_HOST", "localhost"),
			SMTPPort:     getEnvInt("SMTP_PORT", 587),
			SMTPUser:     getEnvString("SMTP_USER", ""),
			SMTPPass:     getEnvString("SMTP_PASS", ""),
			FromEmail:    getEnvString("FROM_EMAIL", "noreply@eventplanner.com"),
			FromName:     getEnvString("FROM_NAME", "Event Planner"),
			ReplyToEmail: getEnvString("REPLY_TO_EMAIL", "support@eventplanner.com"),
			Templates: EmailTemplatesConfig{
				WelcomeTemplate:      getEnvString("WELCOME_TEMPLATE", "welcome.html"),
				ResetPasswordTemplate: getEnvString("RESET_PASSWORD_TEMPLATE", "reset_password.html"),
				EventInviteTemplate:   getEnvString("EVENT_INVITE_TEMPLATE", "event_invite.html"),
				EventReminderTemplate: getEnvString("EVENT_REMINDER_TEMPLATE", "event_reminder.html"),
			},
		},
		Storage: StorageConfig{
			Provider:     getEnvString("STORAGE_PROVIDER", "local"),
			LocalPath:    getEnvString("STORAGE_LOCAL_PATH", "./uploads"),
			MaxFileSize:  getEnvInt64("STORAGE_MAX_FILE_SIZE", 10*1024*1024), // 10MB
			AllowedTypes: strings.Split(getEnvString("STORAGE_ALLOWED_TYPES", "jpg,jpeg,png,gif,pdf"), ","),
			S3Config: S3Config{
				AccessKeyID:     getEnvString("AWS_ACCESS_KEY_ID", ""),
				SecretAccessKey: getEnvString("AWS_SECRET_ACCESS_KEY", ""),
				Region:          getEnvString("AWS_REGION", "us-east-1"),
				Bucket:          getEnvString("S3_BUCKET", ""),
				Endpoint:        getEnvString("S3_ENDPOINT", ""),
			},
		},
		Logging: LoggingConfig{
			Level:     getEnvString("LOG_LEVEL", "info"),
			Format:    getEnvString("LOG_FORMAT", "text"),
			Output:    getEnvString("LOG_OUTPUT", "stdout"),
			Component: getEnvString("LOG_COMPONENT", "event-planner"),
		},
		ExternalAPIs: ExternalAPIConfig{
			GoogleMaps: GoogleMapsConfig{
				APIKey:  getEnvString("GOOGLE_MAPS_API_KEY", ""),
				Enabled: getEnvBool("GOOGLE_MAPS_ENABLED", false),
			},
			PaymentGateway: PaymentGatewayConfig{
				Provider:      getEnvString("PAYMENT_PROVIDER", "stripe"),
				APIKey:        getEnvString("PAYMENT_API_KEY", ""),
				SecretKey:     getEnvString("PAYMENT_SECRET_KEY", ""),
				WebhookSecret: getEnvString("PAYMENT_WEBHOOK_SECRET", ""),
				Enabled:       getEnvBool("PAYMENT_ENABLED", false),
			},
			CalendarAPI: CalendarAPIConfig{
				Provider:     getEnvString("CALENDAR_PROVIDER", "google"),
				ClientID:     getEnvString("CALENDAR_CLIENT_ID", ""),
				ClientSecret: getEnvString("CALENDAR_CLIENT_SECRET", ""),
				RedirectURL:  getEnvString("CALENDAR_REDIRECT_URL", ""),
				Enabled:      getEnvBool("CALENDAR_ENABLED", false),
			},
		},
		Features: FeatureFlags{
			EnableRegistration:    getEnvBool("FEATURE_ENABLE_REGISTRATION", true),
			EnableSocialLogin:     getEnvBool("FEATURE_ENABLE_SOCIAL_LOGIN", false),
			EnablePayments:        getEnvBool("FEATURE_ENABLE_PAYMENTS", false),
			EnableNotifications:   getEnvBool("FEATURE_ENABLE_NOTIFICATIONS", true),
			EnableEventSharing:    getEnvBool("FEATURE_ENABLE_EVENT_SHARING", true),
			EnableAdvancedSearch:  getEnvBool("FEATURE_ENABLE_ADVANCED_SEARCH", true),
			EnableEventTemplates:  getEnvBool("FEATURE_ENABLE_EVENT_TEMPLATES", true),
			EnableMultiLanguage:   getEnvBool("FEATURE_ENABLE_MULTI_LANGUAGE", false),
		},
	}

	// Validate the configuration
	if err := validateConfig(config); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	appConfig = config
	return config, nil
}

// GetConfig returns the loaded configuration
func GetConfig() *Config {
	if appConfig == nil {
		// Try to load config if not already loaded
		config, err := LoadConfig()
		if err != nil {
			panic(fmt.Sprintf("Failed to load configuration: %v", err))
		}
		return config
	}
	return appConfig
}

// validateConfig validates the configuration values
func validateConfig(config *Config) error {
	// Validate required fields
	if config.Database.Password == "" && config.App.Environment == EnvProduction {
		return fmt.Errorf("database password is required in production")
	}

	if config.Auth.JWTSecret == "your-super-secret-jwt-key" && config.App.Environment == EnvProduction {
		return fmt.Errorf("JWT secret must be changed in production")
	}

	// Validate port ranges
	if config.Server.Port < 1 || config.Server.Port > 65535 {
		return fmt.Errorf("server port must be between 1 and 65535")
	}

	// Validate environment
	validEnvs := []Environment{EnvDevelopment, EnvStaging, EnvProduction, EnvTest}
	validEnv := false
	for _, env := range validEnvs {
		if config.App.Environment == env {
			validEnv = true
			break
		}
	}
	if !validEnv {
		return fmt.Errorf("invalid environment: %s", config.App.Environment)
	}

	return nil
}

// GetDSN returns the database connection string
func (c *Config) GetDSN() string {
	switch c.Database.Driver {
	case "postgres":
		return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			c.Database.Host,
			c.Database.Port,
			c.Database.Username,
			c.Database.Password,
			c.Database.Name,
			c.Database.SSLMode,
		)
	case "mysql":
		return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s",
			c.Database.Username,
			c.Database.Password,
			c.Database.Host,
			c.Database.Port,
			c.Database.Name,
		)
	default:
		return ""
	}
}

// GetServerAddress returns the server address string
func (c *Config) GetServerAddress() string {
	return fmt.Sprintf("%s:%d", c.Server.Host, c.Server.Port)
}

// IsProduction returns true if running in production environment
func (c *Config) IsProduction() bool {
	return c.App.Environment == EnvProduction
}

// IsDevelopment returns true if running in development environment
func (c *Config) IsDevelopment() bool {
	return c.App.Environment == EnvDevelopment
}

// Helper functions to read environment variables with defaults

func getEnvString(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvInt64(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.ParseInt(value, 10, 64); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}