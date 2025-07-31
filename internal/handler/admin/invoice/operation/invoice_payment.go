package operation

import (
	"linke/internal/handler/admin/invoice/shared"
	"linke/internal/logger"
	"linke/internal/middleware"
	"linke/internal/model"
	"linke/internal/response"
	"linke/internal/service"

	"github.com/gin-gonic/gin"
)

// InvoicePaymentHandler handles invoice payment operations
type InvoicePaymentHandler struct {
	*shared.BaseHandler
}

// NewInvoicePaymentHandler creates a new invoice payment handler
func NewInvoicePaymentHandler(invoiceService *service.InvoiceService) *InvoicePaymentHandler {
	return &InvoicePaymentHandler{
		BaseHandler: shared.NewBaseHandler(invoiceService),
	}
}

// MarkInvoiceAsPaidRequest represents the request to mark invoice as paid
type MarkInvoiceAsPaidRequest struct {
	PaidAmount       float64 `json:"paid_amount" binding:"required,min=0" example:"29.99"`
	PaymentReference string  `json:"payment_reference,omitempty" example:"REF123456"`
}

// MarkInvoiceAsPaid godoc
// @Summary [Admin] Mark invoice as paid
// @Description Mark an invoice as paid (Admin only)
// @Tags Admin-Invoice-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Invoice ID"
// @Param request body MarkInvoiceAsPaidRequest true "Payment data"
// @Success 200 {object} response.StandardResponse
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/invoices/{id}/mark-paid [post]
func (h *InvoicePaymentHandler) MarkInvoiceAsPaid(c *gin.Context) {
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

	// Check if user is admin
	if !user.IsAdmin() {
		response.Forbidden(c, "Admin access required")
		return
	}

	// Validate invoice ID
	invoiceID, err := h.Validator.ValidateInvoiceID(c)
	if err != nil {
		return // Response already sent by validator
	}

	// Bind request
	var req MarkInvoiceAsPaidRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request data", err.Error())
		return
	}

	// Validate paid amount
	if req.PaidAmount <= 0 {
		response.BadRequest(c, "Paid amount must be greater than 0")
		return
	}

	// Mark invoice as paid
	if err := h.InvoiceService.MarkInvoiceAsPaid(c.Request.Context(), invoiceID, req.PaidAmount, req.PaymentReference); err != nil {
		if err.Error() == "invoice not found" {
			response.NotFound(c, "Invoice not found")
			return
		}
		logger.Error("Failed to mark invoice as paid", 
			logger.Error2("error", err), 
			logger.Uint("invoice_id", invoiceID),
			logger.Uint("admin_id", user.ID))
		response.InternalServerError(c, "Failed to mark invoice as paid", err.Error())
		return
	}

	response.OK(c, "Invoice marked as paid successfully", nil)
}