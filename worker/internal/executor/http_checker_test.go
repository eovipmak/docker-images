package executor

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHTTPChecker_CheckURL_Success(t *testing.T) {
	// Create a test server that returns 200 OK
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))
	defer server.Close()

	checker := NewHTTPChecker()
	ctx := context.Background()

	result := checker.CheckURL(ctx, server.URL, 5*time.Second, "")

	if !result.Success {
		t.Errorf("Expected success=true, got false")
	}
	if result.StatusCode != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", result.StatusCode)
	}
	if result.Error != nil {
		t.Errorf("Expected no error, got: %v", result.Error)
	}
	if result.ResponseTime <= 0 {
		t.Errorf("Expected positive response time, got %v", result.ResponseTime)
	}
}

func TestHTTPChecker_CheckURL_ServerError(t *testing.T) {
	// Create a test server that returns 500 Internal Server Error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	checker := NewHTTPChecker()
	ctx := context.Background()

	result := checker.CheckURL(ctx, server.URL, 5*time.Second, "")

	if result.Success {
		t.Errorf("Expected success=false for 500 error, got true")
	}
	if result.StatusCode != http.StatusInternalServerError {
		t.Errorf("Expected status code 500, got %d", result.StatusCode)
	}
}

func TestHTTPChecker_CheckURL_Redirect(t *testing.T) {
	// Create a test server that redirects
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/redirect" {
			http.Redirect(w, r, "/final", http.StatusFound)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	checker := NewHTTPChecker()
	ctx := context.Background()

	result := checker.CheckURL(ctx, server.URL+"/redirect", 5*time.Second, "")

	if !result.Success {
		t.Errorf("Expected success=true after redirect, got false")
	}
	if result.StatusCode != http.StatusOK {
		t.Errorf("Expected final status code 200, got %d", result.StatusCode)
	}
}

func TestHTTPChecker_CheckURL_Timeout(t *testing.T) {
	// Create a test server that delays response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	checker := NewHTTPChecker()
	ctx := context.Background()

	// Use a very short timeout to trigger timeout
	result := checker.CheckURL(ctx, server.URL, 50*time.Millisecond, "")

	if result.Success {
		t.Errorf("Expected success=false on timeout, got true")
	}
	if result.Error == nil {
		t.Errorf("Expected timeout error, got nil")
	}
}

func TestHTTPChecker_CheckURL_InvalidURL(t *testing.T) {
	checker := NewHTTPChecker()
	ctx := context.Background()

	result := checker.CheckURL(ctx, "://invalid-url", 5*time.Second, "")

	if result.Success {
		t.Errorf("Expected success=false for invalid URL, got true")
	}
	if result.Error == nil {
		t.Errorf("Expected error for invalid URL, got nil")
	}
}

func TestHTTPChecker_CheckURL_ContextCancellation(t *testing.T) {
	// Create a test server that delays
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	checker := NewHTTPChecker()
	ctx, cancel := context.WithCancel(context.Background())

	// Cancel context immediately
	cancel()

	result := checker.CheckURL(ctx, server.URL, 5*time.Second, "")

	if result.Success {
		t.Errorf("Expected success=false on context cancellation, got true")
	}
	if result.Error == nil {
		t.Errorf("Expected error on context cancellation, got nil")
	}
}

func TestHTTPChecker_CheckURL_UserAgent(t *testing.T) {
	// Create a test server that checks user agent
	var receivedUserAgent string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedUserAgent = r.Header.Get("User-Agent")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	checker := NewHTTPChecker()
	ctx := context.Background()

	checker.CheckURL(ctx, server.URL, 5*time.Second, "")

	if receivedUserAgent != "V-Insight-Monitor/1.0" {
		t.Errorf("Expected User-Agent 'V-Insight-Monitor/1.0', got '%s'", receivedUserAgent)
	}
}

func TestHTTPChecker_CheckURL_Keyword(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello World Application"))
	}))
	defer server.Close()

	checker := NewHTTPChecker()
	ctx := context.Background()

	// Test Match
	result := checker.CheckURL(ctx, server.URL, 5*time.Second, "World")
	if !result.Success {
		t.Errorf("Expected success for matching keyword, got error: %v", result.Error)
	}

	// Test No Match
	result = checker.CheckURL(ctx, server.URL, 5*time.Second, "Universe")
	if result.Success {
		t.Errorf("Expected failure for non-matching keyword, got success")
	}
	if result.Error == nil {
		t.Errorf("Expected error for non-matching keyword, got nil")
	}
}
