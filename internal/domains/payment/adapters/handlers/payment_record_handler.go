package handlers

import (
	"strconv"

	"linke/internal/domains/payment/dto"
	"linke/internal/domains/payment/usecases/interfaces"
	userEntities "linke/internal/domains/user/entities"
	"linke/internal/shared/logger"
	"linke/internal/shared/middleware"
	"linke/internal/shared/response"

	"github.com/gin-gonic/gin"
)

// PaymentRecordHandler handles HTTP requests for payment record management
type PaymentRecordHandler struct {
	paymentService interfaces.PaymentService
}

// NewPaymentRecordHandler creates a new payment record handler
func NewPaymentRecordHandler(paymentService interfaces.PaymentService) *PaymentRecordHandler {
	return &PaymentRecordHandler{
		paymentService: paymentService,
	}
}

// GetMyPaymentOrders godoc
// @Summary [User] Get my payment orders
// @Description Get current user's payment orders
// @Tags User-Payment
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param limit query int false "Limit results" minimum(1) maximum(100) example(10)
// @Param offset query int false "Offset results" minimum(0) example(0)
// @Success 200 {object} response.PaginatedResponse{data=[]dto.PaymentRecordResponse}
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /payment/orders/my [get]
func (h *PaymentRecordHandler) GetMyPaymentOrders(c *gin.Context) {
	// Get current user from context
	userValue, exists := c.Get(middleware.AuthContextKey)
	if !exists {
		response.Unauthorized(c, "Authentication required")
		return
	}

	user, ok := userValue.(*userEntities.User)
	if !ok {
		response.Unauthorized(c, "Invalid user context")
		return
	}

	// Parse query parameters
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 100 {
		limit = 10
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	// Get user payment records
	records, totalCount, err := h.paymentService.GetUserPaymentRecords(c.Request.Context(), user.ID, limit, offset)
	if err != nil {
		logger.Error("Failed to get user payment orders", logger.ErrorField(err), logger.Uint("user_id", user.ID))
		response.InternalServerError(c, "Failed to get payment orders", err.Error())
		return
	}

	// Convert to response format
	var recordResponses []*dto.PaymentRecordResponse
	for _, record := range records {
		recordResponses = append(recordResponses, dto.ToPaymentRecordUserResponse(record))
	}

	response.OKPaginated(c, "My payment orders retrieved successfully", recordResponses, totalCount, limit, offset)
}