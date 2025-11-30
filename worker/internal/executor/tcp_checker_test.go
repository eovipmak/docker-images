package executor

import (
	"net"
	"testing"
	"time"
)

func TestTCPChecker_Check_Success(t *testing.T) {
	// Create a test TCP server
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to create test server: %v", err)
	}
	defer listener.Close()

	// Start server in goroutine
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			return // Test might have finished
		}
		conn.Close()
	}()

	// Get the port
	addr := listener.Addr().(*net.TCPAddr)
	host := addr.IP.String()
	port := addr.Port

	checker := NewTCPChecker()
	result := checker.Check(host, port, 5*time.Second)

	if !result.Success {
		t.Errorf("Expected success=true, got false")
	}
	if result.Error != nil {
		t.Errorf("Expected no error, got: %v", result.Error)
	}
	if result.ResponseTime <= 0 {
		t.Errorf("Expected positive response time, got %v", result.ResponseTime)
	}
}

func TestTCPChecker_Check_ConnectionRefused(t *testing.T) {
	checker := NewTCPChecker()
	result := checker.Check("127.0.0.1", 12345, 1*time.Second) // Assuming this port is not open

	if result.Success {
		t.Errorf("Expected success=false for connection refused, got true")
	}
	if result.Error == nil {
		t.Errorf("Expected error for connection refused, got nil")
	}
	if result.ResponseTime <= 0 {
		t.Errorf("Expected positive response time even for failed connections, got %v", result.ResponseTime)
	}
}

func TestTCPChecker_Check_InvalidHost(t *testing.T) {
	checker := NewTCPChecker()
	result := checker.Check("invalid.host.that.does.not.exist", 80, 1*time.Second)

	if result.Success {
		t.Errorf("Expected success=false for invalid host, got true")
	}
	if result.Error == nil {
		t.Errorf("Expected error for invalid host, got nil")
	}
}