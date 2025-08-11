// Package logx provides a structured logging library built on top of Uber's zap logger.
// It offers high-performance, structured logging with additional features like
// sensitive data masking, field-based logging, and easy configuration.
//
// The package provides both a default logger instance and the ability to create
// custom logger instances. All loggers are thread-safe and support concurrent
// logging operations.
package logx

import (
	"fmt"
	"os"
	"sync"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Field represents a structured log field with a key-value pair.
// Fields are used to add structured data to log messages, making them
// easier to parse and analyze.
type Field struct {
	Key   string      // The field key/name
	Value interface{} // The field value
}

// String creates a string field for structured logging.
// This is the preferred way to add string values to log messages.
//
// Example:
//
//	logx.Info("User logged in", logx.String("user_id", "12345"))
func String(key, value string) Field {
	return Field{Key: key, Value: value}
}

// Int creates an integer field for structured logging.
//
// Example:
//
//	logx.Info("Request processed", logx.Int("status_code", 200))
func Int(key string, value int) Field {
	return Field{Key: key, Value: value}
}

// Int64 creates an int64 field for structured logging.
// Useful for large numbers like timestamps or IDs.
//
// Example:
//
//	logx.Info("Event occurred", logx.Int64("timestamp", time.Now().Unix()))
func Int64(key string, value int64) Field {
	return Field{Key: key, Value: value}
}

// Float64 creates a float64 field for structured logging.
// Useful for metrics, percentages, or precise numerical values.
//
// Example:
//
//	logx.Info("Performance metric", logx.Float64("response_time", 0.123))
func Float64(key string, value float64) Field {
	return Field{Key: key, Value: value}
}

// Bool creates a boolean field for structured logging.
// Useful for flags and state indicators.
//
// Example:
//
//	logx.Info("Feature status", logx.Bool("enabled", true))
func Bool(key string, value bool) Field {
	return Field{Key: key, Value: value}
}

// Any creates a field with any value type for structured logging.
// Use this when you need to log complex types or when the type
// is not known at compile time.
//
// Example:
//
//	logx.Info("Complex data", logx.Any("user", userStruct))
func Any(key string, value interface{}) Field {
	return Field{Key: key, Value: value}
}

// ErrorField creates an error field for structured logging.
// This is a convenience function for logging errors with the standard "error" key.
//
// Example:
//
//	if err != nil {
//	    logx.Error("Operation failed", logx.ErrorField(err))
//	}
func ErrorField(err error) Field {
	return Field{Key: "error", Value: err}
}

// Logger wraps the zap logger with additional functionality including
// sensitive data masking and field management. It provides a thread-safe
// interface for structured logging operations.
//
// The Logger maintains a list of fields that are automatically included
// in all log messages, and provides methods for creating child loggers
// with additional fields.
type Logger struct {
	zapLogger *zap.Logger  // The underlying zap logger
	fields    []Field      // Fields to include in all log messages
	mu        sync.RWMutex // Mutex for thread-safe field operations
}

// New creates a new logger instance with the specified configuration.
// The logger is thread-safe and can be used concurrently from multiple goroutines.
//
// If the configuration is invalid or the logger cannot be created,
// an error is returned. Common errors include invalid output paths
// or unsupported log levels.
//
// Example:
//
//	config := &logx.Config{
//	    Level: logx.DebugLevel,
//	    Development: true,
//	}
//	logger, err := logx.New(config)
//	if err != nil {
//	    log.Fatal(err)
//	}
func New(config *Config) (*Logger, error) {
	// Convert our level to zap level
	zapLevel := zapcore.InfoLevel
	switch config.Level {
	case TraceLevel:
		zapLevel = zapcore.DebugLevel // Use Debug level for Trace since zap doesn't have Trace
	case DebugLevel:
		zapLevel = zapcore.DebugLevel
	case InfoLevel:
		zapLevel = zapcore.InfoLevel
	case WarnLevel:
		zapLevel = zapcore.WarnLevel
	case ErrorLevel:
		zapLevel = zapcore.ErrorLevel
	case FatalLevel:
		zapLevel = zapcore.FatalLevel
	}

	// Create encoder config
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.Format(time.RFC3339Nano))
	}
	encoderConfig.LevelKey = "level"
	encoderConfig.MessageKey = "message"
	encoderConfig.CallerKey = "caller"
	encoderConfig.StacktraceKey = "stacktrace"
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder

	// Create core
	var core zapcore.Core
	if config.Development {
		core = zapcore.NewCore(
			zapcore.NewConsoleEncoder(encoderConfig),
			zapcore.AddSync(os.Stdout),
			zapLevel,
		)
	} else {
		// Production configuration
		var output zapcore.WriteSyncer
		if config.OutputPath != "" {
			file, err := os.OpenFile(config.OutputPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				return nil, fmt.Errorf("failed to open log file: %w", err)
			}
			output = zapcore.AddSync(file)
		} else {
			output = zapcore.AddSync(os.Stdout)
		}

		core = zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderConfig),
			output,
			zapLevel,
		)
	}

	// Create zap logger options
	options := []zap.Option{}
	if config.AddCaller {
		options = append(options, zap.AddCaller())
	}
	if config.AddStacktrace {
		options = append(options, zap.AddStacktrace(zapcore.ErrorLevel))
	}

	zapLogger := zap.New(core, options...)

	return &Logger{
		zapLogger: zapLogger,
		fields:    []Field{},
	}, nil
}

// convertFields converts logx fields to zap fields, applying sensitive data masking
func (l *Logger) convertFields(fields []Field) []zap.Field {
	zapFields := make([]zap.Field, 0, len(fields))

	for _, field := range fields {
		// Apply sensitive data masking
		maskedValue := maskSensitiveData(field.Key, field.Value)
		zapFields = append(zapFields, zap.Any(field.Key, maskedValue))
	}

	return zapFields
}

// Trace logs a trace message (most verbose level).
// Trace messages are typically used for detailed debugging and
// are usually disabled in production environments.
//
// The message and fields are automatically masked for sensitive data
// based on the field keys.
func (l *Logger) Trace(msg string, fields ...Field) {
	allFields := append(l.fields, fields...)
	zapFields := l.convertFields(allFields)
	l.zapLogger.Debug(msg, zapFields...)
}

// Debug logs a debug message.
// Debug messages are useful for development and troubleshooting
// but are typically disabled in production environments.
//
// The message and fields are automatically masked for sensitive data
// based on the field keys.
func (l *Logger) Debug(msg string, fields ...Field) {
	allFields := append(l.fields, fields...)
	zapFields := l.convertFields(allFields)
	l.zapLogger.Debug(msg, zapFields...)
}

// Info logs an info message.
// Info messages are used for general application flow and
// important state changes that are not errors.
//
// The message and fields are automatically masked for sensitive data
// based on the field keys.
func (l *Logger) Info(msg string, fields ...Field) {
	allFields := append(l.fields, fields...)
	zapFields := l.convertFields(allFields)
	l.zapLogger.Info(msg, zapFields...)
}

// Warn logs a warning message.
// Warning messages indicate potential issues that should be
// investigated but don't prevent the application from functioning.
//
// The message and fields are automatically masked for sensitive data
// based on the field keys.
func (l *Logger) Warn(msg string, fields ...Field) {
	allFields := append(l.fields, fields...)
	zapFields := l.convertFields(allFields)
	l.zapLogger.Warn(msg, zapFields...)
}

// Error logs an error message.
// Error messages indicate that something has gone wrong and
// should be investigated immediately.
//
// The message and fields are automatically masked for sensitive data
// based on the field keys.
func (l *Logger) Error(msg string, fields ...Field) {
	allFields := append(l.fields, fields...)
	zapFields := l.convertFields(allFields)
	l.zapLogger.Error(msg, zapFields...)
}

// Tracef logs a formatted trace message (most verbose level).
// This method provides printf-style formatting for trace messages.
//
// Example:
//
//	logger.Tracef("Processing user %s with ID %d", username, userID)
func (l *Logger) Tracef(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	l.Trace(msg)
}

// Debugf logs a formatted debug message.
// This method provides printf-style formatting for debug messages.
//
// Example:
//
//	logger.Debugf("Processing request %s with ID %d", requestType, requestID)
func (l *Logger) Debugf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	l.Debug(msg)
}

// Fatal logs a fatal message and then calls os.Exit(1).
// Fatal messages indicate a critical error that prevents the
// application from continuing to run.
//
// The message and fields are automatically masked for sensitive data
// based on the field keys.
func (l *Logger) Fatal(msg string, fields ...Field) {
	allFields := append(l.fields, fields...)
	zapFields := l.convertFields(allFields)
	l.zapLogger.Fatal(msg, zapFields...)
}

// With creates a new logger instance that includes the specified fields
// in all subsequent log messages. This is useful for creating contextual
// loggers that automatically include relevant information.
//
// The returned logger is thread-safe and can be used concurrently.
// The original logger is not modified.
//
// Example:
//
//	userLogger := logger.With(logx.String("user_id", "12345"))
//	userLogger.Info("User action") // Will include user_id in all messages
func (l *Logger) With(fields ...Field) *Logger {
	l.mu.RLock()
	defer l.mu.RUnlock()

	newFields := make([]Field, 0, len(l.fields)+len(fields))
	newFields = append(newFields, l.fields...)
	newFields = append(newFields, fields...)

	return &Logger{
		zapLogger: l.zapLogger,
		fields:    newFields,
	}
}

// Sync flushes any buffered log entries.
// It's important to call this before the application exits
// to ensure all log messages are written.
//
// This method delegates to the underlying zap logger's Sync method.
func (l *Logger) Sync() error {
	return l.zapLogger.Sync()
}

// Infof logs a formatted info message.
// This method provides printf-style formatting for info messages.
//
// Example:
//
//	logger.Infof("User %s logged in from %s", username, ipAddress)
func (l *Logger) Infof(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	l.Info(msg)
}

// Warnf logs a formatted warning message.
// This method provides printf-style formatting for warning messages.
//
// Example:
//
//	logger.Warnf("High memory usage: %d%%", memoryUsage)
func (l *Logger) Warnf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	l.Warn(msg)
}

// Errorf logs a formatted error message.
// This method provides printf-style formatting for error messages.
//
// Example:
//
//	logger.Errorf("Failed to connect to database: %v", err)
func (l *Logger) Errorf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	l.Error(msg)
}

// Fatalf logs a formatted fatal message and then calls os.Exit(1).
// This method provides printf-style formatting for fatal messages.
//
// Example:
//
//	logger.Fatalf("Critical configuration error: %s", configError)
func (l *Logger) Fatalf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	l.Fatal(msg)
}
