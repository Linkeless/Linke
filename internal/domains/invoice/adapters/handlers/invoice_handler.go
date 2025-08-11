package handlers

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	"linke/internal/domains/invoice/dto"
	"linke/internal/domains/invoice/entities"
	"linke/internal/domains/invoice/usecases/interfaces"
	userEntities "linke/internal/domains/user/entities"
	"linke/internal/shared/logger"
	"linke/internal/shared/middleware"
	"linke/internal/shared/response"

	"github.com/gin-gonic/gin"
)

// Admin-only request types removed to keep user handler minimal

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

// Admin create endpoint removed; handled by AdminInvoiceHandler

// GetInvoice godoc
// @Summary [User/Admin] Get invoice by ID
// @Description Get a specific invoice by its ID
// @Tags Invoice
// @Produce json
// @Security BearerAuth
// @Param id path uint true "Invoice ID"
// @Success 200 {object} response.StandardResponse{data=dto.InvoiceResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /invoices/{id} [get]
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

	response.SuccessWithMessage(c, "Invoice retrieved successfully", dto.ToResponse(invoice))
}

// Removed GetInvoiceByNumber to keep surface minimal

// Admin update endpoint removed

// Admin delete endpoint removed

// Admin listing endpoint removed; user listing retained below

// GetUserInvoices godoc
// @Summary [User] Get current user's invoices
// @Description Get a list of invoices for the authenticated user
// @Tags Invoice
// @Produce json
// @Security BearerAuth
// @Param limit query int false "Limit" default(10)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} response.PaginatedResponse{data=[]dto.InvoiceResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// (hidden from docs; admin listing is in AdminInvoiceHandler)
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
	invoiceResponses := make([]*dto.InvoiceResponse, len(invoices))
	for i, invoice := range invoices {
		invoiceResponses[i] = dto.ToResponse(invoice)
	}

	response.OKPaginated(c, "User invoices retrieved successfully", invoiceResponses, total, limit, offset)
}

// Admin mark-as-paid endpoint removed

// Admin mark-as-void endpoint removed

// Admin statistics endpoint removed

// Legacy inline PDF endpoint removed; use DownloadInvoicePDF

// Admin send endpoint removed

// DownloadInvoicePDF godoc
// @Summary [User/Admin] Download invoice as PDF with options
// @Description Download an invoice as PDF with custom template and language options
// @Tags Invoice
// @Accept json
// @Produce application/pdf
// @Security BearerAuth
// @Param id path uint true "Invoice ID"
// @Param template query string false "PDF Template" Enums(default,professional,minimal)
// @Param language query string false "Language" Enums(en,zh,es)
// @Param watermark query string false "Watermark text"
// @Success 200 {file} application/pdf
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /invoices/{id}/download [get]
func (h *InvoiceHandler) DownloadInvoicePDF(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid invoice ID")
		return
	}

	// Check if user has permission to access this invoice
	invoice, err := h.invoiceService.GetInvoice(c.Request.Context(), uint(id))
	if err != nil {
		h.logger.Error("Failed to get invoice for download", logger.ErrorField(err))
		response.NotFound(c, "Invoice not found")
		return
	}

	if !h.canAccessInvoice(c, invoice) {
		response.Forbidden(c, "Access denied")
		return
	}

	// Parse query parameters for PDF options
	template := c.Query("template")
	language := c.Query("language")
	watermark := c.Query("watermark")

	// Validate template if provided
	if template != "" {
		if valid, err := h.invoiceService.ValidateTemplate(c.Request.Context(), template); err != nil || !valid {
			response.BadRequest(c, "Invalid template")
			return
		}
	}

	// Create PDF generation options
	options := &dto.PDFGenerationRequest{
		Template:  template,
		Language:  language,
		Watermark: watermark,
	}

	pdfData, _, err := h.invoiceService.GenerateInvoicePDFWithOptions(c.Request.Context(), uint(id), options)
	if err != nil {
		h.logger.Error("Failed to generate invoice PDF for download", logger.ErrorField(err))
		response.InternalServerError(c, "Failed to generate PDF")
		return
	}

	// Log download activity
	h.logDownloadActivity(c, invoice, template, language)

	// Set response headers
	filename := fmt.Sprintf("invoice_%s.pdf", invoice.InvoiceNumber)
	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Header("Content-Length", strconv.Itoa(len(pdfData)))

	c.Data(200, "application/pdf", pdfData)
}

// Bulk download endpoint removed

// Templates listing endpoint removed

// Languages listing endpoint removed

// Admin send with custom PDF endpoint removed

// User download history endpoint removed

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

// logDownloadActivity logs invoice download activity for audit purposes
func (h *InvoiceHandler) logDownloadActivity(c *gin.Context, invoice *entities.Invoice, template, language string) {
	// Get user info
	userValue, exists := c.Get(middleware.AuthContextKey)
	if !exists {
		return
	}

	user, ok := userValue.(*userEntities.User)
	if !ok {
		return
	}

	// Get client IP
	clientIP := h.getClientIP(c)
	userAgent := c.GetHeader("User-Agent")

	h.logger.Info("Invoice PDF downloaded",
		logger.Uint("user_id", user.ID),
		logger.Uint("invoice_id", invoice.ID),
		logger.String("invoice_number", invoice.InvoiceNumber),
		logger.String("template", template),
		logger.String("language", language),
		logger.String("client_ip", clientIP),
		logger.String("user_agent", userAgent))

	// TODO: Store download record in database for history tracking
}

// getClientIP gets the real client IP address
func (h *InvoiceHandler) getClientIP(c *gin.Context) string {
	// Check X-Forwarded-For header first
	xff := c.GetHeader("X-Forwarded-For")
	if xff != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		if ips := strings.Split(xff, ","); len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// Check X-Real-IP header
	if xri := c.GetHeader("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	ip, _, err := net.SplitHostPort(c.Request.RemoteAddr)
	if err != nil {
		return c.Request.RemoteAddr
	}
	return ip
}

// RegisterRoutes registers all invoice routes
func (h *InvoiceHandler) RegisterRoutes(router *gin.RouterGroup) {
	// Destructive refactor: use plural resource and user read-only endpoints
	invoices := router.Group("/invoices")
	{
		// User-facing read-only
		invoices.GET("", h.GetUserInvoices)
		invoices.GET("/:id", h.GetInvoice)
		invoices.GET("/:id/download", h.DownloadInvoicePDF)
	}
}
