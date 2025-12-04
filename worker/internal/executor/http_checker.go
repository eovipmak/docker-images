package executor

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// HTTPCheckResult represents the result of an HTTP health check
type HTTPCheckResult struct {
	StatusCode   int
	ResponseTime time.Duration
	Error        error
	Success      bool
}

// HTTPChecker performs HTTP health checks
type HTTPChecker struct {
	client *http.Client
}

// NewHTTPChecker creates a new HTTP checker
func NewHTTPChecker() *HTTPChecker {
	return &HTTPChecker{
		client: &http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				// Allow up to 5 redirects
				if len(via) >= 5 {
					return fmt.Errorf("stopped after 5 redirects")
				}
				return nil
			},
		},
	}
}

// CheckURL performs an HTTP health check on the given URL
// Returns status code, response time, and any error encountered
func (c *HTTPChecker) CheckURL(ctx context.Context, url string, timeout time.Duration, keyword string) HTTPCheckResult {
	return c.CheckURLWithExpectedCodes(ctx, url, timeout, keyword, nil)
}

// CheckURLWithExpectedCodes performs an HTTP health check on the given URL with custom expected status codes
// If expectedCodes is nil or empty, defaults to 2xx and 3xx status codes as successful
func (c *HTTPChecker) CheckURLWithExpectedCodes(ctx context.Context, url string, timeout time.Duration, keyword string, expectedCodes []int64) HTTPCheckResult {
	// Create a context with timeout
	checkCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Create HTTP request
	req, err := http.NewRequestWithContext(checkCtx, http.MethodGet, url, nil)
	if err != nil {
		return HTTPCheckResult{
			StatusCode:   0,
			ResponseTime: 0,
			Error:        fmt.Errorf("failed to create request: %w", err),
			Success:      false,
		}
	}

	// Set user agent
	req.Header.Set("User-Agent", "V-Insight-Monitor/1.0")

	// Measure response time
	startTime := time.Now()
	resp, err := c.client.Do(req)
	responseTime := time.Since(startTime)

	if err != nil {
		return HTTPCheckResult{
			StatusCode:   0,
			ResponseTime: responseTime,
			Error:        fmt.Errorf("HTTP request failed: %w", err),
			Success:      false,
		}
	}
	defer resp.Body.Close()

	// Determine success based on expected status codes
	var success bool
	if len(expectedCodes) > 0 {
		// Check if status code is in expected list
		success = false
		for _, code := range expectedCodes {
			if int64(resp.StatusCode) == code {
				success = true
				break
			}
		}
	} else {
		// Default: Consider 2xx and 3xx status codes as successful
		success = resp.StatusCode >= 200 && resp.StatusCode < 400
	}

	if success && keyword != "" {
		// Read body and check for keyword (limit to 1MB to prevent OOM)
		bodyReader := io.LimitReader(resp.Body, 1024*1024)
		bodyBytes, err := io.ReadAll(bodyReader)
		if err != nil {
			return HTTPCheckResult{
				StatusCode:   resp.StatusCode,
				ResponseTime: responseTime,
				Error:        fmt.Errorf("failed to read response body: %w", err),
				Success:      false,
			}
		}

		if !strings.Contains(string(bodyBytes), keyword) {
			return HTTPCheckResult{
				StatusCode:   resp.StatusCode,
				ResponseTime: responseTime,
				Error:        fmt.Errorf("keyword '%s' not found in response", keyword),
				Success:      false,
			}
		}
	}

	return HTTPCheckResult{
		StatusCode:   resp.StatusCode,
		ResponseTime: responseTime,
		Error:        nil,
		Success:      success,
	}
}
