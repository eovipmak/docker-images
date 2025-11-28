package auth

import (
	"testing"
)

func TestDemoUserPasswordHash(t *testing.T) {
	// This is the hash used in migration 000008_seed_demo_user
	demoPasswordHash := "$2a$10$HnStfiQq0siSL9fn7/DIDeV/6WxLzsK1r9/AxRgCzamqSnWi6CPza"
	demoPassword := "Password!"

	// Verify the hash matches the password
	err := VerifyPassword(demoPasswordHash, demoPassword)
	if err != nil {
		t.Errorf("Demo user password hash verification failed: %v", err)
		t.Errorf("Hash: %s", demoPasswordHash)
		t.Errorf("Password: %s", demoPassword)
		
		// Generate a new hash for reference
		newHash, genErr := HashPassword(demoPassword)
		if genErr == nil {
			t.Logf("A valid hash for '%s' would be: %s", demoPassword, newHash)
		}
	} else {
		t.Logf("✅ Demo user password hash verified successfully")
		t.Logf("   Email: test@gmail.com")
		t.Logf("   Password: %s", demoPassword)
	}
}
