package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestGzipCompression_WithAcceptEncoding(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(GzipCompression())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "hello world"})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	if w.Header().Get("Content-Encoding") != "gzip" {
		t.Errorf("Expected Content-Encoding gzip, got %s", w.Header().Get("Content-Encoding"))
	}

	if w.Header().Get("Vary") != "Accept-Encoding" {
		t.Errorf("Expected Vary Accept-Encoding, got %s", w.Header().Get("Vary"))
	}

	// Verify response is gzip compressed
	reader, err := gzip.NewReader(w.Body)
	if err != nil {
		t.Fatalf("Failed to create gzip reader: %v", err)
	}
	defer reader.Close()

	_, err = io.ReadAll(reader)
	if err != nil {
		t.Errorf("Failed to read gzip response: %v", err)
	}
}

func TestGzipCompression_WithoutAcceptEncoding(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(GzipCompression())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "hello world"})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	if w.Header().Get("Content-Encoding") != "" {
		t.Errorf("Expected no Content-Encoding, got %s", w.Header().Get("Content-Encoding"))
	}
}

func TestGzipCompression_SkipsSSEEndpoints(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(GzipCompression())
	router.GET("/stream/events", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "sse endpoint"})
	})

	req := httptest.NewRequest(http.MethodGet, "/stream/events", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	// SSE endpoints should not be compressed
	if w.Header().Get("Content-Encoding") == "gzip" {
		t.Error("SSE endpoint should not be gzip compressed")
	}
}
