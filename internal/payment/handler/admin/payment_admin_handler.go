package admin

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"linke/internal/payment/service/command"
	"linke/internal/payment/service/query"
	"linke/internal/response"
)

// PaymentAdminHandler handles admin payment operations
type PaymentAdminHandler struct {
	commandHandler *command.PaymentCommandHandler
	queryHandler   *query.PaymentQueryHandler
}

// NewPaymentAdminHandler creates a new PaymentAdminHandler
func NewPaymentAdminHandler(
	commandHandler *command.PaymentCommandHandler,
	queryHandler *query.PaymentQueryHandler,
) *PaymentAdminHandler {
	return &PaymentAdminHandler{
		commandHandler: commandHandler,
		queryHandler:   queryHandler,
	}
}

// GetPayment godoc
// @Summary Get payment by ID
// @Description Get detailed payment information by ID (Admin)
// @Tags Admin - Payments
// @Accept json
// @Produce json
// @Param id path int true "Payment ID"
// @Success 200 {object} response.StandardResponse{data=query.PaymentDTO}
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /admin/payments/{id} [get]
// @Security BearerAuth
func (h *PaymentAdminHandler) GetPayment(c *gin.Context) {
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

	response.OK(c, "Payment retrieved successfully", payment)
}

// ListPayments godoc
// @Summary List payments with filters
// @Description List all payments with filtering and pagination (Admin)
// @Tags Admin - Payments
// @Accept json
// @Produce json
// @Param user_id query int false "User ID"
// @Param invoice_id query int false "Invoice ID"
// @Param status query string false "Payment status"
// @Param payment_gateway query string false "Payment gateway"
// @Param payment_method query string false "Payment method"
// @Param currency query string false "Currency"
// @Param min_amount query number false "Minimum amount"
// @Param max_amount query number false "Maximum amount"
// @Param start_date query string false "Start date (YYYY-MM-DD)"
// @Param end_date query string false "End date (YYYY-MM-DD)"
// @Param search query string false "Search in payment number, transaction ID, notes"
// @Param sort_by query string false "Sort field" Enums(created_at,updated_at,amount,status)
// @Param sort_order query string false "Sort order" Enums(asc,desc)
// @Param limit query int false "Limit" default(20)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} response.StandardListResponse{data=query.PaymentListResult}
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /admin/payments [get]
// @Security BearerAuth
func (h *PaymentAdminHandler) ListPayments(c *gin.Context) {
	var queryReq query.ListPaymentsQuery
	if err := c.ShouldBindQuery(&queryReq); err != nil {
		response.BadRequest(c, "Invalid query parameters", err.Error())
		return
	}

	// Set default values
	if queryReq.Limit == 0 {
		queryReq.Limit = 20
	}

	result, err := h.queryHandler.ListPayments(c.Request.Context(), queryReq)
	if err != nil {
		response.InternalServerError(c, "Failed to list payments")
		return
	}

	response.OKPaginated(c, "Payments retrieved successfully", result.Payments, result.TotalCount, result.Limit, result.Offset)
}

// GetPaymentsByUser godoc
// @Summary Get payments by user ID
// @Description Get all payments for a specific user (Admin)
// @Tags Admin - Payments
// @Accept json
// @Produce json
// @Param user_id path int true "User ID"
// @Param limit query int false "Limit" default(20)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} response.StandardListResponse{data=[]query.PaymentDTO}
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /admin/users/{user_id}/payments [get]
// @Security BearerAuth
func (h *PaymentAdminHandler) GetPaymentsByUser(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("user_id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid user ID")
		return
	}

	var queryReq query.GetPaymentsByUserQuery
	if err := c.ShouldBindQuery(&queryReq); err != nil {
		response.BadRequest(c, "Invalid query parameters", err.Error())
		return
	}

	queryReq.UserID = uint(userID)
	if queryReq.Limit == 0 {
		queryReq.Limit = 20
	}

	result, err := h.queryHandler.GetPaymentsByUser(c.Request.Context(), queryReq)
	if err != nil {
		response.InternalServerError(c, "Failed to get payments by user")
		return
	}

	response.OKPaginated(c, "User payments retrieved successfully", result.Payments, result.TotalCount, result.Limit, result.Offset)
}

// GetPaymentStats godoc
// @Summary Get payment statistics
// @Description Get payment statistics with optional filters (Admin)
// @Tags Admin - Payments
// @Accept json
// @Produce json
// @Param user_id query int false "User ID"
// @Param start_date query string false "Start date (YYYY-MM-DD)"
// @Param end_date query string false "End date (YYYY-MM-DD)"
// @Param group_by query string false "Group by" Enums(day,week,month,year)
// @Success 200 {object} response.StandardResponse{data=query.PaymentStatsResult}
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /admin/payments/stats [get]
// @Security BearerAuth
func (h *PaymentAdminHandler) GetPaymentStats(c *gin.Context) {
	var queryReq query.GetPaymentStatsQuery
	if err := c.ShouldBindQuery(&queryReq); err != nil {
		response.BadRequest(c, "Invalid query parameters", err.Error())
		return
	}

	stats, err := h.queryHandler.GetPaymentStats(c.Request.Context(), queryReq)
	if err != nil {
		response.InternalServerError(c, "Failed to get payment stats")
		return
	}

	response.OK(c, "Payment statistics retrieved successfully", stats)
}

// CompletePayment godoc
// @Summary Complete a payment
// @Description Mark a payment as completed (Admin)
// @Tags Admin - Payments
// @Accept json
// @Produce json
// @Param id path int true "Payment ID"
// @Param request body command.CompletePaymentCommand true "Complete payment request"
// @Success 200 {object} response.StandardResponse{data=command.PaymentCommandResult}
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /admin/payments/{id}/complete [post]
// @Security BearerAuth
func (h *PaymentAdminHandler) CompletePayment(c *gin.Context) {
	paymentID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid payment ID")
		return
	}

	var cmd command.CompletePaymentCommand
	if err := c.ShouldBindJSON(&cmd); err != nil {
		response.BadRequest(c, "Invalid request body", err.Error())
		return
	}

	cmd.PaymentID = uint(paymentID)

	result, err := h.commandHandler.CompletePayment(c.Request.Context(), cmd)
	if err != nil {
		response.InternalServerError(c, "Failed to complete payment", err.Error())
		return
	}

	response.OK(c, "Payment completed successfully", result)
}

// FailPayment godoc
// @Summary Fail a payment
// @Description Mark a payment as failed (Admin)
// @Tags Admin - Payments
// @Accept json
// @Produce json
// @Param id path int true "Payment ID"
// @Param request body command.FailPaymentCommand true "Fail payment request"
// @Success 200 {object} response.StandardResponse{data=command.PaymentCommandResult}
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /admin/payments/{id}/fail [post]
// @Security BearerAuth
func (h *PaymentAdminHandler) FailPayment(c *gin.Context) {
	paymentID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid payment ID")
		return
	}

	var cmd command.FailPaymentCommand
	if err := c.ShouldBindJSON(&cmd); err != nil {
		response.BadRequest(c, "Invalid request body", err.Error())
		return
	}

	cmd.PaymentID = uint(paymentID)

	result, err := h.commandHandler.FailPayment(c.Request.Context(), cmd)
	if err != nil {
		response.InternalServerError(c, "Failed to fail payment", err.Error())
		return
	}

	response.OK(c, "Payment failed successfully", result)
}

// RefundPayment godoc
// @Summary Refund a payment
// @Description Process a refund for a completed payment (Admin)
// @Tags Admin - Payments
// @Accept json
// @Produce json
// @Param id path int true "Payment ID"
// @Param request body command.RefundPaymentCommand true "Refund payment request"
// @Success 200 {object} response.StandardResponse{data=command.RefundPaymentResult}
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /admin/payments/{id}/refund [post]
// @Security BearerAuth
func (h *PaymentAdminHandler) RefundPayment(c *gin.Context) {
	paymentID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid payment ID")
		return
	}

	var cmd command.RefundPaymentCommand
	if err := c.ShouldBindJSON(&cmd); err != nil {
		response.BadRequest(c, "Invalid request body", err.Error())
		return
	}

	cmd.PaymentID = uint(paymentID)

	result, err := h.commandHandler.RefundPayment(c.Request.Context(), cmd)
	if err != nil {
		response.InternalServerError(c, "Failed to refund payment", err.Error())
		return
	}

	response.OK(c, "Payment refunded successfully", result)
}

// UpdatePaymentNotes godoc
// @Summary Update payment notes
// @Description Update notes for a payment (Admin)
// @Tags Admin - Payments
// @Accept json
// @Produce json
// @Param id path int true "Payment ID"
// @Param request body command.UpdatePaymentNotesCommand true "Update notes request"
// @Success 200 {object} response.StandardResponse{data=command.PaymentCommandResult}
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /admin/payments/{id}/notes [put]
// @Security BearerAuth
func (h *PaymentAdminHandler) UpdatePaymentNotes(c *gin.Context) {
	paymentID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid payment ID")
		return
	}

	var cmd command.UpdatePaymentNotesCommand
	if err := c.ShouldBindJSON(&cmd); err != nil {
		response.BadRequest(c, "Invalid request body", err.Error())
		return
	}

	cmd.PaymentID = uint(paymentID)

	result, err := h.commandHandler.UpdatePaymentNotes(c.Request.Context(), cmd)
	if err != nil {
		response.InternalServerError(c, "Failed to update payment notes", err.Error())
		return
	}

	response.OK(c, "Payment notes updated successfully", result)
}

// DeletePayment godoc
// @Summary Delete a payment
// @Description Soft delete a payment (Admin)
// @Tags Admin - Payments
// @Accept json
// @Produce json
// @Param id path int true "Payment ID"
// @Success 200 {object} response.StandardResponse{data=command.PaymentCommandResult}
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /admin/payments/{id} [delete]
// @Security BearerAuth
func (h *PaymentAdminHandler) DeletePayment(c *gin.Context) {
	paymentID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid payment ID")
		return
	}

	cmd := command.DeletePaymentCommand{
		PaymentID: uint(paymentID),
	}

	result, err := h.commandHandler.DeletePayment(c.Request.Context(), cmd)
	if err != nil {
		response.InternalServerError(c, "Failed to delete payment", err.Error())
		return
	}

	response.OK(c, "Payment deleted successfully", result)
}
