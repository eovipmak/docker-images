package middleware

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// RateLimiterConfig holds rate limiting configuration
type RateLimiterConfig struct {
	// PerIP is requests per minute per IP
	PerIP int
	// PerUser is requests per hour per user
	PerUser int
}

// ipRateLimiter holds rate limiters for IPs
type ipRateLimiter struct {
	limiters map[string]*rate.Limiter
	mu       sync.RWMutex
	rate     rate.Limit
	burst    int
}

// userRateLimiter holds rate limiters for authenticated users
type userRateLimiter struct {
	limiters map[int]*rate.Limiter
	mu       sync.RWMutex
	rate     rate.Limit
	burst    int
}

// newIPRateLimiter creates a new IP-based rate limiter
func newIPRateLimiter(requestsPerMinute int) *ipRateLimiter {
	return &ipRateLimiter{
		limiters: make(map[string]*rate.Limiter),
		rate:     rate.Limit(requestsPerMinute) / 60, // Convert to per-second rate
		burst:    requestsPerMinute,
	}
}

// newUserRateLimiter creates a new user-based rate limiter
func newUserRateLimiter(requestsPerHour int) *userRateLimiter {
	return &userRateLimiter{
		limiters: make(map[int]*rate.Limiter),
		rate:     rate.Limit(requestsPerHour) / 3600, // Convert to per-second rate
		burst:    requestsPerHour / 60,                // Allow 1/60th of hourly limit as burst
	}
}

// getLimiter gets or creates a rate limiter for an IP
func (l *ipRateLimiter) getLimiter(ip string) *rate.Limiter {
	l.mu.Lock()
	defer l.mu.Unlock()

	limiter, exists := l.limiters[ip]
	if !exists {
		limiter = rate.NewLimiter(l.rate, l.burst)
		l.limiters[ip] = limiter
	}

	return limiter
}

// getLimiter gets or creates a rate limiter for a user
func (l *userRateLimiter) getLimiter(userID int) *rate.Limiter {
	l.mu.Lock()
	defer l.mu.Unlock()

	limiter, exists := l.limiters[userID]
	if !exists {
		limiter = rate.NewLimiter(l.rate, l.burst)
		l.limiters[userID] = limiter
	}

	return limiter
}

// cleanup removes old limiters periodically
func (l *ipRateLimiter) cleanup(interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			l.mu.Lock()
			for ip, limiter := range l.limiters {
				// Remove limiter if it hasn't been used recently
				if limiter.Tokens() == float64(l.burst) {
					delete(l.limiters, ip)
				}
			}
			l.mu.Unlock()
		}
	}()
}

// cleanup removes old limiters periodically
func (l *userRateLimiter) cleanup(interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			l.mu.Lock()
			for userID, limiter := range l.limiters {
				// Remove limiter if it hasn't been used recently
				if limiter.Tokens() == float64(l.burst) {
					delete(l.limiters, userID)
				}
			}
			l.mu.Unlock()
		}
	}()
}

// RateLimiter creates a rate limiting middleware
func RateLimiter(cfg RateLimiterConfig) gin.HandlerFunc {
	ipLimiter := newIPRateLimiter(cfg.PerIP)
	userLimiter := newUserRateLimiter(cfg.PerUser)

	// Start cleanup goroutines
	ipLimiter.cleanup(5 * time.Minute)
	userLimiter.cleanup(10 * time.Minute)

	return func(c *gin.Context) {
		// Get client IP
		clientIP := c.ClientIP()

		// Check IP-based rate limit first
		ipLim := ipLimiter.getLimiter(clientIP)
		if !ipLim.Allow() {
			log.Printf("[RATE_LIMIT] IP rate limit exceeded for %s on %s %s", clientIP, c.Request.Method, c.Request.URL.Path)
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "rate limit exceeded",
				"retry_after": "60s",
			})
			c.Abort()
			return
		}

		// Check user-based rate limit if authenticated
		if userIDValue, exists := c.Get("user_id"); exists {
			if userID, ok := userIDValue.(int); ok {
				userLim := userLimiter.getLimiter(userID)
				if !userLim.Allow() {
					log.Printf("[RATE_LIMIT] User rate limit exceeded for user %d on %s %s", userID, c.Request.Method, c.Request.URL.Path)
					c.JSON(http.StatusTooManyRequests, gin.H{
						"error": "rate limit exceeded",
						"retry_after": "3600s",
					})
					c.Abort()
					return
				}
			}
		}

		c.Next()
	}
}
