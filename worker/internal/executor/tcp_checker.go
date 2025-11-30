package executor

import (
	"fmt"
	"net"
	"time"
)

// TCPCheckResult represents the result of a TCP health check
type TCPCheckResult struct {
	ResponseTime time.Duration
	Error        error
	Success      bool
}

// TCPChecker performs TCP health checks
type TCPChecker struct{}

// NewTCPChecker creates a new TCP checker
func NewTCPChecker() *TCPChecker {
	return &TCPChecker{}
}

// Check performs a TCP health check on the given host and port
// Returns response time and any error encountered
func (c *TCPChecker) Check(host string, port int, timeout time.Duration) TCPCheckResult {
	startTime := time.Now()

	// Create address string
	address := fmt.Sprintf("%s:%d", host, port)

	// Attempt TCP connection with timeout
	conn, err := net.DialTimeout("tcp", address, timeout)
	responseTime := time.Since(startTime)

	if err != nil {
		return TCPCheckResult{
			ResponseTime: responseTime,
			Error:        fmt.Errorf("TCP connection failed: %w", err),
			Success:      false,
		}
	}

	// Close connection immediately after successful connection
	conn.Close()

	return TCPCheckResult{
		ResponseTime: responseTime,
		Error:        nil,
		Success:      true,
	}
}