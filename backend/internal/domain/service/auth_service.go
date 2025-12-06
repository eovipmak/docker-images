package service

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/eovipmak/v-insight/backend/internal/auth"
	"github.com/eovipmak/v-insight/shared/domain/entities"
	"github.com/eovipmak/v-insight/shared/domain/repository"
)

// AuthService handles authentication business logic
type AuthService struct {
	userRepo  repository.UserRepository
	jwtSecret string
}

// NewAuthService creates a new authentication service
func NewAuthService(
	userRepo repository.UserRepository,
	jwtSecret string,
) *AuthService {
	return &AuthService{
		userRepo:  userRepo,
		jwtSecret: jwtSecret,
	}
}

// Register creates a new user, returns JWT token
func (s *AuthService) Register(email, password string) (string, error) {
	// Validate inputs
	if email == "" || password == "" {
		return "", errors.New("email and password are required")
	}

	// Check if user already exists
	existingUser, err := s.userRepo.GetByEmail(email)
	if err == nil && existingUser != nil {
		return "", errors.New("user with this email already exists")
	}

	// Hash password
	hashedPassword, err := auth.HashPassword(password)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	user := &entities.User{
		Email:        email,
		PasswordHash: hashedPassword,
		Role:         "user", // Default role
	}
	if err := s.userRepo.Create(user); err != nil {
		return "", fmt.Errorf("failed to create user: %w", err)
	}

	// Generate JWT token
	token, err := auth.GenerateToken(user.ID, user.Role, s.jwtSecret)
	if err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}

	return token, nil
}

// Login authenticates a user and returns JWT token
func (s *AuthService) Login(email, password string) (string, error) {
	// Validate inputs
	if email == "" || password == "" {
		return "", errors.New("email and password are required")
	}

	// Get user by email
	user, err := s.userRepo.GetByEmail(email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", errors.New("invalid email or password")
		}
		return "", fmt.Errorf("failed to get user: %w", err)
	}

	// Verify password
	if err := auth.VerifyPassword(user.PasswordHash, password); err != nil {
		return "", errors.New("invalid email or password")
	}

	// Generate JWT token
	token, err := auth.GenerateToken(user.ID, user.Role, s.jwtSecret)
	if err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}

	return token, nil
}

// ValidateToken validates a JWT token and returns user ID and role
func (s *AuthService) ValidateToken(tokenString string) (int, string, error) {
	claims, err := auth.ValidateToken(tokenString, s.jwtSecret)
	if err != nil {
		return 0, "", fmt.Errorf("invalid token: %w", err)
	}

	return claims.UserID, claims.Role, nil
}


