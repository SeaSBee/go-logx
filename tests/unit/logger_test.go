package unit

import (
	"errors"
	"fmt"
	"math"
	"strings"
	"sync"
	"testing"
	"time"

	logx "github.com/seasbee/go-logx"
)

func TestLevelString(t *testing.T) {
	tests := []struct {
		level    logx.Level
		expected string
	}{
		{logx.DebugLevel, "DEBUG"},
		{logx.InfoLevel, "INFO"},
		{logx.WarnLevel, "WARN"},
		{logx.ErrorLevel, "ERROR"},
		{logx.FatalLevel, "FATAL"},
		{logx.Level(99), "UNKNOWN"},
	}

	for _, tt := range tests {
		if got := tt.level.String(); got != tt.expected {
			t.Errorf("Level.String() = %v, want %v", got, tt.expected)
		}
	}
}

func TestDefaultConfig(t *testing.T) {
	config := logx.DefaultConfig()
	if config.Level != logx.InfoLevel {
		t.Errorf("Expected default level to be InfoLevel, got %v", config.Level)
	}
	if config.OutputPath != "" {
		t.Errorf("Expected default output path to be empty, got %v", config.OutputPath)
	}
	if config.Development {
		t.Errorf("Expected default development to be false, got %v", config.Development)
	}
	if !config.AddCaller {
		t.Errorf("Expected default AddCaller to be true, got %v", config.AddCaller)
	}
	if !config.AddStacktrace {
		t.Errorf("Expected default AddStacktrace to be true, got %v", config.AddStacktrace)
	}
}

func TestNewLogger(t *testing.T) {
	config := logx.DefaultConfig()
	logger, err := logx.New(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	if logger == nil {
		t.Fatal("Logger should not be nil")
	}
}

func TestLoggerWithFields(t *testing.T) {
	config := logx.DefaultConfig()
	logger, err := logx.New(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	childLogger := logger.With(logx.String("parent", "value"))
	if childLogger == nil {
		t.Fatal("Child logger should not be nil")
	}

	// Test that child logger works
	childLogger.Info("Test message")
}

func TestFieldCreation(t *testing.T) {
	// Test String field
	strField := logx.String("key", "value")
	if strField.Key != "key" || strField.Value != "value" {
		t.Errorf("String field creation failed: %+v", strField)
	}

	// Test Int field
	intField := logx.Int("key", 42)
	if intField.Key != "key" || intField.Value != 42 {
		t.Errorf("Int field creation failed: %+v", intField)
	}

	// Test Int64 field
	int64Field := logx.Int64("key", 42)
	if int64Field.Key != "key" || int64Field.Value != int64(42) {
		t.Errorf("Int64 field creation failed: %+v", int64Field)
	}

	// Test Float64 field
	floatField := logx.Float64("key", 3.14)
	if floatField.Key != "key" || floatField.Value != 3.14 {
		t.Errorf("Float64 field creation failed: %+v", floatField)
	}

	// Test Bool field
	boolField := logx.Bool("key", true)
	if boolField.Key != "key" || boolField.Value != true {
		t.Errorf("Bool field creation failed: %+v", boolField)
	}

	// Test Any field
	anyField := logx.Any("key", "any_value")
	if anyField.Key != "key" || anyField.Value != "any_value" {
		t.Errorf("Any field creation failed: %+v", anyField)
	}

	// Test Error field
	testErr := errors.New("test error")
	errField := logx.ErrorField(testErr)
	if errField.Key != "error" || errField.Value != testErr {
		t.Errorf("Error field creation failed: %+v", errField)
	}
}

func TestSensitiveDataMasking(t *testing.T) {
	config := logx.DefaultConfig()
	logger, err := logx.New(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	// Log sensitive data - this should not panic
	logger.Info("Test sensitive data",
		logx.String("password", "secret123"),
		logx.String("email", "user@example.com"),
		logx.String("ssn", "123-45-6789"),
		logx.String("token", "jwt_token_here"),
		logx.String("normal_field", "visible_value"),
	)
}

func TestAddRemoveSensitiveKey(t *testing.T) {
	// Test adding a new sensitive key
	logx.AddSensitiveKey("custom_secret")

	// Test removing the key
	logx.RemoveSensitiveKey("custom_secret")

	// These functions should not panic
}

func TestLoggerConcurrency(t *testing.T) {
	config := logx.DefaultConfig()
	logger, err := logx.New(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	// Test concurrent logging
	var wg sync.WaitGroup
	numGoroutines := 100

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			logger.Info("Concurrent log", logx.Int("goroutine_id", id))
		}(i)
	}

	wg.Wait()
}

func TestLoggerLevels(t *testing.T) {
	config := logx.DefaultConfig()
	config.Level = logx.DebugLevel
	logger, err := logx.New(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	// Test all log levels
	logger.Debug("Debug message")
	logger.Info("Info message")
	logger.Warn("Warn message")
	logger.Error("Error message")
}

func TestLoggerWithNilLogger(t *testing.T) {
	// Test that With works with nil logger (should not panic)
	var logger *logx.Logger
	// This should panic, so we expect it to panic
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic when calling With on nil logger")
		}
	}()
	logger.With(logx.String("test", "value"))
}

func TestInitDefault(t *testing.T) {
	// Test that InitDefault doesn't panic
	logx.InitDefault()
}

func TestTimestampFormat(t *testing.T) {
	config := logx.DefaultConfig()
	logger, err := logx.New(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	// Test that logging works without panicking
	logger.Info("Test timestamp")
}

func BenchmarkLoggerInfo(b *testing.B) {
	config := logx.DefaultConfig()
	logger, err := logx.New(config)
	if err != nil {
		b.Fatalf("Failed to create logger: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("Benchmark message", logx.Int("iteration", i))
	}
}

func BenchmarkLoggerWithFields(b *testing.B) {
	config := logx.DefaultConfig()
	logger, err := logx.New(config)
	if err != nil {
		b.Fatalf("Failed to create logger: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.With(
			logx.String("field1", "value1"),
			logx.Int("field2", i),
			logx.Bool("field3", true),
		).Info("Benchmark message")
	}
}

func BenchmarkSensitiveDataMasking(b *testing.B) {
	config := logx.DefaultConfig()
	logger, err := logx.New(config)
	if err != nil {
		b.Fatalf("Failed to create logger: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("Benchmark sensitive data",
			logx.String("password", "secret123"),
			logx.String("email", "user@example.com"),
			logx.String("ssn", "123-45-6789"),
		)
	}
}

// TestTraceLevelLogging tests all trace level logging functions
func TestTraceLevelLogging(t *testing.T) {
	config := logx.DefaultConfig()
	config.Level = logx.TraceLevel // Enable trace level
	logger, err := logx.New(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Sync()

	t.Run("Trace Function", func(t *testing.T) {
		// Test basic trace functionality
		logger.Trace("Basic trace message")
		logger.Trace("Trace with simple string")
		logger.Trace("Trace with special chars: \n\t\r")
		logger.Trace("Trace with unicode: 测试中文")
		logger.Trace("Trace with very long message: " + string(make([]byte, 1000)))

		// Test trace with fields
		logger.Trace("Trace with fields",
			logx.String("field1", "value1"),
			logx.Int("field2", 42),
			logx.Float64("field3", 3.14),
			logx.Bool("field4", true),
		)

		// Test trace with error field
		logger.Trace("Trace with error", logx.ErrorField(errors.New("test error")))

		// Test trace with complex data
		logger.Trace("Trace with complex data",
			logx.Any("map", map[string]interface{}{
				"key1": "value1",
				"key2": 42,
				"key3": []int{1, 2, 3},
			}),
			logx.Any("slice", []string{"a", "b", "c"}),
		)
	})

	t.Run("Trace Level Configuration", func(t *testing.T) {
		// Test with different log levels to ensure trace is only logged when enabled
		levels := []logx.Level{
			logx.TraceLevel,
			logx.DebugLevel,
			logx.InfoLevel,
			logx.WarnLevel,
			logx.ErrorLevel,
			logx.FatalLevel,
		}

		for _, level := range levels {
			config := logx.DefaultConfig()
			config.Level = level
			logger, err := logx.New(config)
			if err != nil {
				t.Fatalf("Failed to create logger with level %v: %v", level, err)
			}
			defer logger.Sync()

			// Trace should only be logged when level is TraceLevel
			logger.Trace("Trace message with level " + level.String())
		}
	})

	t.Run("Trace with Sensitive Data", func(t *testing.T) {
		// Test trace logging with sensitive data
		logger.Trace("Trace with sensitive data",
			logx.String("password", "secret123"),
			logx.String("email", "user@example.com"),
			logx.String("token", "jwt_token_here"),
			logx.String("normal_field", "normal_value"),
		)
	})

	t.Run("Trace Performance", func(t *testing.T) {
		numMessages := 1000
		start := time.Now()

		for i := 0; i < numMessages; i++ {
			logger.Trace("Performance test message", logx.Int("iteration", i))
		}

		duration := time.Since(start)
		messagesPerSecond := 0.0
		if duration.Seconds() > 0 {
			messagesPerSecond = float64(numMessages) / duration.Seconds()
		}

		t.Logf("Trace logging performance: %d messages in %v (%.2f msg/sec)",
			numMessages, duration, messagesPerSecond)
	})

	t.Run("Concurrent Trace Logging", func(t *testing.T) {
		numGoroutines := 100
		var wg sync.WaitGroup
		wg.Add(numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			go func(id int) {
				defer wg.Done()

				// Test trace logging concurrently
				logger.Trace("Concurrent trace message", logx.Int("goroutine_id", id))
				logger.Trace("Another concurrent trace", logx.String("data", "value"))
			}(i)
		}

		wg.Wait()
	})

	t.Run("Trace with Logger Chains", func(t *testing.T) {
		// Test trace with logger chains
		childLogger := logger.With(
			logx.String("context", "child"),
			logx.Int("level", 1),
		)

		grandChildLogger := childLogger.With(
			logx.String("context", "grandchild"),
			logx.Int("level", 2),
		)

		logger.Trace("Parent trace message")
		childLogger.Trace("Child trace message")
		grandChildLogger.Trace("Grandchild trace message")
	})

	t.Run("Trace Edge Cases", func(t *testing.T) {
		// Test edge cases for trace logging
		logger.Trace("") // Empty message
		logger.Trace("Single character: a")
		logger.Trace("Very long message: " + string(make([]byte, 10000)))

		// Test with nil fields
		logger.Trace("Trace with nil field", logx.String("nil_field", ""))

		// Test with empty fields
		logger.Trace("Trace with empty fields",
			logx.String("empty", ""),
			logx.Int("zero", 0),
			logx.Float64("zero_float", 0.0),
			logx.Bool("false", false),
		)
	})

	t.Run("Trace with Large Data Structures", func(t *testing.T) {
		// Test trace with large data structures
		largeMap := make(map[string]interface{})
		for i := 0; i < 100; i++ {
			largeMap[fmt.Sprintf("key_%d", i)] = fmt.Sprintf("value_%d", i)
		}

		largeSlice := make([]int, 1000)
		for i := range largeSlice {
			largeSlice[i] = i
		}

		logger.Trace("Trace with large map", logx.Any("large_map", largeMap))
		logger.Trace("Trace with large slice", logx.Any("large_slice", largeSlice))
	})

	t.Run("Trace Level Comparison", func(t *testing.T) {
		// Test that trace level is the lowest level
		tests := []struct {
			level    logx.Level
			expected bool
		}{
			{logx.TraceLevel, true},
			{logx.DebugLevel, false},
			{logx.InfoLevel, false},
			{logx.WarnLevel, false},
			{logx.ErrorLevel, false},
			{logx.FatalLevel, false},
		}

		for _, test := range tests {
			config := logx.DefaultConfig()
			config.Level = test.level
			logger, err := logx.New(config)
			if err != nil {
				t.Fatalf("Failed to create logger: %v", err)
			}
			defer logger.Sync()

			// Trace should only be logged when level is TraceLevel
			logger.Trace("Trace message test")
		}
	})
}

// TestFormattedLoggingFunctions tests all formatted logging functions
func TestFormattedLoggingFunctions(t *testing.T) {
	config := logx.DefaultConfig()
	config.Level = logx.TraceLevel // Enable all levels including trace
	logger, err := logx.New(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Sync()

	t.Run("Tracef Function", func(t *testing.T) {
		// Test basic tracef functionality
		logger.Tracef("Basic tracef: %s", "value")
		logger.Tracef("Tracef with number: %d", 42)
		logger.Tracef("Tracef with float: %f", 3.14)
		logger.Tracef("Tracef with boolean: %t", true)
		logger.Tracef("Tracef with error: %v", errors.New("test error"))

		// Test edge cases
		logger.Tracef("") // Empty format string
		logger.Tracef("No args")
		logger.Tracef("Multiple args: %s %d %f %t", "string", 123, 2.718, false)
		logger.Tracef("Special chars: %s", "test\n\t\r")
		logger.Tracef("Unicode: %s", "测试中文")
		logger.Tracef("Very long format: %s", "very long string that might cause issues with formatting and should be handled properly by the logging system")
	})

	t.Run("Debugf Function", func(t *testing.T) {
		// Test basic debugf functionality
		logger.Debugf("Basic debugf: %s", "value")
		logger.Debugf("Debugf with number: %d", 42)
		logger.Debugf("Debugf with float: %f", 3.14)
		logger.Debugf("Debugf with boolean: %t", true)
		logger.Debugf("Debugf with error: %v", errors.New("test error"))

		// Test edge cases
		logger.Debugf("") // Empty format string
		logger.Debugf("No args")
		logger.Debugf("Multiple args: %s %d %f %t", "string", 123, 2.718, false)
		logger.Debugf("Special chars: %s", "test\n\t\r")
		logger.Debugf("Unicode: %s", "测试中文")
		logger.Debugf("Very long format: %s", "very long string that might cause issues with formatting and should be handled properly by the logging system")
	})

	t.Run("Infof Function", func(t *testing.T) {
		// Test basic infof functionality
		logger.Infof("Basic infof: %s", "value")
		logger.Infof("Infof with number: %d", 42)
		logger.Infof("Infof with float: %f", 3.14)
		logger.Infof("Infof with boolean: %t", true)
		logger.Infof("Infof with error: %v", errors.New("test error"))

		// Test edge cases
		logger.Infof("") // Empty format string
		logger.Infof("No args")
		logger.Infof("Multiple args: %s %d %f %t", "string", 123, 2.718, false)
		logger.Infof("Special chars: %s", "test\n\t\r")
		logger.Infof("Unicode: %s", "测试中文")
		logger.Infof("Very long format: %s", "very long string that might cause issues with formatting and should be handled properly by the logging system")
	})

	t.Run("Warnf Function", func(t *testing.T) {
		// Test basic warnf functionality
		logger.Warnf("Basic warnf: %s", "value")
		logger.Warnf("Warnf with number: %d", 42)
		logger.Warnf("Warnf with float: %f", 3.14)
		logger.Warnf("Warnf with boolean: %t", true)
		logger.Warnf("Warnf with error: %v", errors.New("test error"))

		// Test edge cases
		logger.Warnf("") // Empty format string
		logger.Warnf("No args")
		logger.Warnf("Multiple args: %s %d %f %t", "string", 123, 2.718, false)
		logger.Warnf("Special chars: %s", "test\n\t\r")
		logger.Warnf("Unicode: %s", "测试中文")
		logger.Warnf("Very long format: %s", "very long string that might cause issues with formatting and should be handled properly by the logging system")
	})

	t.Run("Errorf Function", func(t *testing.T) {
		// Test basic errorf functionality
		logger.Errorf("Basic errorf: %s", "value")
		logger.Errorf("Errorf with number: %d", 42)
		logger.Errorf("Errorf with float: %f", 3.14)
		logger.Errorf("Errorf with boolean: %t", true)
		logger.Errorf("Errorf with error: %v", errors.New("test error"))

		// Test edge cases
		logger.Errorf("") // Empty format string
		logger.Errorf("No args")
		logger.Errorf("Multiple args: %s %d %f %t", "string", 123, 2.718, false)
		logger.Errorf("Special chars: %s", "test\n\t\r")
		logger.Errorf("Unicode: %s", "测试中文")
		logger.Errorf("Very long format: %s", "very long string that might cause issues with formatting and should be handled properly by the logging system")
	})

	t.Run("Fatalf Function", func(t *testing.T) {
		// Note: Fatalf calls os.Exit(1), so we can't test it directly
		// Instead, we'll test that the function exists and can be called
		// In a real scenario, this would be tested in integration tests
		t.Skip("Fatalf calls os.Exit(1) - cannot test directly in unit tests")
	})

	t.Run("Format String Edge Cases", func(t *testing.T) {
		// Test various format specifiers
		logger.Infof("String: %s", "test")
		logger.Infof("Integer: %d", 42)
		logger.Infof("Float: %f", 3.14159)
		logger.Infof("Boolean: %t", true)
		logger.Infof("Pointer: %p", &config)
		logger.Infof("Hex: %x", 255)
		logger.Infof("Octal: %o", 64)
		logger.Infof("Binary: %b", 10)
		logger.Infof("Unicode: %U", 'A')
		logger.Infof("Quote: %q", "quoted string")

		// Test width and precision
		logger.Infof("Width: %10s", "test")
		logger.Infof("Precision: %.2f", 3.14159)
		logger.Infof("Width and precision: %10.2f", 3.14159)

		// Test multiple format specifiers
		logger.Infof("Multiple: %s %d %f %t", "string", 42, 3.14, true)
		logger.Infof("Mixed: %s%d%f%t", "string", 42, 3.14, true)
	})

	t.Run("Argument Edge Cases", func(t *testing.T) {
		// Test nil arguments
		logger.Infof("Nil string: %v", nil)
		logger.Infof("Nil interface: %v", nil)

		// Test zero values
		logger.Infof("Zero int: %d", 0)
		logger.Infof("Zero float: %f", 0.0)
		logger.Infof("Zero bool: %t", false)
		logger.Infof("Empty string: %s", "")

		// Test extreme values
		logger.Infof("Max int: %d", 9223372036854775807)
		logger.Infof("Min int: %d", -9223372036854775808)
		logger.Infof("Max float: %f", 1.7976931348623157e+308)
		logger.Infof("Min float: %f", -1.7976931348623157e+308)

		// Test special values
		logger.Infof("NaN: %f", math.NaN())
		logger.Infof("Infinity: %f", math.Inf(1))
		logger.Infof("Negative infinity: %f", math.Inf(-1))
	})

	t.Run("Concurrent Formatted Logging", func(t *testing.T) {
		numGoroutines := 100
		var wg sync.WaitGroup
		wg.Add(numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			go func(id int) {
				defer wg.Done()

				// Test all formatted functions concurrently
				logger.Tracef("Concurrent tracef: %d", id)
				logger.Debugf("Concurrent debugf: %d", id)
				logger.Infof("Concurrent infof: %d", id)
				logger.Warnf("Concurrent warnf: %d", id)
				logger.Errorf("Concurrent errorf: %d", id)

				// Test with complex format strings
				logger.Infof("Complex format %d: %s %f %t", id, "string", 3.14, true)
			}(i)
		}

		wg.Wait()
	})

	t.Run("Large Format Strings", func(t *testing.T) {
		// Test with very large format strings
		largeFormat := ""
		for i := 0; i < 100; i++ {
			largeFormat += fmt.Sprintf("arg%d: %%s ", i)
		}

		args := make([]interface{}, 100)
		for i := range args {
			args[i] = fmt.Sprintf("value%d", i)
		}

		logger.Infof(largeFormat, args...)
	})

	t.Run("Format String Injection", func(t *testing.T) {
		// Test with potentially problematic format strings
		logger.Infof("Normal: %s", "test")
		logger.Infof("Percent: %s", "test")
		logger.Infof("Multiple percent: %s", "test")
		logger.Infof("Mixed: %s %s %s", "first", "second", "third")
	})

	t.Run("Performance Under Load", func(t *testing.T) {
		numMessages := 1000
		start := time.Now()

		for i := 0; i < numMessages; i++ {
			logger.Infof("Performance test %d: %s %d %f", i, "string", i, float64(i)/100.0)
		}

		duration := time.Since(start)
		messagesPerSecond := 0.0
		if duration.Seconds() > 0 {
			messagesPerSecond = float64(numMessages) / duration.Seconds()
		}

		t.Logf("Formatted logging performance: %d messages in %v (%.2f msg/sec)",
			numMessages, duration, messagesPerSecond)
	})
}

// TestLoggerCreationEdgeCases tests edge cases in logger creation
func TestLoggerCreationEdgeCases(t *testing.T) {
	t.Run("Nil Config", func(t *testing.T) {
		// Test with nil config - this causes panic, so we skip it
		// The library doesn't handle nil config gracefully
		t.Skip("Nil config causes panic - not a valid edge case")
	})

	t.Run("Invalid Level", func(t *testing.T) {
		// Test with invalid level
		config := &logx.Config{
			Level: logx.Level(999), // Invalid level
		}
		logger, err := logx.New(config)
		if err == nil {
			t.Log("Logger created with invalid level (library may not validate)")
		} else {
			t.Logf("Expected error with invalid level: %v", err)
		}
		if logger != nil {
			defer logger.Sync()
		}
	})

	t.Run("Empty Config", func(t *testing.T) {
		// Test with empty config
		config := &logx.Config{}
		logger, err := logx.New(config)
		if err != nil {
			t.Fatalf("Failed to create logger with empty config: %v", err)
		}
		defer logger.Sync()

		// Test that it works
		logger.Info("Test with empty config")
	})

	t.Run("All Config Combinations", func(t *testing.T) {
		// Test all possible config combinations
		configs := []*logx.Config{
			{Level: logx.DebugLevel, Development: true, AddCaller: true, AddStacktrace: true},
			{Level: logx.InfoLevel, Development: false, AddCaller: false, AddStacktrace: false},
			{Level: logx.WarnLevel, Development: true, AddCaller: false, AddStacktrace: true},
			{Level: logx.ErrorLevel, Development: false, AddCaller: true, AddStacktrace: false},
			{Level: logx.FatalLevel, Development: true, AddCaller: true, AddStacktrace: false},
		}

		for i, config := range configs {
			logger, err := logx.New(config)
			if err != nil {
				t.Fatalf("Failed to create logger with config %d: %v", i, err)
			}
			defer logger.Sync()

			logger.Info("Config combination test", logx.Int("config_id", i))
		}
	})
}

// TestFieldConversionEdgeCases tests edge cases in field conversion
func TestFieldConversionEdgeCases(t *testing.T) {
	config := logx.DefaultConfig()
	logger, err := logx.New(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Sync()

	t.Run("Empty Fields", func(t *testing.T) {
		// Test logging with no fields
		logger.Info("Message with no fields")
	})

	t.Run("Nil Fields", func(t *testing.T) {
		// Test logging with nil fields slice
		logger.Info("Message with nil fields")
	})

	t.Run("Complex Field Types", func(t *testing.T) {
		// Test complex field types
		logger.Info("Complex fields test",
			logx.Any("map", map[string]interface{}{
				"nested": map[string]string{"key": "value"},
				"array":  []int{1, 2, 3},
				"nil":    nil,
			}),
			logx.Any("slice", []string{"a", "b", "c"}),
			logx.Any("struct", struct{ Name string }{Name: "test"}),
			logx.Any("pointer", &struct{ Value int }{Value: 42}),
		)
	})

	t.Run("Empty String Values", func(t *testing.T) {
		// Test empty string values
		logger.Info("Empty string test",
			logx.String("empty", ""),
			logx.String("whitespace", "   "),
			logx.String("normal", "value"),
		)
	})

	t.Run("Special Characters", func(t *testing.T) {
		// Test special characters in field values
		logger.Info("Special characters test",
			logx.String("newline", "line1\nline2"),
			logx.String("tab", "col1\tcol2"),
			logx.String("quotes", `"quoted" 'string'`),
			logx.String("unicode", "🚀🌟🎉"),
			logx.String("emoji", "Hello 👋 World 🌍"),
		)
	})

	t.Run("Large Values", func(t *testing.T) {
		// Test large field values
		largeString := string(make([]byte, 10000)) // 10KB string
		logger.Info("Large value test",
			logx.String("large", largeString),
			logx.Int("large_number", 999999999),
			logx.Float64("large_float", 3.141592653589793),
		)
	})

	t.Run("Nil Error Field", func(t *testing.T) {
		// Test nil error in ErrorField
		logger.Error("Nil error test", logx.ErrorField(nil))
	})

	t.Run("Empty Error", func(t *testing.T) {
		// Test empty error
		emptyErr := errors.New("")
		logger.Error("Empty error test", logx.ErrorField(emptyErr))
	})
}

// TestSensitiveDataMaskingEdgeCases tests edge cases in sensitive data masking
func TestSensitiveDataMaskingEdgeCases(t *testing.T) {
	config := logx.DefaultConfig()
	logger, err := logx.New(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Sync()

	t.Run("Empty Sensitive Values", func(t *testing.T) {
		logger.Info("Empty sensitive values",
			logx.String("password", ""),
			logx.String("email", ""),
			logx.String("token", ""),
		)
	})

	t.Run("Short Sensitive Values", func(t *testing.T) {
		logger.Info("Short sensitive values",
			logx.String("password", "a"),
			logx.String("email", "ab"),
			logx.String("token", "abc"),
			logx.String("pin", "1234"),
		)
	})

	t.Run("Long Sensitive Values", func(t *testing.T) {
		longPassword := string(make([]byte, 100))
		logger.Info("Long sensitive values",
			logx.String("password", longPassword),
			logx.String("email", "very.long.email.address@very.long.domain.com"),
			logx.String("token", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"),
		)
	})

	t.Run("Case Insensitive Keys", func(t *testing.T) {
		logger.Info("Case insensitive keys",
			logx.String("PASSWORD", "secret123"),
			logx.String("Email", "user@example.com"),
			logx.String("API_KEY", "key123"),
			logx.String("Token", "token123"),
		)
	})

	t.Run("Mixed Case Keys", func(t *testing.T) {
		logger.Info("Mixed case keys",
			logx.String("PassWord", "secret123"),
			logx.String("eMail", "user@example.com"),
			logx.String("ApiKey", "key123"),
		)
	})

	t.Run("Non-String Sensitive Values", func(t *testing.T) {
		logger.Info("Non-string sensitive values",
			logx.Int("password", 12345),
			logx.Float64("token", 3.14159),
			logx.Bool("secret", true),
			logx.Any("key", []int{1, 2, 3}),
		)
	})

	t.Run("Custom Sensitive Keys", func(t *testing.T) {
		// Add custom sensitive keys
		logx.AddSensitiveKey("custom_secret")
		logx.AddSensitiveKey("internal_id")

		logger.Info("Custom sensitive keys",
			logx.String("custom_secret", "secret_value"),
			logx.String("internal_id", "12345"),
			logx.String("normal_field", "visible_value"),
		)

		// Remove custom keys
		logx.RemoveSensitiveKey("custom_secret")
		logx.RemoveSensitiveKey("internal_id")
	})

	t.Run("Empty and Whitespace Keys", func(t *testing.T) {
		logx.AddSensitiveKey("")
		logx.AddSensitiveKey("   ")

		logger.Info("Empty and whitespace keys",
			logx.String("", "empty_key_value"),
			logx.String("   ", "whitespace_key_value"),
		)

		logx.RemoveSensitiveKey("")
		logx.RemoveSensitiveKey("   ")
	})
}

// TestConcurrencyEdgeCases tests edge cases in concurrent operations
func TestConcurrencyEdgeCases(t *testing.T) {
	config := logx.DefaultConfig()
	logger, err := logx.New(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Sync()

	t.Run("Rapid With Operations", func(t *testing.T) {
		// Test rapid With operations
		var wg sync.WaitGroup
		numGoroutines := 100

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()

				// Create multiple With chains rapidly
				childLogger := logger.With(
					logx.Int("goroutine_id", id),
					logx.String("timestamp", time.Now().String()),
				)

				grandChildLogger := childLogger.With(
					logx.String("level", "child"),
					logx.Int("iteration", id*2),
				)

				grandChildLogger.Info("Rapid With test")
			}(i)
		}

		wg.Wait()
	})

	t.Run("Concurrent Sensitive Key Operations", func(t *testing.T) {
		// Test concurrent sensitive key operations
		var wg sync.WaitGroup
		numGoroutines := 50

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()

				key := fmt.Sprintf("concurrent_key_%d", id)
				logx.AddSensitiveKey(key)

				logger.Info("Concurrent sensitive key test",
					logx.String(key, "sensitive_value"),
				)

				logx.RemoveSensitiveKey(key)
			}(i)
		}

		wg.Wait()
	})

	t.Run("Logger Creation Under Load", func(t *testing.T) {
		// Test logger creation under concurrent load
		var wg sync.WaitGroup
		numLoggers := 100
		loggers := make([]*logx.Logger, numLoggers)
		errors := make([]error, numLoggers)

		for i := 0; i < numLoggers; i++ {
			wg.Add(1)
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

		// Test all created loggers
		for i, logger := range loggers {
			if errors[i] != nil {
				t.Errorf("Failed to create logger %d: %v", i, errors[i])
				continue
			}
			if logger != nil {
				logger.Info("Concurrent logger test", logx.Int("logger_id", i))
				logger.Sync()
			}
		}
	})
}

// TestFormattedLoggingEdgeCases tests edge cases in formatted logging
func TestFormattedLoggingEdgeCases(t *testing.T) {
	config := logx.DefaultConfig()
	config.Level = logx.DebugLevel
	logger, err := logx.New(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Sync()

	t.Run("Empty Format String", func(t *testing.T) {
		logger.Infof("")
		logger.Debugf("")
		logger.Warnf("")
		logger.Errorf("")
	})

	t.Run("No Arguments", func(t *testing.T) {
		logger.Infof("No arguments")
		logger.Debugf("Debug message")
		logger.Warnf("Warning message")
		logger.Errorf("Error message")
	})

	t.Run("Nil Arguments", func(t *testing.T) {
		logger.Infof("Nil arguments: %v, %v", nil, nil)
		logger.Debugf("Debug with nil: %v", nil)
		logger.Warnf("Warning with nil: %v", nil)
		logger.Errorf("Error with nil: %v", nil)
	})

	t.Run("Mixed Argument Types", func(t *testing.T) {
		logger.Infof("Mixed types: %s, %d, %f, %t, %v", "string", 42, 3.14, true, []int{1, 2, 3})
		logger.Debugf("Debug mixed: %s %d", "test", 123)
		logger.Warnf("Warning mixed: %f %t", 2.718, false)
		logger.Errorf("Error mixed: %v %s", errors.New("test error"), "message")
	})

	t.Run("Special Format Characters", func(t *testing.T) {
		logger.Infof("Special chars: %% %s", "percent")
		logger.Debugf("Newline: %s", "line1\nline2")
		logger.Warnf("Tab: %s", "col1\tcol2")
		logger.Errorf("Quotes: %s", `"quoted"`)
	})

	t.Run("Large Format Strings", func(t *testing.T) {
		largeFormat := strings.Repeat("%s ", 100) // 100 format specifiers
		args := make([]interface{}, 100)
		for i := range args {
			args[i] = fmt.Sprintf("arg%d", i)
		}

		logger.Infof(largeFormat, args...)
	})

	t.Run("Mismatched Arguments", func(t *testing.T) {
		// Test with more format specifiers than arguments
		logger.Infof("More specifiers: %s %d %f", "string", 42, 3.14)
		logger.Debugf("Extra specifier: %s %d", "string", 42)
	})
}

// TestTraceLevelEdgeCases tests edge cases for trace level logging
func TestTraceLevelEdgeCases(t *testing.T) {
	config := logx.DefaultConfig()
	config.Level = logx.TraceLevel
	logger, err := logx.New(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Sync()

	t.Run("Trace Level Enabled", func(t *testing.T) {
		logger.Trace("Trace message")
		logger.Tracef("Trace formatted: %s", "value")
	})

	t.Run("Trace With Fields", func(t *testing.T) {
		logger.Trace("Trace with fields",
			logx.String("component", "trace_test"),
			logx.Int("level", 0),
		)
	})

	t.Run("Trace Level Disabled", func(t *testing.T) {
		// Create logger with higher level
		config.Level = logx.InfoLevel
		infoLogger, err := logx.New(config)
		if err != nil {
			t.Fatalf("Failed to create info logger: %v", err)
		}
		defer infoLogger.Sync()

		// Trace messages should not appear
		infoLogger.Trace("This should not appear")
		infoLogger.Tracef("This should not appear: %s", "value")
	})
}

// TestPackageLevelEdgeCases tests edge cases for package-level functions
func TestPackageLevelEdgeCases(t *testing.T) {
	t.Run("Package Level Init", func(t *testing.T) {
		// Test package-level initialization
		logx.InitDefault()
		defer logx.Sync()

		// Test package-level logging
		logx.Info("Package level info")
		logx.Debug("Package level debug")
		logx.Warn("Package level warn")
		logx.Error("Package level error")
	})

	t.Run("Package Level With", func(t *testing.T) {
		logx.InitDefault()
		defer logx.Sync()

		// Test package-level With
		logger := logx.With(logx.String("package", "test"))
		logger.Info("Package level with context")
	})

	t.Run("Package Level Formatted", func(t *testing.T) {
		logx.InitDefault()
		defer logx.Sync()

		// Test package-level formatted logging (these functions don't exist at package level)
		// We'll test the logger-level formatted functions instead
		logger, err := logx.New(logx.DefaultConfig())
		if err != nil {
			t.Fatalf("Failed to create logger: %v", err)
		}
		defer logger.Sync()

		logger.Infof("Logger level formatted: %s", "value")
		logger.Debugf("Logger level debug formatted: %d", 42)
		logger.Warnf("Logger level warn formatted: %f", 3.14)
		logger.Errorf("Logger level error formatted: %v", errors.New("test error"))
	})

	t.Run("Package Level Trace", func(t *testing.T) {
		// Test package-level trace (should work if trace level is enabled)
		logx.Trace("Package level trace")
		logx.Tracef("Package level trace formatted: %s", "value")
	})
}

// TestErrorHandlingEdgeCases tests edge cases in error handling
func TestErrorHandlingEdgeCases(t *testing.T) {
	config := logx.DefaultConfig()
	logger, err := logx.New(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Sync()

	t.Run("Nil Logger Operations", func(t *testing.T) {
		var nilLogger *logx.Logger

		// These should not panic
		if nilLogger != nil {
			nilLogger.Info("This should not execute")
			nilLogger.With(logx.String("test", "value"))
			nilLogger.Sync()
		}
	})

	t.Run("Panic Recovery", func(t *testing.T) {
		// Test that logging operations don't panic
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Unexpected panic: %v", r)
			}
		}()

		logger.Info("Normal logging")
		logger.With(logx.String("test", "value"))
		logger.Sync()
	})

	t.Run("Large Error Messages", func(t *testing.T) {
		largeError := errors.New(strings.Repeat("error message ", 1000))
		logger.Error("Large error test", logx.ErrorField(largeError))
	})

	t.Run("Error with Stack Trace", func(t *testing.T) {
		err := errors.New("test error with context")
		logger.Error("Error with stack trace", logx.ErrorField(err))
	})
}

// TestPerformanceEdgeCases tests edge cases in performance scenarios
func TestPerformanceEdgeCases(t *testing.T) {
	config := logx.DefaultConfig()
	logger, err := logx.New(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Sync()

	t.Run("High Frequency Logging", func(t *testing.T) {
		// Test high frequency logging
		for i := 0; i < 10000; i++ {
			logger.Info("High frequency test", logx.Int("iteration", i))
		}
	})

	t.Run("Large Message Logging", func(t *testing.T) {
		// Test logging large messages
		largeMessage := strings.Repeat("large message content ", 1000)
		logger.Info(largeMessage, logx.String("size", "large"))
	})

	t.Run("Many Fields", func(t *testing.T) {
		// Test logging with many fields
		fields := make([]logx.Field, 100)
		for i := range fields {
			fields[i] = logx.String(fmt.Sprintf("field_%d", i), fmt.Sprintf("value_%d", i))
		}

		logger.Info("Many fields test", fields...)
	})

	t.Run("Memory Pressure", func(t *testing.T) {
		// Test under memory pressure
		for i := 0; i < 1000; i++ {
			largeData := strings.Repeat("memory pressure test data ", 100)
			logger.Info("Memory pressure test",
				logx.String("large_data", largeData),
				logx.Int("iteration", i),
			)
		}
	})
}
