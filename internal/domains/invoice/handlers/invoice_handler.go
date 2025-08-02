package handlers

import (
	"strconv"

	"linke/internal/domains/invoice/entities"
	"linke/internal/domains/invoice/usecases/interfaces"
	userEntities "linke/internal/domains/user/entities"
	"linke/internal/shared/logger"
	"linke/internal/shared/middleware"
	"linke/internal/shared/response"

	"github.com/gin-gonic/gin"
)

// InvoiceHandler handles HTTP requests for invoice operations
type InvoiceHandler struct {
	invoiceService interfaces.InvoiceService
	logger         logger.Logger
}

// NewInvoiceHandler creates a new InvoiceHandler
func NewInvoiceHandler(invoiceService interfaces.InvoiceService, logger logger.Logger) *InvoiceHandler {
	return &InvoiceHandler{
		invoiceService: invoiceService,
		logger:         logger,
	}
}

// CreateInvoice godoc
// @Summary [Admin] Create a new invoice
// @Description Create a new invoice for a subscription order
// @Tags Invoice
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param invoice body interfaces.CreateInvoiceRequest true "Invoice data"
// @Success 201 {object} response.StandardResponse{data=entities.InvoiceResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /invoice [post]
func (h *InvoiceHandler) CreateInvoice(c *gin.Context) {
	var req interfaces.CreateInvoiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to bind create invoice request", logger.ErrorField(err))
		response.BadRequest(c, "Invalid request data")
		return
	}

	invoice, err := h.invoiceService.CreateInvoice(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to create invoice", logger.ErrorField(err))
		response.InternalServerError(c, "Failed to create invoice")
		return
	}

	response.CreatedWithMessage(c, "Invoice created successfully", invoice.ToResponse())
}

// GetInvoice godoc
// @Summary [User/Admin] Get invoice by ID
// @Description Get a specific invoice by its ID
// @Tags Invoice
// @Produce json
// @Security BearerAuth
// @Param id path uint true "Invoice ID"
// @Success 200 {object} response.StandardResponse{data=entities.InvoiceResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /invoice/{id} [get]
func (h *InvoiceHandler) GetInvoice(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid invoice ID")
		return
	}

	invoice, err := h.invoiceService.GetInvoice(c.Request.Context(), uint(id))
	if err != nil {
		h.logger.Error("Failed to get invoice", logger.ErrorField(err))
		response.NotFound(c, "Invoice not found")
		return
	}

	// Check if user has permission to view this invoice
	if !h.canAccessInvoice(c, invoice) {
		response.Forbidden(c, "Access denied")
		return
	}

	response.SuccessWithMessage(c, "Invoice retrieved successfully", invoice.ToResponse())
}

// GetInvoiceByNumber godoc
// @Summary [User/Admin] Get invoice by number
// @Description Get a specific invoice by its invoice number
// @Tags Invoice
// @Produce json
// @Security BearerAuth
// @Param number path string true "Invoice Number"
// @Success 200 {object} response.StandardResponse{data=entities.InvoiceResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /invoice/number/{number} [get]
func (h *InvoiceHandler) GetInvoiceByNumber(c *gin.Context) {
	invoiceNumber := c.Param("number")
	if invoiceNumber == "" {
		response.BadRequest(c, "Invoice number is required")
		return
	}

	invoice, err := h.invoiceService.GetInvoiceByNumber(c.Request.Context(), invoiceNumber)
	if err != nil {
		h.logger.Error("Failed to get invoice by number", logger.ErrorField(err))
		response.NotFound(c, "Invoice not found")
		return
	}

	// Check if user has permission to view this invoice
	if !h.canAccessInvoice(c, invoice) {
		response.Forbidden(c, "Access denied")
		return
	}

	response.SuccessWithMessage(c, "Invoice retrieved successfully", invoice.ToResponse())
}

// UpdateInvoice godoc
// @Summary [Admin] Update an invoice
// @Description Update an existing invoice
// @Tags Invoice
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path uint true "Invoice ID"
// @Param invoice body interfaces.UpdateInvoiceRequest true "Invoice update data"
// @Success 200 {object} response.StandardResponse{data=entities.InvoiceResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /invoice/{id} [put]
func (h *InvoiceHandler) UpdateInvoice(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid invoice ID")
		return
	}

	var req interfaces.UpdateInvoiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to bind update invoice request", logger.ErrorField(err))
		response.BadRequest(c, "Invalid request data")
		return
	}

	invoice, err := h.invoiceService.UpdateInvoice(c.Request.Context(), uint(id), &req)
	if err != nil {
		h.logger.Error("Failed to update invoice", logger.ErrorField(err))
		response.InternalServerError(c, "Internal server error")
		return
	}

	response.SuccessWithMessage(c, "Invoice updated successfully", invoice.ToResponse())
}

// DeleteInvoice godoc
// @Summary [Admin] Delete an invoice
// @Description Delete an invoice by ID
// @Tags Invoice
// @Produce json
// @Security BearerAuth
// @Param id path uint true "Invoice ID"
// @Success 200 {object} response.StandardResponse
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /invoice/{id} [delete]
func (h *InvoiceHandler) DeleteInvoice(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid invoice ID")
		return
	}

	err = h.invoiceService.DeleteInvoice(c.Request.Context(), uint(id))
	if err != nil {
		h.logger.Error("Failed to delete invoice", logger.ErrorField(err))
		response.InternalServerError(c, "Internal server error")
		return
	}

	response.SuccessWithMessage(c, "Invoice deleted successfully", nil)
}

// GetInvoices godoc
// @Summary [Admin] Get invoices with filters
// @Description Get a list of invoices with optional filters
// @Tags Invoice
// @Produce json
// @Security BearerAuth
// @Param user_id query uint false "User ID"
// @Param status query string false "Invoice status"
// @Param invoice_type query string false "Invoice type"
// @Param date_from query string false "Date from (YYYY-MM-DD)"
// @Param date_to query string false "Date to (YYYY-MM-DD)"
// @Param limit query int false "Limit" default(10)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} response.PaginatedResponse{data=[]entities.InvoiceResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /invoice [get]
func (h *InvoiceHandler) GetInvoices(c *gin.Context) {
	var req interfaces.GetInvoicesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		h.logger.Error("Failed to bind get invoices request", logger.ErrorField(err))
		response.BadRequest(c, "Invalid query parameters")
		return
	}

	// Set default pagination
	if req.Limit <= 0 {
		req.Limit = 10
	}
	if req.Limit > 100 {
		req.Limit = 100
	}

	invoices, total, err := h.invoiceService.GetInvoices(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to get invoices", logger.ErrorField(err))
		response.InternalServerError(c, "Internal server error")
		return
	}

	// Convert to response format
	invoiceResponses := make([]*entities.InvoiceResponse, len(invoices))
	for i, invoice := range invoices {
		invoiceResponses[i] = invoice.ToResponse()
	}

	response.OKPaginated(c, "Invoices retrieved successfully", invoiceResponses, total, req.Limit, req.Offset)
}

// GetUserInvoices godoc
// @Summary [User] Get current user's invoices
// @Description Get a list of invoices for the authenticated user
// @Tags Invoice
// @Produce json
// @Security BearerAuth
// @Param limit query int false "Limit" default(10)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} response.PaginatedResponse{data=[]entities.InvoiceResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /invoice/user [get]
func (h *InvoiceHandler) GetUserInvoices(c *gin.Context) {
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

	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	invoices, total, err := h.invoiceService.GetUserInvoices(c.Request.Context(), user.ID, limit, offset)
	if err != nil {
		h.logger.Error("Failed to get user invoices", logger.ErrorField(err))
		response.InternalServerError(c, "Internal server error")
		return
	}

	// Convert to response format
	invoiceResponses := make([]*entities.InvoiceResponse, len(invoices))
	for i, invoice := range invoices {
		invoiceResponses[i] = invoice.ToResponse()
	}

	response.OKPaginated(c, "User invoices retrieved successfully", invoiceResponses, total, limit, offset)
}

// MarkInvoiceAsPaid godoc
// @Summary [Admin] Mark invoice as paid
// @Description Mark an invoice as paid with payment date
// @Tags Invoice
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path uint true "Invoice ID"
// @Param payment_date body object{payment_date=string} true "Payment date"
// @Success 200 {object} response.StandardResponse
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /invoice/{id}/mark-paid [put]
func (h *InvoiceHandler) MarkInvoiceAsPaid(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid invoice ID")
		return
	}

	var req struct {
		PaymentDate string `json:"payment_date" binding:"required" example:"2024-01-01"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to bind payment date request", logger.ErrorField(err))
		response.BadRequest(c, "Invalid request data")
		return
	}

	err = h.invoiceService.MarkInvoiceAsPaid(c.Request.Context(), uint(id), req.PaymentDate)
	if err != nil {
		h.logger.Error("Failed to mark invoice as paid", logger.ErrorField(err))
		response.InternalServerError(c, "Internal server error")
		return
	}

	response.SuccessWithMessage(c, "Invoice marked as paid successfully", nil)
}

// MarkInvoiceAsVoid godoc
// @Summary [Admin] Mark invoice as void
// @Description Mark an invoice as void with reason
// @Tags Invoice
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path uint true "Invoice ID"
// @Param reason body object{reason=string} true "Void reason"
// @Success 200 {object} response.StandardResponse
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /invoice/{id}/mark-void [put]
func (h *InvoiceHandler) MarkInvoiceAsVoid(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid invoice ID")
		return
	}

	var req struct {
		Reason string `json:"reason" binding:"required" example:"Customer request"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to bind void reason request", logger.ErrorField(err))
		response.BadRequest(c, "Invalid request data")
		return
	}

	err = h.invoiceService.MarkInvoiceAsVoid(c.Request.Context(), uint(id), req.Reason)
	if err != nil {
		h.logger.Error("Failed to mark invoice as void", logger.ErrorField(err))
		response.InternalServerError(c, "Internal server error")
		return
	}

	response.SuccessWithMessage(c, "Invoice marked as void successfully", nil)
}

// GetInvoiceStatistics godoc
// @Summary [Admin] Get invoice statistics
// @Description Get invoice statistics for a date range
// @Tags Invoice
// @Produce json
// @Security BearerAuth
// @Param from_date query string false "From date (YYYY-MM-DD)"
// @Param to_date query string false "To date (YYYY-MM-DD)"
// @Success 200 {object} response.StandardResponse{data=map[string]interface{}}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /invoice/statistics [get]
func (h *InvoiceHandler) GetInvoiceStatistics(c *gin.Context) {
	fromDate := c.Query("from_date")
	toDate := c.Query("to_date")

	statistics, err := h.invoiceService.GetInvoiceStatistics(c.Request.Context(), fromDate, toDate)
	if err != nil {
		h.logger.Error("Failed to get invoice statistics", logger.ErrorField(err))
		response.InternalServerError(c, "Internal server error")
		return
	}

	response.SuccessWithMessage(c, "Invoice statistics retrieved successfully", statistics)
}

// GenerateInvoicePDF godoc
// @Summary [User/Admin] Generate invoice PDF
// @Description Generate a PDF for the specified invoice
// @Tags Invoice
// @Produce application/pdf
// @Security BearerAuth
// @Param id path uint true "Invoice ID"
// @Success 200 {file} application/pdf
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /invoice/{id}/pdf [get]
func (h *InvoiceHandler) GenerateInvoicePDF(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid invoice ID")
		return
	}

	// Check if user has permission to access this invoice
	invoice, err := h.invoiceService.GetInvoice(c.Request.Context(), uint(id))
	if err != nil {
		h.logger.Error("Failed to get invoice for PDF generation", logger.ErrorField(err))
		response.NotFound(c, "Invoice not found")
		return
	}

	if !h.canAccessInvoice(c, invoice) {
		response.Forbidden(c, "Access denied")
		return
	}

	pdfData, err := h.invoiceService.GenerateInvoicePDF(c.Request.Context(), uint(id))
	if err != nil {
		h.logger.Error("Failed to generate invoice PDF", logger.ErrorField(err))
		response.InternalServerError(c, "Internal server error")
		return
	}

	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", "inline; filename=invoice-"+invoice.InvoiceNumber+".pdf")
	c.Data(200, "application/pdf", pdfData)
}

// SendInvoice godoc
// @Summary [Admin] Send invoice via email
// @Description Send an invoice to the specified email address
// @Tags Invoice
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path uint true "Invoice ID"
// @Param email_request body interfaces.SendInvoiceRequest true "Email details"
// @Success 200 {object} response.StandardResponse
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /invoice/{id}/send [post]
func (h *InvoiceHandler) SendInvoice(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid invoice ID")
		return
	}

	var req interfaces.SendInvoiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to bind send invoice request", logger.ErrorField(err))
		response.BadRequest(c, "Invalid request data")
		return
	}

	err = h.invoiceService.SendInvoice(c.Request.Context(), uint(id), &req)
	if err != nil {
		h.logger.Error("Failed to send invoice", logger.ErrorField(err))
		response.InternalServerError(c, "Internal server error")
		return
	}

	response.SuccessWithMessage(c, "Invoice sent successfully", nil)
}

// Helper methods

// canAccessInvoice checks if the current user can access the specified invoice
func (h *InvoiceHandler) canAccessInvoice(c *gin.Context, invoice *entities.Invoice) bool {
	userValue, exists := c.Get(middleware.AuthContextKey)
	if !exists {
		return false
	}

	user, ok := userValue.(*userEntities.User)
	if !ok {
		return false
	}

	// Admin can access all invoices
	if user.Role == "admin" {
		return true
	}

	// Users can only access their own invoices
	return user.ID == invoice.UserID
}

// RegisterRoutes registers all invoice routes
func (h *InvoiceHandler) RegisterRoutes(router *gin.RouterGroup) {
	invoiceGroup := router.Group("/invoice")
	{
		// Basic routes without middleware for now
		// TODO: Add proper authentication and authorization middleware
		invoiceGroup.POST("", h.CreateInvoice)
		invoiceGroup.GET("", h.GetInvoices)
		invoiceGroup.PUT("/:id", h.UpdateInvoice)
		invoiceGroup.DELETE("/:id", h.DeleteInvoice)
		invoiceGroup.PUT("/:id/mark-paid", h.MarkInvoiceAsPaid)
		invoiceGroup.PUT("/:id/mark-void", h.MarkInvoiceAsVoid)
		invoiceGroup.POST("/:id/send", h.SendInvoice)
		invoiceGroup.GET("/statistics", h.GetInvoiceStatistics)
		invoiceGroup.GET("/:id", h.GetInvoice)
		invoiceGroup.GET("/number/:number", h.GetInvoiceByNumber)
		invoiceGroup.GET("/:id/pdf", h.GenerateInvoicePDF)
		invoiceGroup.GET("/user", h.GetUserInvoices)
	}
}