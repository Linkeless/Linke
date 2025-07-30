package user

import (
	"strconv"

	"linke/internal/logger"
	"linke/internal/middleware"
	"linke/internal/model"
	"linke/internal/response"
	"linke/internal/service"

	"github.com/gin-gonic/gin"
)

type ReferralHandler struct {
	referralService         *service.ReferralService
	referralCampaignService *service.ReferralCampaignService
}

func NewReferralHandler(referralService *service.ReferralService, referralCampaignService *service.ReferralCampaignService) *ReferralHandler {
	return &ReferralHandler{
		referralService:         referralService,
		referralCampaignService: referralCampaignService,
	}
}

// ============= User Referral Management =============

// GetMyReferrals godoc
// @Summary [User] Get my referrals
// @Description Get referrals created by current user
// @Tags User-Referral
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} response.StandardListResponse{data=[]model.ReferralResponse}
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /user/referrals [get]
func (h *ReferralHandler) GetMyReferrals(c *gin.Context) {
	// Get current user from context
	userValue, exists := c.Get(middleware.AuthContextKey)
	if !exists {
		response.Unauthorized(c, "Authentication required")
		return
	}

	user, ok := userValue.(*model.User)
	if !ok {
		response.Unauthorized(c, "Invalid user context")
		return
	}

	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	offset := (page - 1) * limit

	// Get referrals
	referrals, total, err := h.referralService.GetReferralsByReferrer(c.Request.Context(), user.ID, limit, offset)
	if err != nil {
		logger.Error("Failed to get user referrals",
			logger.Uint("user_id", user.ID),
			logger.Error2("error", err),
		)
		response.InternalServerError(c, "Failed to get referrals")
		return
	}

	// Convert to response format
	var referralResponses []*model.ReferralResponse
	for _, referral := range referrals {
		referralResponses = append(referralResponses, referral.ToResponse())
	}

	response.SuccessList(c, referralResponses, page, limit, total)
}

// GetMyReferralStats godoc
// @Summary [User] Get my referral statistics
// @Description Get referral statistics for current user
// @Tags User-Referral
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.StandardResponse{data=map[string]interface{}}
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /user/referrals/stats [get]
func (h *ReferralHandler) GetMyReferralStats(c *gin.Context) {
	// Get current user from context
	userValue, exists := c.Get(middleware.AuthContextKey)
	if !exists {
		response.Unauthorized(c, "Authentication required")
		return
	}

	user, ok := userValue.(*model.User)
	if !ok {
		response.Unauthorized(c, "Invalid user context")
		return
	}

	// Get referral stats
	stats, err := h.referralService.GetReferralStats(c.Request.Context(), user.ID)
	if err != nil {
		logger.Error("Failed to get referral stats",
			logger.Uint("user_id", user.ID),
			logger.Error2("error", err),
		)
		response.InternalServerError(c, "Failed to get referral statistics")
		return
	}

	response.Success(c, stats)
}

// TrackReferralClick godoc
// @Summary [Public] Track referral click
// @Description Track a click on a referral link
// @Tags User-Referral
// @Accept json
// @Produce json
// @Param code path string true "Referral code"
// @Success 200 {object} response.MessageOnlyResponse
// @Failure 400 {object} response.BadRequestResponse
// @Failure 404 {object} response.NotFoundResponse
// @Router /referral/track/{code} [post]
func (h *ReferralHandler) TrackReferralClick(c *gin.Context) {
	referralCode := c.Param("code")
	if referralCode == "" {
		response.BadRequest(c, "Referral code is required")
		return
	}

	// Gather attribution data
	attributionData := map[string]interface{}{
		"ip_address": c.ClientIP(),
		"user_agent": c.GetHeader("User-Agent"),
		"referrer_url": c.GetHeader("Referer"),
		"page_url": c.Request.URL.String(),
	}

	// Track UTM parameters
	if utmSource := c.Query("utm_source"); utmSource != "" {
		attributionData["utm_source"] = utmSource
	}
	if utmCampaign := c.Query("utm_campaign"); utmCampaign != "" {
		attributionData["utm_campaign"] = utmCampaign
	}
	if utmMedium := c.Query("utm_medium"); utmMedium != "" {
		attributionData["utm_medium"] = utmMedium
	}
	if utmTerm := c.Query("utm_term"); utmTerm != "" {
		attributionData["utm_term"] = utmTerm
	}
	if utmContent := c.Query("utm_content"); utmContent != "" {
		attributionData["utm_content"] = utmContent
	}

	// Track click
	if err := h.referralService.TrackReferralClick(c.Request.Context(), referralCode, attributionData); err != nil {
		logger.Error("Failed to track referral click",
			logger.String("referral_code", referralCode),
			logger.Error2("error", err),
		)
		response.NotFound(c, "Referral not found")
		return
	}

	response.SuccessWithMessage(c, "Click tracked successfully", nil)
}

// ============= Public Referral Campaigns =============

// GetPublicReferralCampaigns godoc
// @Summary [Public] Get public referral campaigns
// @Description Get list of active public referral campaigns
// @Tags User-Referral
// @Accept json
// @Produce json
// @Success 200 {object} response.StandardResponse{data=[]model.ReferralCampaignResponse}
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /referral-campaigns [get]
func (h *ReferralHandler) GetPublicReferralCampaigns(c *gin.Context) {
	// Get active campaigns
	campaigns, err := h.referralCampaignService.GetActiveCampaigns(c.Request.Context())
	if err != nil {
		logger.Error("Failed to get public referral campaigns",
			logger.Error2("error", err),
		)
		response.InternalServerError(c, "Failed to get referral campaigns")
		return
	}

	// Convert to public response format
	var campaignResponses []*model.ReferralCampaignResponse
	for _, campaign := range campaigns {
		campaignResponses = append(campaignResponses, campaign.ToPublicResponse())
	}

	response.Success(c, campaignResponses)
}