# Go-LogX Architecture Documentation

## Overview

Go-LogX is a high-performance, structured logging library built on top of Uber's zap logger. It provides a clean, simple API while adding essential features like sensitive data masking, field-based logging, and easy configuration.

## Design Philosophy

### 1. **Simplicity First**
- Clean, intuitive API that's easy to learn and use
- Minimal configuration required to get started
- Sensible defaults that work for most use cases

### 2. **Performance Matters**
- Built on Uber's zap logger for maximum performance
- Zero-allocation field creation where possible
- Efficient memory usage and garbage collection

### 3. **Security by Default**
- Automatic sensitive data masking
- Configurable sensitive field detection
- No accidental data leakage in logs

### 4. **Production Ready**
- Thread-safe concurrent operations
- Structured JSON output for production
- Comprehensive error handling

## Architecture Components

### Core Components

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Package API   │    │   Logger Core   │    │   Zap Logger    │
│   (logx.go)     │───▶│  (logger.go)    │───▶│   (zap)         │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│  Field Helpers  │    │  Field Manager  │    │  Core Logger    │
│  (logger.go)    │    │  (logger.go)    │    │   (zap)         │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │
         ▼                       ▼
┌─────────────────┐    ┌─────────────────┐
│  Data Masking   │    │  Configuration  │
│   (mask.go)     │    │   (logx.go)     │
└─────────────────┘    └─────────────────┘
```

### 1. **Package API Layer** (`logx.go`)
**Purpose**: Provides the public interface for the library

**Key Design Decisions**:
- **Global Default Logger**: Single, thread-safe default logger instance
- **Package-Level Functions**: Simple API that doesn't require logger instances
- **Singleton Pattern**: Ensures consistent logging across the application

**Rationale**:
- Most applications need a single, consistent logging configuration
- Package-level functions reduce boilerplate code
- Thread-safe singleton prevents race conditions

### 2. **Logger Core** (`logger.go`)
**Purpose**: Implements the main logging functionality

**Key Design Decisions**:
- **Wrapper Pattern**: Wraps zap logger to add custom functionality
- **Field Management**: Maintains context fields across log calls
- **Thread Safety**: Uses read-write mutex for concurrent access

**Rationale**:
- Leverages zap's performance while adding custom features
- Context fields enable structured logging patterns
- RWMutex allows concurrent reads, improving performance

### 3. **Data Masking** (`mask.go`)
**Purpose**: Automatically masks sensitive data in log output

**Key Design Decisions**:
- **Automatic Detection**: Predefined list of sensitive field names
- **Configurable**: Runtime addition/removal of sensitive keys
- **Type-Aware Masking**: Different masking strategies for different types

**Rationale**:
- Prevents accidental data leakage in logs
- Configurable for application-specific needs
- Type-aware masking provides better security

## Design Patterns Used

### 1. **Wrapper Pattern**
```go
type Logger struct {
    zapLogger *zap.Logger  // Wrapped component
    fields    []Field      // Additional functionality
    mu        sync.RWMutex // Thread safety
}
```

**Benefits**:
- Reuses proven zap logger functionality
- Adds custom features without modifying zap
- Maintains zap's performance characteristics

### 2. **Builder Pattern**
```go
config := &logx.Config{
    Level: logx.DebugLevel,
    Development: true,
    AddCaller: true,
}
logger, err := logx.New(config)
```

**Benefits**:
- Flexible configuration
- Clear, readable configuration code
- Type-safe configuration

### 3. **Singleton Pattern**
```go
var (
    defaultLogger *Logger
    once          sync.Once
)
```

**Benefits**:
- Ensures single logger instance
- Thread-safe initialization
- Consistent logging across application

### 4. **Factory Pattern**
```go
func New(config *Config) (*Logger, error)
func DefaultConfig() *Config
func NewLogger() (*Logger, error)
```

**Benefits**:
- Encapsulates logger creation logic
- Provides multiple creation strategies
- Easy to extend with new configurations

## Thread Safety Design

### Concurrent Access Strategy

1. **Read-Write Mutex for Fields**
   ```go
   type Logger struct {
       fields []Field
       mu     sync.RWMutex
   }
   ```
   - Multiple readers can access fields concurrently
   - Single writer for field modifications
   - Optimized for read-heavy workloads

2. **Thread-Safe Sensitive Keys**
   ```go
   var (
       sensitiveKeys map[string]bool
       sensitiveKeysMu sync.RWMutex
   )
   ```
   - Concurrent read/write access to sensitive keys
   - Case-insensitive key matching
   - Runtime configuration changes

3. **Immutable Logger Creation**
   ```go
   func (l *Logger) With(fields ...Field) *Logger {
       // Creates new logger instance
       // Original logger unchanged
   }
   ```
   - Child loggers don't modify parent
   - Safe for concurrent use
   - No shared mutable state

## Performance Considerations

### 1. **Zero-Allocation Field Creation**
- Field structs are small and stack-allocated
- No heap allocations for simple field types
- Efficient memory usage

### 2. **Efficient Field Conversion**
```go
func (l *Logger) convertFields(fields []Field) []zap.Field {
    zapFields := make([]zap.Field, 0, len(fields))
    // Pre-allocated slice with exact capacity
}
```

### 3. **Lazy Initialization**
- Default logger initialized only when needed
- Configuration loaded once and cached
- Minimal startup overhead

### 4. **Structured Logging Benefits**
- JSON output enables efficient log parsing
- Field-based logging reduces string formatting
- Better performance than string concatenation

## Security Design

### 1. **Automatic Data Masking**
- Predefined sensitive field detection
- Runtime configuration of sensitive keys
- Type-aware masking strategies

### 2. **Masking Strategies**
```go
// String masking: "password123" → "pa***23"
// Byte masking: []byte("secret") → "se***t"
// Other types: 12345 → "***MASKED***"
```

### 3. **Case-Insensitive Matching**
- Handles variations in field naming
- Robust against naming inconsistencies
- Configurable for application needs

## Configuration Design

### 1. **Hierarchical Configuration**
```go
type Config struct {
    Level         Level  // Logging level
    OutputPath    string // Output destination
    Development   bool   // Development vs production mode
    AddCaller     bool   // Include caller information
    AddStacktrace bool   // Include stack traces
}
```

### 2. **Environment-Aware Defaults**
- **Development**: Console output, verbose formatting
- **Production**: JSON output, structured logging
- **Configurable**: All aspects can be customized

### 3. **Validation and Error Handling**
- Configuration validation on creation
- Clear error messages for invalid configs
- Graceful fallbacks for missing configs

## Extension Points

### 1. **Custom Field Types**
```go
func CustomField(key string, value CustomType) Field {
    return Field{Key: key, Value: value}
}
```

### 2. **Custom Masking Logic**
```go
func AddSensitiveKey(key string)
func RemoveSensitiveKey(key string)
```

### 3. **Custom Logger Creation**
```go
func New(config *Config) (*Logger, error)
```

## Testing Strategy

### 1. **Unit Tests**
- Individual component testing
- Mock dependencies where appropriate
- Edge case coverage

### 2. **Integration Tests**
- End-to-end functionality testing
- Configuration validation
- Performance benchmarks

### 3. **Concurrency Tests**
- Race condition detection
- Thread safety validation
- Stress testing under load

## Future Considerations

### 1. **Potential Enhancements**
- Custom masking strategies
- Log rotation and archival
- Metrics and monitoring integration
- Plugin architecture for custom features

### 2. **Backward Compatibility**
- Semantic versioning
- Deprecation warnings
- Migration guides

### 3. **Performance Optimizations**
- Object pooling for field reuse
- Async logging capabilities
- Compression and batching

## Conclusion

Go-LogX is designed with simplicity, performance, and security in mind. The architecture leverages proven patterns and technologies while adding essential features for modern applications. The modular design allows for easy extension and customization while maintaining a clean, intuitive API.

The library is production-ready and suitable for high-performance applications that require structured logging with automatic sensitive data protection.
