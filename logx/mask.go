package logx

import (
	"strings"
	"sync"
)

// sensitiveKeys contains keys that should be masked
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
	sensitiveKeysMu sync.RWMutex
)

// AddSensitiveKey adds a new sensitive key to be masked
func AddSensitiveKey(key string) {
	sensitiveKeysMu.Lock()
	defer sensitiveKeysMu.Unlock()
	sensitiveKeys[strings.ToLower(key)] = true
}

// RemoveSensitiveKey removes a sensitive key from masking
func RemoveSensitiveKey(key string) {
	sensitiveKeysMu.Lock()
	defer sensitiveKeysMu.Unlock()
	delete(sensitiveKeys, strings.ToLower(key))
}

// isSensitiveKey checks if a key should be masked
func isSensitiveKey(key string) bool {
	sensitiveKeysMu.RLock()
	defer sensitiveKeysMu.RUnlock()
	return sensitiveKeys[strings.ToLower(key)]
}

// maskString masks a string value, showing only first and last characters
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

// maskSensitiveData masks sensitive data based on the field key
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
