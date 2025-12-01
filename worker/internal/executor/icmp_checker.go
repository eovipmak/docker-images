package executor

import (
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"time"
)

// ICMPCheckResult represents the result of an ICMP health check
type ICMPCheckResult struct {
	ResponseTime time.Duration
	Error        error
	Success      bool
}

// ICMPChecker performs ICMP health checks (ping)
type ICMPChecker struct{}

// NewICMPChecker creates a new ICMP checker
func NewICMPChecker() *ICMPChecker {
	return &ICMPChecker{}
}

// Check performs a ping check on the given host
func (c *ICMPChecker) Check(ctx context.Context, host string, timeout time.Duration) ICMPCheckResult {
	// Calculate timeout in seconds for ping command
	timeoutSec := fmt.Sprintf("%.0f", timeout.Seconds())
	if timeoutSec == "0" {
		timeoutSec = "1"
	}

	// Prepare ping command (Linux syntax)
	// -c 1: send 1 packet
	// -W: timeout in seconds
	cmd := exec.CommandContext(ctx, "ping", "-c", "1", "-W", timeoutSec, host)

	startTime := time.Now()
	output, err := cmd.CombinedOutput()
	execDuration := time.Since(startTime)

	if err != nil {
		return ICMPCheckResult{
			ResponseTime: 0,
			Error:        fmt.Errorf("ping failed: %w, output: %s", err, string(output)),
			Success:      false,
		}
	}

	// Parse RTT from output
	// Output format example: "... time=4.96 ms"
	re := regexp.MustCompile(`time=([\d\.]+) ms`)
	matches := re.FindStringSubmatch(string(output))

	var responseTime time.Duration
	if len(matches) > 1 {
		ms, err := strconv.ParseFloat(matches[1], 64)
		if err == nil {
			responseTime = time.Duration(ms * float64(time.Millisecond))
		}
	}

	// Fallback to execution duration if parsing fails or RTT is 0
	// (though execution duration includes process startup overhead)
	if responseTime == 0 {
		responseTime = execDuration
	}

	return ICMPCheckResult{
		ResponseTime: responseTime,
		Error:        nil,
		Success:      true,
	}
}
