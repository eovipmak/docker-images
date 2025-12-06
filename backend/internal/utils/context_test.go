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

func TestGetRole(t *testing.T) {
	tests := []struct {
		name     string
		ctx      context.Context
		expected string
	}{
		{
			name:     "Role exists",
			ctx:      SetRole(context.Background(), "admin"),
			expected: "admin",
		},
		{
			name:     "Role does not exist",
			ctx:      context.Background(),
			expected: "",
		},
		{
			name:     "Wrong type in context",
			ctx:      context.WithValue(context.Background(), roleKey, 123),
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetRole(tt.ctx)
			if result != tt.expected {
				t.Errorf("GetRole() = %v, expected %v", result, tt.expected)
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

func TestSetRole(t *testing.T) {
	ctx := context.Background()
	role := "user"

	ctx = SetRole(ctx, role)
	result := GetRole(ctx)

	if result != role {
		t.Errorf("SetRole/GetRole = %v, expected %v", result, role)
	}
}

func TestUserAndRoleInContext(t *testing.T) {
	ctx := context.Background()
	userID := 111
	role := "admin"

	ctx = SetUserID(ctx, userID)
	ctx = SetRole(ctx, role)

	if GetUserID(ctx) != userID {
		t.Errorf("GetUserID() = %v, expected %v", GetUserID(ctx), userID)
	}

	if GetRole(ctx) != role {
		t.Errorf("GetRole() = %v, expected %v", GetRole(ctx), role)
	}
}
