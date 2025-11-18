package utils

import "context"

// Context keys for storing values in request context
type contextKey string

const (
	userIDKey   contextKey = "user_id"
	tenantIDKey contextKey = "tenant_id"
)

// GetUserID retrieves the user ID from the context
// Returns 0 if not found
func GetUserID(ctx context.Context) int {
	if userID, ok := ctx.Value(userIDKey).(int); ok {
		return userID
	}
	return 0
}

// GetTenantID retrieves the tenant ID from the context
// Returns 0 if not found
func GetTenantID(ctx context.Context) int {
	if tenantID, ok := ctx.Value(tenantIDKey).(int); ok {
		return tenantID
	}
	return 0
}

// SetUserID sets the user ID in the context
func SetUserID(ctx context.Context, userID int) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

// SetTenantID sets the tenant ID in the context
func SetTenantID(ctx context.Context, tenantID int) context.Context {
	return context.WithValue(ctx, tenantIDKey, tenantID)
}
