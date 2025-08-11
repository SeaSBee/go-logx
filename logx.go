// Package logx provides a structured logging library built on top of Uber's zap logger.
// It offers high-performance, structured logging with additional features like
// sensitive data masking, field-based logging, and easy configuration.
//
// The package provides both a default logger instance and the ability to create
// custom logger instances. All loggers are thread-safe and support concurrent
// logging operations.
package logx

import (
	"os"
	"sync"
)

// Level represents the logging level used to control the verbosity of log output.
// Higher levels are more restrictive - only messages at or above the configured
// level will be logged.
type Level int

const (
	// TraceLevel is the most verbose logging level.
	// Trace messages are typically used for detailed debugging and
	// are usually disabled in production environments.
	TraceLevel Level = iota

	// DebugLevel is used for debug messages.
	// Debug messages are useful for development and troubleshooting
	// but are typically disabled in production environments.
	DebugLevel

	// InfoLevel is the default logging level.
	// Info messages are used for general application flow and
	// important state changes that are not errors.
	InfoLevel

	// WarnLevel is used for warning messages.
	// Warning messages indicate potential issues that should be
	// investigated but don't prevent the application from functioning.
	WarnLevel

	// ErrorLevel is used for error messages.
	// Error messages indicate that something has gone wrong and
	// should be investigated immediately.
	ErrorLevel

	// FatalLevel is used for fatal messages.
	// Fatal messages indicate a critical error that prevents the
	// application from continuing to run. Logging a fatal message
	// will cause the application to exit with code 1.
	FatalLevel
)

// String returns the string representation of the logging level.
// This is useful for configuration files and debugging.
//
// Example:
//
//	level := logx.InfoLevel
//	fmt.Println(level.String()) // Output: "INFO"
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

// Config holds the configuration for creating a logger instance.
// All fields are optional and have sensible defaults.
type Config struct {
	// Level specifies the minimum logging level.
	// Only messages at or above this level will be logged.
	// Default: InfoLevel
	Level Level

	// OutputPath specifies the file path for log output.
	// If empty, logs will be written to stdout.
	// Default: "" (stdout)
	OutputPath string

	// Development enables development mode with console output
	// and more verbose formatting. In production, JSON output
	// is used for better parsing.
	// Default: false
	Development bool

	// AddCaller adds the calling function's file name and line number
	// to log messages. This is useful for debugging.
	// Default: true
	AddCaller bool

	// AddStacktrace adds stack traces to error and fatal messages.
	// This can be helpful for debugging but increases log size.
	// Default: true
	AddStacktrace bool
}

// DefaultConfig returns a default configuration suitable for most applications.
// The default configuration uses InfoLevel logging, writes to stdout,
// uses JSON format, and includes caller information and stack traces.
//
// Example:
//
//	config := logx.DefaultConfig()
//	config.Level = logx.DebugLevel  // Override for development
//	logger, err := logx.New(config)
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
	// defaultLogger is the global logger instance used by package-level functions
	defaultLogger *Logger

	// once ensures that the default logger is initialized only once
	once sync.Once
)

// Init initializes the default logger with the given configuration.
// This function can only be called once - subsequent calls will be ignored.
// If initialization fails, an error is returned.
//
// It's recommended to call this function early in your application's
// startup process, typically in main() or init().
//
// Example:
//
//	config := &logx.Config{
//	    Level: logx.DebugLevel,
//	    Development: true,
//	}
//	if err := logx.Init(config); err != nil {
//	    log.Fatal(err)
//	}
func Init(config *Config) error {
	var err error
	once.Do(func() {
		defaultLogger, err = New(config)
	})
	return err
}

// InitDefault initializes the default logger with the default configuration.
// This is a convenience function that calls Init(DefaultConfig()).
//
// Example:
//
//	if err := logx.InitDefault(); err != nil {
//	    log.Fatal(err)
//	}
func InitDefault() error {
	return Init(DefaultConfig())
}

// Trace logs a trace message using the default logger.
// If the default logger is not initialized, the message is ignored.
// Trace messages are typically used for detailed debugging and
// are usually disabled in production environments.
//
// The message and fields are automatically masked for sensitive data
// based on the field keys.
//
// Example:
//
//	logx.Trace("Processing request", logx.String("request_id", "12345"))
func Trace(msg string, fields ...Field) {
	if defaultLogger != nil {
		defaultLogger.Trace(msg, fields...)
	}
}

// Tracef logs a formatted trace message using the default logger.
// This function provides printf-style formatting for trace messages.
//
// Example:
//
//	logx.Tracef("Processing user %s with ID %d", username, userID)
func Tracef(format string, args ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.Tracef(format, args...)
	}
}

// Debugf logs a formatted debug message using the default logger.
// This function provides printf-style formatting for debug messages.
//
// Example:
//
//	logx.Debugf("Processing request %s with ID %d", requestType, requestID)
func Debugf(format string, args ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.Debugf(format, args...)
	}
}

// Debug logs a debug message using the default logger.
// If the default logger is not initialized, the message is ignored.
// Debug messages are useful for development and troubleshooting
// but are typically disabled in production environments.
//
// The message and fields are automatically masked for sensitive data
// based on the field keys.
//
// Example:
//
//	logx.Debug("Database query executed", logx.Int("rows_affected", 5))
func Debug(msg string, fields ...Field) {
	if defaultLogger != nil {
		defaultLogger.Debug(msg, fields...)
	}
}

// Info logs an info message using the default logger.
// If the default logger is not initialized, the message is ignored.
// Info messages are used for general application flow and
// important state changes that are not errors.
//
// The message and fields are automatically masked for sensitive data
// based on the field keys.
//
// Example:
//
//	logx.Info("User logged in", logx.String("user_id", "12345"))
func Info(msg string, fields ...Field) {
	if defaultLogger != nil {
		defaultLogger.Info(msg, fields...)
	}
}

// Warn logs a warning message using the default logger.
// If the default logger is not initialized, the message is ignored.
// Warning messages indicate potential issues that should be
// investigated but don't prevent the application from functioning.
//
// The message and fields are automatically masked for sensitive data
// based on the field keys.
//
// Example:
//
//	logx.Warn("High memory usage detected", logx.Float64("usage_percent", 85.5))
func Warn(msg string, fields ...Field) {
	if defaultLogger != nil {
		defaultLogger.Warn(msg, fields...)
	}
}

// Error logs an error message using the default logger.
// If the default logger is not initialized, the message is ignored.
// Error messages indicate that something has gone wrong and
// should be investigated immediately.
//
// The message and fields are automatically masked for sensitive data
// based on the field keys.
//
// Example:
//
//	logx.Error("Database connection failed", logx.ErrorField(err))
func Error(msg string, fields ...Field) {
	if defaultLogger != nil {
		defaultLogger.Error(msg, fields...)
	}
}

// Fatal logs a fatal message using the default logger and then calls os.Exit(1).
// If the default logger is not initialized, the message is ignored but
// the application will still exit.
// Fatal messages indicate a critical error that prevents the
// application from continuing to run.
//
// The message and fields are automatically masked for sensitive data
// based on the field keys.
//
// Example:
//
//	logx.Fatal("Critical configuration error", logx.String("config_file", "app.conf"))
func Fatal(msg string, fields ...Field) {
	if defaultLogger != nil {
		defaultLogger.Fatal(msg, fields...)
	} else {
		os.Exit(1)
	}
}

// With creates a new logger instance that includes the specified fields
// in all subsequent log messages. This is useful for creating contextual
// loggers that automatically include relevant information.
//
// If the default logger is not initialized, nil is returned.
// The returned logger is thread-safe and can be used concurrently.
//
// Example:
//
//	userLogger := logx.With(logx.String("user_id", "12345"))
//	userLogger.Info("User action") // Will include user_id in all messages
func With(fields ...Field) *Logger {
	if defaultLogger != nil {
		return defaultLogger.With(fields...)
	}
	return nil
}

// Sync flushes any buffered log entries from the default logger.
// It's important to call this before the application exits
// to ensure all log messages are written.
//
// If the default logger is not initialized, this function does nothing.
func Sync() error {
	if defaultLogger != nil {
		return defaultLogger.Sync()
	}
	return nil
}

// NewLogger creates a new logger instance with the default configuration.
// This is a convenience function that calls New(DefaultConfig()).
//
// The returned logger is independent of the default logger and can be
// configured and used separately.
//
// Example:
//
//	logger, err := logx.NewLogger()
//	if err != nil {
//	    log.Fatal(err)
//	}
//	logger.Info("Custom logger message")
func NewLogger() (*Logger, error) {
	return New(DefaultConfig())
}
