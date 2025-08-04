package middleware

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"linke/internal/shared/config"
	"linke/internal/shared/logger"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

// PaymentSecurityMiddleware provides security validation for payment notifications
type PaymentSecurityMiddleware struct {
	config *config.PaymentSecurityConfig
	redis  *redis.Client
}

// NewPaymentSecurityMiddleware creates a new payment security middleware
func NewPaymentSecurityMiddleware(cfg *config.PaymentSecurityConfig, redisClient *redis.Client) *PaymentSecurityMiddleware {
	return &PaymentSecurityMiddleware{
		config: cfg,
		redis:  redisClient,
	}
}

// PaymentNotifySecurityMiddleware returns a middleware that validates payment notifications
func (psm *PaymentSecurityMiddleware) PaymentNotifySecurityMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		gateway := c.Param("gateway")
		if gateway == "" {
			psm.logSecurityEvent(c, "INVALID_GATEWAY", "Missing gateway parameter", "")
			c.String(http.StatusBadRequest, "fail")
			c.Abort()
			return
		}

		// SECURITY: HTTPS requirement check
		if psm.config.RequireHTTPS && c.Request.Header.Get("X-Forwarded-Proto") != "https" && c.Request.TLS == nil {
			psm.logSecurityEvent(c, "HTTPS_REQUIRED", "HTTPS required for payment notifications", gateway)
			c.String(http.StatusForbidden, "fail")
			c.Abort()
			return
		}

		// SECURITY: IP whitelist validation
		if psm.config.EnableIPWhitelist {
			if !psm.validateIPWhitelist(c, gateway) {
				psm.logSecurityEvent(c, "IP_WHITELIST_VIOLATION", "IP not in whitelist", gateway)
				c.String(http.StatusForbidden, "fail")
				c.Abort()
				return
			}
		}

		// SECURITY: Request size validation
		if c.Request.ContentLength > psm.config.MaxRequestSize {
			psm.logSecurityEvent(c, "REQUEST_TOO_LARGE", fmt.Sprintf("Request size %d exceeds limit %d", c.Request.ContentLength, psm.config.MaxRequestSize), gateway)
			c.String(http.StatusBadRequest, "fail")
			c.Abort()
			return
		}

		// SECURITY: Replay attack protection
		if psm.config.EnableReplayProtection {
			if !psm.validateReplayProtection(c, gateway) {
				psm.logSecurityEvent(c, "REPLAY_ATTACK_DETECTED", "Duplicate or expired request", gateway)
				c.String(http.StatusBadRequest, "fail")
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

// SignatureValidationMiddleware validates payment gateway signatures
func (psm *PaymentSecurityMiddleware) SignatureValidationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		gateway := c.Param("gateway")

		if !psm.config.RequireSignature {
			c.Next()
			return
		}

		// Get raw request data for signature validation
		var requestData map[string]interface{}

		contentType := c.GetHeader("Content-Type")
		if contentType == "application/json" {
			if err := c.ShouldBindJSON(&requestData); err != nil {
				psm.logSecurityEvent(c, "INVALID_JSON", "Failed to parse JSON", gateway)
				c.String(http.StatusBadRequest, "fail")
				c.Abort()
				return
			}
		} else {
			// Parse form data
			if err := c.Request.ParseForm(); err != nil {
				psm.logSecurityEvent(c, "INVALID_FORM", "Failed to parse form data", gateway)
				c.String(http.StatusBadRequest, "fail")
				c.Abort()
				return
			}

			requestData = make(map[string]interface{})
			for key, values := range c.Request.PostForm {
				if len(values) > 0 {
					requestData[key] = values[0]
				}
			}
		}

		// Validate signature based on gateway
		if !psm.validateSignature(c, gateway, requestData) {
			psm.logSecurityEvent(c, "SIGNATURE_VALIDATION_FAILED", "Invalid signature", gateway)
			c.String(http.StatusForbidden, "fail")
			c.Abort()
			return
		}

		// Store validated data in context for later use
		c.Set("payment_request_data", requestData)
		c.Next()
	}
}

// validateIPWhitelist checks if the client IP is in the whitelist for the gateway
func (psm *PaymentSecurityMiddleware) validateIPWhitelist(c *gin.Context, gateway string) bool {
	clientIP := c.ClientIP()

	var whitelist []string
	switch gateway {
	case "epay":
		whitelist = psm.config.EpayIPWhitelist
	case "epusdt":
		whitelist = psm.config.EpusdtIPWhitelist
	default:
		return false
	}

	if len(whitelist) == 0 {
		return true // No whitelist configured, allow all
	}

	clientIPAddr := net.ParseIP(clientIP)
	if clientIPAddr == nil {
		return false
	}

	for _, allowedIP := range whitelist {
		// Support both single IP and CIDR notation
		if strings.Contains(allowedIP, "/") {
			_, ipNet, err := net.ParseCIDR(allowedIP)
			if err != nil {
				continue
			}
			if ipNet.Contains(clientIPAddr) {
				return true
			}
		} else {
			if clientIP == allowedIP {
				return true
			}
		}
	}

	return false
}

// validateReplayProtection prevents replay attacks using Redis for deduplication
func (psm *PaymentSecurityMiddleware) validateReplayProtection(c *gin.Context, gateway string) bool {
	// Create a unique request identifier
	requestID := psm.generateRequestID(c, gateway)

	// Check if request exists in Redis
	key := fmt.Sprintf("payment_notify:%s:%s", gateway, requestID)
	exists, err := psm.redis.Exists(c.Request.Context(), key).Result()
	if err != nil {
		logger.Error("Failed to check replay protection", logger.Error2("error", err))
		return false // Fail safe: reject request if we can't check
	}

	if exists > 0 {
		return false // Request already processed
	}

	// Store request ID with expiration
	expiration := time.Duration(psm.config.ReplayTimeWindowMinutes) * time.Minute
	err = psm.redis.Set(c.Request.Context(), key, time.Now().Unix(), expiration).Err()
	if err != nil {
		logger.Error("Failed to store replay protection key", logger.Error2("error", err))
		return false // Fail safe: reject request if we can't store
	}

	return true
}

// generateRequestID creates a unique identifier for the request
func (psm *PaymentSecurityMiddleware) generateRequestID(c *gin.Context, gateway string) string {
	// Include timestamp, IP, and request content hash for uniqueness
	timestamp := strconv.FormatInt(time.Now().Unix()/60, 10) // Round to minute for time window
	clientIP := c.ClientIP()

	// Create content hash (simplified - real implementation should include request body)
	contentHash := fmt.Sprintf("%s-%s-%s-%s",
		c.Request.Method,
		c.Request.URL.Path,
		clientIP,
		timestamp)

	hasher := sha256.New()
	hasher.Write([]byte(contentHash))
	return hex.EncodeToString(hasher.Sum(nil))[:16] // Use first 16 chars
}

// validateSignature validates the payment gateway signature
func (psm *PaymentSecurityMiddleware) validateSignature(c *gin.Context, gateway string, data map[string]interface{}) bool {
	var signKey string
	switch gateway {
	case "epay":
		signKey = psm.config.EpaySignKey
	case "epusdt":
		signKey = psm.config.EpusdtSignKey
	default:
		return false
	}

	if signKey == "" {
		logger.Warn("No signature key configured for gateway", logger.String("gateway", gateway))
		return false
	}

	switch gateway {
	case "epay":
		return psm.validateEpaySignature(data, signKey)
	case "epusdt":
		return psm.validateEpusdtSignature(data, signKey)
	default:
		return false
	}
}

// validateEpaySignature validates Epay gateway signature (MD5-based)
func (psm *PaymentSecurityMiddleware) validateEpaySignature(data map[string]interface{}, signKey string) bool {
	receivedSign, exists := data["sign"].(string)
	if !exists || receivedSign == "" {
		return false
	}

	// Remove sign from data for calculation
	signData := make(map[string]interface{})
	for k, v := range data {
		if k != "sign" && k != "sign_type" {
			signData[k] = v
		}
	}

	// Generate signature
	expectedSign := psm.generateEpaySignature(signData, signKey)
	return strings.EqualFold(receivedSign, expectedSign)
}

// generateEpaySignature generates Epay signature using MD5
func (psm *PaymentSecurityMiddleware) generateEpaySignature(data map[string]interface{}, signKey string) string {
	// Sort keys
	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Build query string
	var params []string
	for _, k := range keys {
		if v := data[k]; v != nil && fmt.Sprintf("%v", v) != "" {
			params = append(params, fmt.Sprintf("%s=%v", k, v))
		}
	}

	// Add key and calculate MD5
	queryString := strings.Join(params, "&") + signKey
	hasher := md5.New()
	hasher.Write([]byte(queryString))
	return strings.ToUpper(hex.EncodeToString(hasher.Sum(nil)))
}

// validateEpusdtSignature validates EPUSDT gateway signature (HMAC-SHA256)
func (psm *PaymentSecurityMiddleware) validateEpusdtSignature(data map[string]interface{}, signKey string) bool {
	receivedSign, exists := data["sign"].(string)
	if !exists || receivedSign == "" {
		return false
	}

	// Remove sign from data for calculation
	signData := make(map[string]interface{})
	for k, v := range data {
		if k != "sign" {
			signData[k] = v
		}
	}

	// Generate signature
	expectedSign := psm.generateEpusdtSignature(signData, signKey)
	return hmac.Equal([]byte(receivedSign), []byte(expectedSign))
}

// generateEpusdtSignature generates EPUSDT signature using HMAC-SHA256
func (psm *PaymentSecurityMiddleware) generateEpusdtSignature(data map[string]interface{}, signKey string) string {
	// Sort keys
	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Build query string
	var params []string
	for _, k := range keys {
		if v := data[k]; v != nil && fmt.Sprintf("%v", v) != "" {
			params = append(params, fmt.Sprintf("%s=%v", k, v))
		}
	}

	queryString := strings.Join(params, "&")

	// Calculate HMAC-SHA256
	mac := hmac.New(sha256.New, []byte(signKey))
	mac.Write([]byte(queryString))
	return hex.EncodeToString(mac.Sum(nil))
}

// logSecurityEvent logs security-related events with context
func (psm *PaymentSecurityMiddleware) logSecurityEvent(c *gin.Context, eventType, message, gateway string) {
	logger.Warn("Payment security event",
		logger.String("event_type", eventType),
		logger.String("message", message),
		logger.String("gateway", gateway),
		logger.String("client_ip", c.ClientIP()),
		logger.String("user_agent", c.GetHeader("User-Agent")),
		logger.String("path", c.Request.URL.Path),
		logger.String("method", c.Request.Method),
		logger.Int64("content_length", c.Request.ContentLength))
}
