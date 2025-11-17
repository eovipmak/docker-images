package repository

import "github.com/eovipmak/v-insight/backend/internal/domain/entities"

// UserRepository defines the interface for user data operations
type UserRepository interface {
	// Create creates a new user
	Create(user *entities.User) error

	// GetByID retrieves a user by their ID
	GetByID(id int) (*entities.User, error)

	// GetByEmail retrieves a user by their email address
	GetByEmail(email string) (*entities.User, error)

	// Update updates an existing user
	Update(user *entities.User) error
}
