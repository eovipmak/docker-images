package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHashPassword(t *testing.T) {
	password := "testpassword123"

	hash, err := HashPassword(password)

	assert.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.NotEqual(t, password, hash, "Hash should not equal plain password")
	
	// Hash should start with bcrypt prefix
	assert.Contains(t, hash, "$2a$", "Hash should be a bcrypt hash")
}

func TestHashPassword_EmptyPassword(t *testing.T) {
	hash, err := HashPassword("")

	// bcrypt can hash empty strings
	assert.NoError(t, err)
	assert.NotEmpty(t, hash)
}

func TestHashPassword_LongPassword(t *testing.T) {
	// Create a very long password (72 bytes is bcrypt limit)
	longPassword := string(make([]byte, 100))
	for i := range longPassword {
		longPassword = longPassword[:i] + "a"
	}

	hash, err := HashPassword(longPassword)

	assert.NoError(t, err)
	assert.NotEmpty(t, hash)
}

func TestVerifyPassword_Success(t *testing.T) {
	password := "testpassword123"
	hash, err := HashPassword(password)
	assert.NoError(t, err)

	err = VerifyPassword(hash, password)
	assert.NoError(t, err, "Should verify correct password")
}

func TestVerifyPassword_WrongPassword(t *testing.T) {
	password := "testpassword123"
	wrongPassword := "wrongpassword"
	
	hash, err := HashPassword(password)
	assert.NoError(t, err)

	err = VerifyPassword(hash, wrongPassword)
	assert.Error(t, err, "Should fail with wrong password")
}

func TestVerifyPassword_InvalidHash(t *testing.T) {
	password := "testpassword123"
	invalidHash := "not-a-valid-hash"

	err := VerifyPassword(invalidHash, password)
	assert.Error(t, err, "Should fail with invalid hash")
}

func TestVerifyPassword_EmptyPassword(t *testing.T) {
	password := "testpassword123"
	hash, err := HashPassword(password)
	assert.NoError(t, err)

	err = VerifyPassword(hash, "")
	assert.Error(t, err, "Should fail with empty password")
}

func TestHashPassword_Uniqueness(t *testing.T) {
	password := "testpassword123"

	hash1, err1 := HashPassword(password)
	hash2, err2 := HashPassword(password)

	assert.NoError(t, err1)
	assert.NoError(t, err2)
	
	// Due to salt, same password should produce different hashes
	assert.NotEqual(t, hash1, hash2, "Same password should produce different hashes due to salt")
	
	// But both should verify correctly
	assert.NoError(t, VerifyPassword(hash1, password))
	assert.NoError(t, VerifyPassword(hash2, password))
}

func TestHashAndVerify_RoundTrip(t *testing.T) {
	testCases := []string{
		"simple",
		"with spaces",
		"with-special-chars!@#$%^&*()",
		"123456789",
		"MixedCasePassword",
		"🔒emoji🔑password",
	}

	for _, password := range testCases {
		t.Run(password, func(t *testing.T) {
			hash, err := HashPassword(password)
			assert.NoError(t, err)
			assert.NotEmpty(t, hash)

			err = VerifyPassword(hash, password)
			assert.NoError(t, err, "Should verify password: %s", password)
		})
	}
}
