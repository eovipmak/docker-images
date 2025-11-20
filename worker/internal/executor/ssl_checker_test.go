package executor

import (
	"testing"
	"time"
)

func TestSSLChecker_CheckSSL(t *testing.T) {
	checker := NewSSLChecker(10 * time.Second)

	tests := []struct {
		name     string
		hostname string
		wantErr  bool
	}{
		{
			name:     "valid HTTPS site",
			hostname: "https://www.google.com",
			wantErr:  false,
		},
		{
			name:     "invalid hostname",
			hostname: "https://this-domain-should-not-exist-12345.com",
			wantErr:  true,
		},
		{
			name:     "malformed URL",
			hostname: "not-a-valid-url",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := checker.CheckSSL(tt.hostname)
			
			if tt.wantErr && result.Error == nil {
				t.Errorf("CheckSSL() expected error but got none")
			}
			
			if !tt.wantErr {
				if result.Error != nil {
					t.Errorf("CheckSSL() unexpected error: %v", result.Error)
				}
				
				if !result.Valid {
					t.Errorf("CheckSSL() expected valid certificate but got invalid")
				}
				
				if result.ExpiresAt.IsZero() {
					t.Errorf("CheckSSL() expected non-zero expiry date")
				}
				
				// Certificate should not expire within the next day for google.com
				if result.DaysUntil < 1 {
					t.Errorf("CheckSSL() expected DaysUntil > 1, got %d", result.DaysUntil)
				}
			}
		})
	}
}
