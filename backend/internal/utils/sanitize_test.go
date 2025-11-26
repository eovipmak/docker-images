package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSanitizeString(t *testing.T) {
	t.Run("trims whitespace", func(t *testing.T) {
		result := SanitizeString("  hello world  ")
		assert.Equal(t, "hello world", result)
	})

	t.Run("escapes HTML", func(t *testing.T) {
		result := SanitizeString("<script>alert('xss')</script>")
		assert.Equal(t, "&lt;script&gt;alert(&#39;xss&#39;)&lt;/script&gt;", result)
	})

	t.Run("handles both trim and escape", func(t *testing.T) {
		result := SanitizeString("  <b>bold</b>  ")
		assert.Equal(t, "&lt;b&gt;bold&lt;/b&gt;", result)
	})

	t.Run("handles empty string", func(t *testing.T) {
		result := SanitizeString("")
		assert.Equal(t, "", result)
	})

	t.Run("handles plain text", func(t *testing.T) {
		result := SanitizeString("plain text")
		assert.Equal(t, "plain text", result)
	})
}

func TestValidateStringLength(t *testing.T) {
	t.Run("validates length within range", func(t *testing.T) {
		assert.True(t, ValidateStringLength("hello", 1, 10))
		assert.True(t, ValidateStringLength("hello", 5, 5))
	})

	t.Run("rejects length below minimum", func(t *testing.T) {
		assert.False(t, ValidateStringLength("hi", 3, 10))
	})

	t.Run("rejects length above maximum", func(t *testing.T) {
		assert.False(t, ValidateStringLength("hello world", 1, 5))
	})

	t.Run("handles empty string", func(t *testing.T) {
		assert.True(t, ValidateStringLength("", 0, 10))
		assert.False(t, ValidateStringLength("", 1, 10))
	})
}

func TestSanitizeAndValidate(t *testing.T) {
	t.Run("sanitizes and validates successfully", func(t *testing.T) {
		result, valid := SanitizeAndValidate("  hello  ", 1, 10)
		assert.True(t, valid)
		assert.Equal(t, "hello", result)
	})

	t.Run("sanitizes but fails validation", func(t *testing.T) {
		result, valid := SanitizeAndValidate("  hi  ", 5, 10)
		assert.False(t, valid)
		assert.Equal(t, "hi", result)
	})

	t.Run("sanitizes HTML and validates", func(t *testing.T) {
		result, valid := SanitizeAndValidate("<script>test</script>", 1, 100)
		assert.True(t, valid)
		assert.Contains(t, result, "&lt;script&gt;")
	})
}
