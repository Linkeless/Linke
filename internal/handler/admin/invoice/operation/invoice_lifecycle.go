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

// InvoiceLifecycleHandler handles invoice lifecycle operations
type InvoiceLifecycleHandler struct {
	*shared.BaseHandler
}

// NewInvoiceLifecycleHandler creates a new invoice lifecycle handler
func NewInvoiceLifecycleHandler(invoiceService *service.InvoiceService) *InvoiceLifecycleHandler {
	return &InvoiceLifecycleHandler{
		BaseHandler: shared.NewBaseHandler(invoiceService),
	}
}

// VoidInvoiceRequest represents the request to void an invoice
type VoidInvoiceRequest struct {
	Reason string `json:"reason" binding:"required" example:"Customer request"`
}

// VoidInvoice godoc
// @Summary [Admin] Void invoice
// @Description Void an invoice (Admin only)
// @Tags Admin-Invoice-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Invoice ID"
// @Param request body VoidInvoiceRequest true "Void reason"
// @Success 200 {object} response.StandardResponse
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/invoices/{id}/void [post]
func (h *InvoiceLifecycleHandler) VoidInvoice(c *gin.Context) {
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
	var req VoidInvoiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request data", err.Error())
		return
	}

	// Validate void reason
	if err := h.Validator.ValidateVoidReason(req.Reason); err != nil {
		response.BadRequest(c, "Invalid void reason", err.Error())
		return
	}

	// Void invoice
	if err := h.InvoiceService.VoidInvoice(c.Request.Context(), invoiceID, req.Reason); err != nil {
		if err.Error() == "invoice not found" {
			response.NotFound(c, "Invoice not found")
			return
		}
		logger.Error("Failed to void invoice", 
			logger.Error2("error", err), 
			logger.Uint("invoice_id", invoiceID),
			logger.Uint("admin_id", user.ID))
		response.InternalServerError(c, "Failed to void invoice", err.Error())
		return
	}

	response.OK(c, "Invoice voided successfully", nil)
}

// DeleteInvoice godoc
// @Summary [Admin] Delete invoice
// @Description Soft delete an invoice (Admin only)
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
// @Router /admin/invoices/{id} [delete]
func (h *InvoiceLifecycleHandler) DeleteInvoice(c *gin.Context) {
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

	// Delete invoice
	if err := h.InvoiceService.DeleteInvoice(c.Request.Context(), invoiceID); err != nil {
		if err.Error() == "invoice not found" {
			response.NotFound(c, "Invoice not found")
			return
		}
		logger.Error("Failed to delete invoice", 
			logger.Error2("error", err), 
			logger.Uint("invoice_id", invoiceID),
			logger.Uint("admin_id", user.ID))
		response.InternalServerError(c, "Failed to delete invoice", err.Error())
		return
	}

	response.OK(c, "Invoice deleted successfully", nil)
}