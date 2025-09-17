# Go-LogX Usage Guide

## Table of Contents
1. [Quick Start](#quick-start)
2. [Basic Usage](#basic-usage)
3. [Configuration](#configuration)
4. [Structured Logging](#structured-logging)
5. [Sensitive Data Masking](#sensitive-data-masking)
6. [Advanced Features](#advanced-features)
7. [Best Practices](#best-practices)
8. [Common Patterns](#common-patterns)
9. [Troubleshooting](#troubleshooting)

## Quick Start

### Installation
```bash
go get github.com/seasbee/go-logx
```

### Basic Setup
```go
package main

import (
    "github.com/seasbee/go-logx"
)

func main() {
    // Initialize with default configuration
    if err := logx.InitDefault(); err != nil {
        panic(err)
    }
    defer logx.Sync()

    // Start logging
    logx.Info("Application started")
    logx.Info("User logged in", logx.String("user_id", "12345"))
}
```

## Basic Usage

### Package-Level Functions
The simplest way to use go-logx is through package-level functions:

```go
// Initialize the logger
logx.InitDefault()

// Basic logging
logx.Trace("Detailed debugging information")
logx.Debug("Debug information")
logx.Info("General information")
logx.Warn("Warning message")
logx.Error("Error occurred")
logx.Fatal("Critical error - application will exit")

// Formatted logging
logx.Infof("User %s logged in from %s", username, ipAddress)
logx.Errorf("Failed to connect to database: %v", err)
```

### Logger Instances
For more control, create custom logger instances:

```go
// Create a custom logger
config := &logx.Config{
    Level:       logx.DebugLevel,
    Development: true,
    AddCaller:   true,
}

logger, err := logx.New(config)
if err != nil {
    log.Fatal(err)
}
defer logger.Sync()

// Use the logger
logger.Info("Custom logger message")
logger.With(logx.String("component", "auth")).Info("Authentication event")
```

## Configuration

### Default Configuration
```go
config := logx.DefaultConfig()
// Equivalent to:
// &logx.Config{
//     Level:         logx.InfoLevel,
//     OutputPath:    "",           // stdout
//     Development:   false,        // JSON output
//     AddCaller:     true,         // Include caller info
//     AddStacktrace: true,         // Include stack traces
// }
```

### Custom Configuration
```go
config := &logx.Config{
    Level:         logx.DebugLevel,    // Set minimum log level
    OutputPath:    "/var/log/app.log", // Write to file
    Development:   true,               // Console output for development
    AddCaller:     true,               // Include file:line information
    AddStacktrace: false,              // Disable stack traces
}

if err := logx.Init(config); err != nil {
    log.Fatal(err)
}
```

### Environment-Based Configuration
```go
func setupLogger() error {
    config := logx.DefaultConfig()
    
    // Override based on environment
    switch os.Getenv("ENV") {
    case "development":
        config.Level = logx.DebugLevel
        config.Development = true
    case "production":
        config.Level = logx.InfoLevel
        config.Development = false
        config.OutputPath = "/var/log/app.log"
    case "staging":
        config.Level = logx.WarnLevel
        config.Development = false
    }
    
    return logx.Init(config)
}
```

## Structured Logging

### Field Types
go-logx provides type-safe field constructors:

```go
// String fields
logx.Info("User action", logx.String("user_id", "12345"))

// Numeric fields
logx.Info("Request processed", logx.Int("status_code", 200))
logx.Info("Performance metric", logx.Float64("response_time", 0.123))
logx.Info("Large number", logx.Int64("timestamp", time.Now().Unix()))

// Boolean fields
logx.Info("Feature status", logx.Bool("enabled", true))

// Error fields
if err != nil {
    logx.Error("Operation failed", logx.ErrorField(err))
}

// Any type
logx.Info("Complex data", logx.Any("user", userStruct))
```

### Contextual Logging
Create loggers with persistent context:

```go
// Create a logger with user context
userLogger := logx.With(
    logx.String("user_id", "12345"),
    logx.String("session_id", "abc123"),
)

// All messages from this logger will include the context
userLogger.Info("User performed action")
userLogger.Info("Another user action")

// Output:
// {"level":"INFO","message":"User performed action","user_id":"12345","session_id":"abc123"}
// {"level":"INFO","message":"Another user action","user_id":"12345","session_id":"abc123"}
```

### Nested Context
```go
// Create base logger with application context
appLogger := logx.With(
    logx.String("app", "myapp"),
    logx.String("version", "1.0.0"),
)

// Create user-specific logger
userLogger := appLogger.With(
    logx.String("user_id", "12345"),
)

// Create request-specific logger
requestLogger := userLogger.With(
    logx.String("request_id", "req-456"),
)

// All context is preserved
requestLogger.Info("Request processed")
// Output: {"level":"INFO","message":"Request processed","app":"myapp","version":"1.0.0","user_id":"12345","request_id":"req-456"}
```

## Sensitive Data Masking

### Automatic Masking
go-logx automatically masks sensitive data based on field names:

```go
// These will be automatically masked
logx.Info("Login attempt", logx.String("password", "secret123"))
logx.Info("API call", logx.String("token", "abc123"))
logx.Info("User data", logx.String("ssn", "123-45-6789"))

// Output:
// {"level":"INFO","message":"Login attempt","password":"se***23"}
// {"level":"INFO","message":"API call","token":"ab***23"}
// {"level":"INFO","message":"User data","ssn":"12***89"}
```

### Custom Sensitive Keys
```go
// Add custom sensitive field names
logx.AddSensitiveKey("my_secret_field")
logx.AddSensitiveKey("internal_token")

// These will now be masked
logx.Info("Custom secret", logx.String("my_secret_field", "sensitive_value"))
logx.Info("Internal call", logx.String("internal_token", "secret123"))

// Remove keys from sensitive list
logx.RemoveSensitiveKey("email") // Don't mask email addresses
```

### Masking Behavior
```go
// String masking examples
logx.Info("Short", logx.String("password", "ab"))      // "***"
logx.Info("Medium", logx.String("password", "abc"))    // "a***c"
logx.Info("Long", logx.String("password", "abcdef"))   // "ab***ef"

// Other types
logx.Info("Bytes", logx.Any("token", []byte("secret"))) // "se***t"
logx.Info("Number", logx.Any("secret", 12345))          // "***MASKED***"
```

## Advanced Features

### Custom Logger Creation
```go
// Create multiple loggers for different components
authLogger, _ := logx.New(&logx.Config{
    Level:       logx.DebugLevel,
    Development: true,
})
defer authLogger.Sync()

dbLogger, _ := logx.New(&logx.Config{
    Level:      logx.InfoLevel,
    OutputPath: "/var/log/db.log",
})
defer dbLogger.Sync()

// Use component-specific loggers
authLogger.Info("Authentication successful")
dbLogger.Info("Database query executed")
```

### File Output
```go
config := &logx.Config{
    Level:      logx.InfoLevel,
    OutputPath: "/var/log/application.log",
}

if err := logx.Init(config); err != nil {
    log.Fatal(err)
}
```

### Development vs Production
```go
func setupLogger(isDevelopment bool) error {
    config := &logx.Config{
        Level: logx.InfoLevel,
    }
    
    if isDevelopment {
        config.Development = true  // Console output
        config.Level = logx.DebugLevel
    } else {
        config.Development = false // JSON output
        config.OutputPath = "/var/log/app.log"
    }
    
    return logx.Init(config)
}
```

## Best Practices

### 1. Initialize Early
```go
func main() {
    // Initialize logger as early as possible
    if err := logx.InitDefault(); err != nil {
        panic(err)
    }
    defer logx.Sync()
    
    // Rest of your application
}
```

### 2. Use Structured Fields
```go
// Good: Structured logging
logx.Info("User logged in", 
    logx.String("user_id", userID),
    logx.String("ip_address", ipAddress),
    logx.String("user_agent", userAgent),
)

// Avoid: String concatenation
logx.Info("User " + userID + " logged in from " + ipAddress)
```

### 3. Appropriate Log Levels
```go
// TRACE: Very detailed debugging
logx.Trace("Entering function", logx.String("function", "processRequest"))

// DEBUG: General debugging
logx.Debug("Processing request", logx.String("request_id", reqID))

// INFO: General application flow
logx.Info("User logged in", logx.String("user_id", userID))

// WARN: Potential issues
logx.Warn("High memory usage", logx.Float64("usage_percent", 85.5))

// ERROR: Something went wrong
logx.Error("Database connection failed", logx.ErrorField(err))

// FATAL: Critical error, application will exit
logx.Fatal("Configuration file not found")
```

### 4. Contextual Logging
```go
// Create contextual loggers for different parts of your application
func (s *Server) handleRequest(w http.ResponseWriter, r *http.Request) {
    requestLogger := logx.With(
        logx.String("request_id", generateRequestID()),
        logx.String("method", r.Method),
        logx.String("path", r.URL.Path),
    )
    
    requestLogger.Info("Request started")
    defer requestLogger.Info("Request completed")
    
    // Handle request...
}
```

### 5. Error Handling
```go
// Always check for initialization errors
if err := logx.Init(config); err != nil {
    log.Fatal("Failed to initialize logger:", err)
}

// Use ErrorField for errors
if err := someOperation(); err != nil {
    logx.Error("Operation failed", logx.ErrorField(err))
}
```

## Common Patterns

### Request Logging
```go
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    start := time.Now()
    
    // Create request logger
    logger := logx.With(
        logx.String("request_id", uuid.New().String()),
        logx.String("method", r.Method),
        logx.String("path", r.URL.Path),
        logx.String("user_agent", r.UserAgent()),
    )
    
    logger.Info("Request started")
    
    // Handle request
    h.handleRequest(w, r)
    
    // Log completion
    logger.Info("Request completed",
        logx.Duration("duration", time.Since(start)),
    )
}
```

### Database Operations
```go
func (db *Database) Query(query string, args ...interface{}) (*sql.Rows, error) {
    logger := logx.With(
        logx.String("operation", "query"),
        logx.String("query", query),
    )
    
    logger.Debug("Executing database query")
    
    rows, err := db.db.Query(query, args...)
    if err != nil {
        logger.Error("Database query failed", logx.ErrorField(err))
        return nil, err
    }
    
    logger.Debug("Database query successful")
    return rows, nil
}
```

### Background Jobs
```go
func (j *JobProcessor) ProcessJob(jobID string) {
    logger := logx.With(
        logx.String("job_id", jobID),
        logx.String("processor", "background"),
    )
    
    logger.Info("Starting job processing")
    
    defer func() {
        if r := recover(); r != nil {
            logger.Error("Job processing panicked", logx.Any("panic", r))
        }
    }()
    
    // Process job...
    logger.Info("Job processing completed")
}
```

### Configuration Validation
```go
func validateConfig(config *Config) error {
    logger := logx.With(logx.String("component", "config"))
    
    if config.Port == 0 {
        logger.Error("Invalid port configuration")
        return errors.New("port must be specified")
    }
    
    if config.DatabaseURL == "" {
        logger.Error("Database URL not configured")
        return errors.New("database URL must be specified")
    }
    
    logger.Info("Configuration validated successfully")
    return nil
}
```

## Troubleshooting

### Common Issues

#### 1. Logger Not Initialized
```go
// Problem: Logs are not appearing
logx.Info("This won't appear if logger isn't initialized")

// Solution: Always initialize
if err := logx.InitDefault(); err != nil {
    panic(err)
}
```

#### 2. File Permission Issues
```go
// Problem: Cannot write to log file
config := &logx.Config{
    OutputPath: "/var/log/app.log", // May not have write permission
}

// Solution: Check permissions or use relative path
config := &logx.Config{
    OutputPath: "./logs/app.log", // Relative to current directory
}
```

#### 3. Memory Leaks
```go
// Problem: Not calling Sync() can cause memory leaks
logger, _ := logx.New(config)
// Missing: defer logger.Sync()

// Solution: Always defer Sync()
logger, _ := logx.New(config)
defer logger.Sync()
```

#### 4. Sensitive Data Exposure
```go
// Problem: Sensitive data in logs
logx.Info("API key", logx.String("key", "secret-api-key"))

// Solution: Use different field names or remove from sensitive list
logx.Info("API key", logx.String("api_key_hash", hash(apiKey)))
```

### Debug Mode
```go
// Enable debug mode for troubleshooting
config := &logx.Config{
    Level:       logx.TraceLevel,  // Most verbose
    Development: true,              // Console output
    AddCaller:   true,              // Include caller info
    AddStacktrace: true,            // Include stack traces
}

logx.Init(config)
```

### Performance Monitoring
```go
// Monitor logging performance
start := time.Now()
for i := 0; i < 1000; i++ {
    logx.Info("Performance test", logx.Int("iteration", i))
}
duration := time.Since(start)

logx.Info("Logging performance test completed",
    logx.Duration("total_time", duration),
    logx.Float64("avg_time_per_log", float64(duration)/1000),
)
```

This comprehensive usage guide covers all aspects of using the go-logx library effectively. The examples demonstrate real-world scenarios and best practices for production applications.
