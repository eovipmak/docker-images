package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

const testSecret = "test-secret-key-for-testing"

func TestGenerateToken_Success(t *testing.T) {
	userID := 123
	tenantID := 456

	token, err := GenerateToken(userID, tenantID, testSecret)

	assert.NoError(t, err)
	assert.NotEmpty(t, token)
	
	// Token should have 3 parts separated by dots (header.payload.signature)
	// This is a basic JWT format check
	assert.Contains(t, token, ".")
}

func TestGenerateToken_ZeroIDs(t *testing.T) {
	token, err := GenerateToken(0, 0, testSecret)

	assert.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestGenerateToken_EmptySecret(t *testing.T) {
	token, err := GenerateToken(1, 1, "")

	assert.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestValidateToken_Success(t *testing.T) {
	userID := 123
	tenantID := 456

	token, err := GenerateToken(userID, tenantID, testSecret)
	assert.NoError(t, err)

	claims, err := ValidateToken(token, testSecret)

	assert.NoError(t, err)
	assert.NotNil(t, claims)
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, tenantID, claims.TenantID)
}

func TestValidateToken_InvalidToken(t *testing.T) {
	invalidToken := "invalid.token.string"

	claims, err := ValidateToken(invalidToken, testSecret)

	assert.Error(t, err)
	assert.Nil(t, claims)
}

func TestValidateToken_WrongSecret(t *testing.T) {
	userID := 123
	tenantID := 456

	token, err := GenerateToken(userID, tenantID, testSecret)
	assert.NoError(t, err)

	wrongSecret := "wrong-secret"
	claims, err := ValidateToken(token, wrongSecret)

	assert.Error(t, err)
	assert.Nil(t, claims)
}

func TestValidateToken_EmptyToken(t *testing.T) {
	claims, err := ValidateToken("", testSecret)

	assert.Error(t, err)
	assert.Nil(t, claims)
}

func TestValidateToken_MalformedToken(t *testing.T) {
	malformedTokens := []string{
		"not-a-jwt",
		"header.payload",  // Missing signature
		".....",           // Just dots
		"",
		"Bearer token",    // Wrong format
	}

	for _, token := range malformedTokens {
		t.Run(token, func(t *testing.T) {
			claims, err := ValidateToken(token, testSecret)
			assert.Error(t, err)
			assert.Nil(t, claims)
		})
	}
}

func TestValidateToken_ExpiredToken(t *testing.T) {
	// Create a token with past expiration
	claims := &Claims{
		UserID:   123,
		TenantID: 456,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)), // Expired 1 hour ago
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(testSecret))
	assert.NoError(t, err)

	validatedClaims, err := ValidateToken(tokenString, testSecret)

	assert.Error(t, err)
	assert.Nil(t, validatedClaims)
}

func TestTokenRoundTrip(t *testing.T) {
	testCases := []struct {
		name     string
		userID   int
		tenantID int
	}{
		{"positive IDs", 123, 456},
		{"zero IDs", 0, 0},
		{"large IDs", 999999, 888888},
		{"negative IDs", -1, -2},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			token, err := GenerateToken(tc.userID, tc.tenantID, testSecret)
			assert.NoError(t, err)
			assert.NotEmpty(t, token)

			claims, err := ValidateToken(token, testSecret)
			assert.NoError(t, err)
			assert.NotNil(t, claims)
			assert.Equal(t, tc.userID, claims.UserID)
			assert.Equal(t, tc.tenantID, claims.TenantID)
		})
	}
}

func TestClaims_ExpirationTime(t *testing.T) {
	userID := 123
	tenantID := 456

	beforeGeneration := time.Now().UTC().Truncate(time.Second).Add(-1 * time.Second)
	token, err := GenerateToken(userID, tenantID, testSecret)
	assert.NoError(t, err)

	claims, err := ValidateToken(token, testSecret)
	assert.NoError(t, err)

	// Token should expire approximately 24 hours from now
	expectedExpiration := beforeGeneration.Add(24 * time.Hour)
	actualExpiration := claims.ExpiresAt.Time

	// Allow 1 minute tolerance for test execution time
	diff := actualExpiration.Sub(expectedExpiration)
	assert.Less(t, diff.Abs(), 1*time.Minute, "Expiration should be ~24 hours from creation")
}

func TestClaims_IssuedAt(t *testing.T) {
	userID := 123
	tenantID := 456

	beforeGeneration := time.Now().UTC().Truncate(time.Second).Add(-1 * time.Second)
	token, err := GenerateToken(userID, tenantID, testSecret)
	afterGeneration := time.Now().UTC().Truncate(time.Second).Add(1 * time.Second)
	assert.NoError(t, err)

	claims, err := ValidateToken(token, testSecret)
	assert.NoError(t, err)

	issuedAt := claims.IssuedAt.Time
	t.Logf("beforeGeneration=%s, issuedAt=%s, afterGeneration=%s", beforeGeneration.Format(time.RFC3339Nano), issuedAt.Format(time.RFC3339Nano), afterGeneration.Format(time.RFC3339Nano))
	assert.True(t, issuedAt.After(beforeGeneration) || issuedAt.Equal(beforeGeneration))
	assert.True(t, issuedAt.Before(afterGeneration) || issuedAt.Equal(afterGeneration))
}

func TestValidateToken_InvalidSigningMethod(t *testing.T) {
	// Create a token with RS256 instead of HS256
	claims := &Claims{
		UserID:   123,
		TenantID: 456,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	// Create token with "none" algorithm (insecure)
	token := jwt.NewWithClaims(jwt.SigningMethodNone, claims)
	tokenString, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
	assert.NoError(t, err)

	// Should fail validation because we expect HMAC signing
	validatedClaims, err := ValidateToken(tokenString, testSecret)
	assert.Error(t, err)
	assert.Nil(t, validatedClaims)
}
