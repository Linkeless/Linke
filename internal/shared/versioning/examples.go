package versioning

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	
	"linke/internal/shared/logger"
	"linke/internal/shared/response"
)

// ExampleHandlers demonstrates how to create version-specific handlers
type ExampleHandlers struct {
	logger logger.Logger
}

// NewExampleHandlers creates new example handlers
func NewExampleHandlers(log logger.Logger) *ExampleHandlers {
	return &ExampleHandlers{
		logger: log,
	}
}

// RegisterExampleRoutes demonstrates how to register version-specific routes
func (eh *ExampleHandlers) RegisterExampleRoutes(versionRouter *VersionRouter) {
	v1 := NewVersion(1, 0, 0)
	v2 := NewVersion(2, 0, 0)
	
	// User endpoints with different implementations for v1 and v2
	versionRouter.GET("/users/:id", v1, eh.GetUserV1)
	versionRouter.GET("/users/:id", v2, eh.GetUserV2)
	
	// Profile endpoints - v2 has enhanced features
	versionRouter.GET("/profile", v1, eh.GetProfileV1)
	versionRouter.GET("/profile", v2, eh.GetProfileV2)
	
	// Subscription endpoints - v2 has breaking changes
	versionRouter.GET("/subscriptions", v1, eh.GetSubscriptionsV1)
	versionRouter.GET("/subscriptions", v2, eh.GetSubscriptionsV2)
	
	// New endpoint only available in v2
	versionRouter.GET("/analytics", v2, eh.GetAnalyticsV2)
}

// GetUserV1 handles user retrieval for API v1
func (eh *ExampleHandlers) GetUserV1(c *gin.Context) {
	userID := c.Param("id")
	
	eh.logger.Info("Processing user request with v1 format",
		zap.String("user_id", userID),
		zap.String("version", "1.0.0"),
	)
	
	// V1 format: simple user data
	userData := map[string]any{
		"id":    userID,
		"name":  "John Doe",
		"email": "john@example.com",
	}
	
	response.Success(c, userData)
}

// GetUserV2 handles user retrieval for API v2
func (eh *ExampleHandlers) GetUserV2(c *gin.Context) {
	userID := c.Param("id")
	
	eh.logger.Info("Processing user request with v2 format",
		zap.String("user_id", userID),
		zap.String("version", "2.0.0"),
	)
	
	// V2 format: enhanced user data with additional fields
	userData := map[string]any{
		"user_id":    userID,  // field name changed from "id"
		"full_name":  "John Doe", // field name changed from "name"
		"email":      "john@example.com",
		"created_at": "2024-01-01T00:00:00Z", // new field
		"updated_at": "2024-01-01T00:00:00Z", // new field
		"profile": map[string]any{ // nested profile data
			"avatar_url": "https://example.com/avatar.jpg",
			"bio":        "Software Engineer",
		},
	}
	
	response.Success(c, userData)
}

// GetProfileV1 handles profile retrieval for API v1
func (eh *ExampleHandlers) GetProfileV1(c *gin.Context) {
	eh.logger.Info("Processing profile request with v1 format")
	
	profileData := map[string]any{
		"user_id": "123",
		"name":    "John Doe",
		"email":   "john@example.com",
		"plan":    "basic",
	}
	
	response.Success(c, profileData)
}

// GetProfileV2 handles profile retrieval for API v2
func (eh *ExampleHandlers) GetProfileV2(c *gin.Context) {
	eh.logger.Info("Processing profile request with v2 format")
	
	// V2 has enhanced profile with more detailed information
	profileData := map[string]any{
		"user_id":    "123",
		"full_name":  "John Doe",
		"email":      "john@example.com",
		"subscription": map[string]any{ // enhanced subscription info
			"plan_name":    "Basic Plan",
			"plan_id":      "basic",
			"status":       "active",
			"expires_at":   "2024-12-31T23:59:59Z",
			"auto_renew":   true,
			"billing_cycle": "monthly",
		},
		"preferences": map[string]any{ // new in v2
			"language":     "en",
			"timezone":     "UTC",
			"notifications": true,
		},
		"stats": map[string]any{ // new in v2
			"total_usage":   "1.2GB",
			"usage_limit":   "10GB",
			"last_login":    "2024-01-15T10:30:00Z",
		},
	}
	
	response.Success(c, profileData)
}

// GetSubscriptionsV1 handles subscription listing for API v1
func (eh *ExampleHandlers) GetSubscriptionsV1(c *gin.Context) {
	eh.logger.Info("Processing subscriptions request with v1 format")
	
	subscriptions := []map[string]any{
		{
			"id":         "sub_123",
			"plan_id":    "basic",
			"status":     "active",
			"expires_at": "2024-12-31T23:59:59Z",
		},
	}
	
	response.Success(c, subscriptions)
}

// GetSubscriptionsV2 handles subscription listing for API v2
func (eh *ExampleHandlers) GetSubscriptionsV2(c *gin.Context) {
	eh.logger.Info("Processing subscriptions request with v2 format")
	
	// V2 has breaking changes: different field names and structure
	subscriptions := []map[string]any{
		{
			"subscription_id":   "sub_123", // field name changed
			"subscription_plan": map[string]any{ // nested plan info
				"plan_id":      "basic",
				"plan_name":    "Basic Plan",
				"price":        9.99,
				"currency":     "USD",
				"billing_cycle": "monthly",
				"features": []string{
					"10GB traffic",
					"Basic support",
					"5 devices",
				},
			},
			"subscription_status": "active", // field name changed
			"expiry_date":        "2024-12-31T23:59:59Z", // field name changed
			"auto_renewal":       true, // new field
			"usage": map[string]any{ // new nested usage info
				"current_usage": "1.2GB",
				"usage_limit":   "10GB",
				"usage_percent": 12.0,
			},
			"billing_history": []map[string]any{ // new field
				{
					"invoice_id":   "inv_456",
					"amount":       9.99,
					"currency":     "USD",
					"paid_at":      "2024-01-01T00:00:00Z",
					"status":       "paid",
				},
			},
		},
	}
	
	response.Success(c, subscriptions)
}

// GetAnalyticsV2 is only available in API v2
func (eh *ExampleHandlers) GetAnalyticsV2(c *gin.Context) {
	eh.logger.Info("Processing analytics request (v2 only)")
	
	analytics := map[string]any{
		"usage_analytics": map[string]any{
			"daily_usage": []map[string]any{
				{"date": "2024-01-01", "usage": "0.5GB"},
				{"date": "2024-01-02", "usage": "0.3GB"},
				{"date": "2024-01-03", "usage": "0.4GB"},
			},
			"peak_usage_time": "18:00-20:00",
			"average_daily":   "0.4GB",
		},
		"performance_metrics": map[string]any{
			"avg_latency":    "45ms",
			"uptime":         "99.9%",
			"error_rate":     "0.01%",
		},
		"geographic_distribution": []map[string]any{
			{"country": "US", "usage_percent": 60.0},
			{"country": "CA", "usage_percent": 25.0},
			{"country": "UK", "usage_percent": 15.0},
		},
	}
	
	response.Success(c, analytics)
}

// SharedVersionHandler demonstrates a handler that works across versions with adaptation
type SharedVersionHandler struct {
	logger logger.Logger
}

// NewSharedVersionHandler creates a new shared version handler
func NewSharedVersionHandler(log logger.Logger) *SharedVersionHandler {
	return &SharedVersionHandler{
		logger: log,
	}
}

// GetServerStatus works across versions but adapts response format
func (svh *SharedVersionHandler) GetServerStatus(c *gin.Context) {
	versionCtx, exists := GetVersionFromContext(c)
	if !exists {
		response.ErrorJSON(c, http.StatusInternalServerError, response.ErrorResponse{
			Error:   "version_context_missing",
			Message: "Version context not found",
		})
		return
	}
	
	version := versionCtx.ResolvedVersion
	
	// Core data that's the same across versions
	coreStatus := map[string]any{
		"status":    "healthy",
		"timestamp": "2024-01-01T12:00:00Z",
		"uptime":    "99.9%",
	}
	
	// Adapt response based on version
	if version.Major == 1 {
		// V1 format: simple status
		svh.logger.Info("Serving server status in v1 format")
		response.Success(c, coreStatus)
	} else {
		// V2 format: enhanced status with more details
		svh.logger.Info("Serving server status in v2 format")
		enhancedStatus := map[string]any{
			"server_status":   coreStatus["status"],
			"last_checked":    coreStatus["timestamp"],
			"availability":    coreStatus["uptime"],
			"system_info": map[string]any{
				"version":     "2.1.0",
				"environment": "production",
				"region":      "us-east-1",
			},
			"metrics": map[string]any{
				"cpu_usage":    "15.5%",
				"memory_usage": "342MB",
				"disk_usage":   "45%",
			},
		}
		response.Success(c, enhancedStatus)
	}
}

// BackwardCompatibilityHandler demonstrates backward compatibility patterns
type BackwardCompatibilityHandler struct {
	logger logger.Logger
}

// NewBackwardCompatibilityHandler creates a new backward compatibility handler
func NewBackwardCompatibilityHandler(log logger.Logger) *BackwardCompatibilityHandler {
	return &BackwardCompatibilityHandler{
		logger: log,
	}
}

// GetSettings demonstrates how to maintain backward compatibility
func (bch *BackwardCompatibilityHandler) GetSettings(c *gin.Context) {
	versionCtx, exists := GetVersionFromContext(c)
	if !exists {
		response.ErrorJSON(c, http.StatusInternalServerError, response.ErrorResponse{
			Error:   "version_context_missing",
			Message: "Version context not found",
		})
		return
	}
	
	version := versionCtx.ResolvedVersion
	
	// Get the latest data format
	latestSettings := map[string]any{
		"user_preferences": map[string]any{
			"theme":              "dark",
			"language":           "en",
			"notifications":      true,
			"auto_backup":        true,
			"two_factor_enabled": false,
		},
		"account_settings": map[string]any{
			"privacy_level":      "medium",
			"data_sharing":       false,
			"marketing_emails":   true,
		},
		"subscription_settings": map[string]any{
			"auto_renew":         true,
			"billing_email":      "billing@example.com",
			"payment_method_id":  "pm_123456",
		},
	}
	
	if version.Major == 1 {
		// V1 expects flat structure
		bch.logger.Info("Adapting settings response for v1 compatibility")
		
		flatSettings := map[string]any{
			"theme":              latestSettings["user_preferences"].(map[string]any)["theme"],
			"language":           latestSettings["user_preferences"].(map[string]any)["language"],
			"notifications":      latestSettings["user_preferences"].(map[string]any)["notifications"],
			"auto_renew":         latestSettings["subscription_settings"].(map[string]any)["auto_renew"],
			"privacy_level":      latestSettings["account_settings"].(map[string]any)["privacy_level"],
		}
		
		response.Success(c, flatSettings)
	} else {
		// V2+ can handle the full nested structure
		bch.logger.Info("Serving settings in enhanced v2+ format")
		response.Success(c, latestSettings)
	}
}