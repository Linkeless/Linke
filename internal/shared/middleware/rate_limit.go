package middleware

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"linke/internal/shared/logger"
	
	"github.com/gin-gonic/gin"
)

// RateLimiter implements a simple token bucket rate limiter
type RateLimiter struct {
	clients map[string]*ClientLimiter
	mutex   sync.RWMutex
	rate    int           // requests per minute
	burst   int           // maximum burst size
	window  time.Duration // time window for rate limiting
}

// ClientLimiter represents a rate limiter for a specific client
type ClientLimiter struct {
	tokens    int
	lastSeen  time.Time
	mutex     sync.Mutex
}

// NewRateLimiter creates a new rate limiter
// rate: requests per minute
// burst: maximum burst size
func NewRateLimiter(rate, burst int) *RateLimiter {
	return &RateLimiter{
		clients: make(map[string]*ClientLimiter),
		rate:    rate,
		burst:   burst,
		window:  time.Minute,
	}
}

// Allow checks if a request should be allowed for the given client ID
func (rl *RateLimiter) Allow(clientID string) bool {
	rl.mutex.RLock()
	limiter, exists := rl.clients[clientID]
	rl.mutex.RUnlock()

	if !exists {
		rl.mutex.Lock()
		// Double-check after acquiring write lock
		if limiter, exists = rl.clients[clientID]; !exists {
			limiter = &ClientLimiter{
				tokens:   rl.burst,
				lastSeen: time.Now(),
			}
			rl.clients[clientID] = limiter
		}
		rl.mutex.Unlock()
	}

	limiter.mutex.Lock()
	defer limiter.mutex.Unlock()

	now := time.Now()
	elapsed := now.Sub(limiter.lastSeen)
	
	// Add tokens based on elapsed time
	tokensToAdd := int(elapsed.Minutes() * float64(rl.rate))
	limiter.tokens += tokensToAdd
	
	if limiter.tokens > rl.burst {
		limiter.tokens = rl.burst
	}
	
	limiter.lastSeen = now

	if limiter.tokens > 0 {
		limiter.tokens--
		return true
	}

	return false
}

// CleanupExpiredClients removes clients that haven't been seen for a while
func (rl *RateLimiter) CleanupExpiredClients() {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	cutoff := time.Now().Add(-1 * time.Hour) // Remove clients not seen in 1 hour
	
	for clientID, limiter := range rl.clients {
		limiter.mutex.Lock()
		if limiter.lastSeen.Before(cutoff) {
			delete(rl.clients, clientID)
		}
		limiter.mutex.Unlock()
	}
}

// RateLimit creates a rate limiting middleware
func RateLimit(rate, burst int) gin.HandlerFunc {
	limiter := NewRateLimiter(rate, burst)
	
	// Start cleanup goroutine
	go func() {
		ticker := time.NewTicker(10 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			limiter.CleanupExpiredClients()
		}
	}()

	return func(c *gin.Context) {
		clientID := getClientID(c)
		
		if !limiter.Allow(clientID) {
			logger.Warn("Rate limit exceeded",
				logger.String("client_ip", c.ClientIP()),
				logger.String("path", c.Request.URL.Path),
				logger.String("method", c.Request.Method),
				logger.String("client_id", clientID))
			
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded",
				"code":  "RATE_LIMIT_EXCEEDED",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// getClientID generates a client identifier for rate limiting
func getClientID(c *gin.Context) string {
	// Use IP address as the primary identifier
	clientIP := c.ClientIP()
	
	// For authenticated requests, use user ID if available
	if userID, exists := c.Get("user_id"); exists {
		return fmt.Sprintf("user:%v", userID)
	}
	
	return fmt.Sprintf("ip:%s", clientIP)
}

// Specific rate limiters for different endpoint types

// AuthRateLimit creates a rate limiter for authentication endpoints (stricter)
func AuthRateLimit() gin.HandlerFunc {
	return RateLimit(20, 5) // 20 requests per minute, burst of 5
}

// PaymentRateLimit creates a rate limiter for payment endpoints (very strict)
func PaymentRateLimit() gin.HandlerFunc {
	return RateLimit(10, 2) // 10 requests per minute, burst of 2
}

// APIRateLimit creates a rate limiter for general API endpoints
func APIRateLimit() gin.HandlerFunc {
	return RateLimit(100, 20) // 100 requests per minute, burst of 20
}

// AdminRateLimit creates a rate limiter for admin endpoints
func AdminRateLimit() gin.HandlerFunc {
	return RateLimit(200, 50) // 200 requests per minute, burst of 50
}