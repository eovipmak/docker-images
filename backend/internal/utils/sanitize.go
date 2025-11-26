package utils

import (
	"html"
	"strings"
)

// SanitizeString sanitizes a string by trimming whitespace and escaping HTML
func SanitizeString(s string) string {
	// Trim leading/trailing whitespace
	s = strings.TrimSpace(s)
	
	// Escape HTML to prevent XSS
	s = html.EscapeString(s)
	
	return s
}

// ValidateStringLength validates that a string is within min and max length
func ValidateStringLength(s string, min, max int) bool {
	length := len(s)
	return length >= min && length <= max
}

// SanitizeAndValidate sanitizes a string and validates its length
func SanitizeAndValidate(s string, min, max int) (string, bool) {
	sanitized := SanitizeString(s)
	valid := ValidateStringLength(sanitized, min, max)
	return sanitized, valid
}
