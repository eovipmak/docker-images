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

	// Create TLS configuration
	conf := &tls.Config{
		// We want to verify the certificate but also get details even if invalid
		InsecureSkipVerify: false,
		ServerName:         host,
	}

	// Set up dialer with timeout
	dialer := &net.Dialer{
		Timeout: c.timeout,
	}

	// Connect to the server
	conn, err := tls.DialWithDialer(dialer, "tcp", net.JoinHostPort(host, port), conf)
	if err != nil {
		// Try again with InsecureSkipVerify to get certificate details even if invalid
		conf.InsecureSkipVerify = true
		conn, err = tls.DialWithDialer(dialer, "tcp", net.JoinHostPort(host, port), conf)
		if err != nil {
			return SSLCheckResult{
				Valid: false,
				Error: fmt.Errorf("failed to connect: %w", err),
			}
		}
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

	// Check if certificate is valid
	now := time.Now()
	valid := true
	var validationErr error

	// Check if certificate has expired or not yet valid
	if now.Before(cert.NotBefore) {
		valid = false
		validationErr = fmt.Errorf("certificate not yet valid")
	} else if now.After(cert.NotAfter) {
		valid = false
		validationErr = fmt.Errorf("certificate has expired")
	}

	// Verify certificate chain
	opts := &tls.Config{
		ServerName: host,
	}
	if err := cert.VerifyHostname(opts.ServerName); err != nil {
		valid = false
		if validationErr == nil {
			validationErr = fmt.Errorf("hostname verification failed: %w", err)
		}
	}

	// Calculate days until expiry
	daysUntil := int(time.Until(cert.NotAfter).Hours() / 24)

	// Extract issuer and subject information
	issuer := cert.Issuer.CommonName
	subject := cert.Subject.CommonName

	return SSLCheckResult{
		Valid:     valid,
		ExpiresAt: cert.NotAfter,
		DaysUntil: daysUntil,
		Error:     validationErr,
		Issuer:    issuer,
		Subject:   subject,
	}
}
