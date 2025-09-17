# Extending Go-LogX

## Table of Contents
1. [Overview](#overview)
2. [Custom Field Types](#custom-field-types)
3. [Custom Masking Strategies](#custom-masking-strategies)
4. [Custom Logger Implementations](#custom-logger-implementations)
5. [Custom Output Formats](#custom-output-formats)
6. [Custom Configuration](#custom-configuration)
7. [Plugin Architecture](#plugin-architecture)
8. [Performance Extensions](#performance-extensions)
9. [Integration Examples](#integration-examples)

## Overview

Go-LogX is designed to be extensible while maintaining its core simplicity and performance. This guide covers various ways to extend the library to meet your specific needs.

## Custom Field Types

### Creating Custom Field Constructors

You can create custom field constructors for your specific data types:

```go
package myapp

import (
    "time"
    "github.com/seasbee/go-logx"
)

// Custom field for time.Duration
func Duration(key string, value time.Duration) logx.Field {
    return logx.Field{Key: key, Value: value}
}

// Custom field for UUID
func UUID(key string, value [16]byte) logx.Field {
    return logx.Field{Key: key, Value: value}
}

// Custom field for IP addresses
func IP(key string, value net.IP) logx.Field {
    return logx.Field{Key: key, Value: value.String()}
}

// Custom field for HTTP status codes
func StatusCode(key string, value int) logx.Field {
    return logx.Field{Key: key, Value: value}
}

// Usage
func main() {
    logx.InitDefault()
    
    logx.Info("Request processed",
        Duration("response_time", 150*time.Millisecond),
        StatusCode("status", 200),
        IP("client_ip", net.ParseIP("192.168.1.1")),
    )
}
```

### Complex Object Fields

For complex objects, you can create custom serialization:

```go
// User object with custom logging
type User struct {
    ID       string
    Name     string
    Email    string
    IsActive bool
}

// Custom field for User objects
func UserField(key string, user User) logx.Field {
    return logx.Field{
        Key: key,
        Value: map[string]interface{}{
            "id":        user.ID,
            "name":      user.Name,
            "email":     user.Email,
            "is_active": user.IsActive,
        },
    }
}

// Usage
user := User{ID: "123", Name: "John Doe", Email: "john@example.com", IsActive: true}
logx.Info("User created", UserField("user", user))
```

### Structured Data Fields

For structured data like JSON or XML:

```go
// JSON field constructor
func JSON(key string, data interface{}) logx.Field {
    jsonBytes, _ := json.Marshal(data)
    return logx.Field{Key: key, Value: string(jsonBytes)}
}

// XML field constructor
func XML(key string, data interface{}) logx.Field {
    xmlBytes, _ := xml.Marshal(data)
    return logx.Field{Key: key, Value: string(xmlBytes)}
}

// Usage
data := map[string]interface{}{
    "user_id": "123",
    "action":  "login",
    "timestamp": time.Now(),
}

logx.Info("API request", JSON("payload", data))
```

## Custom Masking Strategies

### Extending Sensitive Data Detection

You can extend the sensitive data detection with custom patterns:

```go
package myapp

import (
    "regexp"
    "github.com/seasbee/go-logx"
)

// Custom sensitive key patterns
var customSensitivePatterns = []*regexp.Regexp{
    regexp.MustCompile(`(?i)credit_card`),
    regexp.MustCompile(`(?i)ssn`),
    regexp.MustCompile(`(?i)social_security`),
    regexp.MustCompile(`(?i)driver_license`),
    regexp.MustCompile(`(?i)passport`),
}

// Initialize custom sensitive patterns
func init() {
    for _, pattern := range customSensitivePatterns {
        // Add pattern-based keys to sensitive list
        logx.AddSensitiveKey(pattern.String())
    }
}

// Custom masking function for specific data types
func maskCreditCard(value string) string {
    if len(value) < 4 {
        return "***"
    }
    return value[:4] + "****" + value[len(value)-4:]
}

// Custom field constructor with automatic masking
func CreditCard(key, value string) logx.Field {
    return logx.Field{Key: key, Value: maskCreditCard(value)}
}
```

### Custom Masking for Complex Objects

```go
// Custom masking for user objects
func maskUser(user User) User {
    return User{
        ID:       user.ID,
        Name:     maskString(user.Name),
        Email:    maskEmail(user.Email),
        IsActive: user.IsActive,
    }
}

func maskString(value string) string {
    if len(value) <= 2 {
        return "***"
    }
    return value[:1] + "***" + value[len(value)-1:]
}

func maskEmail(email string) string {
    parts := strings.Split(email, "@")
    if len(parts) != 2 {
        return "***"
    }
    username := parts[0]
    domain := parts[1]
    
    if len(username) <= 1 {
        username = "***"
    } else {
        username = username[:1] + "***"
    }
    
    return username + "@" + domain
}

// Safe user field constructor
func SafeUserField(key string, user User) logx.Field {
    return logx.Field{Key: key, Value: maskUser(user)}
}
```

## Custom Logger Implementations

### Creating Specialized Loggers

You can create specialized loggers for different use cases:

```go
// Audit logger for security events
type AuditLogger struct {
    logger *logx.Logger
}

func NewAuditLogger(config *logx.Config) (*AuditLogger, error) {
    logger, err := logx.New(config)
    if err != nil {
        return nil, err
    }
    
    return &AuditLogger{logger: logger}, nil
}

func (a *AuditLogger) LogAccess(userID, resource, action string, success bool) {
    a.logger.Info("Access attempt",
        logx.String("event_type", "access"),
        logx.String("user_id", userID),
        logx.String("resource", resource),
        logx.String("action", action),
        logx.Bool("success", success),
        logx.String("timestamp", time.Now().Format(time.RFC3339)),
    )
}

func (a *AuditLogger) LogSecurityEvent(eventType, description string, severity string) {
    a.logger.Warn("Security event",
        logx.String("event_type", eventType),
        logx.String("description", description),
        logx.String("severity", severity),
        logx.String("timestamp", time.Now().Format(time.RFC3339)),
    )
}

// Usage
auditLogger, _ := NewAuditLogger(&logx.Config{
    Level:      logx.InfoLevel,
    OutputPath: "/var/log/audit.log",
})

auditLogger.LogAccess("user123", "/api/users", "GET", true)
auditLogger.LogSecurityEvent("failed_login", "Multiple failed login attempts", "high")
```

### Performance Logger

```go
// Performance logger for metrics
type PerformanceLogger struct {
    logger *logx.Logger
}

func NewPerformanceLogger(config *logx.Config) (*PerformanceLogger, error) {
    logger, err := logx.New(config)
    if err != nil {
        return nil, err
    }
    
    return &PerformanceLogger{logger: logger}, nil
}

func (p *PerformanceLogger) LogMetric(name string, value float64, tags map[string]string) {
    fields := []logx.Field{
        logx.String("metric_name", name),
        logx.Float64("value", value),
        logx.String("timestamp", time.Now().Format(time.RFC3339)),
    }
    
    for key, value := range tags {
        fields = append(fields, logx.String(key, value))
    }
    
    p.logger.Info("Metric recorded", fields...)
}

func (p *PerformanceLogger) LogLatency(operation string, duration time.Duration, success bool) {
    p.logger.Info("Operation latency",
        logx.String("operation", operation),
        Duration("duration", duration),
        logx.Bool("success", success),
    )
}

// Usage
perfLogger, _ := NewPerformanceLogger(&logx.Config{
    Level:      logx.InfoLevel,
    OutputPath: "/var/log/performance.log",
})

perfLogger.LogMetric("response_time", 150.5, map[string]string{
    "endpoint": "/api/users",
    "method":   "GET",
})

perfLogger.LogLatency("database_query", 25*time.Millisecond, true)
```

## Custom Output Formats

### Custom Encoder

You can create custom encoders for different output formats:

```go
// Custom encoder for CSV output
type CSVEncoder struct {
    writer io.Writer
}

func NewCSVEncoder(writer io.Writer) *CSVEncoder {
    return &CSVEncoder{writer: writer}
}

func (c *CSVEncoder) EncodeEntry(entry logx.LogEntry) error {
    // Convert log entry to CSV format
    csvLine := fmt.Sprintf("%s,%s,%s,%s\n",
        entry.Timestamp.Format("2006-01-02 15:04:05"),
        entry.Level,
        entry.Message,
        c.formatFields(entry.Fields),
    )
    
    _, err := c.writer.Write([]byte(csvLine))
    return err
}

func (c *CSVEncoder) formatFields(fields []logx.Field) string {
    var parts []string
    for _, field := range fields {
        parts = append(parts, fmt.Sprintf("%s=%v", field.Key, field.Value))
    }
    return strings.Join(parts, ";")
}

// Custom logger with CSV output
type CSVLogger struct {
    encoder *CSVEncoder
    logger  *logx.Logger
}

func NewCSVLogger(filePath string) (*CSVLogger, error) {
    file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    if err != nil {
        return nil, err
    }
    
    encoder := NewCSVEncoder(file)
    
    // Create custom config with file output
    config := &logx.Config{
        Level:      logx.InfoLevel,
        OutputPath: filePath,
    }
    
    logger, err := logx.New(config)
    if err != nil {
        return nil, err
    }
    
    return &CSVLogger{
        encoder: encoder,
        logger:  logger,
    }, nil
}
```

## Custom Configuration

### Environment-Specific Configuration

```go
// Configuration builder with environment support
type ConfigBuilder struct {
    config *logx.Config
}

func NewConfigBuilder() *ConfigBuilder {
    return &ConfigBuilder{
        config: logx.DefaultConfig(),
    }
}

func (b *ConfigBuilder) SetLevel(level logx.Level) *ConfigBuilder {
    b.config.Level = level
    return b
}

func (b *ConfigBuilder) SetOutputPath(path string) *ConfigBuilder {
    b.config.OutputPath = path
    return b
}

func (b *ConfigBuilder) SetDevelopment(dev bool) *ConfigBuilder {
    b.config.Development = dev
    return b
}

func (b *ConfigBuilder) Build() *logx.Config {
    return b.config
}

// Environment-aware configuration
func LoadConfigFromEnvironment() *logx.Config {
    builder := NewConfigBuilder()
    
    // Set level from environment
    if levelStr := os.Getenv("LOG_LEVEL"); levelStr != "" {
        switch strings.ToUpper(levelStr) {
        case "TRACE":
            builder.SetLevel(logx.TraceLevel)
        case "DEBUG":
            builder.SetLevel(logx.DebugLevel)
        case "INFO":
            builder.SetLevel(logx.InfoLevel)
        case "WARN":
            builder.SetLevel(logx.WarnLevel)
        case "ERROR":
            builder.SetLevel(logx.ErrorLevel)
        case "FATAL":
            builder.SetLevel(logx.FatalLevel)
        }
    }
    
    // Set output path from environment
    if outputPath := os.Getenv("LOG_OUTPUT_PATH"); outputPath != "" {
        builder.SetOutputPath(outputPath)
    }
    
    // Set development mode from environment
    if dev := os.Getenv("LOG_DEVELOPMENT"); dev == "true" {
        builder.SetDevelopment(true)
    }
    
    return builder.Build()
}

// Usage
config := LoadConfigFromEnvironment()
logx.Init(config)
```

## Plugin Architecture

### Creating Plugins

You can create plugins to extend functionality:

```go
// Plugin interface
type Plugin interface {
    Name() string
    Initialize(config map[string]interface{}) error
    ProcessEntry(entry *logx.LogEntry) error
    Shutdown() error
}

// Metrics plugin
type MetricsPlugin struct {
    metrics map[string]int64
    mutex   sync.RWMutex
}

func NewMetricsPlugin() *MetricsPlugin {
    return &MetricsPlugin{
        metrics: make(map[string]int64),
    }
}

func (m *MetricsPlugin) Name() string {
    return "metrics"
}

func (m *MetricsPlugin) Initialize(config map[string]interface{}) error {
    // Initialize metrics collection
    return nil
}

func (m *MetricsPlugin) ProcessEntry(entry *logx.LogEntry) error {
    m.mutex.Lock()
    defer m.mutex.Unlock()
    
    // Count log entries by level
    m.metrics[entry.Level.String()]++
    
    return nil
}

func (m *MetricsPlugin) Shutdown() error {
    // Export final metrics
    return nil
}

func (m *MetricsPlugin) GetMetrics() map[string]int64 {
    m.mutex.RLock()
    defer m.mutex.RUnlock()
    
    result := make(map[string]int64)
    for k, v := range m.metrics {
        result[k] = v
    }
    return result
}

// Plugin manager
type PluginManager struct {
    plugins []Plugin
}

func NewPluginManager() *PluginManager {
    return &PluginManager{
        plugins: make([]Plugin, 0),
    }
}

func (pm *PluginManager) RegisterPlugin(plugin Plugin) {
    pm.plugins = append(pm.plugins, plugin)
}

func (pm *PluginManager) InitializePlugins(configs map[string]map[string]interface{}) error {
    for _, plugin := range pm.plugins {
        if config, exists := configs[plugin.Name()]; exists {
            if err := plugin.Initialize(config); err != nil {
                return fmt.Errorf("failed to initialize plugin %s: %w", plugin.Name(), err)
            }
        }
    }
    return nil
}

func (pm *PluginManager) ProcessEntry(entry *logx.LogEntry) error {
    for _, plugin := range pm.plugins {
        if err := plugin.ProcessEntry(entry); err != nil {
            return fmt.Errorf("plugin %s failed to process entry: %w", plugin.Name(), err)
        }
    }
    return nil
}

func (pm *PluginManager) Shutdown() error {
    for _, plugin := range pm.plugins {
        if err := plugin.Shutdown(); err != nil {
            return fmt.Errorf("plugin %s failed to shutdown: %w", plugin.Name(), err)
        }
    }
    return nil
}
```

## Performance Extensions

### Object Pooling

For high-performance applications, you can implement object pooling:

```go
// Field pool for reducing allocations
type FieldPool struct {
    pool sync.Pool
}

func NewFieldPool() *FieldPool {
    return &FieldPool{
        pool: sync.Pool{
            New: func() interface{} {
                return &logx.Field{}
            },
        },
    }
}

func (fp *FieldPool) Get() *logx.Field {
    return fp.pool.Get().(*logx.Field)
}

func (fp *FieldPool) Put(field *logx.Field) {
    // Reset field
    field.Key = ""
    field.Value = nil
    fp.pool.Put(field)
}

// High-performance logger with pooling
type PooledLogger struct {
    logger    *logx.Logger
    fieldPool *FieldPool
}

func NewPooledLogger(config *logx.Config) (*PooledLogger, error) {
    logger, err := logx.New(config)
    if err != nil {
        return nil, err
    }
    
    return &PooledLogger{
        logger:    logger,
        fieldPool: NewFieldPool(),
    }, nil
}

func (pl *PooledLogger) InfoWithPooledFields(msg string, fields ...*logx.Field) {
    // Convert pooled fields to regular fields
    regularFields := make([]logx.Field, len(fields))
    for i, field := range fields {
        regularFields[i] = *field
    }
    
    pl.logger.Info(msg, regularFields...)
    
    // Return fields to pool
    for _, field := range fields {
        pl.fieldPool.Put(field)
    }
}

// Usage
pooledLogger, _ := NewPooledLogger(&logx.Config{Level: logx.InfoLevel})

// Get fields from pool
userIDField := pooledLogger.fieldPool.Get()
userIDField.Key = "user_id"
userIDField.Value = "12345"

actionField := pooledLogger.fieldPool.Get()
actionField.Key = "action"
actionField.Value = "login"

// Log with pooled fields
pooledLogger.InfoWithPooledFields("User action", userIDField, actionField)
```

## Integration Examples

### HTTP Middleware

```go
// HTTP logging middleware
func LoggingMiddleware(logger *logx.Logger) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            start := time.Now()
            
            // Create request logger
            requestLogger := logger.With(
                logx.String("request_id", uuid.New().String()),
                logx.String("method", r.Method),
                logx.String("path", r.URL.Path),
                logx.String("user_agent", r.UserAgent()),
                IP("client_ip", getClientIP(r)),
            )
            
            requestLogger.Info("Request started")
            
            // Wrap response writer to capture status code
            wrappedWriter := &responseWriter{ResponseWriter: w, statusCode: 200}
            
            // Call next handler
            next.ServeHTTP(wrappedWriter, r)
            
            // Log completion
            requestLogger.Info("Request completed",
                StatusCode("status_code", wrappedWriter.statusCode),
                Duration("duration", time.Since(start)),
            )
        })
    }
}

type responseWriter struct {
    http.ResponseWriter
    statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
    rw.statusCode = code
    rw.ResponseWriter.WriteHeader(code)
}

// Usage
logger, _ := logx.New(&logx.Config{Level: logx.InfoLevel})
handler := LoggingMiddleware(logger)(yourHandler)
```

### Database Integration

```go
// Database logger wrapper
type DBLogger struct {
    db     *sql.DB
    logger *logx.Logger
}

func NewDBLogger(dsn string, logger *logx.Logger) (*DBLogger, error) {
    db, err := sql.Open("mysql", dsn)
    if err != nil {
        return nil, err
    }
    
    return &DBLogger{
        db:     db,
        logger: logger,
    }, nil
}

func (dbl *DBLogger) Query(query string, args ...interface{}) (*sql.Rows, error) {
    start := time.Now()
    
    queryLogger := dbl.logger.With(
        logx.String("operation", "query"),
        logx.String("query", query),
    )
    
    queryLogger.Debug("Executing database query")
    
    rows, err := dbl.db.Query(query, args...)
    
    if err != nil {
        queryLogger.Error("Database query failed",
            logx.ErrorField(err),
            Duration("duration", time.Since(start)),
        )
        return nil, err
    }
    
    queryLogger.Debug("Database query successful",
        Duration("duration", time.Since(start)),
    )
    
    return rows, nil
}

func (dbl *DBLogger) Exec(query string, args ...interface{}) (sql.Result, error) {
    start := time.Now()
    
    execLogger := dbl.logger.With(
        logx.String("operation", "exec"),
        logx.String("query", query),
    )
    
    execLogger.Debug("Executing database command")
    
    result, err := dbl.db.Exec(query, args...)
    
    if err != nil {
        execLogger.Error("Database command failed",
            logx.ErrorField(err),
            Duration("duration", time.Since(start)),
        )
        return nil, err
    }
    
    execLogger.Debug("Database command successful",
        Duration("duration", time.Since(start)),
    )
    
    return result, nil
}
```

This comprehensive extension guide demonstrates how to extend go-logx for various use cases while maintaining the library's core principles of simplicity and performance.
