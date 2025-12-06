package entities

import (
	"time"
)

// StatusPage represents a public status page configuration
type StatusPage struct {
	ID            string    `db:"id" json:"id"`
	UserID        int       `db:"user_id" json:"user_id"`
	Slug          string    `db:"slug" json:"slug"`
	Name          string    `db:"name" json:"name"`
	PublicEnabled bool      `db:"public_enabled" json:"public_enabled"`
	CreatedAt     time.Time `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time `db:"updated_at" json:"updated_at"`
}