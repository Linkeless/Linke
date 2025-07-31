package management

import (
	"linke/internal/handler/admin/invoice/shared"
	"linke/internal/logger"
	"linke/internal/middleware"
	"linke/internal/model"
	"linke/internal/response"
	"linke/internal/service"

	"github.com/gin-gonic/gin"
)

// InvoiceCRUDHandler handles invoice CRUD operations
type InvoiceCRUDHandler struct {
	*shared.BaseHandler
}

// NewInvoiceCRUDHandler creates a new invoice CRUD handler
func NewInvoiceCRUDHandler(invoiceService *service.InvoiceService) *InvoiceCRUDHandler {
	return &InvoiceCRUDHandler{
		BaseHandler: shared.NewBaseHandler(invoiceService),
	}
}

// CreateInvoiceFromOrderRequest represents the request to create invoice from order
type CreateInvoiceFromOrderRequest struct {
	OrderID  uint                          `json:"order_id" binding:"required"`
	Options  *service.CreateInvoiceRequest `json:"options,omitempty"`
	AutoSend bool                          `json:"auto_send,omitempty"`
}

// CreateInvoice godoc
// @Summary [Admin] Create invoice
// @Description Create a new invoice (Admin only)
// @Tags Admin-Invoice-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param invoice body service.CreateInvoiceRequest true "Invoice data"
// @Success 201 {object} response.StandardResponse{data=model.InvoiceResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/invoices [post]
func (h *InvoiceCRUDHandler) CreateInvoice(c *gin.Context) {
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

	// Bind request
	var req service.CreateInvoiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request data", err.Error())
		return
	}

	// Create invoice
	invoice, err := h.InvoiceService.CreateInvoice(c.Request.Context(), &req)
	if err != nil {
		logger.Error("Failed to create invoice", logger.Error2("error", err), logger.Uint("admin_id", user.ID))
		response.InternalServerError(c, "Failed to create invoice", err.Error())
		return
	}

	response.CreatedWithMessage(c, "Invoice created successfully", invoice.ToResponse())
}

// CreateInvoiceFromOrder godoc
// @Summary [Admin] Create invoice from order
// @Description Create an invoice from an existing subscription order (Admin only)
// @Tags Admin-Invoice-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateInvoiceFromOrderRequest true "Invoice from order data"
// @Success 201 {object} response.StandardResponse{data=model.InvoiceResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/invoices/from-order [post]
func (h *InvoiceCRUDHandler) CreateInvoiceFromOrder(c *gin.Context) {
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

	// Bind request
	var req CreateInvoiceFromOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request data", err.Error())
		return
	}

	// Validate order ID
	if err := h.Validator.ValidateOrderID(req.OrderID); err != nil {
		response.BadRequest(c, "Invalid order ID", err.Error())
		return
	}

	// Set auto_send in options if provided
	if req.Options != nil {
		req.Options.AutoSend = req.AutoSend
	}

	// Create invoice from order
	invoice, err := h.InvoiceService.CreateInvoiceFromOrder(c.Request.Context(), req.OrderID, req.Options)
	if err != nil {
		logger.Error("Failed to create invoice from order", 
			logger.Error2("error", err), 
			logger.Uint("admin_id", user.ID),
			logger.Uint("order_id", req.OrderID))
		response.InternalServerError(c, "Failed to create invoice from order", err.Error())
		return
	}

	response.CreatedWithMessage(c, "Invoice created from order successfully", invoice.ToResponse())
}

// UpdateInvoice godoc
// @Summary [Admin] Update invoice
// @Description Update an invoice (Admin only)
// @Tags Admin-Invoice-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Invoice ID"
// @Param invoice body service.UpdateInvoiceRequest true "Updated invoice data"
// @Success 200 {object} response.StandardResponse{data=model.InvoiceResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/invoices/{id} [put]
func (h *InvoiceCRUDHandler) UpdateInvoice(c *gin.Context) {
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
	var req service.UpdateInvoiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request data", err.Error())
		return
	}

	// Update invoice
	invoice, err := h.InvoiceService.UpdateInvoice(c.Request.Context(), invoiceID, &req)
	if err != nil {
		if err.Error() == "invoice not found" {
			response.NotFound(c, "Invoice not found")
			return
		}
		logger.Error("Failed to update invoice", 
			logger.Error2("error", err), 
			logger.Uint("invoice_id", invoiceID),
			logger.Uint("admin_id", user.ID))
		response.InternalServerError(c, "Failed to update invoice", err.Error())
		return
	}

	response.OK(c, "Invoice updated successfully", invoice.ToResponse())
}