package api

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"linke/internal/payment/service/command"
	"linke/internal/payment/service/query"
	"linke/internal/response"
)

// PaymentHandler handles user payment operations
type PaymentHandler struct {
	commandHandler *command.PaymentCommandHandler
	queryHandler   *query.PaymentQueryHandler
}

// NewPaymentHandler creates a new PaymentHandler
func NewPaymentHandler(
	commandHandler *command.PaymentCommandHandler,
	queryHandler *query.PaymentQueryHandler,
) *PaymentHandler {
	return &PaymentHandler{
		commandHandler: commandHandler,
		queryHandler:   queryHandler,
	}
}

// CreatePayment godoc
// @Summary Create a new payment
// @Description Create a new payment for an invoice
// @Tags User - Payments
// @Accept json
// @Produce json
// @Param request body command.CreatePaymentCommand true "Create payment request"
// @Success 201 {object} response.StandardResponse{data=command.CreatePaymentResult}
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/payments [post]
// @Security BearerAuth
func (h *PaymentHandler) CreatePayment(c *gin.Context) {
	var cmd command.CreatePaymentCommand
	if err := c.ShouldBindJSON(&cmd); err != nil {
		response.BadRequest(c, "Invalid request body", err.Error())
		return
	}

	// TODO: Get user ID from JWT token
	// For now, using the provided user ID
	
	result, err := h.commandHandler.CreatePayment(c.Request.Context(), cmd)
	if err != nil {
		response.InternalServerError(c, "Failed to create payment", err.Error())
		return
	}

	response.CreatedWithMessage(c, "Payment created successfully", result)
}

// GetPayment godoc
// @Summary Get payment by ID
// @Description Get payment details by ID (User can only access their own payments)
// @Tags User - Payments
// @Accept json
// @Produce json
// @Param id path int true "Payment ID"
// @Success 200 {object} response.StandardResponse{data=query.PaymentDTO}
// @Failure 400 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/payments/{id} [get]
// @Security BearerAuth
func (h *PaymentHandler) GetPayment(c *gin.Context) {
	paymentID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid payment ID")
		return
	}

	queryReq := query.GetPaymentQuery{
		PaymentID: uint(paymentID),
	}

	payment, err := h.queryHandler.GetPayment(c.Request.Context(), queryReq)
	if err != nil {
		if err.Error() == "payment not found" {
			response.NotFound(c, "Payment not found")
			return
		}
		response.InternalServerError(c, "Failed to get payment")
		return
	}

	// TODO: Check if payment belongs to the current user
	// userID := getUserIDFromContext(c)
	// if payment.UserID != userID {
	//     response.ForbiddenResponse(c, "You can only access your own payments", nil)
	//     return
	// }

	response.OK(c, "Payment retrieved successfully", payment)
}

// GetPaymentByNumber godoc
// @Summary Get payment by number
// @Description Get payment details by payment number (User can only access their own payments)
// @Tags User - Payments
// @Accept json
// @Produce json
// @Param number path string true "Payment Number"
// @Success 200 {object} response.StandardResponse{data=query.PaymentDTO}
// @Failure 400 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/payments/number/{number} [get]
// @Security BearerAuth
func (h *PaymentHandler) GetPaymentByNumber(c *gin.Context) {
	paymentNumber := c.Param("number")
	if paymentNumber == "" {
		response.BadRequest(c, "Payment number is required")
		return
	}

	queryReq := query.GetPaymentByNumberQuery{
		PaymentNumber: paymentNumber,
	}

	payment, err := h.queryHandler.GetPaymentByNumber(c.Request.Context(), queryReq)
	if err != nil {
		if err.Error() == "payment not found" {
			response.NotFound(c, "Payment not found")
			return
		}
		response.InternalServerError(c, "Failed to get payment")
		return
	}

	// TODO: Check if payment belongs to the current user
	// userID := getUserIDFromContext(c)
	// if payment.UserID != userID {
	//     response.ForbiddenResponse(c, "You can only access your own payments", nil)
	//     return
	// }

	response.OK(c, "Payment retrieved successfully", payment)
}

// ListMyPayments godoc
// @Summary List user's payments
// @Description List current user's payments with pagination
// @Tags User - Payments
// @Accept json
// @Produce json
// @Param status query string false "Payment status"
// @Param payment_method query string false "Payment method"
// @Param currency query string false "Currency"
// @Param start_date query string false "Start date (YYYY-MM-DD)"
// @Param end_date query string false "End date (YYYY-MM-DD)"
// @Param sort_by query string false "Sort field" Enums(created_at,updated_at,amount)
// @Param sort_order query string false "Sort order" Enums(asc,desc)
// @Param limit query int false "Limit" default(20)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} response.StandardListResponse{data=[]query.PaymentDTO}
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/payments/my [get]
// @Security BearerAuth
func (h *PaymentHandler) ListMyPayments(c *gin.Context) {
	// TODO: Get user ID from JWT token
	// userID := getUserIDFromContext(c)
	userID := uint(1) // Placeholder

	var queryReq query.GetPaymentsByUserQuery
	if err := c.ShouldBindQuery(&queryReq); err != nil {
		response.BadRequest(c, "Invalid query parameters", err.Error())
		return
	}

	queryReq.UserID = userID
	if queryReq.Limit == 0 {
		queryReq.Limit = 20
	}

	result, err := h.queryHandler.GetPaymentsByUser(c.Request.Context(), queryReq)
	if err != nil {
		response.InternalServerError(c, "Failed to get payments")
		return
	}

	response.OKPaginated(c, "Payments retrieved successfully", result.Payments, result.TotalCount, result.Limit, result.Offset)
}

// GetPaymentsByInvoice godoc
// @Summary Get payments by invoice ID
// @Description Get all payments for a specific invoice (User can only access their own invoices)
// @Tags User - Payments
// @Accept json
// @Produce json
// @Param invoice_id path int true "Invoice ID"
// @Success 200 {object} response.StandardResponse{data=[]query.PaymentDTO}
// @Failure 400 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/invoices/{invoice_id}/payments [get]
// @Security BearerAuth
func (h *PaymentHandler) GetPaymentsByInvoice(c *gin.Context) {
	invoiceID, err := strconv.ParseUint(c.Param("invoice_id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid invoice ID")
		return
	}

	queryReq := query.GetPaymentsByInvoiceQuery{
		InvoiceID: uint(invoiceID),
	}

	payments, err := h.queryHandler.GetPaymentsByInvoice(c.Request.Context(), queryReq)
	if err != nil {
		response.InternalServerError(c, "Failed to get payments by invoice")
		return
	}

	// TODO: Check if invoice belongs to the current user
	// This should be done by checking the invoice ownership first

	response.OK(c, "Invoice payments retrieved successfully", payments)
}

// CancelPayment godoc
// @Summary Cancel a payment
// @Description Cancel a pending payment (User can only cancel their own payments)
// @Tags User - Payments
// @Accept json
// @Produce json
// @Param id path int true "Payment ID"
// @Param request body command.CancelPaymentCommand true "Cancel payment request"
// @Success 200 {object} response.StandardResponse{data=command.PaymentCommandResult}
// @Failure 400 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/payments/{id}/cancel [post]
// @Security BearerAuth
func (h *PaymentHandler) CancelPayment(c *gin.Context) {
	paymentID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid payment ID")
		return
	}

	var cmd command.CancelPaymentCommand
	if err := c.ShouldBindJSON(&cmd); err != nil {
		response.BadRequest(c, "Invalid request body", err.Error())
		return
	}

	cmd.PaymentID = uint(paymentID)

	// TODO: Check if payment belongs to the current user before cancelling
	// This should be done in the command handler or as a pre-check here

	result, err := h.commandHandler.CancelPayment(c.Request.Context(), cmd)
	if err != nil {
		response.InternalServerError(c, "Failed to cancel payment", err.Error())
		return
	}

	response.OK(c, "Payment cancelled successfully", result)
}

// GetMyPaymentStats godoc
// @Summary Get user's payment statistics
// @Description Get payment statistics for the current user
// @Tags User - Payments
// @Accept json
// @Produce json
// @Param start_date query string false "Start date (YYYY-MM-DD)"
// @Param end_date query string false "End date (YYYY-MM-DD)"
// @Param group_by query string false "Group by" Enums(day,week,month,year)
// @Success 200 {object} response.StandardResponse{data=query.PaymentStatsResult}
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/payments/my/stats [get]
// @Security BearerAuth
func (h *PaymentHandler) GetMyPaymentStats(c *gin.Context) {
	// TODO: Get user ID from JWT token
	// userID := getUserIDFromContext(c)
	userID := uint(1) // Placeholder

	var queryReq query.GetPaymentStatsQuery
	if err := c.ShouldBindQuery(&queryReq); err != nil {
		response.BadRequest(c, "Invalid query parameters", err.Error())
		return
	}

	queryReq.UserID = &userID

	stats, err := h.queryHandler.GetPaymentStats(c.Request.Context(), queryReq)
	if err != nil {
		response.InternalServerError(c, "Failed to get payment stats")
		return
	}

	response.OK(c, "Payment statistics retrieved successfully", stats)
}

// UpdatePaymentDetails godoc
// @Summary Update payment details
// @Description Update payment gateway details like URLs and intent ID
// @Tags User - Payments
// @Accept json
// @Produce json
// @Param id path int true "Payment ID"
// @Param request body command.UpdatePaymentDetailsCommand true "Update payment details request"
// @Success 200 {object} response.StandardResponse{data=command.PaymentCommandResult}
// @Failure 400 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/payments/{id}/details [put]
// @Security BearerAuth
func (h *PaymentHandler) UpdatePaymentDetails(c *gin.Context) {
	paymentID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid payment ID")
		return
	}

	var cmd command.UpdatePaymentDetailsCommand
	if err := c.ShouldBindJSON(&cmd); err != nil {
		response.BadRequest(c, "Invalid request body", err.Error())
		return
	}

	cmd.PaymentID = uint(paymentID)

	// TODO: Check if payment belongs to the current user
	// This should be done in the command handler or as a pre-check here

	result, err := h.commandHandler.UpdatePaymentDetails(c.Request.Context(), cmd)
	if err != nil {
		response.InternalServerError(c, "Failed to update payment details", err.Error())
		return
	}

	response.OK(c, "Payment details updated successfully", result)
}

// Helper functions (TODO: Implement these based on your authentication system)

// func getUserIDFromContext(c *gin.Context) uint {
//     // Extract user ID from JWT token or session
//     // This is a placeholder implementation
//     return 1
// }