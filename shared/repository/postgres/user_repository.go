package postgres

import (
	"database/sql"
	"fmt"


	"github.com/eovipmak/v-insight/shared/domain/entities"
	"github.com/eovipmak/v-insight/shared/domain/repository"
	"github.com/jmoiron/sqlx"
)

// userRepository implements the UserRepository interface using PostgreSQL
type userRepository struct {
	db *sqlx.DB
}

// NewUserRepository creates a new PostgreSQL user repository
func NewUserRepository(db *sqlx.DB) repository.UserRepository {
	return &userRepository{db: db}
}

// Create creates a new user in the database
func (r *userRepository) Create(user *entities.User) error {
	query := `
		INSERT INTO users (email, password_hash, role, created_at, updated_at)
		VALUES ($1, $2, $3, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRow(query, user.Email, user.PasswordHash, user.Role).Scan(
		&user.ID,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// GetByID retrieves a user by their ID
func (r *userRepository) GetByID(id int) (*entities.User, error) {
	user := &entities.User{}
	query := `
		SELECT id, email, password_hash, role, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	err := r.db.Get(user, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

// GetByEmail retrieves a user by their email address
func (r *userRepository) GetByEmail(email string) (*entities.User, error) {
	user := &entities.User{}
	query := `
		SELECT id, email, password_hash, role, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	err := r.db.Get(user, query, email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

// Update updates an existing user
func (r *userRepository) Update(user *entities.User) error {
	query := `
		UPDATE users
		SET email = $1, password_hash = $2, role = $3, updated_at = NOW()
		WHERE id = $4
		RETURNING updated_at
	`

	err := r.db.QueryRow(query, user.Email, user.PasswordHash, user.Role, user.ID).Scan(&user.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("user not found: %w", err)
		}
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

// Delete deletes a user by their ID
func (r *userRepository) Delete(id int) error {
	// First delete all dependent resources
	// Since we use ON DELETE CASCADE in db schema for user_id, we might not need to manually delete dependent resources if configured correctly.
	// But let's assume standard cascade or manual cleanup.
	// Given previously:
	/*
		ALTER TABLE monitors ADD CONSTRAINT fk_monitors_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
		ALTER TABLE alert_rules ADD CONSTRAINT fk_alert_rules_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
		...
	*/
	// So we can just delete the user.

	query := `DELETE FROM users WHERE id = $1`
	
	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("user not found or already deleted")
	}

	return nil
}

// List retrieves all users
func (r *userRepository) List() ([]*entities.User, error) {
	users := []*entities.User{}
	query := `
		SELECT id, email, password_hash, role, created_at, updated_at
		FROM users
		ORDER BY id
	`

	err := r.db.Select(&users, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	return users, nil
}
