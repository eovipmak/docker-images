package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"golang.org/x/crypto/bcrypt"
)

// HashPassword hashes a password using bcrypt
func HashPassword(password string) (string, error) {
	// Use default cost (10)
	// Pre-hash long passwords with SHA-256 to avoid bcrypt length limit
	if len(password) > 72 {
		sum := sha256.Sum256([]byte(password))
		password = hex.EncodeToString(sum[:])
	}

	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}

// VerifyPassword verifies a password against a hash
func VerifyPassword(hashedPassword, password string) error {
	// Try raw password first
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err == nil {
		return nil
	}

	// If password is long (>72) or raw attempt failed, try SHA-256 pre-hash
	if len(password) > 72 {
		sum := sha256.Sum256([]byte(password))
		hashed := hex.EncodeToString(sum[:])
		return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(hashed))
	}

	return err
}
