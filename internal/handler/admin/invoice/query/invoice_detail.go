package query

import (
	"linke/internal/handler/admin/invoice/shared"
	"linke/internal/logger"
	"linke/internal/middleware"
	"linke/internal/model"
	"linke/internal/response"
	"linke/internal/service"

	"github.com/gin-gonic/gin"
)

// InvoiceDetailHandler handles invoice detail operations
type InvoiceDetailHandler struct {
	*shared.BaseHandler
}

// NewInvoiceDetailHandler creates a new invoice detail handler
func NewInvoiceDetailHandler(invoiceService *service.InvoiceService) *InvoiceDetailHandler {
	return &InvoiceDetailHandler{
		BaseHandler: shared.NewBaseHandler(invoiceService),
	}
}

// GetInvoice godoc
// @Summary [Admin] Get invoice by ID
// @Description Get an invoice by ID with full details (Admin only)
// @Tags Admin-Invoice-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Invoice ID"
// @Success 200 {object} response.StandardResponse{data=model.InvoiceResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/invoices/{id} [get]
func (h *InvoiceDetailHandler) GetInvoice(c *gin.Context) {
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

	// Get invoice with relations
	invoice, err := h.InvoiceService.GetInvoiceWithRelations(c.Request.Context(), invoiceID)
	if err != nil {
		if err.Error() == "invoice not found" {
			response.NotFound(c, "Invoice not found")
			return
		}
		logger.Error("Failed to get invoice", logger.Error2("error", err), logger.Uint("invoice_id", invoiceID))
		response.InternalServerError(c, "Failed to get invoice", err.Error())
		return
	}

	response.OK(c, "Invoice retrieved successfully", invoice.ToResponse())
}