package executor

import (
	"context"
	"fmt"
	"net/http"
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
func (c *HTTPChecker) CheckURL(ctx context.Context, url string, timeout time.Duration) HTTPCheckResult {
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

	// Consider 2xx and 3xx status codes as successful
	success := resp.StatusCode >= 200 && resp.StatusCode < 400

	return HTTPCheckResult{
		StatusCode:   resp.StatusCode,
		ResponseTime: responseTime,
		Error:        nil,
		Success:      success,
	}
}
