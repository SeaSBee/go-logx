package logx

import (
	"fmt"
	"os"
	"sync"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Field represents a structured log field
type Field struct {
	Key   string
	Value interface{}
}

// String creates a string field
func String(key, value string) Field {
	return Field{Key: key, Value: value}
}

// Int creates an integer field
func Int(key string, value int) Field {
	return Field{Key: key, Value: value}
}

// Int64 creates an int64 field
func Int64(key string, value int64) Field {
	return Field{Key: key, Value: value}
}

// Float64 creates a float64 field
func Float64(key string, value float64) Field {
	return Field{Key: key, Value: value}
}

// Bool creates a boolean field
func Bool(key string, value bool) Field {
	return Field{Key: key, Value: value}
}

// Any creates a field with any value
func Any(key string, value interface{}) Field {
	return Field{Key: key, Value: value}
}

// ErrorField creates an error field
func ErrorField(err error) Field {
	return Field{Key: "error", Value: err}
}

// Logger wraps the zap logger with additional functionality
type Logger struct {
	zapLogger *zap.Logger
	fields    []Field
	mu        sync.RWMutex
}

// New creates a new logger instance
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
		core = zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderConfig),
			zapcore.AddSync(os.Stdout),
			zapLevel,
		)
	}

	// Create zap logger options
	opts := []zap.Option{
		zap.AddCallerSkip(1),
	}

	if config.AddCaller {
		opts = append(opts, zap.AddCaller())
	}

	if config.AddStacktrace {
		opts = append(opts, zap.AddStacktrace(zapcore.ErrorLevel))
	}

	// Create zap logger
	zapLogger := zap.New(core, opts...)

	return &Logger{
		zapLogger: zapLogger,
		fields:    make([]Field, 0),
	}, nil
}

// convertFields converts our Field type to zap.Field
func (l *Logger) convertFields(fields []Field) []zap.Field {
	zapFields := make([]zap.Field, 0, len(fields))

	for _, field := range fields {
		// Apply masking for sensitive data
		maskedValue := maskSensitiveData(field.Key, field.Value)

		switch v := maskedValue.(type) {
		case string:
			zapFields = append(zapFields, zap.String(field.Key, v))
		case int:
			zapFields = append(zapFields, zap.Int(field.Key, v))
		case int64:
			zapFields = append(zapFields, zap.Int64(field.Key, v))
		case float64:
			zapFields = append(zapFields, zap.Float64(field.Key, v))
		case bool:
			zapFields = append(zapFields, zap.Bool(field.Key, v))
		case error:
			zapFields = append(zapFields, zap.Error(v))
		default:
			zapFields = append(zapFields, zap.Any(field.Key, v))
		}
	}

	return zapFields
}

// Trace logs a trace message (most verbose level)
func (l *Logger) Trace(msg string, fields ...Field) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	allFields := append(l.fields, fields...)
	zapFields := l.convertFields(allFields)
	l.zapLogger.Debug(msg, zapFields...) // Use Debug level for Trace since zap doesn't have Trace
}

// Debug logs a debug message
func (l *Logger) Debug(msg string, fields ...Field) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	allFields := append(l.fields, fields...)
	zapFields := l.convertFields(allFields)
	l.zapLogger.Debug(msg, zapFields...)
}

// Info logs an info message
func (l *Logger) Info(msg string, fields ...Field) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	allFields := append(l.fields, fields...)
	zapFields := l.convertFields(allFields)
	l.zapLogger.Info(msg, zapFields...)
}

// Warn logs a warning message
func (l *Logger) Warn(msg string, fields ...Field) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	allFields := append(l.fields, fields...)
	zapFields := l.convertFields(allFields)
	l.zapLogger.Warn(msg, zapFields...)
}

// Error logs an error message
func (l *Logger) Error(msg string, fields ...Field) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	allFields := append(l.fields, fields...)
	zapFields := l.convertFields(allFields)
	l.zapLogger.Error(msg, zapFields...)
}

// Tracef logs a formatted trace message (most verbose level)
func (l *Logger) Tracef(format string, args ...interface{}) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	msg := fmt.Sprintf(format, args...)
	zapFields := l.convertFields(l.fields)
	l.zapLogger.Debug(msg, zapFields...) // Use Debug level for Trace since zap doesn't have Trace
}

// Debugf logs a formatted debug message
func (l *Logger) Debugf(format string, args ...interface{}) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	msg := fmt.Sprintf(format, args...)
	zapFields := l.convertFields(l.fields)
	l.zapLogger.Debug(msg, zapFields...)
}

// Fatal logs a fatal message
func (l *Logger) Fatal(msg string, fields ...Field) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	allFields := append(l.fields, fields...)
	zapFields := l.convertFields(allFields)
	l.zapLogger.Fatal(msg, zapFields...)
}

// With returns a new logger with the given fields
func (l *Logger) With(fields ...Field) *Logger {
	l.mu.RLock()
	defer l.mu.RUnlock()

	newFields := make([]Field, len(l.fields)+len(fields))
	copy(newFields, l.fields)
	copy(newFields[len(l.fields):], fields)

	return &Logger{
		zapLogger: l.zapLogger,
		fields:    newFields,
	}
}

// Sync flushes any buffered log entries
func (l *Logger) Sync() error {
	return l.zapLogger.Sync()
}

// Infof logs a formatted info message
func (l *Logger) Infof(format string, args ...interface{}) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	msg := fmt.Sprintf(format, args...)
	zapFields := l.convertFields(l.fields)
	l.zapLogger.Info(msg, zapFields...)
}

// Warnf logs a formatted warning message
func (l *Logger) Warnf(format string, args ...interface{}) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	msg := fmt.Sprintf(format, args...)
	zapFields := l.convertFields(l.fields)
	l.zapLogger.Warn(msg, zapFields...)
}

// Errorf logs a formatted error message
func (l *Logger) Errorf(format string, args ...interface{}) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	msg := fmt.Sprintf(format, args...)
	zapFields := l.convertFields(l.fields)
	l.zapLogger.Error(msg, zapFields...)
}

// Fatalf logs a formatted fatal message
func (l *Logger) Fatalf(format string, args ...interface{}) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	msg := fmt.Sprintf(format, args...)
	zapFields := l.convertFields(l.fields)
	l.zapLogger.Fatal(msg, zapFields...)
}
