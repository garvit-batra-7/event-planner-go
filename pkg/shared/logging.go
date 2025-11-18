package shared

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
	"time"
)

// LogLevel represents the logging level
type LogLevel string

const (
	LevelDebug LogLevel = "DEBUG"
	LevelInfo  LogLevel = "INFO"
	LevelWarn  LogLevel = "WARN"
	LevelError LogLevel = "ERROR"
)

// Logger wraps slog.Logger with additional functionality for event planning context
type Logger struct {
	*slog.Logger
	component string
}

// LoggerConfig holds configuration for the logger
type LoggerConfig struct {
	Level     LogLevel
	Format    string // "json" or "text"
	Output    string // "stdout", "stderr", or file path
	Component string // Component name for structured logging
}

var (
	defaultLogger *Logger
	logFile       *os.File
	lastConfig    LoggerConfig
)

// InitLogger initializes the global logger with the provided configuration
func InitLogger(config LoggerConfig) error {
	var writer io.Writer
	var err error

	// Determine output destination
	switch strings.ToLower(config.Output) {
	case "", "stdout":
		writer = os.Stdout
	case "stderr":
		writer = os.Stderr
	default:
		// Assume it's a file path
		if err := os.MkdirAll(filepath.Dir(config.Output), 0755); err != nil {
			return fmt.Errorf("failed to create log directory: %w", err)
		}
		logFile, err = os.OpenFile(config.Output, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return fmt.Errorf("failed to open log file: %w", err)
		}
		writer = logFile
	}

	// Convert LogLevel to slog.Level
	var slogLevel slog.Level
	switch config.Level {
	case LevelDebug:
		slogLevel = slog.LevelDebug
	case LevelInfo:
		slogLevel = slog.LevelInfo
	case LevelWarn:
		slogLevel = slog.LevelWarn
	case LevelError:
		slogLevel = slog.LevelError
	default:
		slogLevel = slog.LevelInfo
	}

	// Create handler options
	opts := &slog.HandlerOptions{
		Level:     slogLevel,
		AddSource: true, // <-- include file:line of caller
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// Uniform timestamp key + RFC3339 format
			if a.Key == slog.TimeKey {
				return slog.Attr{
					Key:   "timestamp",
					Value: slog.StringValue(a.Value.Time().Format(time.RFC3339)),
				}
			}
			return a
		},
	}

	// Create appropriate handler
	var handler slog.Handler
	switch strings.ToLower(config.Format) {
	case "json":
		handler = slog.NewJSONHandler(writer, opts)
	default:
		handler = slog.NewTextHandler(writer, opts)
	}

	// Create the logger
	logger := slog.New(handler)

	// Add component context if provided
	if config.Component != "" {
		logger = logger.With("component", config.Component)
	}

	defaultLogger = &Logger{
		Logger:    logger,
		component: config.Component,
	}
	lastConfig = config

	return nil
}

// GetLogger returns the default logger instance
func GetLogger() *Logger {
	if defaultLogger == nil {
		// Initialize with default config if not already initialized
		_ = InitLogger(LoggerConfig{
			Level:     LevelInfo,
			Format:    "text",
			Output:    "stdout",
			Component: "event-planner",
		})
	}
	return defaultLogger
}

// NewLogger creates a new logger instance with a specific component name
func NewLogger(component string) *Logger {
	base := GetLogger()
	return &Logger{
		Logger:    base.Logger.With("component", component),
		component: component,
	}
}

// WithContext adds contextual fields to the logger
func (l *Logger) WithContext(fields ...any) *Logger {
	return &Logger{
		Logger:    l.Logger.With(fields...),
		component: l.component,
	}
}

// WithStack attaches a stack trace to the log fields (useful for panics/errors)
func (l *Logger) WithStack(fields ...any) []any {
	stack := debug.Stack()
	out := make([]any, 0, len(fields)+2)
	out = append(out, fields...)
	out = append(out, "stack", string(stack))
	return out
}

// --- Event / User / API / DB convenience ---

func (l *Logger) LogEvent(level LogLevel, eventID string, message string, fields ...any) {
	args := append([]any{"event_id", eventID}, fields...)
	l.logWithLevel(level, message, args...)
}

func (l *Logger) LogUser(level LogLevel, userID string, action string, message string, fields ...any) {
	args := append([]any{
		"user_id", userID,
		"action",  action,
	}, fields...)
	l.logWithLevel(level, message, args...)
}

func (l *Logger) LogAPI(level LogLevel, method, endpoint string, statusCode int, duration time.Duration, fields ...any) {
	args := append([]any{
		"method",       method,
		"endpoint",     endpoint,
		"status_code",  statusCode,
		"duration_ms",  duration.Milliseconds(),
	}, fields...)
	l.logWithLevel(level, "API request processed", args...)
}

func (l *Logger) LogDB(level LogLevel, operation, table string, duration time.Duration, fields ...any) {
	args := append([]any{
		"db_operation", operation,
		"table",        table,
		"duration_ms",  duration.Milliseconds(),
	}, fields...)
	l.logWithLevel(level, "Database operation", args...)
}

func (l *Logger) LogError(err error, message string, fields ...any) {
	args := append([]any{"error", err.Error()}, fields...)
	l.logWithLevel(LevelError, message, args...)
}

// logWithLevel is a helper method to log with the specified level
func (l *Logger) logWithLevel(level LogLevel, message string, fields ...any) {
	switch level {
	case LevelDebug:
		l.Debug(message, fields...)
	case LevelInfo:
		l.Info(message, fields...)
	case LevelWarn:
		l.Warn(message, fields...)
	case LevelError:
		l.Error(message, fields...)
	default:
		l.Info(message, fields...)
	}
}

// Convenience methods for common log levels

func (l *Logger) Debug(message string, fields ...any) { l.Logger.Debug(message, fields...) }
func (l *Logger) Info(message string, fields ...any)  { l.Logger.Info(message, fields...) }
func (l *Logger) Warn(message string, fields ...any)  { l.Logger.Warn(message, fields...) }
func (l *Logger) Error(message string, fields ...any) { l.Logger.Error(message, fields...) }

// Package-level convenience functions

func Debug(message string, fields ...any) { GetLogger().Debug(message, fields...) }
func Info(message string, fields ...any)  { GetLogger().Info(message, fields...) }
func Warn(message string, fields ...any)  { GetLogger().Warn(message, fields...) }
func Error(message string, fields ...any) { GetLogger().Error(message, fields...) }

func LogEvent(level LogLevel, eventID string, message string, fields ...any) {
	GetLogger().LogEvent(level, eventID, message, fields...)
}

func LogUser(level LogLevel, userID string, action string, message string, fields ...any) {
	GetLogger().LogUser(level, userID, action, message, fields...)
}

func LogAPI(level LogLevel, method, endpoint string, statusCode int, duration time.Duration, fields ...any) {
	GetLogger().LogAPI(level, method, endpoint, statusCode, duration, fields...)
}

func LogDB(level LogLevel, operation, table string, duration time.Duration, fields ...any) {
	GetLogger().LogDB(level, operation, table, duration, fields...)
}

func LogError(err error, message string, fields ...any) {
	GetLogger().LogError(err, message, fields...)
}

// Cleanup closes any open log files
func Cleanup() error {
	if logFile != nil {
		return logFile.Close()
	}
	return nil
}

// SetLogLevel dynamically changes the log level (note: this creates a new handler)
func SetLogLevel(level LogLevel) error {
	if defaultLogger == nil {
		return fmt.Errorf("logger not initialized")
	}
	config := lastConfig
	config.Level = level
	return InitLogger(config)
}

// --- Gin panic capture integration ---

// writerAdapter lets gin.RecoveryWithWriter write into slog with a stack trace.
type writerAdapter struct{}

func (writerAdapter) Write(p []byte) (int, error) {
	// p already contains the stack produced by gin's recovery. Log it as an error:
	GetLogger().Error("Recovered panic", "stack", string(p))
	return len(p), nil
}

// GinRecoveryWriter returns a writer suitable for gin.RecoveryWithWriter().
func GinRecoveryWriter() io.Writer { return writerAdapter{} }
