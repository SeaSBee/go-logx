// Package logx provides a structured logging library built on top of Uber's zap logger.
// It offers high-performance, structured logging with additional features like
// sensitive data masking, field-based logging, and easy configuration.
//
// The package provides both a default logger instance and the ability to create
// custom logger instances. All loggers are thread-safe and support concurrent
// logging operations.
package logx

import (
	"strings"
	"sync"
)

// sensitiveKeys contains a set of field keys that should be automatically masked
// in log output to prevent sensitive data exposure. The keys are stored in
// lowercase for case-insensitive matching.
//
// This map includes common sensitive field names such as passwords, tokens,
// API keys, personal information, and authentication data.
var sensitiveKeys = map[string]bool{
	"password":      true,
	"passwd":        true,
	"pass":          true,
	"ssn":           true,
	"token":         true,
	"apikey":        true,
	"api_key":       true,
	"secret":        true,
	"key":           true,
	"email":         true,
	"phone":         true,
	"credit_card":   true,
	"cc":            true,
	"cvv":           true,
	"pin":           true,
	"auth":          true,
	"authorization": true,
	"bearer":        true,
	"jwt":           true,
}

var (
	// sensitiveKeysMu protects concurrent access to the sensitiveKeys map
	sensitiveKeysMu sync.RWMutex
)

// AddSensitiveKey adds a new sensitive key to the list of fields that should be masked.
// The key is converted to lowercase for case-insensitive matching.
// This function is thread-safe and can be called concurrently.
//
// Use this function to add custom sensitive field names that are specific to
// your application, such as custom API keys or proprietary sensitive data.
//
// Example:
//
//	logx.AddSensitiveKey("my_custom_secret")
//	logx.Info("API call", logx.String("my_custom_secret", "sensitive_value"))
//	// Output: {"message":"API call","my_custom_secret":"s***e"}
func AddSensitiveKey(key string) {
	sensitiveKeysMu.Lock()
	defer sensitiveKeysMu.Unlock()
	sensitiveKeys[strings.ToLower(key)] = true
}

// RemoveSensitiveKey removes a sensitive key from the list of fields that should be masked.
// The key is converted to lowercase for case-insensitive matching.
// This function is thread-safe and can be called concurrently.
//
// Use this function to remove keys from the sensitive list if they are no longer
// considered sensitive in your application context.
//
// Example:
//
//	logx.RemoveSensitiveKey("email")  // Don't mask email addresses
//	logx.Info("User info", logx.String("email", "user@example.com"))
//	// Output: {"message":"User info","email":"user@example.com"}
func RemoveSensitiveKey(key string) {
	sensitiveKeysMu.Lock()
	defer sensitiveKeysMu.Unlock()
	delete(sensitiveKeys, strings.ToLower(key))
}

// isSensitiveKey checks if a key should be masked based on the sensitive keys list.
// The check is case-insensitive for better matching.
// This function is thread-safe and can be called concurrently.
//
// The function returns true if the key (or any of its variations) is in the
// sensitive keys list, false otherwise.
func isSensitiveKey(key string) bool {
	sensitiveKeysMu.RLock()
	defer sensitiveKeysMu.RUnlock()
	return sensitiveKeys[strings.ToLower(key)]
}

// maskString masks a string value by showing only the first and last characters,
// replacing the middle with asterisks. This provides a balance between
// security (hiding sensitive data) and usability (allowing some identification).
//
// Masking rules:
// - Empty strings remain empty
// - Strings of length 1-2 are replaced with "***"
// - Strings of length 3-4 show first and last character: "a***b"
// - Longer strings show first 2 and last 2 characters: "ab***cd"
//
// Example:
//
//	maskString("")        // ""
//	maskString("a")       // "***"
//	maskString("ab")      // "***"
//	maskString("abc")     // "a***c"
//	maskString("abcd")    // "a***d"
//	maskString("abcdef")  // "ab***ef"
//	maskString("password123") // "pa***23"
func maskString(value string) string {
	if len(value) == 0 {
		return ""
	}
	if len(value) <= 2 {
		return "***"
	}
	if len(value) <= 4 {
		return value[:1] + "***" + value[len(value)-1:]
	}
	return value[:2] + "***" + value[len(value)-2:]
}

// maskSensitiveData masks sensitive data based on the field key.
// If the key is in the sensitive keys list, the value is masked according to its type.
// This function is called automatically by the logger for all field values.
//
// Supported types:
// - string: masked using maskString function
// - []byte: converted to string and masked
// - other types: replaced with "***MASKED***"
//
// The function returns the original value unchanged if the key is not sensitive.
//
// Example:
//
//	maskSensitiveData("password", "secret123")     // "se***23"
//	maskSensitiveData("username", "john_doe")      // "john_doe" (not masked)
//	maskSensitiveData("token", []byte("abc123"))   // "ab***23"
//	maskSensitiveData("secret", 12345)             // "***MASKED***"
func maskSensitiveData(key string, value interface{}) interface{} {
	if !isSensitiveKey(key) {
		return value
	}

	switch v := value.(type) {
	case string:
		if v == "" {
			return v
		}
		return maskString(v)
	case []byte:
		if len(v) == 0 {
			return v
		}
		return maskString(string(v))
	default:
		// For other types, return a generic mask
		return "***MASKED***"
	}
}
