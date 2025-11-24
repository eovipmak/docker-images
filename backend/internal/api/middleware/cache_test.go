package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestCacheHeaders_GetRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(CacheHeaders())
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	cacheControl := w.Header().Get("Cache-Control")
	if cacheControl == "" {
		t.Error("Expected Cache-Control header to be set")
	}
}

func TestCacheHeaders_PostRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(CacheHeaders())
	router.POST("/api/v1/monitors", func(c *gin.Context) {
		c.JSON(http.StatusCreated, gin.H{"id": "123"})
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/monitors", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, w.Code)
	}

	cacheControl := w.Header().Get("Cache-Control")
	if cacheControl != "no-store" {
		t.Errorf("Expected Cache-Control no-store for POST, got %s", cacheControl)
	}
}

func TestCacheHeaders_SSEEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(CacheHeaders())
	router.GET("/api/v1/stream/events", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"event": "data"})
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/stream/events", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	cacheControl := w.Header().Get("Cache-Control")
	if cacheControl != "no-cache, no-store, must-revalidate" {
		t.Errorf("Expected no-cache for SSE, got %s", cacheControl)
	}
}

func TestCacheHeaders_APIEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(CacheHeaders())
	router.GET("/api/v1/monitors", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"monitors": []string{}})
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/monitors", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	cacheControl := w.Header().Get("Cache-Control")
	// Should have private cache with max-age for API endpoints
	if cacheControl == "" {
		t.Error("Expected Cache-Control header for API endpoint")
	}
}

func TestCacheHeadersWithConfig(t *testing.T) {
	gin.SetMode(gin.TestMode)

	config := CacheConfig{
		DefaultMaxAge: 10,
		HealthMaxAge:  120,
		APIMaxAge:     60,
	}

	router := gin.New()
	router.Use(CacheHeadersWithConfig(config))
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	cacheControl := w.Header().Get("Cache-Control")
	expected := "public, max-age=120"
	if cacheControl != expected {
		t.Errorf("Expected Cache-Control %s, got %s", expected, cacheControl)
	}
}

func TestETagMiddleware_SetsETag(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(ETagMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "hello"})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	etag := w.Header().Get("ETag")
	if etag == "" {
		t.Error("Expected ETag header to be set")
	}

	// ETag should be a weak validator
	if len(etag) < 4 || etag[:3] != `W/"` {
		t.Errorf("Expected weak ETag format, got %s", etag)
	}
}

func TestETagMiddleware_SkipsNonGetRequests(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(ETagMiddleware())
	router.POST("/test", func(c *gin.Context) {
		c.JSON(http.StatusCreated, gin.H{"id": "123"})
	})

	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, w.Code)
	}

	etag := w.Header().Get("ETag")
	if etag != "" {
		t.Errorf("Expected no ETag for POST request, got %s", etag)
	}
}

func TestGenerateETag(t *testing.T) {
	etag1 := generateETag("/test", 100)
	etag2 := generateETag("/test", 100)
	etag3 := generateETag("/test", 200)
	etag4 := generateETag("/other", 100)

	// Same input should produce same ETag
	if etag1 != etag2 {
		t.Errorf("Same input should produce same ETag: %s != %s", etag1, etag2)
	}

	// Different size should produce different ETag
	if etag1 == etag3 {
		t.Errorf("Different size should produce different ETag: %s == %s", etag1, etag3)
	}

	// Different path should produce different ETag
	if etag1 == etag4 {
		t.Errorf("Different path should produce different ETag: %s == %s", etag1, etag4)
	}

	// Should be weak ETag format
	if len(etag1) < 4 || etag1[:3] != `W/"` {
		t.Errorf("Expected weak ETag format, got %s", etag1)
	}
}
