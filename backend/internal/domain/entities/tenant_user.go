package entities

import "time"

// TenantUser represents the many-to-many relationship between tenants and users
type TenantUser struct {
	TenantID  int       `db:"tenant_id" json:"tenant_id"`
	UserID    int       `db:"user_id" json:"user_id"`
	Role      string    `db:"role" json:"role"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}
