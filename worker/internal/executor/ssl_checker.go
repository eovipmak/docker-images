package executor

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/url"
	"time"
)

// SSLCheckResult represents the result of an SSL certificate check
type SSLCheckResult struct {
	Valid      bool
	ExpiresAt  time.Time
	DaysUntil  int
	Error      error
	Issuer     string
	Subject    string
}

// SSLChecker performs SSL certificate checks
type SSLChecker struct {
	timeout time.Duration
}

// NewSSLChecker creates a new SSL checker with the specified timeout
func NewSSLChecker(timeout time.Duration) *SSLChecker {
	return &SSLChecker{
		timeout: timeout,
	}
}

// CheckSSL checks the SSL certificate for the given hostname
// It returns certificate expiry information and validity status
func (c *SSLChecker) CheckSSL(hostname string) SSLCheckResult {
	// Parse URL to extract hostname and port
	parsedURL, err := url.Parse(hostname)
	if err != nil {
		return SSLCheckResult{
			Valid: false,
			Error: fmt.Errorf("failed to parse URL: %w", err),
		}
	}

	// Extract host and port
	host := parsedURL.Hostname()
	port := parsedURL.Port()
	
	// Default to port 443 if not specified
	if port == "" {
		port = "443"
	}

	// Set up dialer with timeout
	dialer := &net.Dialer{
		Timeout: c.timeout,
	}

	// First, try to connect with proper certificate verification
	conf := &tls.Config{
		InsecureSkipVerify: false,
		ServerName:         host,
	}

	// Track whether the verified handshake succeeded
	verifiedHandshake := false
	var verificationErr error

	conn, err := tls.DialWithDialer(dialer, "tcp", net.JoinHostPort(host, port), conf)
	if err != nil {
		// Verified handshake failed - store the error
		verificationErr = err
		
		// Try again with InsecureSkipVerify to get certificate details for diagnostics
		conf.InsecureSkipVerify = true
		conn, err = tls.DialWithDialer(dialer, "tcp", net.JoinHostPort(host, port), conf)
		if err != nil {
			return SSLCheckResult{
				Valid: false,
				Error: fmt.Errorf("failed to connect: %w", err),
			}
		}
	} else {
		// Verified handshake succeeded
		verifiedHandshake = true
	}
	defer conn.Close()

	// Get connection state
	state := conn.ConnectionState()
	
	// Check if we have any certificates
	if len(state.PeerCertificates) == 0 {
		return SSLCheckResult{
			Valid: false,
			Error: fmt.Errorf("no certificates found"),
		}
	}

	// Get the leaf certificate (the server's certificate)
	cert := state.PeerCertificates[0]

	// Calculate days until expiry
	daysUntil := int(time.Until(cert.NotAfter).Hours() / 24)

	// Extract issuer and subject information
	issuer := cert.Issuer.CommonName
	subject := cert.Subject.CommonName

	// Only mark as valid if the verified handshake succeeded
	// This ensures we don't mark certificates as valid when they failed TLS verification
	if !verifiedHandshake {
		return SSLCheckResult{
			Valid:     false,
			ExpiresAt: cert.NotAfter,
			DaysUntil: daysUntil,
			Error:     verificationErr,
			Issuer:    issuer,
			Subject:   subject,
		}
	}

	// Verified handshake succeeded - certificate is valid
	return SSLCheckResult{
		Valid:     true,
		ExpiresAt: cert.NotAfter,
		DaysUntil: daysUntil,
		Error:     nil,
		Issuer:    issuer,
		Subject:   subject,
	}
}
