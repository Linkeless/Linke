package admin

import (
	"strconv"

	"linke/internal/logger"
	"linke/internal/middleware"
	"linke/internal/model"
	"linke/internal/response"
	"linke/internal/service"

	"github.com/gin-gonic/gin"
)

type AdminInvoiceHandler struct {
	invoiceService *service.InvoiceService
}

func NewAdminInvoiceHandler(invoiceService *service.InvoiceService) *AdminInvoiceHandler {
	return &AdminInvoiceHandler{
		invoiceService: invoiceService,
	}
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
func (h *AdminInvoiceHandler) CreateInvoice(c *gin.Context) {
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
	invoice, err := h.invoiceService.CreateInvoice(c.Request.Context(), &req)
	if err != nil {
		logger.Error("Failed to create invoice", logger.Error2("error", err), logger.Uint("admin_id", user.ID))
		response.InternalServerError(c, "Failed to create invoice", err.Error())
		return
	}

	response.CreatedWithMessage(c, "Invoice created successfully", invoice.ToResponse())
}

// CreateInvoiceFromOrderRequest represents the request to create invoice from order
type CreateInvoiceFromOrderRequest struct {
	OrderID  uint                          `json:"order_id" binding:"required"`
	Options  *service.CreateInvoiceRequest `json:"options,omitempty"`
	AutoSend bool                          `json:"auto_send,omitempty"`
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
func (h *AdminInvoiceHandler) CreateInvoiceFromOrder(c *gin.Context) {
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

	// Set auto_send in options if provided
	if req.Options != nil {
		req.Options.AutoSend = req.AutoSend
	}

	// Create invoice from order
	invoice, err := h.invoiceService.CreateInvoiceFromOrder(c.Request.Context(), req.OrderID, req.Options)
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

// ListInvoices godoc
// @Summary [Admin] List all invoices
// @Description Get all invoices with filtering (Admin only)
// @Tags Admin-Invoice-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param user_id query int false "Filter by user ID"
// @Param status query string false "Filter by status" Enums(draft,sent,paid,overdue,cancelled,voided)
// @Param invoice_type query string false "Filter by invoice type" Enums(standard,proforma,credit_note)
// @Param currency query string false "Filter by currency"
// @Param start_date query string false "Start date filter (YYYY-MM-DD)"
// @Param end_date query string false "End date filter (YYYY-MM-DD)"
// @Param overdue query bool false "Filter overdue invoices"
// @Param search query string false "Search in invoice number, billing name, email"
// @Param sort_by query string false "Sort by field" Enums(created_at,issued_at,due_at,paid_at,total_amount,status) default(created_at)
// @Param sort_order query string false "Sort order" Enums(asc,desc) default(desc)
// @Param limit query int false "Limit results" minimum(1) maximum(100) example(10)
// @Param offset query int false "Offset results" minimum(0) example(0)
// @Success 200 {object} response.PaginatedResponse{data=[]model.InvoiceResponse}
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/invoices [get]
func (h *AdminInvoiceHandler) ListInvoices(c *gin.Context) {
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

	// Bind query parameters
	var req service.GetInvoicesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "Invalid query parameters", err.Error())
		return
	}

	// Set defaults
	if req.Limit <= 0 || req.Limit > 100 {
		req.Limit = 10
	}
	if req.Offset < 0 {
		req.Offset = 0
	}

	// Get invoices
	invoices, totalCount, err := h.invoiceService.GetInvoices(c.Request.Context(), &req)
	if err != nil {
		logger.Error("Failed to get invoices", logger.Error2("error", err))
		response.InternalServerError(c, "Failed to get invoices", err.Error())
		return
	}

	// Convert to response format
	var invoiceResponses []*model.InvoiceResponse
	for _, invoice := range invoices {
		invoiceResponses = append(invoiceResponses, invoice.ToResponse())
	}

	response.OKPaginated(c, "Invoices retrieved successfully", invoiceResponses, totalCount, req.Limit, req.Offset)
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
func (h *AdminInvoiceHandler) GetInvoice(c *gin.Context) {
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

	// Parse invoice ID
	invoiceIDStr := c.Param("id")
	invoiceID, err := strconv.ParseUint(invoiceIDStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid invoice ID", "Invoice ID must be a valid number")
		return
	}

	// Get invoice with relations
	invoice, err := h.invoiceService.GetInvoiceWithRelations(c.Request.Context(), uint(invoiceID))
	if err != nil {
		if err.Error() == "invoice not found" {
			response.NotFound(c, "Invoice not found")
			return
		}
		logger.Error("Failed to get invoice", logger.Error2("error", err), logger.Uint("invoice_id", uint(invoiceID)))
		response.InternalServerError(c, "Failed to get invoice", err.Error())
		return
	}

	response.OK(c, "Invoice retrieved successfully", invoice.ToResponse())
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
func (h *AdminInvoiceHandler) UpdateInvoice(c *gin.Context) {
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

	// Parse invoice ID
	invoiceIDStr := c.Param("id")
	invoiceID, err := strconv.ParseUint(invoiceIDStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid invoice ID", "Invoice ID must be a valid number")
		return
	}

	// Bind request
	var req service.UpdateInvoiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request data", err.Error())
		return
	}

	// Update invoice
	invoice, err := h.invoiceService.UpdateInvoice(c.Request.Context(), uint(invoiceID), &req)
	if err != nil {
		if err.Error() == "invoice not found" {
			response.NotFound(c, "Invoice not found")
			return
		}
		logger.Error("Failed to update invoice", 
			logger.Error2("error", err), 
			logger.Uint("invoice_id", uint(invoiceID)),
			logger.Uint("admin_id", user.ID))
		response.InternalServerError(c, "Failed to update invoice", err.Error())
		return
	}

	response.OK(c, "Invoice updated successfully", invoice.ToResponse())
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
func (h *AdminInvoiceHandler) SendInvoice(c *gin.Context) {
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

	// Parse invoice ID
	invoiceIDStr := c.Param("id")
	invoiceID, err := strconv.ParseUint(invoiceIDStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid invoice ID", "Invoice ID must be a valid number")
		return
	}

	// Send invoice
	if err := h.invoiceService.SendInvoice(c.Request.Context(), uint(invoiceID)); err != nil {
		if err.Error() == "invoice not found" {
			response.NotFound(c, "Invoice not found")
			return
		}
		logger.Error("Failed to send invoice", 
			logger.Error2("error", err), 
			logger.Uint("invoice_id", uint(invoiceID)),
			logger.Uint("admin_id", user.ID))
		response.InternalServerError(c, "Failed to send invoice", err.Error())
		return
	}

	response.OK(c, "Invoice sent successfully", nil)
}

// MarkInvoiceAsPaidRequest represents the request to mark invoice as paid
type MarkInvoiceAsPaidRequest struct {
	PaymentMethod    string `json:"payment_method" binding:"required" example:"bank_transfer"`
	PaymentReference string `json:"payment_reference,omitempty" example:"REF123456"`
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
func (h *AdminInvoiceHandler) MarkInvoiceAsPaid(c *gin.Context) {
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

	// Parse invoice ID
	invoiceIDStr := c.Param("id")
	invoiceID, err := strconv.ParseUint(invoiceIDStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid invoice ID", "Invoice ID must be a valid number")
		return
	}

	// Bind request
	var req MarkInvoiceAsPaidRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request data", err.Error())
		return
	}

	// Mark invoice as paid
	if err := h.invoiceService.MarkInvoiceAsPaid(c.Request.Context(), uint(invoiceID), req.PaymentMethod, req.PaymentReference); err != nil {
		if err.Error() == "invoice not found" {
			response.NotFound(c, "Invoice not found")
			return
		}
		logger.Error("Failed to mark invoice as paid", 
			logger.Error2("error", err), 
			logger.Uint("invoice_id", uint(invoiceID)),
			logger.Uint("admin_id", user.ID))
		response.InternalServerError(c, "Failed to mark invoice as paid", err.Error())
		return
	}

	response.OK(c, "Invoice marked as paid successfully", nil)
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
func (h *AdminInvoiceHandler) VoidInvoice(c *gin.Context) {
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

	// Parse invoice ID
	invoiceIDStr := c.Param("id")
	invoiceID, err := strconv.ParseUint(invoiceIDStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid invoice ID", "Invoice ID must be a valid number")
		return
	}

	// Bind request
	var req VoidInvoiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request data", err.Error())
		return
	}

	// Void invoice
	if err := h.invoiceService.VoidInvoice(c.Request.Context(), uint(invoiceID), req.Reason); err != nil {
		if err.Error() == "invoice not found" {
			response.NotFound(c, "Invoice not found")
			return
		}
		logger.Error("Failed to void invoice", 
			logger.Error2("error", err), 
			logger.Uint("invoice_id", uint(invoiceID)),
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
func (h *AdminInvoiceHandler) DeleteInvoice(c *gin.Context) {
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

	// Parse invoice ID
	invoiceIDStr := c.Param("id")
	invoiceID, err := strconv.ParseUint(invoiceIDStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid invoice ID", "Invoice ID must be a valid number")
		return
	}

	// Delete invoice
	if err := h.invoiceService.DeleteInvoice(c.Request.Context(), uint(invoiceID)); err != nil {
		if err.Error() == "invoice not found" {
			response.NotFound(c, "Invoice not found")
			return
		}
		logger.Error("Failed to delete invoice", 
			logger.Error2("error", err), 
			logger.Uint("invoice_id", uint(invoiceID)),
			logger.Uint("admin_id", user.ID))
		response.InternalServerError(c, "Failed to delete invoice", err.Error())
		return
	}

	response.OK(c, "Invoice deleted successfully", nil)
}