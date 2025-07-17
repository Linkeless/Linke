package handler

import (
	"strconv"

	"linke/internal/logger"
	"linke/internal/middleware"
	"linke/internal/model"
	"linke/internal/response"
	"linke/internal/service"

	"github.com/gin-gonic/gin"
)

type InviteCodeHandler struct {
	inviteCodeService      *service.InviteCodeService
	inviteCodeUsageService *service.InviteCodeUsageService
}

func NewInviteCodeHandler(inviteCodeService *service.InviteCodeService, inviteCodeUsageService *service.InviteCodeUsageService) *InviteCodeHandler {
	return &InviteCodeHandler{
		inviteCodeService:      inviteCodeService,
		inviteCodeUsageService: inviteCodeUsageService,
	}
}

// CreateInviteCode godoc
// @Summary [User] Create invite code
// @Description Create a new invite code
// @Tags invite-codes
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param invite_code body service.CreateInviteCodeRequest true "Invite code data"
// @Success 201 {object} response.StandardResponse{data=model.InviteCodeResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /invite-codes [post]
func (h *InviteCodeHandler) CreateInviteCode(c *gin.Context) {
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

	var req service.CreateInviteCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	inviteCode, err := h.inviteCodeService.CreateInviteCode(c.Request.Context(), user.ID, &req)
	if err != nil {
		logger.Error("Failed to create invite code",
			logger.Uint("user_id", user.ID),
			logger.Error2("error", err),
		)
		response.BadRequest(c, err.Error())
		return
	}

	response.Created(c, inviteCode.ToResponse())
}


// GetInviteCode godoc
// @Summary [User] Get invite code by ID
// @Description Get invite code details by ID (only creator or admin can access)
// @Tags invite-codes
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Invite code ID"
// @Success 200 {object} response.StandardResponse{data=model.InviteCodeResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Router /invite-codes/{id} [get]
func (h *InviteCodeHandler) GetInviteCode(c *gin.Context) {
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

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid invite code ID")
		return
	}

	inviteCode, err := h.inviteCodeService.GetInviteCodeByIDWithRelations(c.Request.Context(), uint(id))
	if err != nil {
		logger.Error("Failed to get invite code",
			logger.Uint("invite_code_id", uint(id)),
			logger.Error2("error", err),
		)
		response.NotFound(c, "Invite code not found")
		return
	}

	// Check if user is the creator or admin
	if inviteCode.CreatedByID != user.ID && !user.IsAdmin() {
		response.Forbidden(c, "You can only access your own invite codes")
		return
	}

	response.Success(c, inviteCode.ToResponse())
}

// ValidateInviteCode godoc
// @Summary [Public] Validate invite code
// @Description Validate if an invite code can be used
// @Tags invite-codes
// @Accept json
// @Produce json
// @Param code path string true "Invite code"
// @Success 200 {object} response.StandardResponse{data=model.InviteCodeResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 404 {object} response.NotFoundResponse
// @Router /invite-codes/validate/{code} [get]
func (h *InviteCodeHandler) ValidateInviteCode(c *gin.Context) {
	code := c.Param("code")
	if code == "" {
		response.BadRequest(c, "Invite code is required")
		return
	}

	inviteCode, err := h.inviteCodeService.ValidateInviteCode(c.Request.Context(), code)
	if err != nil {
		logger.Warn("Invite code validation failed",
			logger.String("code", code),
			logger.Error2("error", err),
		)
		response.NotFound(c, err.Error())
		return
	}

	response.Success(c, inviteCode.ToPublicResponse())
}

// UpdateInviteCodeStatus godoc
// @Summary [User] Update invite code status
// @Description Update the status of an invite code (only creator or admin can update)
// @Tags invite-codes
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Invite code ID"
// @Param status body map[string]string true "New status"
// @Success 200 {object} response.StandardResponse{data=model.InviteCodeResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Router /invite-codes/{id}/status [put]
func (h *InviteCodeHandler) UpdateInviteCodeStatus(c *gin.Context) {
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

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid invite code ID")
		return
	}

	var req struct {
		Status string `json:"status" binding:"required,oneof=active disabled"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// Check if user owns the invite code
	inviteCode, err := h.inviteCodeService.GetInviteCodeByID(c.Request.Context(), uint(id))
	if err != nil {
		response.NotFound(c, "Invite code not found")
		return
	}

	if inviteCode.CreatedByID != user.ID && !user.IsAdmin() {
		response.Forbidden(c, "You can only update your own invite codes")
		return
	}

	updatedCode, err := h.inviteCodeService.UpdateInviteCodeStatus(c.Request.Context(), uint(id), req.Status)
	if err != nil {
		logger.Error("Failed to update invite code status",
			logger.Uint("invite_code_id", uint(id)),
			logger.String("status", req.Status),
			logger.Error2("error", err),
		)
		response.InternalServerError(c, "Failed to update invite code status")
		return
	}

	response.Success(c, updatedCode.ToResponse())
}

// DeleteInviteCode godoc
// @Summary [User] Delete invite code
// @Description Delete an invite code (only creator or admin can delete)
// @Tags invite-codes
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Invite code ID"
// @Success 200 {object} response.MessageOnlyResponse
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Router /invite-codes/{id} [delete]
func (h *InviteCodeHandler) DeleteInviteCode(c *gin.Context) {
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

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid invite code ID")
		return
	}

	// Check if user owns the invite code
	inviteCode, err := h.inviteCodeService.GetInviteCodeByID(c.Request.Context(), uint(id))
	if err != nil {
		response.NotFound(c, "Invite code not found")
		return
	}

	if inviteCode.CreatedByID != user.ID && !user.IsAdmin() {
		response.Forbidden(c, "You can only delete your own invite codes")
		return
	}

	if err := h.inviteCodeService.DeleteInviteCode(c.Request.Context(), uint(id)); err != nil {
		logger.Error("Failed to delete invite code",
			logger.Uint("invite_code_id", uint(id)),
			logger.Error2("error", err),
		)
		response.InternalServerError(c, "Failed to delete invite code")
		return
	}

	response.SuccessWithMessage(c, "Invite code deleted successfully", nil)
}

// GetInviteCodeStats godoc
// @Summary [Admin] Get invite code statistics
// @Description Get statistics about invite codes (admin only)
// @Tags invite-codes
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.StandardResponse{data=map[string]interface{}}
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/invite-codes/stats [get]
func (h *InviteCodeHandler) GetInviteCodeStats(c *gin.Context) {
	stats, err := h.inviteCodeService.GetInviteCodeStats(c.Request.Context())
	if err != nil {
		logger.Error("Failed to get invite code stats",
			logger.Error2("error", err),
		)
		response.InternalServerError(c, "Failed to get invite code statistics")
		return
	}

	response.Success(c, stats)
}

// ListAllInviteCodes godoc
// @Summary [Admin] List all invite codes
// @Description Get list of all invite codes with pagination (admin only)
// @Tags invite-codes
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} response.StandardListResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/invite-codes [get]
func (h *InviteCodeHandler) ListAllInviteCodes(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	offset := (page - 1) * limit

	codes, total, err := h.inviteCodeService.ListAllInviteCodes(c.Request.Context(), limit, offset)
	if err != nil {
		logger.Error("Failed to list all invite codes",
			logger.Error2("error", err),
		)
		response.InternalServerError(c, "Failed to list invite codes")
		return
	}

	// Convert to response
	var responseData []*model.InviteCodeResponse
	for _, code := range codes {
		responseData = append(responseData, code.ToResponse())
	}

	response.SuccessList(c, responseData, page, limit, total)
}

// GetMyInviteCodes godoc
// @Summary [User] Get my invite codes
// @Description Get invite codes created by current user
// @Tags invite-codes
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} response.StandardListResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /invite-codes/my [get]
func (h *InviteCodeHandler) GetMyInviteCodes(c *gin.Context) {
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

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	offset := (page - 1) * limit

	codes, total, err := h.inviteCodeService.ListInviteCodesByCreator(c.Request.Context(), user.ID, limit, offset)
	if err != nil {
		logger.Error("Failed to get user invite codes",
			logger.Uint("user_id", user.ID),
			logger.Error2("error", err),
		)
		response.InternalServerError(c, "Failed to get invite codes")
		return
	}

	// Convert to response
	var responseData []*model.InviteCodeResponse
	for _, code := range codes {
		responseData = append(responseData, code.ToResponse())
	}

	response.SuccessList(c, responseData, page, limit, total)
}

// GetInviteCodeUsages godoc
// @Summary [User] Get invite code usages
// @Description Get usage records for a specific invite code (only creator or admin can access)
// @Tags invite-codes
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Invite code ID"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} response.StandardListResponse
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Router /invite-codes/{id}/usages [get]
func (h *InviteCodeHandler) GetInviteCodeUsages(c *gin.Context) {
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

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid invite code ID")
		return
	}

	// Check if user owns the invite code
	inviteCode, err := h.inviteCodeService.GetInviteCodeByID(c.Request.Context(), uint(id))
	if err != nil {
		response.NotFound(c, "Invite code not found")
		return
	}

	if inviteCode.CreatedByID != user.ID && !user.IsAdmin() {
		response.Forbidden(c, "You can only access your own invite codes")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	offset := (page - 1) * limit

	usages, total, err := h.inviteCodeUsageService.GetUsagesByInviteCode(c.Request.Context(), uint(id), limit, offset)
	if err != nil {
		logger.Error("Failed to get invite code usages",
			logger.Uint("invite_code_id", uint(id)),
			logger.Error2("error", err),
		)
		response.InternalServerError(c, "Failed to get invite code usages")
		return
	}

	// Load related data
	if err := h.inviteCodeUsageService.LoadRelatedData(c.Request.Context(), usages); err != nil {
		logger.Error("Failed to load related data for usages",
			logger.Uint("invite_code_id", uint(id)),
			logger.Error2("error", err),
		)
	}

	// Convert to response
	var responseData []*model.InviteCodeUsageResponse
	for _, usage := range usages {
		responseData = append(responseData, usage.ToResponse())
	}

	response.SuccessList(c, responseData, page, limit, total)
}