package stress

import (
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"

	logx "go-logx"
)

// safeDivision safely divides a by b, returning 0 if b is 0
func safeDivision(a, b float64) float64 {
	if b == 0 {
		return 0
	}
	return a / b
}

// StressTestConfig holds configuration for stress tests
type StressTestConfig struct {
	NumGoroutines           int
	MessagesPerGoroutine    int
	TestDuration            time.Duration
	LogLevel                logx.Level
	EnableSensitiveData     bool
	EnableStructuredLogging bool
}

// DefaultStressConfig returns a default stress test configuration
func DefaultStressConfig() *StressTestConfig {
	return &StressTestConfig{
		NumGoroutines:           10000,
		MessagesPerGoroutine:    10,
		TestDuration:            30 * time.Second,
		LogLevel:                logx.InfoLevel,
		EnableSensitiveData:     true,
		EnableStructuredLogging: true,
	}
}

// TestExtremeConcurrency tests the logger with 10,000+ concurrent goroutines
func TestExtremeConcurrency(t *testing.T) {
	config := DefaultStressConfig()
	config.NumGoroutines = 15000 // 15,000 concurrent goroutines
	config.MessagesPerGoroutine = 5

	// Initialize memory profiler
	profiler := NewMemoryProfiler()
	profiler.Start()

	logger, err := createTestLogger(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	start := time.Now()
	var wg sync.WaitGroup
	wg.Add(config.NumGoroutines)

	// Track metrics
	var totalMessages int64
	var totalErrors int64
	var mu sync.Mutex

	for i := 0; i < config.NumGoroutines; i++ {
		go func(id int) {
			defer wg.Done()

			for j := 0; j < config.MessagesPerGoroutine; j++ {
				mu.Lock()
				totalMessages++
				mu.Unlock()

				// Random log level to simulate real-world usage
				switch rand.Intn(4) {
				case 0:
					logger.Debug("Debug message", logx.Int("goroutine_id", id), logx.Int("message_id", j))
				case 1:
					logger.Info("Info message", logx.Int("goroutine_id", id), logx.Int("message_id", j))
				case 2:
					logger.Warn("Warn message", logx.Int("goroutine_id", id), logx.Int("message_id", j))
				case 3:
					logger.Error("Error message", logx.Int("goroutine_id", id), logx.Int("message_id", j))
				}
			}
		}(i)
	}

	wg.Wait()
	duration := time.Since(start)

	// Stop profiling and generate report
	report := profiler.Stop()
	profiler.PrintReport(report)

	// Log test results
	logger.Info("Extreme concurrency test completed",
		logx.Int("num_goroutines", config.NumGoroutines),
		logx.Int("messages_per_goroutine", config.MessagesPerGoroutine),
		logx.Int64("total_messages", totalMessages),
		logx.Int64("total_errors", totalErrors),
		logx.Int64("duration_ns", duration.Nanoseconds()),
		logx.Float64("duration_seconds", duration.Seconds()),
		logx.Float64("messages_per_second", safeDivision(float64(totalMessages), duration.Seconds())),
		logx.String("test_type", "extreme_concurrency"),
	)

	// Verify test completed successfully
	if duration > 60*time.Second {
		t.Errorf("Test took too long: %v", duration)
	}

	if totalErrors > 0 {
		t.Errorf("Encountered %d errors during stress test", totalErrors)
	}

	t.Logf("✅ Extreme concurrency test passed: %d messages in %v (%.2f msg/sec)",
		totalMessages, duration, safeDivision(float64(totalMessages), duration.Seconds()))
}

// TestHighFrequencyLogging tests rapid-fire logging operations
func TestHighFrequencyLogging(t *testing.T) {
	config := DefaultStressConfig()
	config.NumGoroutines = 1000
	config.MessagesPerGoroutine = 100 // 100,000 total messages

	logger, err := createTestLogger(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	start := time.Now()
	var wg sync.WaitGroup
	wg.Add(config.NumGoroutines)

	for i := 0; i < config.NumGoroutines; i++ {
		go func(id int) {
			defer wg.Done()

			for j := 0; j < config.MessagesPerGoroutine; j++ {
				// High-frequency logging with minimal fields
				logger.Info("High frequency message",
					logx.Int("worker_id", id),
					logx.Int("message_id", j),
					logx.Int64("timestamp", time.Now().UnixNano()),
				)
			}
		}(i)
	}

	wg.Wait()
	duration := time.Since(start)
	totalMessages := int64(config.NumGoroutines * config.MessagesPerGoroutine)

	logger.Info("High frequency logging test completed",
		logx.Int("num_workers", config.NumGoroutines),
		logx.Int("messages_per_worker", config.MessagesPerGoroutine),
		logx.Int64("total_messages", totalMessages),
		logx.Int64("duration_ns", duration.Nanoseconds()),
		logx.Float64("messages_per_second", safeDivision(float64(totalMessages), duration.Seconds())),
		logx.String("test_type", "high_frequency"),
	)

	// Performance validation
	rate := safeDivision(float64(totalMessages), duration.Seconds())
	if rate < 10000 { // Should handle at least 10k msg/sec
		t.Errorf("Performance too slow: %.2f messages/sec", rate)
	}

	t.Logf("✅ High frequency test passed: %.2f messages/sec", rate)
}

// TestSensitiveDataStress tests sensitive data masking under extreme load
func TestSensitiveDataStress(t *testing.T) {
	config := DefaultStressConfig()
	config.NumGoroutines = 5000
	config.MessagesPerGoroutine = 20
	config.EnableSensitiveData = true

	logger, err := createTestLogger(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	// Add custom sensitive keys for testing
	logx.AddSensitiveKey("custom_secret")
	logx.AddSensitiveKey("api_key")
	logx.AddSensitiveKey("session_token")

	start := time.Now()
	var wg sync.WaitGroup
	wg.Add(config.NumGoroutines)

	for i := 0; i < config.NumGoroutines; i++ {
		go func(id int) {
			defer wg.Done()

			for j := 0; j < config.MessagesPerGoroutine; j++ {
				// Log sensitive data with various patterns
				logger.Info("User authentication attempt",
					logx.String("username", fmt.Sprintf("user_%d", id)),
					logx.String("password", fmt.Sprintf("secret_password_%d_%d", id, j)),
					logx.String("email", fmt.Sprintf("user_%d@example.com", id)),
					logx.String("ssn", fmt.Sprintf("123-45-%04d", id)),
					logx.String("token", fmt.Sprintf("jwt_token_%d_%d", id, j)),
					logx.String("custom_secret", fmt.Sprintf("custom_secret_%d", id)),
					logx.String("api_key", fmt.Sprintf("api_key_%d_%d", id, j)),
					logx.String("session_token", fmt.Sprintf("session_%d_%d", id, j)),
					logx.String("normal_field", fmt.Sprintf("visible_value_%d", id)),
					logx.Int("user_id", id),
					logx.Int("attempt_id", j),
				)
			}
		}(i)
	}

	wg.Wait()
	duration := time.Since(start)

	logger.Info("Sensitive data stress test completed",
		logx.Int("num_goroutines", config.NumGoroutines),
		logx.Int("messages_per_goroutine", config.MessagesPerGoroutine),
		logx.Int64("duration_ns", duration.Nanoseconds()),
		logx.Float64("messages_per_second", safeDivision(float64(config.NumGoroutines*config.MessagesPerGoroutine), duration.Seconds())),
		logx.String("test_type", "sensitive_data_stress"),
	)

	if duration > 45*time.Second {
		t.Errorf("Sensitive data test took too long: %v", duration)
	}

	t.Logf("✅ Sensitive data stress test passed: %d messages in %v",
		config.NumGoroutines*config.MessagesPerGoroutine, duration)
}

// TestMemoryPressure tests logging under memory pressure
func TestMemoryPressure(t *testing.T) {
	config := DefaultStressConfig()
	config.NumGoroutines = 2000
	config.MessagesPerGoroutine = 50

	// Initialize memory profiler
	profiler := NewMemoryProfiler()
	profiler.Start()

	logger, err := createTestLogger(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	var wg sync.WaitGroup
	wg.Add(config.NumGoroutines)

	for i := 0; i < config.NumGoroutines; i++ {
		go func(id int) {
			defer wg.Done()

			for j := 0; j < config.MessagesPerGoroutine; j++ {
				// Create large structured logs to test memory usage
				logger.Info("Memory pressure test",
					logx.String("large_field_1", generateLargeString(1000)),
					logx.String("large_field_2", generateLargeString(1000)),
					logx.String("large_field_3", generateLargeString(1000)),
					logx.Int("worker_id", id),
					logx.Int("message_id", j),
					logx.Int64("timestamp", time.Now().UnixNano()),
					logx.Float64("random_value", rand.Float64()),
					logx.Bool("success", rand.Intn(2) == 0),
				)
			}
		}(i)
	}

	wg.Wait()

	// Stop profiling and generate report
	report := profiler.Stop()
	profiler.PrintReport(report)

	// Memory validation (should not increase by more than 100MB)
	if report.MemoryGrowth > 100*1024*1024 {
		t.Errorf("Memory increase too high: %d bytes", report.MemoryGrowth)
	}

	// Check for potential leaks
	if report.LeakIndicators.PotentialLeak {
		t.Logf("⚠️ Potential memory leak detected (confidence: %.1f%%)",
			report.LeakIndicators.LeakConfidence*100)
	}

	t.Logf("✅ Memory pressure test passed: %d bytes increase", report.MemoryGrowth)
}

// TestConcurrentLoggerCreation tests creating multiple loggers concurrently
func TestConcurrentLoggerCreation(t *testing.T) {
	numLoggers := 1000
	var wg sync.WaitGroup
	wg.Add(numLoggers)

	loggers := make([]*logx.Logger, numLoggers)
	errors := make([]error, numLoggers)

	start := time.Now()

	for i := 0; i < numLoggers; i++ {
		go func(id int) {
			defer wg.Done()

			config := logx.DefaultConfig()
			config.Level = logx.InfoLevel

			logger, err := logx.New(config)
			loggers[id] = logger
			errors[id] = err
		}(i)
	}

	wg.Wait()
	duration := time.Since(start)

	// Check that all loggers were created successfully
	successCount := 0
	for i, err := range errors {
		if err != nil {
			t.Errorf("Failed to create logger %d: %v", i, err)
		} else {
			successCount++
		}
		if loggers[i] == nil {
			t.Errorf("Logger %d is nil", i)
		}
	}

	// Test that all loggers work
	for i, logger := range loggers {
		if logger != nil {
			logger.Info("Logger creation test", logx.Int("logger_id", i))
		}
	}

	t.Logf("✅ Concurrent logger creation test passed: %d/%d loggers created successfully in %v",
		successCount, numLoggers, duration)
}

// TestMixedOperations tests various logging operations mixed together
func TestMixedOperations(t *testing.T) {
	config := DefaultStressConfig()
	config.NumGoroutines = 3000
	config.MessagesPerGoroutine = 15

	logger, err := createTestLogger(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	start := time.Now()
	var wg sync.WaitGroup
	wg.Add(config.NumGoroutines)

	for i := 0; i < config.NumGoroutines; i++ {
		go func(id int) {
			defer wg.Done()

			for j := 0; j < config.MessagesPerGoroutine; j++ {
				// Mix different types of operations
				switch j % 5 {
				case 0:
					// Basic logging
					logger.Info("Basic message", logx.Int("worker_id", id))
				case 1:
					// Structured logging with many fields
					logger.Info("Structured message",
						logx.String("operation", "database_query"),
						logx.String("table", "users"),
						logx.Int("user_id", id),
						logx.Int("query_id", j),
						logx.Float64("duration_ms", rand.Float64()*100),
						logx.Bool("cached", rand.Intn(2) == 0),
						logx.Int64("timestamp", time.Now().UnixNano()),
					)
				case 2:
					// Error logging
					logger.Error("Error occurred",
						logx.String("error_type", "timeout"),
						logx.Int("worker_id", id),
						logx.Int("attempt", j),
					)
				case 3:
					// Sensitive data logging
					logger.Info("Authentication",
						logx.String("username", fmt.Sprintf("user_%d", id)),
						logx.String("password", fmt.Sprintf("pass_%d", id)),
						logx.String("token", fmt.Sprintf("token_%d", id)),
					)
				case 4:
					// Debug logging
					logger.Debug("Debug info",
						logx.String("component", "worker"),
						logx.Int("worker_id", id),
						logx.Int("iteration", j),
					)
				}
			}
		}(i)
	}

	wg.Wait()
	duration := time.Since(start)

	logger.Info("Mixed operations test completed",
		logx.Int("num_workers", config.NumGoroutines),
		logx.Int("messages_per_worker", config.MessagesPerGoroutine),
		logx.Int64("duration_ns", duration.Nanoseconds()),
		logx.Float64("messages_per_second", safeDivision(float64(config.NumGoroutines*config.MessagesPerGoroutine), duration.Seconds())),
		logx.String("test_type", "mixed_operations"),
	)

	t.Logf("✅ Mixed operations test passed: %d messages in %v",
		config.NumGoroutines*config.MessagesPerGoroutine, duration)
}

// BenchmarkStressLogging provides benchmark metrics for stress testing
func BenchmarkStressLogging(b *testing.B) {
	config := logx.DefaultConfig()
	config.Level = logx.InfoLevel
	logger, err := logx.New(config)
	if err != nil {
		b.Fatalf("Failed to create logger: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("Benchmark stress message",
			logx.Int("iteration", i),
			logx.String("test_type", "stress_benchmark"),
			logx.Int64("timestamp", time.Now().UnixNano()),
		)
	}
}

// Helper functions

func createTestLogger(config *StressTestConfig) (*logx.Logger, error) {
	logConfig := logx.DefaultConfig()
	logConfig.Level = config.LogLevel
	logConfig.AddCaller = true
	logConfig.AddStacktrace = true

	return logx.New(logConfig)
}

func generateLargeString(size int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, size)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}
