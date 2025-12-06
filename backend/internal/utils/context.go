package utils

import "context"

// Context keys for storing values in request context
type contextKey string

const (
	userIDKey contextKey = "user_id"
	roleKey   contextKey = "role"
)

// GetUserID retrieves the user ID from the context
// Returns 0 if not found
func GetUserID(ctx context.Context) int {
	if userID, ok := ctx.Value(userIDKey).(int); ok {
		return userID
	}
	return 0
}

// GetRole retrieves the user role from the context
// Returns empty string if not found
func GetRole(ctx context.Context) string {
	if role, ok := ctx.Value(roleKey).(string); ok {
		return role
	}
	return ""
}

// SetUserID sets the user ID in the context
func SetUserID(ctx context.Context, userID int) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

// SetRole sets the user role in the context
func SetRole(ctx context.Context, role string) context.Context {
	return context.WithValue(ctx, roleKey, role)
}
