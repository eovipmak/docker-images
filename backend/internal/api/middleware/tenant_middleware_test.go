package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/eovipmak/v-insight/shared/domain/entities"
	"github.com/gin-gonic/gin"
)

// mockTenantUserRepository is a mock implementation for testing
type mockTenantUserRepository struct {
	hasAccessFunc func(userID, tenantID int) (bool, error)
}

func (m *mockTenantUserRepository) AddUserToTenant(tenantUser *entities.TenantUser) error {
	return nil
}

func (m *mockTenantUserRepository) HasAccess(userID, tenantID int) (bool, error) {
	if m.hasAccessFunc != nil {
		return m.hasAccessFunc(userID, tenantID)
	}
	return true, nil
}

func TestTenantRequired_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &mockTenantUserRepository{
		hasAccessFunc: func(userID, tenantID int) (bool, error) {
			return true, nil
		},
	}

	middleware := NewTenantMiddleware(mockRepo)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		// Simulate AuthRequired middleware setting user and tenant
		c.Set("user_id", 123)
		c.Set("tenant_id", 456)
		c.Next()
	})
	router.Use(middleware.TenantRequired())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestTenantRequired_WithXTenantIDHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &mockTenantUserRepository{
		hasAccessFunc: func(userID, tenantID int) (bool, error) {
			if userID == 123 && tenantID == 789 {
				return true, nil
			}
			return false, nil
		},
	}

	middleware := NewTenantMiddleware(mockRepo)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", 123)
		c.Set("tenant_id", 456)
		c.Next()
	})
	router.Use(middleware.TenantRequired())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-Tenant-ID", "789")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d. Response: %s", http.StatusOK, w.Code, w.Body.String())
	}
}

func TestTenantRequired_NoUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &mockTenantUserRepository{}
	middleware := NewTenantMiddleware(mockRepo)

	router := gin.New()
	router.Use(middleware.TenantRequired())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestTenantRequired_NoTenantID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &mockTenantUserRepository{}
	middleware := NewTenantMiddleware(mockRepo)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", 123)
		c.Next()
	})
	router.Use(middleware.TenantRequired())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestTenantRequired_NoAccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &mockTenantUserRepository{
		hasAccessFunc: func(userID, tenantID int) (bool, error) {
			return false, nil
		},
	}

	middleware := NewTenantMiddleware(mockRepo)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", 123)
		c.Set("tenant_id", 456)
		c.Next()
	})
	router.Use(middleware.TenantRequired())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status %d, got %d", http.StatusForbidden, w.Code)
	}
}

func TestTenantRequired_RepositoryError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &mockTenantUserRepository{
		hasAccessFunc: func(userID, tenantID int) (bool, error) {
			return false, errors.New("database error")
		},
	}

	middleware := NewTenantMiddleware(mockRepo)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", 123)
		c.Set("tenant_id", 456)
		c.Next()
	})
	router.Use(middleware.TenantRequired())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestTenantRequired_InvalidXTenantIDHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &mockTenantUserRepository{}
	middleware := NewTenantMiddleware(mockRepo)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", 123)
		c.Set("tenant_id", 456)
		c.Next()
	})
	router.Use(middleware.TenantRequired())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-Tenant-ID", "invalid")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}
