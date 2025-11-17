package service

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/eovipmak/v-insight/backend/internal/auth"
	"github.com/eovipmak/v-insight/backend/internal/domain/entities"
	"github.com/eovipmak/v-insight/backend/internal/domain/repository"
)

// AuthService handles authentication business logic
type AuthService struct {
	userRepo       repository.UserRepository
	tenantRepo     repository.TenantRepository
	tenantUserRepo repository.TenantUserRepository
	jwtSecret      string
}

// NewAuthService creates a new authentication service
func NewAuthService(
	userRepo repository.UserRepository,
	tenantRepo repository.TenantRepository,
	tenantUserRepo repository.TenantUserRepository,
	jwtSecret string,
) *AuthService {
	return &AuthService{
		userRepo:       userRepo,
		tenantRepo:     tenantRepo,
		tenantUserRepo: tenantUserRepo,
		jwtSecret:      jwtSecret,
	}
}

// Register creates a new user and default tenant, returns JWT token
func (s *AuthService) Register(email, password, tenantName string) (string, error) {
	// Validate inputs
	if email == "" || password == "" || tenantName == "" {
		return "", errors.New("email, password, and tenant name are required")
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
	}
	if err := s.userRepo.Create(user); err != nil {
		return "", fmt.Errorf("failed to create user: %w", err)
	}

	// Create tenant slug from name (simple slug generation)
	slug := generateSlug(tenantName)

	// Create tenant
	tenant := &entities.Tenant{
		Name:    tenantName,
		Slug:    slug,
		OwnerID: user.ID,
	}
	if err := s.tenantRepo.Create(tenant); err != nil {
		return "", fmt.Errorf("failed to create tenant: %w", err)
	}

	// Add user to tenant as owner
	tenantUser := &entities.TenantUser{
		TenantID: tenant.ID,
		UserID:   user.ID,
		Role:     "owner",
	}
	if err := s.tenantUserRepo.AddUserToTenant(tenantUser); err != nil {
		return "", fmt.Errorf("failed to add user to tenant: %w", err)
	}

	// Generate JWT token
	token, err := auth.GenerateToken(user.ID, tenant.ID, s.jwtSecret)
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

	// Get user's tenants to get the first tenant ID for the token
	tenants, err := s.tenantRepo.GetUserTenants(user.ID)
	if err != nil {
		return "", fmt.Errorf("failed to get user tenants: %w", err)
	}

	if len(tenants) == 0 {
		return "", errors.New("user has no associated tenants")
	}

	// Use the first tenant for the token
	tenantID := tenants[0].ID

	// Generate JWT token
	token, err := auth.GenerateToken(user.ID, tenantID, s.jwtSecret)
	if err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}

	return token, nil
}

// ValidateToken validates a JWT token and returns user ID and tenant ID
func (s *AuthService) ValidateToken(tokenString string) (int, int, error) {
	claims, err := auth.ValidateToken(tokenString, s.jwtSecret)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid token: %w", err)
	}

	return claims.UserID, claims.TenantID, nil
}

// generateSlug creates a URL-friendly slug from a string
func generateSlug(s string) string {
	// Convert to lowercase and replace spaces with hyphens
	slug := strings.ToLower(s)
	slug = strings.ReplaceAll(slug, " ", "-")
	// Remove special characters (keep only alphanumeric and hyphens)
	var result strings.Builder
	for _, char := range slug {
		if (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') || char == '-' {
			result.WriteRune(char)
		}
	}
	return result.String()
}
