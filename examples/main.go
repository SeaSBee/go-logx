package main

import (
	"errors"
	logx "go-logx"
	"sync"
	"time"
)

func main() {
	// Example 1: Basic usage with default configuration
	logx.InitDefault()

	logx.Info("Application started")
	logx.Debug("Debug information", logx.String("component", "main"))
	logx.Warn("Warning message", logx.Int("warning_code", 1001))
	logx.Error("Error occurred", logx.String("error_type", "validation"))

	// Example 2: Custom configuration
	customConfig := &logx.Config{
		Level:         logx.DebugLevel,
		Development:   true,
		AddCaller:     true,
		AddStacktrace: true,
	}

	logger, err := logx.New(customConfig)
	if err != nil {
		logx.Fatal("Failed to create custom logger", logx.ErrorField(err))
	}

	logger.Info("Custom logger initialized")
	logger.Debug("Debug message from custom logger")

	// Example 3: Structured logging with fields
	logx.Info("User action",
		logx.String("action", "login"),
		logx.String("user_id", "user123"),
		logx.String("ip_address", "192.168.1.1"),
		logx.Int("response_time_ms", 150),
		logx.Bool("success", true),
	)

	// Example 4: Sensitive data masking
	logx.Info("User login attempt",
		logx.String("username", "john.doe"),
		logx.String("password", "secretpassword123"), // This will be masked
		logx.String("email", "john.doe@example.com"), // This will be masked
		logx.String("ssn", "123-45-6789"),            // This will be masked
		logx.String("token", "jwt_token_here"),       // This will be masked
		logx.String("normal_field", "visible_value"), // This will not be masked
	)

	// Example 5: Error logging
	err = errors.New("database connection failed")
	logx.Error("Database operation failed",
		logx.ErrorField(err),
		logx.String("operation", "user_lookup"),
		logx.Int("retry_count", 3),
	)

	// Example 6: Logger with context
	userLogger := logx.With(
		logx.String("user_id", "user456"),
		logx.String("session_id", "sess_789"),
		logx.String("request_id", "req_123"),
	)

	userLogger.Info("User performed action", logx.String("action", "update_profile"))
	userLogger.Warn("User action warning", logx.String("warning", "profile_update_rate_limit"))

	// Example 7: Concurrent logging
	concurrentLoggingExample()

	// Example 8: Performance logging
	performanceLoggingExample()

	// Example 9: Custom sensitive keys
	logx.AddSensitiveKey("custom_secret")
	logx.Info("Custom sensitive data",
		logx.String("custom_secret", "very_secret_value"), // This will be masked
		logx.String("public_data", "visible_value"),       // This will not be masked
	)

	// Example 10: Different log levels
	logger.Debug("Debug message - only visible in debug mode")
	logger.Info("Info message - visible in info mode and above")
	logger.Warn("Warning message - visible in warn mode and above")
	logger.Error("Error message - visible in error mode and above")

	// Sync before exit
	logx.Sync()
}

func concurrentLoggingExample() {
	logx.Info("Starting concurrent logging example")

	var wg sync.WaitGroup
	numGoroutines := 10

	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()

			// Each goroutine logs with its own context
			goroutineLogger := logx.With(
				logx.Int("goroutine_id", id),
				logx.String("worker_type", "concurrent"),
			)

			for j := 0; j < 5; j++ {
				goroutineLogger.Info("Concurrent log message",
					logx.Int("message_number", j),
					logx.String("timestamp", time.Now().Format(time.RFC3339)),
				)
				time.Sleep(10 * time.Millisecond)
			}
		}(i)
	}

	wg.Wait()
	logx.Info("Concurrent logging example completed")
}

func performanceLoggingExample() {
	logx.Info("Starting performance logging example")

	start := time.Now()

	// Simulate some work
	time.Sleep(100 * time.Millisecond)

	// Log performance metrics
	duration := time.Since(start)
	logx.Info("Operation completed",
		logx.String("operation", "performance_test"),
		logx.Int64("duration_ns", duration.Nanoseconds()),
		logx.Float64("duration_ms", float64(duration.Nanoseconds())/1e6),
		logx.String("status", "success"),
	)

	// Log with different data types
	logx.Info("Data processing completed",
		logx.Int("records_processed", 1000),
		logx.Int64("total_bytes", 1024*1024),
		logx.Float64("processing_rate_mbps", 10.5),
		logx.Bool("cache_hit", true),
		logx.String("data_format", "JSON"),
	)

	logx.Info("Performance logging example completed")
}
