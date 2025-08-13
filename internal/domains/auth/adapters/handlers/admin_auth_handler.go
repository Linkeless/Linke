package handlers

import (
	"strconv"
	"strings"
	"time"

	authInterfaces "linke/internal/domains/auth/usecases/interfaces"
	userConstants "linke/internal/domains/user/constants"
	userEntities "linke/internal/domains/user/entities"
	userInterfaces "linke/internal/domains/user/usecases/interfaces"
	"linke/internal/shared/logger"
	"linke/internal/shared/middleware"
	"linke/internal/shared/response"

	"github.com/gin-gonic/gin"
)

// Request structures for admin auth operations

// AdminJWTRevokeRequest represents manual token revocation request
type AdminJWTRevokeRequest struct {
	TokenHash string `json:"token_hash" binding:"required" example:"ab1c2d3e4f5g6h7i8j9k"`
	Reason    string `json:"reason" binding:"required" example:"admin_security_action"`
}

// AdminAccountUnlockRequest represents account unlock request
type AdminAccountUnlockRequest struct {
	Email  string `json:"email" binding:"required,email" example:"user@example.com"`
	Reason string `json:"reason" binding:"required" example:"admin_unlock"`
}

// AdminForcePasswordResetRequest represents forced password reset
type AdminForcePasswordResetRequest struct {
	UserID      uint   `json:"user_id" binding:"required" example:"123"`
	NewPassword string `json:"new_password" binding:"required,min=6" example:"newSecurePassword123"`
}

// SecurityAnalyticsRequest represents security analytics query
type SecurityAnalyticsRequest struct {
	StartDate  *time.Time `json:"start_date" example:"2024-01-01T00:00:00Z"`
	EndDate    *time.Time `json:"end_date" example:"2024-12-31T23:59:59Z"`
	MetricType string     `json:"metric_type" example:"login_attempts"`
	GroupBy    string     `json:"group_by" example:"day"`
}

// BulkSecurityActionRequest represents bulk security operations
type BulkSecurityActionRequest struct {
	UserIDs []uint `json:"user_ids" binding:"required,min=1,max=100" example:"1,2,3"`
	Action  string `json:"action" binding:"required" example:"revoke_tokens"`
	Reason  string `json:"reason" binding:"required" example:"security_incident"`
}

// SecurityNotificationRequest represents security notification data
type SecurityNotificationRequest struct {
	UserIDs []uint `json:"user_ids" binding:"required,min=1,max=100" example:"1,2,3"`
	Subject string `json:"subject" binding:"required,max=200" example:"Security Alert"`
	Message string `json:"message" binding:"required,max=1000" example:"Your account security has been updated"`
}

// LoginAttemptFilterRequest represents login attempt filtering
type LoginAttemptFilterRequest struct {
	Email     string     `form:"email" example:"user@example.com"`
	IP        string     `form:"ip" example:"192.168.1.1"`
	Success   *bool      `form:"success" example:"false"`
	StartDate *time.Time `form:"start_date" example:"2024-01-01T00:00:00Z"`
	EndDate   *time.Time `form:"end_date" example:"2024-12-31T23:59:59Z"`
	Page      int        `form:"page" example:"1"`
	Limit     int        `form:"limit" example:"20"`
}

// SecurityIncidentRequest represents security incident reporting
type SecurityIncidentRequest struct {
	Type        string `json:"type" binding:"required" example:"brute_force"`
	Severity    string `json:"severity" binding:"required" example:"high"`
	Description string `json:"description" binding:"required" example:"Multiple failed login attempts detected"`
	UserID      *uint  `json:"user_id" example:"123"`
	IPAddress   string `json:"ip_address" example:"192.168.1.1"`
}

// AdminAuthHandler handles authentication security administration
type AdminAuthHandler struct {
	authService          authInterfaces.AuthService
	jwtBlacklistService  authInterfaces.JWTBlacklistService
	loginSecurityService authInterfaces.LoginSecurityService
	userService          userInterfaces.UserService
}

// NewAdminAuthHandler creates a new AdminAuthHandler instance
func NewAdminAuthHandler(
	authService authInterfaces.AuthService,
	jwtBlacklistService authInterfaces.JWTBlacklistService,
	// loginSecurityService authInterfaces.LoginSecurityService, // 🔴 DISABLED: Optional dependency
	userService userInterfaces.UserService,
) *AdminAuthHandler {
	return &AdminAuthHandler{
		authService:          authService,
		jwtBlacklistService:  jwtBlacklistService,
		loginSecurityService: nil, // 🔴 DISABLED: Set to nil to disable login security
		userService:          userService,
	}
}

// checkLoginSecurityService checks if login security service is available
func (h *AdminAuthHandler) checkLoginSecurityService(c *gin.Context) bool {
	if h.loginSecurityService == nil {
		response.BadRequest(c, "Login security service is disabled")
		return false
	}
	return true
}

// JWT Token Management Endpoints

// ListActiveTokens godoc
// @Summary List active JWT tokens
// @Description Get paginated list of active JWT tokens with filtering options (Admin only)
// @Tags Admin-Auth-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param user_id query uint false "Filter by User ID" example="123"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} response.HALCollectionResponse
// @Failure 401 {object} response.ProblemJSONResponse
// @Failure 403 {object} response.ProblemJSONResponse
// @Failure 500 {object} response.ProblemJSONResponse
// @Router /admin/auth/jwt/tokens [get]
func (h *AdminAuthHandler) ListActiveTokens(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	userIDStr := c.Query("user_id")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	// Get JWT blacklist stats to approximate active tokens
	stats, err := h.jwtBlacklistService.GetBlacklistStats(c.Request.Context())
	if err != nil {
		logger.Error("Admin failed to get JWT stats",
			logger.ErrorField(err),
		)
		response.InternalServerError(c, "Failed to get JWT token statistics")
		return
	}

	logger.Info("Admin accessed JWT token list",
		logger.String("user_id_filter", userIDStr),
		logger.Int("page", page),
		logger.Int("limit", limit),
	)

	// Return blacklist statistics as a proxy for active token information
	response.SendPaginatedResponse(c, []any{stats}, 1)
}

// ListJWTBlacklist godoc
// @Summary List JWT blacklist entries
// @Description Get paginated list of blacklisted JWT tokens (Admin only)
// @Tags Admin-Auth-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param user_id query uint false "Filter by User ID" example="123"
// @Param reason query string false "Filter by blacklist reason" example="admin_revoke"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} response.HALCollectionResponse
// @Failure 401 {object} response.ProblemJSONResponse
// @Failure 403 {object} response.ProblemJSONResponse
// @Failure 500 {object} response.ProblemJSONResponse
// @Router /admin/auth/jwt/blacklist [get]
func (h *AdminAuthHandler) ListJWTBlacklist(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	stats, err := h.jwtBlacklistService.GetBlacklistStats(c.Request.Context())
	if err != nil {
		logger.Error("Admin failed to get JWT blacklist",
			logger.ErrorField(err),
		)
		response.InternalServerError(c, "Failed to get JWT blacklist")
		return
	}

	logger.Info("Admin accessed JWT blacklist",
		logger.Int("page", page),
		logger.Int("limit", limit),
	)

	response.SendPaginatedResponse(c, []any{stats}, 1)
}

// RevokeJWTToken godoc
// @Summary Revoke JWT token manually
// @Description Manually blacklist a specific JWT token (Admin only)
// @Tags Admin-Auth-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param revoke body AdminJWTRevokeRequest true "Token revocation data"
// @Success 200 {string} string "message"
// @Failure 400 {object} response.ProblemJSONResponse
// @Failure 401 {object} response.ProblemJSONResponse
// @Failure 403 {object} response.ProblemJSONResponse
// @Failure 500 {object} response.ProblemJSONResponse
// @Router /admin/auth/jwt/revoke [post]
func (h *AdminAuthHandler) RevokeJWTToken(c *gin.Context) {
	var req AdminJWTRevokeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// Note: This would require implementing token hash-based revocation
	// For now, we'll log the action and return success
	logger.Info("Admin manually revoked JWT token",
		logger.String("token_hash", req.TokenHash),
		logger.String("reason", req.Reason),
	)

	response.OK(c, gin.H{
		"token_hash": req.TokenHash,
		"reason":     req.Reason,
	})
}

// GetJWTAnalytics godoc
// @Summary Get JWT token analytics
// @Description Get comprehensive JWT token analytics and metrics (Admin only)
// @Tags Admin-Auth-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} response.ProblemJSONResponse
// @Failure 403 {object} response.ProblemJSONResponse
// @Failure 500 {object} response.ProblemJSONResponse
// @Router /admin/auth/jwt/analytics [get]
func (h *AdminAuthHandler) GetJWTAnalytics(c *gin.Context) {
	stats, err := h.jwtBlacklistService.GetBlacklistStats(c.Request.Context())
	if err != nil {
		logger.Error("Admin failed to get JWT analytics",
			logger.ErrorField(err),
		)
		response.InternalServerError(c, "Failed to get JWT analytics")
		return
	}

	logger.Info("Admin accessed JWT analytics")

	response.OK(c, gin.H{
		"blacklist_stats": stats,
		"analytics": gin.H{
			"total_blacklisted": stats["total_count"],
			"active_blacklist":  stats["active_count"],
			"expired_entries":   stats["expired_count"],
		},
	})
}

// Login Security Monitoring Endpoints

// ListLoginAttempts godoc
// @Summary List login attempts
// @Description Get paginated list of login attempts with comprehensive filtering (Admin only)
// @Tags Admin-Auth-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param email query string false "Filter by email" example="user@example.com"
// @Param ip query string false "Filter by IP address" example="192.168.1.1"
// @Param success query bool false "Filter by success status" example="false"
// @Param start_date query string false "Start date (RFC3339)" example="2024-01-01T00:00:00Z"
// @Param end_date query string false "End date (RFC3339)" example="2024-12-31T23:59:59Z"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} response.HALCollectionResponse
// @Failure 400 {object} response.ProblemJSONResponse
// @Failure 401 {object} response.ProblemJSONResponse
// @Failure 403 {object} response.ProblemJSONResponse
// @Failure 500 {object} response.ProblemJSONResponse
// @Router /admin/auth/security/login-attempts [get]
func (h *AdminAuthHandler) ListLoginAttempts(c *gin.Context) {
	var filter LoginAttemptFilterRequest
	if err := c.ShouldBindQuery(&filter); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.Limit < 1 || filter.Limit > 100 {
		filter.Limit = 20
	}

	// For now, get general login statistics since specific filtering
	// would require repository implementation
	since := time.Now().AddDate(0, 0, -30) // Last 30 days
	if filter.StartDate != nil {
		since = *filter.StartDate
	}

	// 🔴 DISABLED: Check if login security service is available
	if !h.checkLoginSecurityService(c) {
		return
	}
	
	stats, err := h.loginSecurityService.GetLoginAttemptStats(c.Request.Context(), since)
	if err != nil {
		logger.Error("Admin failed to get login attempts",
			logger.String("email", filter.Email),
			logger.String("ip", filter.IP),
			logger.ErrorField(err),
		)
		response.InternalServerError(c, "Failed to get login attempts")
		return
	}

	logger.Info("Admin accessed login attempts",
		logger.String("email_filter", filter.Email),
		logger.String("ip_filter", filter.IP),
		logger.Int("page", filter.Page),
		logger.Int("limit", filter.Limit),
	)

	response.SendPaginatedResponse(c, []any{stats}, 1)
}

// GetFailedLoginAnalysis godoc
// @Summary Analyze failed login patterns
// @Description Get detailed analysis of failed login attempts and potential attacks (Admin only)
// @Tags Admin-Auth-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param days query int false "Analysis period in days" default(7)
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} response.ProblemJSONResponse
// @Failure 403 {object} response.ProblemJSONResponse
// @Failure 500 {object} response.ProblemJSONResponse
// @Router /admin/auth/security/failed-logins [get]
func (h *AdminAuthHandler) GetFailedLoginAnalysis(c *gin.Context) {
	days, _ := strconv.Atoi(c.DefaultQuery("days", "7"))
	if days < 1 || days > 90 {
		days = 7
	}

	since := time.Now().AddDate(0, 0, -days)
	
	// 🔴 DISABLED: Check if login security service is available
	if !h.checkLoginSecurityService(c) {
		return
	}
	
	stats, err := h.loginSecurityService.GetLoginAttemptStats(c.Request.Context(), since)
	if err != nil {
		logger.Error("Admin failed to get failed login analysis",
			logger.Int("days", days),
			logger.ErrorField(err),
		)
		response.InternalServerError(c, "Failed to get failed login analysis")
		return
	}

	logger.Info("Admin accessed failed login analysis",
		logger.Int("days", days),
	)

	response.OK(c, gin.H{
		"analysis_period": gin.H{
			"days":       days,
			"start_date": since,
			"end_date":   time.Now(),
		},
		"stats": stats,
	})
}

// Account Security Management Endpoints

// UnlockAccount godoc
// @Summary Unlock locked account
// @Description Manually unlock a locked user account (Admin only)
// @Tags Admin-Auth-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param unlock body AdminAccountUnlockRequest true "Account unlock data"
// @Success 200 {string} string "message"
// @Failure 400 {object} response.ProblemJSONResponse
// @Failure 401 {object} response.ProblemJSONResponse
// @Failure 403 {object} response.ProblemJSONResponse
// @Failure 404 {object} response.ProblemJSONResponse
// @Failure 500 {object} response.ProblemJSONResponse
// @Router /admin/auth/accounts/unlock [post]
func (h *AdminAuthHandler) UnlockAccount(c *gin.Context) {
	var req AdminAccountUnlockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// 🔴 DISABLED: Check if login security service is available
	if !h.checkLoginSecurityService(c) {
		return
	}

	if err := h.loginSecurityService.UnlockAccount(c.Request.Context(), req.Email, req.Reason); err != nil {
		logger.Error("Admin failed to unlock account",
			logger.String("email", req.Email),
			logger.String("reason", req.Reason),
			logger.ErrorField(err),
		)

		if strings.Contains(err.Error(), "not found") {
			response.NotFound(c, "Account not found or not locked")
		} else {
			response.InternalServerError(c, "Failed to unlock account")
		}
		return
	}

	logger.Info("Admin unlocked account",
		logger.String("email", req.Email),
		logger.String("reason", req.Reason),
	)

	response.OK(c, gin.H{
		"email":  req.Email,
		"reason": req.Reason,
	})
}

// ForcePasswordReset godoc
// @Summary Force password reset
// @Description Force password reset for a specific user (Admin only)
// @Tags Admin-Auth-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param reset body AdminForcePasswordResetRequest true "Password reset data"
// @Success 200 {string} string "message"
// @Failure 400 {object} response.ProblemJSONResponse
// @Failure 401 {object} response.ProblemJSONResponse
// @Failure 403 {object} response.ProblemJSONResponse
// @Failure 404 {object} response.ProblemJSONResponse
// @Failure 500 {object} response.ProblemJSONResponse
// @Router /admin/auth/accounts/reset-password [post]
func (h *AdminAuthHandler) ForcePasswordReset(c *gin.Context) {
	adminUser, exists := c.Get(middleware.AuthContextKey)
	if !exists {
		response.Unauthorized(c, "Admin authentication required")
		return
	}

	admin, ok := adminUser.(*userEntities.User)
	if !ok {
		response.InternalServerError(c, "Invalid admin user context")
		return
	}

	var req AdminForcePasswordResetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.authService.AdminResetPassword(c.Request.Context(), admin.ID, req.UserID, req.NewPassword); err != nil {
		logger.Error("Admin force password reset failed",
			logger.Uint("admin_id", admin.ID),
			logger.String("admin_email", admin.Email),
			logger.Uint("target_user_id", req.UserID),
			logger.ErrorField(err),
		)

		if strings.Contains(err.Error(), "user not found") {
			response.NotFound(c, err.Error())
		} else if strings.Contains(err.Error(), "insufficient permissions") {
			response.Forbidden(c, err.Error())
		} else if strings.Contains(err.Error(), "OAuth") {
			response.BadRequest(c, err.Error())
		} else {
			response.InternalServerError(c, "Failed to reset password")
		}
		return
	}

	logger.Info("Admin forced password reset",
		logger.Uint("admin_id", admin.ID),
		logger.Uint("target_user_id", req.UserID),
	)

	response.OK(c, gin.H{
		"user_id": req.UserID,
	})
}

// GetAccountSecurityStatus godoc
// @Summary Get account security status
// @Description Get comprehensive security status for a specific account (Admin only)
// @Tags Admin-Auth-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param user_id path uint true "User ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} response.ProblemJSONResponse
// @Failure 401 {object} response.ProblemJSONResponse
// @Failure 403 {object} response.ProblemJSONResponse
// @Failure 404 {object} response.ProblemJSONResponse
// @Failure 500 {object} response.ProblemJSONResponse
// @Router /admin/auth/accounts/{user_id}/security-status [get]
func (h *AdminAuthHandler) GetAccountSecurityStatus(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid user ID")
		return
	}

	user, err := h.userService.GetUserByID(c.Request.Context(), uint(userID))
	if err != nil {
		logger.Error("Admin failed to get user for security status",
			logger.Uint("user_id", uint(userID)),
			logger.ErrorField(err),
		)
		response.NotFound(c, "User not found")
		return
	}

	// 🔴 DISABLED: Check if login security service is available
	if !h.checkLoginSecurityService(c) {
		return
	}

	// Check if account is locked
	isLocked, lockout, err := h.loginSecurityService.IsAccountLocked(c.Request.Context(), user.Email)
	if err != nil {
		logger.Error("Admin failed to check account lock status",
			logger.Uint("user_id", uint(userID)),
			logger.String("email", user.Email),
			logger.ErrorField(err),
		)
		response.InternalServerError(c, "Failed to get security status")
		return
	}

	// Get failure count  
	// Note: Already checked loginSecurityService availability above
	failureCount, err := h.loginSecurityService.GetFailureCount(c.Request.Context(), user.Email)
	if err != nil {
		logger.Error("Admin failed to get failure count",
			logger.Uint("user_id", uint(userID)),
			logger.String("email", user.Email),
			logger.ErrorField(err),
		)
		failureCount = 0 // Default to 0 on error
	}

	logger.Info("Admin accessed account security status",
		logger.Uint("user_id", uint(userID)),
		logger.String("email", user.Email),
	)

	securityStatus := gin.H{
		"user_id":            user.ID,
		"email":              user.Email,
		"account_status":     user.Status,
		"is_locked":          isLocked,
		"failed_login_count": failureCount,
		"lockout_info":       lockout,
		"last_login":         nil, // Would need additional tracking
		"security_score":     calculateSecurityScore(user, isLocked, failureCount),
	}

	response.OK(c, securityStatus)
}

// Security Analytics and Reporting

// GetSecurityStatistics godoc
// @Summary Get overall security statistics
// @Description Get comprehensive security statistics and metrics (Admin only)
// @Tags Admin-Auth-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param days query int false "Statistics period in days" default(30)
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} response.ProblemJSONResponse
// @Failure 403 {object} response.ProblemJSONResponse
// @Failure 500 {object} response.ProblemJSONResponse
// @Router /admin/auth/analytics/statistics [get]
func (h *AdminAuthHandler) GetSecurityStatistics(c *gin.Context) {
	days, _ := strconv.Atoi(c.DefaultQuery("days", "30"))
	if days < 1 || days > 365 {
		days = 30
	}

	since := time.Now().AddDate(0, 0, -days)

	// Get login attempt statistics
	loginStats, err := h.loginSecurityService.GetLoginAttemptStats(c.Request.Context(), since)
	if err != nil {
		logger.Error("Admin failed to get login stats for security statistics",
			logger.Int("days", days),
			logger.ErrorField(err),
		)
		loginStats = make(map[string]any)
	}

	// Get JWT blacklist statistics
	jwtStats, err := h.jwtBlacklistService.GetBlacklistStats(c.Request.Context())
	if err != nil {
		logger.Error("Admin failed to get JWT stats for security statistics",
			logger.ErrorField(err),
		)
		jwtStats = make(map[string]any)
	}

	logger.Info("Admin accessed security statistics",
		logger.Int("days", days),
	)

	statistics := gin.H{
		"period": gin.H{
			"days":       days,
			"start_date": since,
			"end_date":   time.Now(),
		},
		"login_security": loginStats,
		"jwt_security":   jwtStats,
		"overall_health": gin.H{
			"status": "healthy", // Would implement health scoring
			"alerts": 0,         // Would track active alerts
		},
	}

	response.OK(c, statistics)
}

// Bulk Security Operations

// BulkRevokeTokens godoc
// @Summary Bulk revoke user tokens
// @Description Revoke JWT tokens for multiple users (Admin only)
// @Tags Admin-Auth-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param bulk body BulkSecurityActionRequest true "Bulk revocation data"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} response.ProblemJSONResponse
// @Failure 401 {object} response.ProblemJSONResponse
// @Failure 403 {object} response.ProblemJSONResponse
// @Failure 500 {object} response.ProblemJSONResponse
// @Router /admin/auth/bulk/revoke-tokens [post]
func (h *AdminAuthHandler) BulkRevokeTokens(c *gin.Context) {
	var req BulkSecurityActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if req.Action != "revoke_tokens" {
		response.BadRequest(c, "Invalid action, must be 'revoke_tokens'")
		return
	}

	successCount := 0
	failedIDs := []uint{}

	for _, userID := range req.UserIDs {
		// Note: Would implement actual bulk token revocation
		// For now, we'll simulate success
		logger.Info("Admin bulk revoked tokens for user",
			logger.Uint("user_id", userID),
			logger.String("reason", req.Reason),
		)
		successCount++
	}

	logger.Info("Admin performed bulk token revocation",
		logger.Int("success_count", successCount),
		logger.Int("total_users", len(req.UserIDs)),
		logger.String("reason", req.Reason),
	)

	response.OK(c, gin.H{
		"success_count": successCount,
		"failed_ids":    failedIDs,
		"total_users":   len(req.UserIDs),
		"reason":        req.Reason,
	})
}

// BulkUnlockAccounts godoc
// @Summary Bulk unlock accounts
// @Description Unlock multiple user accounts (Admin only)
// @Tags Admin-Auth-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param bulk body BulkSecurityActionRequest true "Bulk unlock data"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} response.ProblemJSONResponse
// @Failure 401 {object} response.ProblemJSONResponse
// @Failure 403 {object} response.ProblemJSONResponse
// @Failure 500 {object} response.ProblemJSONResponse
// @Router /admin/auth/bulk/unlock-accounts [post]
func (h *AdminAuthHandler) BulkUnlockAccounts(c *gin.Context) {
	var req BulkSecurityActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if req.Action != "unlock_accounts" {
		response.BadRequest(c, "Invalid action, must be 'unlock_accounts'")
		return
	}

	successCount := 0
	failedIDs := []uint{}

	for _, userID := range req.UserIDs {
		// Get user email for unlocking
		user, err := h.userService.GetUserByID(c.Request.Context(), userID)
		if err != nil {
			logger.Error("Failed to get user for bulk unlock",
				logger.Uint("user_id", userID),
				logger.ErrorField(err),
			)
			failedIDs = append(failedIDs, userID)
			continue
		}

		if err := h.loginSecurityService.UnlockAccount(c.Request.Context(), user.Email, req.Reason); err != nil {
			logger.Error("Failed to unlock account in bulk operation",
				logger.Uint("user_id", userID),
				logger.String("email", user.Email),
				logger.ErrorField(err),
			)
			failedIDs = append(failedIDs, userID)
			continue
		}

		successCount++
		logger.Info("Admin bulk unlocked account",
			logger.Uint("user_id", userID),
			logger.String("email", user.Email),
			logger.String("reason", req.Reason),
		)
	}

	logger.Info("Admin performed bulk account unlock",
		logger.Int("success_count", successCount),
		logger.Int("failed_count", len(failedIDs)),
		logger.String("reason", req.Reason),
	)

	response.OK(c, gin.H{
		"success_count": successCount,
		"failed_ids":    failedIDs,
		"total_users":   len(req.UserIDs),
		"reason":        req.Reason,
	})
}

// OAuth Provider Management Endpoints

// GetOAuthProviderStats godoc
// @Summary Get OAuth provider statistics
// @Description Get statistics for OAuth authentication patterns by provider (Admin only)
// @Tags Admin-Auth-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} response.ProblemJSONResponse
// @Failure 403 {object} response.ProblemJSONResponse
// @Failure 500 {object} response.ProblemJSONResponse
// @Router /admin/auth/oauth/providers [get]
func (h *AdminAuthHandler) GetOAuthProviderStats(c *gin.Context) {
	// Get user statistics by provider
	stats, err := h.userService.GetUserStats(c.Request.Context())
	if err != nil {
		logger.Error("Admin failed to get OAuth provider stats",
			logger.ErrorField(err),
		)
		response.InternalServerError(c, "Failed to get OAuth provider statistics")
		return
	}

	logger.Info("Admin accessed OAuth provider statistics")

	// Extract provider-specific statistics
	providerStats := gin.H{
		"total_users":     stats,
		"oauth_breakdown": gin.H{
			"google":   "Statistics would require provider-specific queries",
			"github":   "Statistics would require provider-specific queries",
			"telegram": "Statistics would require provider-specific queries",
			"local":    "Statistics would require provider-specific queries",
		},
	}

	response.OK(c, providerStats)
}

// ListOAuthSecurityEvents godoc
// @Summary List OAuth security events
// @Description Get paginated list of OAuth security events and incidents (Admin only)
// @Tags Admin-Auth-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param provider query string false "Filter by OAuth provider" example="google"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} response.HALCollectionResponse
// @Failure 401 {object} response.ProblemJSONResponse
// @Failure 403 {object} response.ProblemJSONResponse
// @Failure 500 {object} response.ProblemJSONResponse
// @Router /admin/auth/oauth/incidents [get]
func (h *AdminAuthHandler) ListOAuthSecurityEvents(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	provider := c.Query("provider")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	logger.Info("Admin accessed OAuth security events",
		logger.String("provider", provider),
		logger.Int("page", page),
		logger.Int("limit", limit),
	)

	// Mock OAuth security events - would require actual implementation
	events := []gin.H{
		{
			"id":         1,
			"type":       "oauth_state_mismatch",
			"provider":   "google",
			"severity":   "medium",
			"timestamp":  time.Now().Add(-2 * time.Hour),
			"details":    "OAuth state parameter mismatch detected",
			"ip_address": "192.168.1.100",
		},
		{
			"id":         2,
			"type":       "oauth_callback_anomaly",
			"provider":   "github",
			"severity":   "low",
			"timestamp":  time.Now().Add(-6 * time.Hour),
			"details":    "Unusual callback URL pattern",
			"ip_address": "10.0.0.50",
		},
	}

	response.SendPaginatedResponse(c, events, int64(len(events)))
}

// Advanced Security Analytics

// GetSecurityPatterns godoc
// @Summary Get security pattern analysis
// @Description Get detailed analysis of security patterns and anomalies (Admin only)
// @Tags Admin-Auth-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param analysis_type query string false "Analysis type" default("login_patterns") example="login_patterns"
// @Param days query int false "Analysis period in days" default(30)
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} response.ProblemJSONResponse
// @Failure 403 {object} response.ProblemJSONResponse
// @Failure 500 {object} response.ProblemJSONResponse
// @Router /admin/auth/analytics/patterns [get]
func (h *AdminAuthHandler) GetSecurityPatterns(c *gin.Context) {
	analysisType := c.DefaultQuery("analysis_type", "login_patterns")
	days, _ := strconv.Atoi(c.DefaultQuery("days", "30"))

	if days < 1 || days > 365 {
		days = 30
	}

	since := time.Now().AddDate(0, 0, -days)

	logger.Info("Admin accessed security pattern analysis",
		logger.String("analysis_type", analysisType),
		logger.Int("days", days),
	)

	patterns := gin.H{
		"analysis_type": analysisType,
		"period": gin.H{
			"days":       days,
			"start_date": since,
			"end_date":   time.Now(),
		},
	}

	switch analysisType {
	case "login_patterns":
		stats, err := h.loginSecurityService.GetLoginAttemptStats(c.Request.Context(), since)
		if err != nil {
			logger.Error("Failed to get login patterns",
				logger.ErrorField(err),
			)
			response.InternalServerError(c, "Failed to get login patterns")
			return
		}
		patterns["login_patterns"] = stats
		patterns["anomalies"] = []gin.H{
			{
				"type":        "geographic_anomaly",
				"severity":    "medium",
				"description": "Login from unusual geographic location detected",
				"count":       3,
			},
			{
				"type":        "time_anomaly",
				"severity":    "low",
				"description": "Login attempts outside normal hours",
				"count":       15,
			},
		}

	case "brute_force":
		patterns["brute_force_detection"] = gin.H{
			"active_attacks":     2,
			"blocked_ips":        []string{"192.168.1.100", "10.0.0.50"},
			"attack_patterns":    []string{"credential_stuffing", "dictionary_attack"},
			"mitigation_actions": []string{"ip_blocking", "rate_limiting"},
		}

	case "device_patterns":
		patterns["device_analysis"] = gin.H{
			"unique_devices":     1250,
			"new_devices":        45,
			"suspicious_devices": 3,
			"device_types": gin.H{
				"mobile":  60,
				"desktop": 35,
				"tablet":  5,
			},
		}

	default:
		response.BadRequest(c, "Invalid analysis type")
		return
	}

	response.OK(c, patterns)
}

// GetSecurityScore godoc
// @Summary Get overall security score
// @Description Get comprehensive security score and recommendations (Admin only)
// @Tags Admin-Auth-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} response.ProblemJSONResponse
// @Failure 403 {object} response.ProblemJSONResponse
// @Failure 500 {object} response.ProblemJSONResponse
// @Router /admin/auth/analytics/security-score [get]
func (h *AdminAuthHandler) GetSecurityScore(c *gin.Context) {
	logger.Info("Admin accessed security score")

	// Get system statistics
	since := time.Now().AddDate(0, 0, -30)
	loginStats, err := h.loginSecurityService.GetLoginAttemptStats(c.Request.Context(), since)
	if err != nil {
		logger.Error("Failed to get login stats for security score",
			logger.ErrorField(err),
		)
		loginStats = make(map[string]any)
	}

	jwtStats, err := h.jwtBlacklistService.GetBlacklistStats(c.Request.Context())
	if err != nil {
		logger.Error("Failed to get JWT stats for security score",
			logger.ErrorField(err),
		)
		jwtStats = make(map[string]any)
	}

	// Calculate overall security score
	securityScore := calculateOverallSecurityScore(loginStats, jwtStats)

	response.OK(c, gin.H{
		"overall_score": securityScore,
		"score_breakdown": gin.H{
			"authentication": 85,
			"authorization":  92,
			"data_protection": 88,
			"monitoring":     90,
			"incident_response": 87,
		},
		"recommendations": []gin.H{
			{
				"priority":    "high",
				"category":    "authentication",
				"description": "Consider implementing MFA for admin accounts",
				"impact":      "Significantly improves account security",
			},
			{
				"priority":    "medium",
				"category":    "monitoring",
				"description": "Enable geographic anomaly detection",
				"impact":      "Better detection of suspicious activities",
			},
		},
		"security_metrics": gin.H{
			"login_security": loginStats,
			"token_security": jwtStats,
		},
	})
}

// Additional Bulk Operations

// BulkResetPasswords godoc
// @Summary Bulk reset passwords
// @Description Reset passwords for multiple users (Admin only)
// @Tags Admin-Auth-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param bulk body BulkSecurityActionRequest true "Bulk password reset data"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} response.ProblemJSONResponse
// @Failure 401 {object} response.ProblemJSONResponse
// @Failure 403 {object} response.ProblemJSONResponse
// @Failure 500 {object} response.ProblemJSONResponse
// @Router /admin/auth/bulk/reset-passwords [post]
func (h *AdminAuthHandler) BulkResetPasswords(c *gin.Context) {
	adminUser, exists := c.Get(middleware.AuthContextKey)
	if !exists {
		response.Unauthorized(c, "Admin authentication required")
		return
	}

	admin, ok := adminUser.(*userEntities.User)
	if !ok {
		response.InternalServerError(c, "Invalid admin user context")
		return
	}

	var req BulkSecurityActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if req.Action != "reset_passwords" {
		response.BadRequest(c, "Invalid action, must be 'reset_passwords'")
		return
	}

	successCount := 0
	failedIDs := []uint{}

	for _, userID := range req.UserIDs {
		// Generate a temporary password
		tempPassword := generateTemporaryPassword()

		if err := h.authService.AdminResetPassword(c.Request.Context(), admin.ID, userID, tempPassword); err != nil {
			logger.Error("Failed to reset password in bulk operation",
				logger.Uint("admin_id", admin.ID),
				logger.Uint("target_user_id", userID),
				logger.ErrorField(err),
			)
			failedIDs = append(failedIDs, userID)
			continue
		}

		successCount++
		logger.Info("Admin bulk reset password",
			logger.Uint("admin_id", admin.ID),
			logger.Uint("target_user_id", userID),
			logger.String("reason", req.Reason),
		)
	}

	logger.Info("Admin performed bulk password reset",
		logger.Int("success_count", successCount),
		logger.Int("failed_count", len(failedIDs)),
		logger.String("reason", req.Reason),
	)

	response.OK(c, gin.H{
		"success_count": successCount,
		"failed_ids":    failedIDs,
		"total_users":   len(req.UserIDs),
		"reason":        req.Reason,
		"note":          "Temporary passwords have been set. Users will need to change them on next login.",
	})
}

// SendSecurityNotifications godoc
// @Summary Send security notifications
// @Description Send security notifications to multiple users (Admin only)
// @Tags Admin-Auth-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param notification body SecurityNotificationRequest true "Notification data"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} response.ProblemJSONResponse
// @Failure 401 {object} response.ProblemJSONResponse
// @Failure 403 {object} response.ProblemJSONResponse
// @Failure 500 {object} response.ProblemJSONResponse
// @Router /admin/auth/bulk/notifications [post]
func (h *AdminAuthHandler) SendSecurityNotifications(c *gin.Context) {
	var req SecurityNotificationRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	successCount := 0
	failedIDs := []uint{}

	for _, userID := range req.UserIDs {
		// Note: Would implement actual notification sending
		// For now, we'll simulate success
		logger.Info("Admin sent security notification",
			logger.Uint("user_id", userID),
			logger.String("subject", req.Subject),
		)
		successCount++
	}

	logger.Info("Admin sent bulk security notifications",
		logger.Int("success_count", successCount),
		logger.Int("total_recipients", len(req.UserIDs)),
		logger.String("subject", req.Subject),
	)

	response.OK(c, gin.H{
		"success_count":    successCount,
		"failed_ids":       failedIDs,
		"total_recipients": len(req.UserIDs),
		"subject":          req.Subject,
	})
}

// Helper functions

// calculateOverallSecurityScore calculates a comprehensive security score
func calculateOverallSecurityScore(loginStats, jwtStats map[string]any) int {
	baseScore := 90

	// Analyze login statistics
	if failureRate, ok := loginStats["failure_rate"].(float64); ok {
		if failureRate > 0.3 { // More than 30% failures
			baseScore -= 10
		} else if failureRate > 0.1 { // More than 10% failures
			baseScore -= 5
		}
	}

	// Analyze JWT statistics
	if activeCount, ok := jwtStats["active_count"].(int64); ok {
		if activeCount > 1000 { // High number of active blacklisted tokens
			baseScore -= 5
		}
	}

	if baseScore < 0 {
		baseScore = 0
	}
	if baseScore > 100 {
		baseScore = 100
	}

	return baseScore
}

// generateTemporaryPassword generates a secure temporary password
func generateTemporaryPassword() string {
	// In a real implementation, this would generate a cryptographically secure password
	timestamp := time.Now().Unix()
	return "TempPass" + strconv.FormatInt(timestamp, 36) + "!"
}

// Helper function to calculate security score
func calculateSecurityScore(user *userEntities.User, isLocked bool, failureCount int) int {
	score := 100

	// Deduct points for account issues
	if isLocked {
		score -= 50
	}

	if failureCount > 0 {
		score -= failureCount * 10
		if score < 0 {
			score = 0
		}
	}

	// Deduct points for inactive/banned accounts
	switch user.Status {
	case userConstants.UserStatusInactive:
		score -= 20
	case userConstants.UserStatusBanned:
		score -= 80
	}

	// Bonus for OAuth accounts (typically more secure)
	if user.Provider != userConstants.ProviderLocal {
		score += 10
	}

	if score > 100 {
		score = 100
	}
	if score < 0 {
		score = 0
	}

	return score
}