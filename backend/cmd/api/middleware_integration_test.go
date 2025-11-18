package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/eovipmak/v-insight/backend/internal/api/handlers"
	"github.com/eovipmak/v-insight/backend/internal/api/middleware"
	"github.com/eovipmak/v-insight/backend/internal/domain/entities"
	"github.com/eovipmak/v-insight/backend/internal/domain/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
)

// Mock repositories
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(user *entities.User) error {
	args := m.Called(user)
	if args.Get(0) != nil {
		return args.Error(0)
	}
	user.ID = 1
	return nil
}

func (m *MockUserRepository) GetByID(id int) (*entities.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(email string) (*entities.User, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.User), args.Error(1)
}

func (m *MockUserRepository) Update(user *entities.User) error {
	args := m.Called(user)
	return args.Error(0)
}

type MockTenantRepository struct {
	mock.Mock
}

func (m *MockTenantRepository) Create(tenant *entities.Tenant) error {
	args := m.Called(tenant)
	if args.Get(0) != nil {
		return args.Error(0)
	}
	tenant.ID = 1
	return nil
}

func (m *MockTenantRepository) GetByID(id int) (*entities.Tenant, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Tenant), args.Error(1)
}

func (m *MockTenantRepository) GetBySlug(slug string) (*entities.Tenant, error) {
	args := m.Called(slug)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Tenant), args.Error(1)
}

func (m *MockTenantRepository) GetUserTenants(userID int) ([]*entities.Tenant, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.Tenant), args.Error(1)
}

type MockTenantUserRepository struct {
	mock.Mock
}

func (m *MockTenantUserRepository) AddUserToTenant(tenantUser *entities.TenantUser) error {
	args := m.Called(tenantUser)
	return args.Error(0)
}

func (m *MockTenantUserRepository) HasAccess(userID, tenantID int) (bool, error) {
	args := m.Called(userID, tenantID)
	return args.Bool(0), args.Error(1)
}

// TestMiddlewareChain tests the authentication and tenant middleware chain
func TestMiddlewareChain(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup mocks
	userRepo := new(MockUserRepository)
	tenantRepo := new(MockTenantRepository)
	tenantUserRepo := new(MockTenantUserRepository)

	// Initialize services
	authService := service.NewAuthService(userRepo, tenantRepo, tenantUserRepo, "test-secret")
	authHandler := handlers.NewAuthHandler(authService, userRepo)

	// Initialize middleware
	authMiddleware := middleware.NewAuthMiddleware(authService)
	tenantMiddleware := middleware.NewTenantMiddleware(tenantUserRepo)

	// Setup router
	router := gin.New()

	// Auth routes
	auth := router.Group("/auth")
	{
		auth.POST("/register", authHandler.Register)
	}

	// Protected routes
	protected := router.Group("/protected")
	protected.Use(authMiddleware.AuthRequired(), tenantMiddleware.TenantRequired())
	{
		protected.GET("/info", func(c *gin.Context) {
			tenantIDValue, _ := c.Get("tenant_id")
			c.JSON(200, gin.H{
				"message":   "success",
				"tenant_id": tenantIDValue,
			})
		})
	}

	// Test 1: Register a user
	userRepo.On("GetByEmail", "test@example.com").Return(nil, nil).Once()
	userRepo.On("Create", mock.AnythingOfType("*entities.User")).Return(nil)
	tenantRepo.On("Create", mock.AnythingOfType("*entities.Tenant")).Return(nil)
	tenantUserRepo.On("AddUserToTenant", mock.AnythingOfType("*entities.TenantUser")).Return(nil)

	registerBody := map[string]string{
		"email":       "test@example.com",
		"password":    "password123",
		"tenant_name": "Test Tenant",
	}
	registerJSON, _ := json.Marshal(registerBody)
	registerReq := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBuffer(registerJSON))
	registerReq.Header.Set("Content-Type", "application/json")
	registerW := httptest.NewRecorder()
	router.ServeHTTP(registerW, registerReq)

	if registerW.Code != http.StatusCreated {
		t.Fatalf("Register failed with status %d: %s", registerW.Code, registerW.Body.String())
	}

	// Extract token from response
	var registerResp map[string]interface{}
	json.Unmarshal(registerW.Body.Bytes(), &registerResp)
	token, ok := registerResp["token"].(string)
	if !ok || token == "" {
		t.Fatal("No token in register response")
	}

	// Test 2: Access protected route with valid token and tenant access
	tenantUserRepo.On("HasAccess", 1, 1).Return(true, nil)

	protectedReq := httptest.NewRequest(http.MethodGet, "/protected/info", nil)
	protectedReq.Header.Set("Authorization", "Bearer "+token)
	protectedW := httptest.NewRecorder()
	router.ServeHTTP(protectedW, protectedReq)

	if protectedW.Code != http.StatusOK {
		t.Errorf("Protected route failed with status %d: %s", protectedW.Code, protectedW.Body.String())
	}

	// Test 3: Access protected route with X-Tenant-ID header
	tenantUserRepo.On("HasAccess", 1, 1).Return(true, nil)

	headerReq := httptest.NewRequest(http.MethodGet, "/protected/info", nil)
	headerReq.Header.Set("Authorization", "Bearer "+token)
	headerReq.Header.Set("X-Tenant-ID", "1")
	headerW := httptest.NewRecorder()
	router.ServeHTTP(headerW, headerReq)

	if headerW.Code != http.StatusOK {
		t.Errorf("Protected route with X-Tenant-ID failed with status %d: %s", headerW.Code, headerW.Body.String())
	}

	// Test 4: Access protected route without token
	noTokenReq := httptest.NewRequest(http.MethodGet, "/protected/info", nil)
	noTokenW := httptest.NewRecorder()
	router.ServeHTTP(noTokenW, noTokenReq)

	if noTokenW.Code != http.StatusUnauthorized {
		t.Errorf("Expected unauthorized without token, got %d", noTokenW.Code)
	}

	// Test 5: Access protected route with no tenant access
	tenantUserRepo.On("HasAccess", 1, 2).Return(false, nil)

	noAccessReq := httptest.NewRequest(http.MethodGet, "/protected/info", nil)
	noAccessReq.Header.Set("Authorization", "Bearer "+token)
	noAccessReq.Header.Set("X-Tenant-ID", "2")
	noAccessW := httptest.NewRecorder()
	router.ServeHTTP(noAccessW, noAccessReq)

	if noAccessW.Code != http.StatusForbidden {
		t.Errorf("Expected forbidden without tenant access, got %d", noAccessW.Code)
	}
}
