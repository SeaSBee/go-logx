package integration

import (
	"errors"
	"fmt"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	logx "go-logx"
)

func TestHighConcurrencyLogging(t *testing.T) {
	config := logx.DefaultConfig()
	config.Level = logx.DebugLevel
	logger, err := logx.New(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	// Simulate 10,000 concurrent goroutines
	numGoroutines := 10000
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	start := time.Now()

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()

			// Log multiple messages per goroutine
			logger.Debug("debug message", logx.Int("goroutine_id", id), logx.String("level", "debug"))
			logger.Info("info message", logx.Int("goroutine_id", id), logx.String("level", "info"))
			logger.Warn("warn message", logx.Int("goroutine_id", id), logx.String("level", "warn"))
			logger.Error("error message", logx.Int("goroutine_id", id), logx.String("level", "error"))
		}(i)
	}

	wg.Wait()
	duration := time.Since(start)

	// Log performance metrics
	logger.Info("Concurrency test completed",
		logx.Int("num_goroutines", numGoroutines),
		logx.Int64("duration_ns", duration.Nanoseconds()),
		logx.Float64("duration_ms", float64(duration.Nanoseconds())/1e6),
		logx.String("test_type", "high_concurrency"),
	)

	// Verify that the test completed without panicking
	if duration > 30*time.Second {
		t.Errorf("Test took too long: %v", duration)
	}
}

func TestConcurrentSensitiveDataMasking(t *testing.T) {
	config := logx.DefaultConfig()
	logger, err := logx.New(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	// Test concurrent sensitive data logging
	numGoroutines := 1000
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()

			// Log sensitive data concurrently
			logger.Info("User login attempt",
				logx.String("username", "user123"),
				logx.String("password", "secretpassword123"),
				logx.String("email", "user123@example.com"),
				logx.String("ssn", "123-45-6789"),
				logx.String("token", "jwt_token_here"),
				logx.Int("user_id", id),
			)
		}(i)
	}

	wg.Wait()

	// Test should complete without panicking
	logger.Info("Concurrent sensitive data masking test completed")
}

func TestLoggerPerformanceUnderLoad(t *testing.T) {
	config := logx.DefaultConfig()
	config.Level = logx.DebugLevel
	logger, err := logx.New(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	// Simulate high-frequency logging
	numMessages := 10000
	start := time.Now()

	for i := 0; i < numMessages; i++ {
		logger.Info("Performance test message",
			logx.Int("message_id", i),
			logx.String("timestamp", time.Now().Format(time.RFC3339Nano)),
			logx.Float64("random_value", float64(i)/100.0),
		)
	}

	duration := time.Since(start)
	rate := 0.0
	if duration.Seconds() > 0 {
		rate = float64(numMessages) / duration.Seconds()
	}

	logger.Info("Performance test completed",
		logx.Int("num_messages", numMessages),
		logx.Int64("duration_ns", duration.Nanoseconds()),
		logx.Float64("messages_per_second", rate),
		logx.String("test_type", "performance"),
	)

	// Verify reasonable performance (should handle at least 1000 msg/sec)
	if rate < 1000 {
		t.Errorf("Performance too slow: %.2f messages/sec", rate)
	}
}

func TestConcurrentLoggerCreation(t *testing.T) {
	// Test creating multiple loggers concurrently
	numLoggers := 100
	var wg sync.WaitGroup
	wg.Add(numLoggers)

	loggers := make([]*logx.Logger, numLoggers)
	errors := make([]error, numLoggers)

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

	// Check that all loggers were created successfully
	for i, err := range errors {
		if err != nil {
			t.Errorf("Failed to create logger %d: %v", i, err)
		}
		if loggers[i] == nil {
			t.Errorf("Logger %d is nil", i)
		}
	}

	// Test that all loggers work
	for i, logger := range loggers {
		if logger != nil {
			logger.Info("Logger test", logx.Int("logger_id", i))
		}
	}
}

func TestConcurrentFieldOperations(t *testing.T) {
	config := logx.DefaultConfig()
	logger, err := logx.New(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	// Test concurrent field creation and logging
	numOperations := 1000
	var wg sync.WaitGroup
	wg.Add(numOperations)

	for i := 0; i < numOperations; i++ {
		go func(id int) {
			defer wg.Done()

			// Create fields concurrently
			fields := []logx.Field{
				logx.String("operation_id", "op123"),
				logx.Int("user_id", id),
				logx.Float64("score", float64(id)/100.0),
				logx.Bool("success", id%2 == 0),
				logx.Int64("timestamp", time.Now().UnixNano()),
			}

			// Log with fields
			logger.Info("Concurrent field operation", fields...)
		}(i)
	}

	wg.Wait()

	logger.Info("Concurrent field operations test completed")
}

// TestIntegrationEdgeCases tests edge cases in integration scenarios
func TestIntegrationEdgeCases(t *testing.T) {
	t.Run("High Load with Edge Cases", testHighLoadEdgeCases)
	t.Run("Memory Pressure Edge Cases", testMemoryPressureEdgeCases)
	t.Run("Concurrency Edge Cases", testConcurrencyEdgeCases)
	t.Run("Configuration Edge Cases", testConfigurationEdgeCases)
	t.Run("Sensitive Data Edge Cases", testSensitiveDataEdgeCases)
	t.Run("Error Handling Edge Cases", testErrorHandlingEdgeCases)
	t.Run("Performance Edge Cases", testPerformanceEdgeCases)
}

// testHighLoadEdgeCases tests edge cases under high load
func testHighLoadEdgeCases(t *testing.T) {
	config := logx.DefaultConfig()
	config.Level = logx.DebugLevel
	logger, err := logx.New(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Sync()

	t.Run("Extreme Concurrency with Edge Cases", func(t *testing.T) {
		numGoroutines := 5000
		messagesPerGoroutine := 20
		var wg sync.WaitGroup
		wg.Add(numGoroutines)

		start := time.Now()

		for i := 0; i < numGoroutines; i++ {
			go func(id int) {
				defer wg.Done()

				for j := 0; j < messagesPerGoroutine; j++ {
					// Test various edge cases in high concurrency
					switch j % 7 {
					case 0:
						// Empty message
						logger.Info("")
					case 1:
						// Very long message
						longMsg := strings.Repeat("very long message content ", 100)
						logger.Info(longMsg, logx.Int("goroutine_id", id))
					case 2:
						// Many fields
						fields := make([]logx.Field, 50)
						for k := range fields {
							fields[k] = logx.String(fmt.Sprintf("field_%d", k), fmt.Sprintf("value_%d", k))
						}
						logger.Info("Many fields test", fields...)
					case 3:
						// Sensitive data
						logger.Info("Sensitive data test",
							logx.String("password", "secret123"),
							logx.String("email", "user@example.com"),
							logx.String("token", "jwt_token_here"),
						)
					case 4:
						// Complex nested data
						logger.Info("Complex data test",
							logx.Any("nested", map[string]interface{}{
								"level1": map[string]interface{}{
									"level2": map[string]string{"key": "value"},
									"array":  []int{1, 2, 3, 4, 5},
								},
								"simple": "value",
							}),
						)
					case 5:
						// Error logging
						logger.Error("Error test", logx.ErrorField(errors.New("test error")))
					case 6:
						// Formatted logging
						logger.Infof("Formatted test: %s, %d, %f", "string", 42, 3.14)
					}
				}
			}(i)
		}

		wg.Wait()
		duration := time.Since(start)
		totalMessages := numGoroutines * messagesPerGoroutine
		messagesPerSecond := 0.0
		if duration.Seconds() > 0 {
			messagesPerSecond = float64(totalMessages) / duration.Seconds()
		}

		t.Logf("High load edge cases completed: %d messages in %v (%.2f msg/sec)",
			totalMessages, duration, messagesPerSecond)
	})

	t.Run("Mixed Operations Under Load", func(t *testing.T) {
		numGoroutines := 1000
		var wg sync.WaitGroup
		wg.Add(numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			go func(id int) {
				defer wg.Done()

				// Create child logger with context
				childLogger := logger.With(
					logx.Int("goroutine_id", id),
					logx.String("timestamp", time.Now().String()),
				)

				// Perform mixed operations
				childLogger.Debug("Debug message")
				childLogger.Info("Info message")
				childLogger.Warn("Warning message")
				childLogger.Error("Error message")

				// Test formatted logging
				childLogger.Infof("Formatted: %s", "value")
				childLogger.Errorf("Error formatted: %v", errors.New("test error"))

				// Test sensitive data
				childLogger.Info("Sensitive", logx.String("password", "secret"))

				// Test complex data
				childLogger.Info("Complex", logx.Any("data", map[string]interface{}{
					"nested": "value",
					"array":  []int{1, 2, 3},
				}))
			}(i)
		}

		wg.Wait()
		t.Logf("Mixed operations under load completed: %d goroutines", numGoroutines)
	})
}

// testMemoryPressureEdgeCases tests edge cases under memory pressure
func testMemoryPressureEdgeCases(t *testing.T) {
	config := logx.DefaultConfig()
	config.Level = logx.InfoLevel
	logger, err := logx.New(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Sync()

	t.Run("Large Data Structures", func(t *testing.T) {
		// Test with very large data structures
		largeMap := make(map[string]interface{})
		for i := 0; i < 1000; i++ {
			largeMap[fmt.Sprintf("key_%d", i)] = strings.Repeat("value_", 100)
		}

		logger.Info("Large map test", logx.Any("large_map", largeMap))

		// Test with large slices
		largeSlice := make([]string, 10000)
		for i := range largeSlice {
			largeSlice[i] = fmt.Sprintf("item_%d", i)
		}

		logger.Info("Large slice test", logx.Any("large_slice", largeSlice))
	})

	t.Run("Memory Leak Prevention", func(t *testing.T) {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		initialAlloc := m.Alloc

		// Create many loggers and use them
		for i := 0; i < 1000; i++ {
			config := logx.DefaultConfig()
			config.Level = logx.InfoLevel

			tempLogger, err := logx.New(config)
			if err != nil {
				t.Fatalf("Failed to create temp logger: %v", err)
			}

			tempLogger.Info("Memory leak test", logx.Int("iteration", i))
			tempLogger.Sync()
		}

		// Force garbage collection
		runtime.GC()

		runtime.ReadMemStats(&m)
		finalAlloc := m.Alloc
		memoryIncrease := finalAlloc - initialAlloc

		t.Logf("Memory leak test completed. Memory increase: %d bytes", memoryIncrease)

		if memoryIncrease > 10*1024*1024 { // 10MB threshold
			t.Logf("⚠️ High memory increase detected: %d bytes", memoryIncrease)
		}
	})

	t.Run("String Interning Pressure", func(t *testing.T) {
		// Test with many unique strings to pressure string interning
		for i := 0; i < 10000; i++ {
			uniqueString := fmt.Sprintf("unique_string_%d_%s", i, time.Now().String())
			logger.Info("String interning test", logx.String("unique", uniqueString))
		}
	})
}

// testConcurrencyEdgeCases tests edge cases in concurrent operations
func testConcurrencyEdgeCases(t *testing.T) {
	config := logx.DefaultConfig()
	config.Level = logx.DebugLevel
	logger, err := logx.New(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Sync()

	t.Run("Rapid Logger Creation", func(t *testing.T) {
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

		// Test all created loggers
		successCount := 0
		for i, logger := range loggers {
			if errors[i] != nil {
				t.Errorf("Failed to create logger %d: %v", i, errors[i])
				continue
			}
			if logger != nil {
				logger.Info("Rapid creation test", logx.Int("logger_id", i))
				logger.Sync()
				successCount++
			}
		}

		t.Logf("Rapid logger creation completed: %d/%d successful in %v",
			successCount, numLoggers, duration)
	})

	t.Run("Concurrent Sensitive Key Operations", func(t *testing.T) {
		numOperations := 1000
		var wg sync.WaitGroup
		wg.Add(numOperations)

		for i := 0; i < numOperations; i++ {
			go func(id int) {
				defer wg.Done()

				key := fmt.Sprintf("concurrent_key_%d", id)

				// Add sensitive key
				logx.AddSensitiveKey(key)

				// Use the key in logging
				logger.Info("Concurrent sensitive key test",
					logx.String(key, "sensitive_value"),
					logx.String("normal_key", "normal_value"),
				)

				// Remove the key
				logx.RemoveSensitiveKey(key)
			}(i)
		}

		wg.Wait()
		t.Logf("Concurrent sensitive key operations completed: %d operations", numOperations)
	})

	t.Run("Logger Chain Stress", func(t *testing.T) {
		numChains := 500
		var wg sync.WaitGroup
		wg.Add(numChains)

		for i := 0; i < numChains; i++ {
			go func(id int) {
				defer wg.Done()

				// Create a chain of loggers
				level1 := logger.With(logx.Int("level", 1), logx.Int("chain_id", id))
				level2 := level1.With(logx.String("context", "level2"))
				level3 := level2.With(logx.Bool("final", true))

				// Use all levels
				level1.Info("Level 1 message")
				level2.Info("Level 2 message")
				level3.Info("Level 3 message")
			}(i)
		}

		wg.Wait()
		t.Logf("Logger chain stress completed: %d chains", numChains)
	})
}

// testConfigurationEdgeCases tests edge cases in configuration
func testConfigurationEdgeCases(t *testing.T) {
	t.Run("All Configuration Combinations", func(t *testing.T) {
		// Test all possible configuration combinations
		configs := []*logx.Config{
			{Level: logx.TraceLevel, Development: true, AddCaller: true, AddStacktrace: true},
			{Level: logx.DebugLevel, Development: false, AddCaller: true, AddStacktrace: false},
			{Level: logx.InfoLevel, Development: true, AddCaller: false, AddStacktrace: true},
			{Level: logx.WarnLevel, Development: false, AddCaller: false, AddStacktrace: false},
			{Level: logx.ErrorLevel, Development: true, AddCaller: true, AddStacktrace: false},
			{Level: logx.FatalLevel, Development: false, AddCaller: true, AddStacktrace: true},
		}

		for i, config := range configs {
			logger, err := logx.New(config)
			if err != nil {
				t.Fatalf("Failed to create logger with config %d: %v", i, err)
			}
			defer logger.Sync()

			// Test all log levels
			logger.Trace("Trace message")
			logger.Debug("Debug message")
			logger.Info("Info message")
			logger.Warn("Warning message")
			logger.Error("Error message")

			// Test formatted logging
			logger.Tracef("Trace formatted: %s", "value")
			logger.Debugf("Debug formatted: %d", 42)
			logger.Infof("Info formatted: %f", 3.14)
			logger.Warnf("Warning formatted: %t", true)
			logger.Errorf("Error formatted: %v", errors.New("test error"))

			t.Logf("Configuration %d tested successfully", i)
		}
	})

	t.Run("Configuration Under Load", func(t *testing.T) {
		numConfigs := 100
		var wg sync.WaitGroup
		wg.Add(numConfigs)

		for i := 0; i < numConfigs; i++ {
			go func(id int) {
				defer wg.Done()

				config := logx.DefaultConfig()
				config.Level = logx.Level(id % 6) // Cycle through all levels
				config.Development = id%2 == 0
				config.AddCaller = id%3 == 0
				config.AddStacktrace = id%4 == 0

				logger, err := logx.New(config)
				if err != nil {
					t.Errorf("Failed to create logger with config %d: %v", id, err)
					return
				}
				defer logger.Sync()

				logger.Info("Configuration under load test", logx.Int("config_id", id))
			}(i)
		}

		wg.Wait()
		t.Logf("Configuration under load completed: %d configurations", numConfigs)
	})
}

// testSensitiveDataEdgeCases tests edge cases in sensitive data handling
func testSensitiveDataEdgeCases(t *testing.T) {
	config := logx.DefaultConfig()
	config.Level = logx.InfoLevel
	logger, err := logx.New(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Sync()

	t.Run("All Sensitive Key Types", func(t *testing.T) {
		// Test all built-in sensitive keys
		sensitiveKeys := []string{
			"password", "passwd", "pass", "ssn", "token", "apikey", "api_key",
			"secret", "key", "email", "phone", "credit_card", "cc", "cvv",
			"pin", "auth", "authorization", "bearer", "jwt",
		}

		for _, key := range sensitiveKeys {
			logger.Info("Sensitive key test",
				logx.String(key, "sensitive_value"),
				logx.String("normal_key", "normal_value"),
			)
		}
	})

	t.Run("Case Sensitivity", func(t *testing.T) {
		// Test case sensitivity in sensitive keys
		variations := []string{
			"PASSWORD", "Password", "password", "pAsSwOrD",
			"EMAIL", "Email", "email", "eMaIl",
			"TOKEN", "Token", "token", "tOkEn",
		}

		for _, key := range variations {
			logger.Info("Case sensitivity test",
				logx.String(key, "sensitive_value"),
			)
		}
	})

	t.Run("Dynamic Sensitive Keys", func(t *testing.T) {
		// Test dynamic addition and removal of sensitive keys
		for i := 0; i < 100; i++ {
			key := fmt.Sprintf("dynamic_key_%d", i)

			// Add key
			logx.AddSensitiveKey(key)

			// Log with the key
			logger.Info("Dynamic key test",
				logx.String(key, "sensitive_value"),
				logx.String("normal_key", "normal_value"),
			)

			// Remove key
			logx.RemoveSensitiveKey(key)

			// Log again (should not be masked)
			logger.Info("After removal test",
				logx.String(key, "sensitive_value"),
			)
		}
	})

	t.Run("Sensitive Data Types", func(t *testing.T) {
		// Test different data types for sensitive fields
		logger.Info("Sensitive data types test",
			logx.String("password", "string_password"),
			logx.Int("password", 12345),
			logx.Float64("token", 3.14159),
			logx.Bool("secret", true),
			logx.Any("key", []int{1, 2, 3}),
			logx.Any("auth", map[string]string{"type": "bearer"}),
		)
	})

	t.Run("Concurrent Sensitive Key Management", func(t *testing.T) {
		numOperations := 500
		var wg sync.WaitGroup
		wg.Add(numOperations)

		for i := 0; i < numOperations; i++ {
			go func(id int) {
				defer wg.Done()

				key := fmt.Sprintf("concurrent_sensitive_%d", id)

				// Add key
				logx.AddSensitiveKey(key)

				// Use in multiple loggers
				for j := 0; j < 10; j++ {
					logger.Info("Concurrent sensitive test",
						logx.String(key, fmt.Sprintf("value_%d", j)),
					)
				}

				// Remove key
				logx.RemoveSensitiveKey(key)
			}(i)
		}

		wg.Wait()
		t.Logf("Concurrent sensitive key management completed: %d operations", numOperations)
	})
}

// testErrorHandlingEdgeCases tests edge cases in error handling
func testErrorHandlingEdgeCases(t *testing.T) {
	config := logx.DefaultConfig()
	config.Level = logx.ErrorLevel
	logger, err := logx.New(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Sync()

	t.Run("Various Error Types", func(t *testing.T) {
		// Test different types of errors
		errors := []error{
			errors.New("simple error"),
			errors.New(""),
			nil,
			fmt.Errorf("formatted error: %s", "value"),
		}

		for i, err := range errors {
			logger.Error("Error type test", logx.ErrorField(err), logx.Int("error_id", i))
		}
	})

	t.Run("Large Error Messages", func(t *testing.T) {
		// Test with very large error messages
		largeErrorMsg := strings.Repeat("large error message content ", 1000)
		largeError := errors.New(largeErrorMsg)

		logger.Error("Large error test", logx.ErrorField(largeError))
	})

	t.Run("Error with Context", func(t *testing.T) {
		// Test errors with additional context
		err := errors.New("context error")
		logger.Error("Error with context",
			logx.ErrorField(err),
			logx.String("context", "additional information"),
			logx.Int("error_code", 500),
			logx.String("user_id", "user123"),
		)
	})

	t.Run("Concurrent Error Logging", func(t *testing.T) {
		numErrors := 1000
		var wg sync.WaitGroup
		wg.Add(numErrors)

		for i := 0; i < numErrors; i++ {
			go func(id int) {
				defer wg.Done()

				err := fmt.Errorf("concurrent error %d", id)
				logger.Error("Concurrent error test",
					logx.ErrorField(err),
					logx.Int("error_id", id),
				)
			}(i)
		}

		wg.Wait()
		t.Logf("Concurrent error logging completed: %d errors", numErrors)
	})
}

// testPerformanceEdgeCases tests edge cases in performance scenarios
func testPerformanceEdgeCases(t *testing.T) {
	config := logx.DefaultConfig()
	config.Level = logx.InfoLevel
	logger, err := logx.New(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Sync()

	t.Run("High Frequency with Edge Cases", func(t *testing.T) {
		numMessages := 50000
		start := time.Now()

		for i := 0; i < numMessages; i++ {
			// Vary the message type to test different code paths
			switch i % 8 {
			case 0:
				logger.Info("High frequency test", logx.Int("iteration", i))
			case 1:
				logger.Infof("High frequency formatted: %d", i)
			case 2:
				logger.Info("High frequency with sensitive data",
					logx.String("password", "secret123"),
					logx.Int("iteration", i),
				)
			case 3:
				logger.Info("High frequency with complex data",
					logx.Any("data", map[string]interface{}{
						"iteration": i,
						"nested":    map[string]int{"value": i * 2},
					}),
				)
			case 4:
				logger.Error("High frequency error", logx.ErrorField(errors.New("test error")))
			case 5:
				logger.Info("High frequency with many fields",
					logx.String("field1", "value1"),
					logx.Int("field2", i),
					logx.Float64("field3", float64(i)/100.0),
					logx.Bool("field4", i%2 == 0),
				)
			case 6:
				logger.Warn("High frequency warning", logx.Int("iteration", i))
			case 7:
				logger.Debug("High frequency debug", logx.Int("iteration", i))
			}
		}

		duration := time.Since(start)
		messagesPerSecond := 0.0
		if duration.Seconds() > 0 {
			messagesPerSecond = float64(numMessages) / duration.Seconds()
		}

		t.Logf("High frequency edge cases completed: %d messages in %v (%.2f msg/sec)",
			numMessages, duration, messagesPerSecond)
	})

	t.Run("Memory Usage Under Load", func(t *testing.T) {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		initialAlloc := m.Alloc

		// Perform intensive logging operations
		for i := 0; i < 10000; i++ {
			largeData := strings.Repeat("memory pressure test data ", 50)
			logger.Info("Memory usage test",
				logx.String("large_data", largeData),
				logx.Int("iteration", i),
				logx.Any("complex_data", map[string]interface{}{
					"nested": map[string]string{
						"key1": "value1",
						"key2": "value2",
						"key3": "value3",
					},
					"array": []int{1, 2, 3, 4, 5},
				}),
			)
		}

		// Force garbage collection
		runtime.GC()

		runtime.ReadMemStats(&m)
		finalAlloc := m.Alloc
		memoryIncrease := finalAlloc - initialAlloc

		t.Logf("Memory usage under load completed. Memory increase: %d bytes", memoryIncrease)

		if memoryIncrease > 50*1024*1024 { // 50MB threshold
			t.Logf("⚠️ High memory increase detected: %d bytes", memoryIncrease)
		}
	})

	t.Run("CPU Intensive Operations", func(t *testing.T) {
		start := time.Now()

		// Perform CPU-intensive logging operations
		for i := 0; i < 10000; i++ {
			// Create complex data structures
			complexData := make(map[string]interface{})
			for j := 0; j < 100; j++ {
				complexData[fmt.Sprintf("key_%d", j)] = fmt.Sprintf("value_%d_%d", i, j)
			}

			logger.Info("CPU intensive test",
				logx.Any("complex_data", complexData),
				logx.Int("iteration", i),
				logx.String("timestamp", time.Now().String()),
			)
		}

		duration := time.Since(start)
		t.Logf("CPU intensive operations completed in %v", duration)
	})
}
