package utils

import (
	"context"
	"testing"
)

func TestGetUserID(t *testing.T) {
	tests := []struct {
		name     string
		ctx      context.Context
		expected int
	}{
		{
			name:     "User ID exists",
			ctx:      SetUserID(context.Background(), 123),
			expected: 123,
		},
		{
			name:     "User ID does not exist",
			ctx:      context.Background(),
			expected: 0,
		},
		{
			name:     "Wrong type in context",
			ctx:      context.WithValue(context.Background(), userIDKey, "not-an-int"),
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetUserID(tt.ctx)
			if result != tt.expected {
				t.Errorf("GetUserID() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestGetTenantID(t *testing.T) {
	tests := []struct {
		name     string
		ctx      context.Context
		expected int
	}{
		{
			name:     "Tenant ID exists",
			ctx:      SetTenantID(context.Background(), 456),
			expected: 456,
		},
		{
			name:     "Tenant ID does not exist",
			ctx:      context.Background(),
			expected: 0,
		},
		{
			name:     "Wrong type in context",
			ctx:      context.WithValue(context.Background(), tenantIDKey, "not-an-int"),
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetTenantID(tt.ctx)
			if result != tt.expected {
				t.Errorf("GetTenantID() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestSetUserID(t *testing.T) {
	ctx := context.Background()
	userID := 789

	ctx = SetUserID(ctx, userID)
	result := GetUserID(ctx)

	if result != userID {
		t.Errorf("SetUserID/GetUserID = %v, expected %v", result, userID)
	}
}

func TestSetTenantID(t *testing.T) {
	ctx := context.Background()
	tenantID := 101

	ctx = SetTenantID(ctx, tenantID)
	result := GetTenantID(ctx)

	if result != tenantID {
		t.Errorf("SetTenantID/GetTenantID = %v, expected %v", result, tenantID)
	}
}

func TestBothIDsInContext(t *testing.T) {
	ctx := context.Background()
	userID := 111
	tenantID := 222

	ctx = SetUserID(ctx, userID)
	ctx = SetTenantID(ctx, tenantID)

	if GetUserID(ctx) != userID {
		t.Errorf("GetUserID() = %v, expected %v", GetUserID(ctx), userID)
	}

	if GetTenantID(ctx) != tenantID {
		t.Errorf("GetTenantID() = %v, expected %v", GetTenantID(ctx), tenantID)
	}
}
