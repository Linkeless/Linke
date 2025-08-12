package handlers

import (
	"strconv"
	"strings"

	"linke/internal/domains/referral/constants"
	"linke/internal/domains/referral/dto"
	referralInterfaces "linke/internal/domains/referral/usecases/interfaces"
	"linke/internal/shared/logger"
	"linke/internal/shared/response"

	"github.com/gin-gonic/gin"
)

// AdminReferralHandler handles admin referral operations
type AdminReferralHandler struct {
	referralService         referralInterfaces.ReferralService
	referralCampaignService referralInterfaces.ReferralCampaignService
	inviteCodeService       referralInterfaces.InviteCodeService
}

// NewAdminReferralHandler creates a new AdminReferralHandler
func NewAdminReferralHandler(
	referralService referralInterfaces.ReferralService,
	referralCampaignService referralInterfaces.ReferralCampaignService,
	inviteCodeService referralInterfaces.InviteCodeService,
) *AdminReferralHandler {
	return &AdminReferralHandler{
		referralService:         referralService,
		referralCampaignService: referralCampaignService,
		inviteCodeService:       inviteCodeService,
	}
}

// CreateReferral godoc
// @Summary Create new referral
// @Description Create a new referral relationship (Admin only)
// @Tags Admin-Referral-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param referral body dto.CreateReferralRequest true "Referral creation data"
// @Success 201 {object} response.StandardResponse{data=dto.ReferralResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 409 {object} response.ConflictResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/referrals [post]
func (h *AdminReferralHandler) CreateReferral(c *gin.Context) {
	var createReq dto.CreateReferralRequest
	if err := c.ShouldBindJSON(&createReq); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// Convert request to service request
	serviceReq := &dto.CreateReferralRequest{
		ReferrerID:      createReq.ReferrerID,
		RefereeID:       createReq.RefereeID,
		InviteCodeID:    createReq.InviteCodeID,
		ReferralSource:  createReq.ReferralSource,
		ReferralChannel: createReq.ReferralChannel,
		ReferralCode:    createReq.ReferralCode,
		CampaignID:      createReq.CampaignID,
		ConversionValue: createReq.ConversionValue,
		ConversionType:  createReq.ConversionType,
		ExpirationDays:  createReq.ExpirationDays,
		AttributionData: createReq.AttributionData,
	}

	referral, err := h.referralService.CreateReferral(c.Request.Context(), serviceReq)
	if err != nil {
		logger.Error("Admin failed to create referral",
			logger.Uint("referrer_id", createReq.ReferrerID),
			logger.Uint("referee_id", createReq.RefereeID),
			logger.ErrorField(err),
		)

		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "already exists") {
			response.Conflict(c, "Referral relationship already exists")
			return
		}

		response.InternalServerError(c, "Failed to create referral")
		return
	}

	logger.Info("Admin created new referral",
		logger.Uint("referral_id", referral.ID),
		logger.Uint("referrer_id", referral.ReferrerID),
		logger.Uint("referee_id", referral.RefereeID),
		logger.String("admin_action", "create_referral"),
	)

	response.Created(c, dto.ToReferralResponse(referral))
}

// ListReferrals godoc
// @Summary List all referrals
// @Description Get paginated list of all referrals with filtering (Admin only)
// @Tags Admin-Referral-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param status query string false "Referral status" Enums(pending,confirmed,rewarded,cancelled)
// @Param reward_status query string false "Reward status" Enums(pending,earned,paid,cancelled)
// @Param campaign_id query int false "Campaign ID filter"
// @Param referrer_id query int false "Referrer user ID filter"
// @Param referee_id query int false "Referee user ID filter"
// @Success 200 {object} response.StandardListResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/referrals [get]
func (h *AdminReferralHandler) ListReferrals(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	offset := (page - 1) * limit

	// Parse query parameters
	var campaignID *uint
	if campaignIDStr := c.Query("campaign_id"); campaignIDStr != "" {
		if id, err := strconv.ParseUint(campaignIDStr, 10, 32); err == nil {
			campaignIDVal := uint(id)
			campaignID = &campaignIDVal
		}
	}

	referrerID, _ := strconv.ParseUint(c.Query("referrer_id"), 10, 32)
	refereeID, _ := strconv.ParseUint(c.Query("referee_id"), 10, 32)

	serviceReq := &dto.GetReferralsRequest{
		ReferrerID:   uint(referrerID),
		RefereeID:    uint(refereeID),
		Status:       c.Query("status"),
		RewardStatus: c.Query("reward_status"),
		CampaignID:   campaignID,
		Limit:        limit,
		Offset:       offset,
	}

	referrals, total, err := h.referralService.GetReferrals(c.Request.Context(), serviceReq)
	if err != nil {
		logger.Error("Admin failed to list referrals", logger.ErrorField(err))
		response.InternalServerError(c, "Failed to list referrals")
		return
	}

	// Convert to responses
	var referralResponses []*dto.ReferralResponse
	for _, referral := range referrals {
		referralResponses = append(referralResponses, dto.ToReferralResponse(referral))
	}

	response.SuccessList(c, referralResponses, page, limit, total)
}

// GetReferral godoc
// @Summary Get referral information
// @Description Get referral details by referral ID (Admin only)
// @Tags Admin-Referral-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Referral ID"
// @Success 200 {object} response.StandardResponse{data=dto.ReferralResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Router /admin/referrals/{id} [get]
func (h *AdminReferralHandler) GetReferral(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid referral ID")
		return
	}

	referral, err := h.referralService.GetReferral(c.Request.Context(), uint(id))
	if err != nil {
		logger.Error("Admin failed to get referral",
			logger.Uint("referral_id", uint(id)),
			logger.ErrorField(err),
		)
		response.NotFound(c, "Referral not found")
		return
	}

	response.Success(c, dto.ToReferralResponse(referral))
}

// UpdateReferral godoc
// @Summary Update referral information
// @Description Update referral details (Admin only)
// @Tags Admin-Referral-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Referral ID"
// @Param referral body dto.UpdateReferralRequest true "Referral update data"
// @Success 200 {object} response.StandardResponse{data=dto.ReferralResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/referrals/{id} [put]
func (h *AdminReferralHandler) UpdateReferral(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid referral ID")
		return
	}

	var updateReq dto.UpdateReferralRequest
	if err := c.ShouldBindJSON(&updateReq); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// Convert to service request
	serviceReq := &dto.UpdateReferralRequest{
		Status:          updateReq.Status,
		RefereeStatus:   updateReq.RefereeStatus,
		RewardStatus:    updateReq.RewardStatus,
		RewardAmount:    updateReq.RewardAmount,
		ConversionValue: updateReq.ConversionValue,
		ConversionType:  updateReq.ConversionType,
	}

	referral, err := h.referralService.UpdateReferral(c.Request.Context(), uint(id), serviceReq)
	if err != nil {
		logger.Error("Admin failed to update referral",
			logger.Uint("referral_id", uint(id)),
			logger.ErrorField(err),
		)
		response.InternalServerError(c, "Failed to update referral")
		return
	}

	logger.Info("Admin updated referral",
		logger.Uint("referral_id", uint(id)),
		logger.String("admin_action", "update_referral"),
	)

	response.Success(c, dto.ToReferralResponse(referral))
}

// ApproveReferral godoc
// @Summary Approve referral
// @Description Approve a pending referral and set reward amount (Admin only)
// @Tags Admin-Referral-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Referral ID"
// @Param approve body dto.ApproveReferralRequest true "Approval data"
// @Success 200 {object} response.StandardResponse{data=dto.ReferralResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Router /admin/referrals/{id}/approve [post]
func (h *AdminReferralHandler) ApproveReferral(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid referral ID")
		return
	}

	var approveReq dto.ApproveReferralRequest
	if err := c.ShouldBindJSON(&approveReq); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	err = h.referralService.ConfirmReferral(c.Request.Context(), uint(id))
	if err != nil {
		logger.Error("Admin failed to approve referral",
			logger.Uint("referral_id", uint(id)),
			logger.ErrorField(err),
		)
		response.NotFound(c, "Referral not found")
		return
	}

	// If reward amount is specified, process the reward
	if approveReq.RewardAmount != nil && *approveReq.RewardAmount > 0 {
		err = h.referralService.ProcessReferralReward(c.Request.Context(), uint(id), *approveReq.RewardAmount)
		if err != nil {
			logger.Error("Admin failed to process referral reward after approval",
				logger.Uint("referral_id", uint(id)),
				logger.Float64("reward_amount", *approveReq.RewardAmount),
				logger.ErrorField(err),
			)
		}
	}

	// Get updated referral
	referral, err := h.referralService.GetReferral(c.Request.Context(), uint(id))
	if err != nil {
		response.InternalServerError(c, "Failed to get updated referral")
		return
	}

	logger.Info("Admin approved referral",
		logger.Uint("referral_id", uint(id)),
		logger.String("note", approveReq.Note),
		logger.String("admin_action", "approve_referral"),
	)

	response.Success(c, dto.ToReferralResponse(referral))
}

// RejectReferral godoc
// @Summary Reject referral
// @Description Reject a pending referral (Admin only)
// @Tags Admin-Referral-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Referral ID"
// @Param reject body dto.RejectReferralRequest true "Rejection data"
// @Success 200 {object} response.StandardResponse{data=dto.ReferralResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Router /admin/referrals/{id}/reject [post]
func (h *AdminReferralHandler) RejectReferral(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid referral ID")
		return
	}

	var rejectReq dto.RejectReferralRequest
	if err := c.ShouldBindJSON(&rejectReq); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// Update referral status to cancelled
	cancelledStatus := constants.ReferralStatusCancelled
	cancelledRewardStatus := constants.RewardStatusCancelled
	updateReq := &dto.UpdateReferralRequest{
		Status:       &cancelledStatus,
		RewardStatus: &cancelledRewardStatus,
	}

	referral, err := h.referralService.UpdateReferral(c.Request.Context(), uint(id), updateReq)
	if err != nil {
		logger.Error("Admin failed to reject referral",
			logger.Uint("referral_id", uint(id)),
			logger.String("reason", rejectReq.Reason),
			logger.ErrorField(err),
		)
		response.NotFound(c, "Referral not found")
		return
	}

	logger.Info("Admin rejected referral",
		logger.Uint("referral_id", uint(id)),
		logger.String("reason", rejectReq.Reason),
		logger.String("note", rejectReq.Note),
		logger.String("admin_action", "reject_referral"),
	)

	response.Success(c, dto.ToReferralResponse(referral))
}

// ProcessReferralPayout godoc
// @Summary Process referral payout
// @Description Process payout for a referral reward (Admin only)
// @Tags Admin-Referral-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Referral ID"
// @Param payout body dto.PayoutReferralRequest true "Payout data"
// @Success 200 {object} response.StandardResponse{data=dto.ReferralResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Router /admin/referrals/{id}/payout [post]
func (h *AdminReferralHandler) ProcessReferralPayout(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid referral ID")
		return
	}

	var payoutReq dto.PayoutReferralRequest
	if err := c.ShouldBindJSON(&payoutReq); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// Mark referral as paid
	err = h.referralService.MarkReferralAsPaid(c.Request.Context(), uint(id))
	if err != nil {
		logger.Error("Admin failed to process referral payout",
			logger.Uint("referral_id", uint(id)),
			logger.Float64("amount", payoutReq.Amount),
			logger.ErrorField(err),
		)
		response.NotFound(c, "Referral not found or not eligible for payout")
		return
	}

	// Get updated referral
	referral, err := h.referralService.GetReferral(c.Request.Context(), uint(id))
	if err != nil {
		response.InternalServerError(c, "Failed to get updated referral")
		return
	}

	logger.Info("Admin processed referral payout",
		logger.Uint("referral_id", uint(id)),
		logger.Float64("amount", payoutReq.Amount),
		logger.String("payment_method", payoutReq.PaymentMethod),
		logger.String("admin_action", "process_payout"),
	)

	response.Success(c, dto.ToReferralResponse(referral))
}

// SearchReferrals godoc
// @Summary Search referrals
// @Description Search referrals by various criteria (Admin only)
// @Tags Admin-Referral-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param q query string false "Search query"
// @Param status query string false "Referral status" Enums(pending,confirmed,rewarded,cancelled)
// @Param reward_status query string false "Reward status" Enums(pending,earned,paid,cancelled)
// @Param referrer_id query int false "Referrer user ID"
// @Param referee_id query int false "Referee user ID"
// @Param campaign_id query int false "Campaign ID"
// @Param date_from query string false "Date from" format(date-time)
// @Param date_to query string false "Date to" format(date-time)
// @Param min_reward query number false "Minimum reward amount"
// @Param max_reward query number false "Maximum reward amount"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} response.SearchResponse
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/referrals/search [get]
func (h *AdminReferralHandler) SearchReferrals(c *gin.Context) {
	var searchReq dto.SearchReferralsRequest
	if err := c.ShouldBindQuery(&searchReq); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if searchReq.Query == "" {
		response.BadRequest(c, "Search query is required")
		return
	}

	if searchReq.Page == 0 {
		searchReq.Page = 1
	}
	if searchReq.Limit == 0 || searchReq.Limit > 100 {
		searchReq.Limit = 10
	}

	offset := (searchReq.Page - 1) * searchReq.Limit

	// Build service request
	serviceReq := &dto.GetReferralsRequest{
		ReferrerID:   searchReq.ReferrerID,
		RefereeID:    searchReq.RefereeID,
		Status:       searchReq.Status,
		RewardStatus: searchReq.RewardStatus,
		CampaignID:   searchReq.CampaignID,
		Limit:        searchReq.Limit,
		Offset:       offset,
	}

	referrals, total, err := h.referralService.GetReferrals(c.Request.Context(), serviceReq)
	if err != nil {
		logger.Error("Admin failed to search referrals",
			logger.String("query", searchReq.Query),
			logger.ErrorField(err),
		)
		response.InternalServerError(c, "Failed to search referrals")
		return
	}

	// Convert to responses
	var referralResponses []*dto.ReferralResponse
	for _, referral := range referrals {
		referralResponses = append(referralResponses, dto.ToReferralResponse(referral))
	}

	response.SuccessListWithExtra(c, "Search completed", referralResponses, searchReq.Page, searchReq.Limit, total, gin.H{
		"query": searchReq.Query,
	})
}

// GetReferralStatistics godoc
// @Summary Get referral statistics
// @Description Get overall referral system statistics (Admin only)
// @Tags Admin-Referral-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.StandardResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/referrals/statistics [get]
func (h *AdminReferralHandler) GetReferralStatistics(c *gin.Context) {
	stats, err := h.referralService.GetSystemReferralStatistics(c.Request.Context())
	if err != nil {
		logger.Error("Admin failed to get referral statistics", logger.ErrorField(err))
		response.InternalServerError(c, "Failed to get referral statistics")
		return
	}

	response.Success(c, stats)
}

// GetReferralAnalytics godoc
// @Summary Get referral analytics
// @Description Get detailed analytics for referral performance (Admin only)
// @Tags Admin-Referral-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param period query string false "Analytics period" Enums(7d,30d,90d,1y) default(30d)
// @Success 200 {object} response.StandardResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/referrals/analytics [get]
func (h *AdminReferralHandler) GetReferralAnalytics(c *gin.Context) {
	period := c.DefaultQuery("period", "30d")

	// This would be implemented in the service layer
	analytics := gin.H{
		"period":               period,
		"total_referrals":      0,
		"active_referrals":     0,
		"total_conversions":    0,
		"conversion_rate":      0.0,
		"total_rewards_paid":   0.0,
		"average_reward":       0.0,
		"top_referrers":        []gin.H{},
		"conversion_trends":    []gin.H{},
		"campaign_performance": []gin.H{},
		"fraud_indicators": gin.H{
			"suspicious_patterns": 0,
			"duplicate_referrals": 0,
			"blocked_attempts":    0,
		},
	}

	logger.Info("Admin requested referral analytics",
		logger.String("period", period),
		logger.String("admin_action", "get_referral_analytics"),
	)

	response.Success(c, analytics)
}

// BulkApproveReferrals godoc
// @Summary Bulk approve referrals
// @Description Approve multiple referrals at once (Admin only)
// @Tags Admin-Referral-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param bulk body dto.BulkReferralRequest true "Bulk approval data"
// @Success 200 {object} response.StandardResponse
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/referrals/bulk/approve [post]
func (h *AdminReferralHandler) BulkApproveReferrals(c *gin.Context) {
	var bulkReq dto.BulkReferralRequest
	if err := c.ShouldBindJSON(&bulkReq); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if bulkReq.Action != "approve" {
		response.BadRequest(c, "Invalid action for bulk approval")
		return
	}

	successCount := 0
	failedIDs := make([]uint, 0)

	for _, id := range bulkReq.IDs {
		err := h.referralService.ConfirmReferral(c.Request.Context(), id)
		if err != nil {
			logger.Error("Failed to approve referral in bulk operation",
				logger.Uint("referral_id", id),
				logger.ErrorField(err),
			)
			failedIDs = append(failedIDs, id)
			continue
		}

		// Process reward if amount is specified
		if bulkReq.Amount > 0 {
			err = h.referralService.ProcessReferralReward(c.Request.Context(), id, bulkReq.Amount)
			if err != nil {
				logger.Error("Failed to process reward in bulk approval",
					logger.Uint("referral_id", id),
					logger.Float64("amount", bulkReq.Amount),
					logger.ErrorField(err),
				)
			}
		}

		successCount++
	}

	logger.Info("Admin executed bulk referral approval",
		logger.Int("requested_count", len(bulkReq.IDs)),
		logger.Int("success_count", successCount),
		logger.Int("failed_count", len(failedIDs)),
		logger.String("admin_action", "bulk_approve_referrals"),
	)

	response.SuccessWithMessage(c, "Bulk referral approval completed", gin.H{
		"requested_count": len(bulkReq.IDs),
		"success_count":   successCount,
		"failed_count":    len(failedIDs),
		"failed_ids":      failedIDs,
	})
}

// BulkProcessPayouts godoc
// @Summary Bulk process referral payouts
// @Description Process payouts for multiple referrals at once (Admin only)
// @Tags Admin-Referral-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param bulk body dto.BulkReferralRequest true "Bulk payout data"
// @Success 200 {object} response.StandardResponse
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/referrals/bulk/payout [post]
func (h *AdminReferralHandler) BulkProcessPayouts(c *gin.Context) {
	var bulkReq dto.BulkReferralRequest
	if err := c.ShouldBindJSON(&bulkReq); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if bulkReq.Action != "payout" {
		response.BadRequest(c, "Invalid action for bulk payout")
		return
	}

	successCount := 0
	failedIDs := make([]uint, 0)

	for _, id := range bulkReq.IDs {
		err := h.referralService.MarkReferralAsPaid(c.Request.Context(), id)
		if err != nil {
			logger.Error("Failed to process payout in bulk operation",
				logger.Uint("referral_id", id),
				logger.ErrorField(err),
			)
			failedIDs = append(failedIDs, id)
			continue
		}
		successCount++
	}

	logger.Info("Admin executed bulk referral payout",
		logger.Int("requested_count", len(bulkReq.IDs)),
		logger.Int("success_count", successCount),
		logger.Int("failed_count", len(failedIDs)),
		logger.String("admin_action", "bulk_process_payouts"),
	)

	response.SuccessWithMessage(c, "Bulk referral payout completed", gin.H{
		"requested_count": len(bulkReq.IDs),
		"success_count":   successCount,
		"failed_count":    len(failedIDs),
		"failed_ids":      failedIDs,
	})
}

// Campaign Management Methods

// ListCampaigns godoc
// @Summary List referral campaigns
// @Description Get paginated list of all referral campaigns (Admin only)
// @Tags Admin-Referral-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param status query string false "Campaign status" Enums(active,paused,ended)
// @Success 200 {object} response.StandardListResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/referrals/campaigns [get]
func (h *AdminReferralHandler) ListCampaigns(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	offset := (page - 1) * limit

	serviceReq := &dto.GetReferralCampaignsRequest{
		Status: c.Query("status"),
		Limit:  limit,
		Offset: offset,
	}

	campaigns, total, err := h.referralCampaignService.GetReferralCampaigns(c.Request.Context(), serviceReq)
	if err != nil {
		logger.Error("Admin failed to list referral campaigns", logger.ErrorField(err))
		response.InternalServerError(c, "Failed to list referral campaigns")
		return
	}

	// Convert to responses
	var campaignResponses []*dto.ReferralCampaignResponse
	for _, campaign := range campaigns {
		campaignResponses = append(campaignResponses, dto.ToReferralCampaignResponse(campaign))
	}

	response.SuccessList(c, campaignResponses, page, limit, total)
}

// CreateCampaign godoc
// @Summary Create referral campaign
// @Description Create a new referral campaign (Admin only)
// @Tags Admin-Referral-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param campaign body dto.CreateReferralCampaignRequest true "Campaign creation data"
// @Success 201 {object} response.StandardResponse{data=dto.ReferralCampaignResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/referrals/campaigns [post]
func (h *AdminReferralHandler) CreateCampaign(c *gin.Context) {
	var createReq dto.CreateReferralCampaignRequest
	if err := c.ShouldBindJSON(&createReq); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	campaign, err := h.referralCampaignService.CreateReferralCampaign(c.Request.Context(), &createReq)
	if err != nil {
		logger.Error("Admin failed to create referral campaign",
			logger.String("name", createReq.Name),
			logger.ErrorField(err),
		)
		response.InternalServerError(c, "Failed to create referral campaign")
		return
	}

	logger.Info("Admin created new referral campaign",
		logger.Uint("campaign_id", campaign.ID),
		logger.String("name", campaign.Name),
		logger.String("admin_action", "create_campaign"),
	)

	response.Created(c, dto.ToReferralCampaignResponse(campaign))
}

// Invite Code Management Methods

// ListInviteCodes godoc
// @Summary List invite codes
// @Description Get paginated list of all invite codes (Admin only)
// @Tags Admin-Referral-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param status query string false "Invite code status" Enums(active,used,disabled)
// @Success 200 {object} response.StandardListResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/referrals/invite-codes [get]
func (h *AdminReferralHandler) ListInviteCodes(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	// This would call the invite code service
	// For now, return empty response
	response.SuccessList(c, []gin.H{}, page, limit, 0)
}
