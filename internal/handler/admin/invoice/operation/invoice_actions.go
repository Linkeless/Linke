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

// InvoiceActionHandler handles invoice action operations
type InvoiceActionHandler struct {
	*shared.BaseHandler
}

// NewInvoiceActionHandler creates a new invoice action handler
func NewInvoiceActionHandler(invoiceService *service.InvoiceService) *InvoiceActionHandler {
	return &InvoiceActionHandler{
		BaseHandler: shared.NewBaseHandler(invoiceService),
	}
}

// SendInvoice godoc
// @Summary [Admin] Send invoice
// @Description Send an invoice to the customer (Admin only)
// @Tags Admin-Invoice-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Invoice ID"
// @Success 200 {object} response.StandardResponse
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/invoices/{id}/send [post]
func (h *InvoiceActionHandler) SendInvoice(c *gin.Context) {
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

	// Send invoice
	if err := h.InvoiceService.SendInvoice(c.Request.Context(), invoiceID); err != nil {
		if err.Error() == "invoice not found" {
			response.NotFound(c, "Invoice not found")
			return
		}
		logger.Error("Failed to send invoice", 
			logger.Error2("error", err), 
			logger.Uint("invoice_id", invoiceID),
			logger.Uint("admin_id", user.ID))
		response.InternalServerError(c, "Failed to send invoice", err.Error())
		return
	}

	response.OK(c, "Invoice sent successfully", nil)
}