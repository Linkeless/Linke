package admin

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

// ============= Admin Referral Management =============

// ListAllReferrals godoc
// @Summary [Admin] List all referrals
// @Description Get paginated list of all referrals (Admin only)
// @Tags Admin-Referral-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param status query string false "Filter by status"
// @Param source query string false "Filter by referral source"
// @Param campaign_id query int false "Filter by campaign ID"
// @Success 200 {object} response.StandardListResponse{data=[]model.ReferralResponse}
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/referrals [get]
func (h *ReferralHandler) ListAllReferrals(c *gin.Context) {
	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	// TODO: Implement admin referral listing with filters
	// For now, return empty response
	response.SuccessList(c, []*model.ReferralResponse{}, page, limit, 0)
}

// GetReferral godoc
// @Summary [Admin] Get referral by ID
// @Description Get referral details by ID (Admin only)
// @Tags Admin-Referral-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Referral ID"
// @Success 200 {object} response.StandardResponse{data=model.ReferralResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Router /admin/referrals/{id} [get]
func (h *ReferralHandler) GetReferral(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid referral ID")
		return
	}

	// Get referral with relations
	referral, err := h.referralService.GetReferralWithRelations(c.Request.Context(), uint(id))
	if err != nil {
		logger.Error("Failed to get referral",
			logger.Uint("referral_id", uint(id)),
			logger.Error2("error", err),
		)
		response.NotFound(c, "Referral not found")
		return
	}

	response.Success(c, referral.ToResponse())
}

// ============= Referral Campaign Management =============

// CreateReferralCampaign godoc
// @Summary [Admin] Create referral campaign
// @Description Create a new referral campaign (Admin only)
// @Tags Admin-Referral-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param campaign body service.CreateReferralCampaignRequest true "Campaign data"
// @Success 201 {object} response.StandardResponse{data=model.ReferralCampaignResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/referral-campaigns [post]
func (h *ReferralHandler) CreateReferralCampaign(c *gin.Context) {
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

	var req service.CreateReferralCampaignRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// Create campaign
	campaign, err := h.referralCampaignService.CreateReferralCampaign(c.Request.Context(), user.ID, &req)
	if err != nil {
		logger.Error("Failed to create referral campaign",
			logger.Uint("admin_id", user.ID),
			logger.Error2("error", err),
		)
		response.BadRequest(c, err.Error())
		return
	}

	response.Created(c, campaign.ToResponse())
}

// ListReferralCampaigns godoc
// @Summary [Admin] List referral campaigns
// @Description Get paginated list of referral campaigns (Admin only)
// @Tags Admin-Referral-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param status query string false "Filter by status"
// @Param campaign_type query string false "Filter by campaign type"
// @Param is_public query bool false "Filter by public visibility"
// @Success 200 {object} response.StandardListResponse{data=[]model.ReferralCampaignResponse}
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/referral-campaigns [get]
func (h *ReferralHandler) ListReferralCampaigns(c *gin.Context) {
	// Parse query parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	offset := (page - 1) * limit

	req := &service.GetReferralCampaignsRequest{
		Status:       c.Query("status"),
		CampaignType: c.Query("campaign_type"),
		Limit:        limit,
		Offset:       offset,
	}

	// Handle is_public filter
	if isPublicStr := c.Query("is_public"); isPublicStr != "" {
		if isPublic, err := strconv.ParseBool(isPublicStr); err == nil {
			req.IsPublic = &isPublic
		}
	}

	// Get campaigns
	campaigns, total, err := h.referralCampaignService.GetReferralCampaigns(c.Request.Context(), req)
	if err != nil {
		logger.Error("Failed to get referral campaigns",
			logger.Error2("error", err),
		)
		response.InternalServerError(c, "Failed to get referral campaigns")
		return
	}

	// Convert to response format
	var campaignResponses []*model.ReferralCampaignResponse
	for _, campaign := range campaigns {
		campaignResponses = append(campaignResponses, campaign.ToResponse())
	}

	response.SuccessList(c, campaignResponses, page, limit, total)
}

// GetReferralCampaign godoc
// @Summary [Admin] Get referral campaign by ID
// @Description Get referral campaign details by ID (Admin only)
// @Tags Admin-Referral-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Campaign ID"
// @Success 200 {object} response.StandardResponse{data=model.ReferralCampaignResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Router /admin/referral-campaigns/{id} [get]
func (h *ReferralHandler) GetReferralCampaign(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid campaign ID")
		return
	}

	// Get campaign
	campaign, err := h.referralCampaignService.GetReferralCampaignByID(c.Request.Context(), uint(id))
	if err != nil {
		logger.Error("Failed to get referral campaign",
			logger.Uint("campaign_id", uint(id)),
			logger.Error2("error", err),
		)
		response.NotFound(c, "Campaign not found")
		return
	}

	response.Success(c, campaign.ToResponse())
}

// UpdateReferralCampaign godoc
// @Summary [Admin] Update referral campaign
// @Description Update referral campaign by ID (Admin only)
// @Tags Admin-Referral-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Campaign ID"
// @Param campaign body service.UpdateReferralCampaignRequest true "Updated campaign data"
// @Success 200 {object} response.StandardResponse{data=model.ReferralCampaignResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Router /admin/referral-campaigns/{id} [put]
func (h *ReferralHandler) UpdateReferralCampaign(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid campaign ID")
		return
	}

	var req service.UpdateReferralCampaignRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// Update campaign
	campaign, err := h.referralCampaignService.UpdateReferralCampaign(c.Request.Context(), uint(id), &req)
	if err != nil {
		logger.Error("Failed to update referral campaign",
			logger.Uint("campaign_id", uint(id)),
			logger.Error2("error", err),
		)
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, campaign.ToResponse())
}

// DeleteReferralCampaign godoc
// @Summary [Admin] Delete referral campaign
// @Description Delete referral campaign by ID (Admin only)
// @Tags Admin-Referral-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Campaign ID"
// @Success 200 {object} response.MessageOnlyResponse
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Router /admin/referral-campaigns/{id} [delete]
func (h *ReferralHandler) DeleteReferralCampaign(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid campaign ID")
		return
	}

	// Delete campaign
	if err := h.referralCampaignService.DeleteReferralCampaign(c.Request.Context(), uint(id)); err != nil {
		logger.Error("Failed to delete referral campaign",
			logger.Uint("campaign_id", uint(id)),
			logger.Error2("error", err),
		)
		response.NotFound(c, "Campaign not found")
		return
	}

	response.SuccessWithMessage(c, "Campaign deleted successfully", nil)
}

// GetReferralCampaignStats godoc
// @Summary [Admin] Get referral campaign statistics
// @Description Get statistics for a referral campaign (Admin only)
// @Tags Admin-Referral-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Campaign ID"
// @Success 200 {object} response.StandardResponse{data=map[string]interface{}}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Router /admin/referral-campaigns/{id}/stats [get]
func (h *ReferralHandler) GetReferralCampaignStats(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid campaign ID")
		return
	}

	// Get campaign stats
	stats, err := h.referralCampaignService.GetCampaignStats(c.Request.Context(), uint(id))
	if err != nil {
		logger.Error("Failed to get referral campaign stats",
			logger.Uint("campaign_id", uint(id)),
			logger.Error2("error", err),
		)
		response.NotFound(c, "Campaign not found")
		return
	}

	response.Success(c, stats)
}