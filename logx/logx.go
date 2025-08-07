package logx

import (
	"os"
	"sync"
)

// Level represents the logging level
type Level int

const (
	TraceLevel Level = iota
	DebugLevel
	InfoLevel
	WarnLevel
	ErrorLevel
	FatalLevel
)

// String returns the string representation of the level
func (l Level) String() string {
	switch l {
	case TraceLevel:
		return "TRACE"
	case DebugLevel:
		return "DEBUG"
	case InfoLevel:
		return "INFO"
	case WarnLevel:
		return "WARN"
	case ErrorLevel:
		return "ERROR"
	case FatalLevel:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

// Config holds the logger configuration
type Config struct {
	Level         Level
	OutputPath    string // empty for stdout
	Development   bool
	AddCaller     bool
	AddStacktrace bool
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		Level:         InfoLevel,
		OutputPath:    "",
		Development:   false,
		AddCaller:     true,
		AddStacktrace: true,
	}
}

var (
	defaultLogger *Logger
	once          sync.Once
)

// Init initializes the default logger with the given configuration
func Init(config *Config) error {
	var err error
	once.Do(func() {
		defaultLogger, err = New(config)
	})
	return err
}

// InitDefault initializes the default logger with default configuration
func InitDefault() error {
	return Init(DefaultConfig())
}

// Trace logs a trace message (most verbose level)
func Trace(msg string, fields ...Field) {
	if defaultLogger != nil {
		defaultLogger.Trace(msg, fields...)
	}
}

// Tracef logs a formatted trace message (most verbose level)
func Tracef(format string, args ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.Tracef(format, args...)
	}
}

// Debug logs a debug message
func Debug(msg string, fields ...Field) {
	if defaultLogger != nil {
		defaultLogger.Debug(msg, fields...)
	}
}

// Info logs an info message
func Info(msg string, fields ...Field) {
	if defaultLogger != nil {
		defaultLogger.Info(msg, fields...)
	}
}

// Warn logs a warning message
func Warn(msg string, fields ...Field) {
	if defaultLogger != nil {
		defaultLogger.Warn(msg, fields...)
	}
}

// Error logs an error message
func Error(msg string, fields ...Field) {
	if defaultLogger != nil {
		defaultLogger.Error(msg, fields...)
	}
}

// Fatal logs a fatal message and exits
func Fatal(msg string, fields ...Field) {
	if defaultLogger != nil {
		defaultLogger.Fatal(msg, fields...)
	}
	os.Exit(1)
}

// With returns a logger with the given fields
func With(fields ...Field) *Logger {
	if defaultLogger != nil {
		return defaultLogger.With(fields...)
	}
	return nil
}

// Sync flushes any buffered log entries
func Sync() error {
	if defaultLogger != nil {
		return defaultLogger.Sync()
	}
	return nil
}

// NewLogger creates a new logger instance with default configuration
// This function is provided for compatibility with api-gateway packages
func NewLogger() (*Logger, error) {
	return New(DefaultConfig())
}
