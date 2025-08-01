package api

import (
	"time"

	"github.com/gin-gonic/gin"

	"linke/internal/invoice/handler/dto"
	"linke/internal/invoice/service/query"
	"linke/internal/response"
)

// InvoiceHandler handles user operations for invoices
type InvoiceHandler struct {
	invoiceQueryHandler *query.InvoiceQueryHandler
}

// NewInvoiceHandler creates a new invoice handler
func NewInvoiceHandler(invoiceQueryHandler *query.InvoiceQueryHandler) *InvoiceHandler {
	return &InvoiceHandler{
		invoiceQueryHandler: invoiceQueryHandler,
	}
}

// GetInvoice handles retrieving a specific invoice for the authenticated user
// @Summary Get user's invoice by ID
// @Description Get detailed information about a specific invoice belonging to the authenticated user
// @Tags user,invoices
// @Accept json
// @Produce json
// @Param id path string true "Invoice ID"
// @Success 200 {object} response.StandardResponse{data=dto.InvoiceResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/user/invoices/{id} [get]
// @Security BearerAuth
func (h *InvoiceHandler) GetInvoice(c *gin.Context) {
	invoiceID := c.Param("id")
	if invoiceID == "" {
		response.BadRequest(c, "Invoice ID is required")
		return
	}

	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	userIDUint, ok := userID.(uint)
	if !ok {
		response.InternalServerError(c, "Invalid user ID format")
		return
	}

	query_ := query.GetUserInvoiceQuery{
		InvoiceID: invoiceID,
		UserID:    userIDUint,
	}

	invoice, err := h.invoiceQueryHandler.GetUserInvoice(c.Request.Context(), query_)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	invoiceResponse := dto.FromInvoice(invoice)
	response.OK(c, "Invoice retrieved successfully", invoiceResponse)
}

// ListUserInvoices handles listing invoices for the authenticated user
// @Summary List user's invoices
// @Description Get a paginated list of invoices belonging to the authenticated user
// @Tags user,invoices
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param size query int false "Page size" default(20)
// @Param status query []string false "Filter by status" Enums(draft,sent,paid,overdue,voided)
// @Param type query string false "Filter by type" Enums(standard,refund,proforma)
// @Param date_from query string false "Filter from date (RFC3339)"
// @Param date_to query string false "Filter to date (RFC3339)"
// @Param search query string false "Search term"
// @Param sort_by query string false "Sort field" Enums(created_at,updated_at,due_date,total_amount) default(created_at)
// @Param sort_order query string false "Sort order" Enums(asc,desc) default(desc)
// @Success 200 {object} response.StandardListResponse{data=dto.InvoiceListResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/user/invoices [get]
// @Security BearerAuth
func (h *InvoiceHandler) ListUserInvoices(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	userIDUint, ok := userID.(uint)
	if !ok {
		response.InternalServerError(c, "Invalid user ID format")
		return
	}

	var req dto.InvoiceListQuery
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "Invalid query parameters", err.Error())
		return
	}

	// Force the user ID to be the authenticated user's ID
	req.UserID = &userIDUint

	query_ := query.ListInvoicesQuery{
		Page:      req.Page,
		Size:      req.Size,
		UserID:    req.UserID,
		Status:    req.Status,
		Type:      req.Type,
		DateFrom:  req.DateFrom,
		DateTo:    req.DateTo,
		Search:    req.Search,
		SortBy:    req.SortBy,
		SortOrder: req.SortOrder,
	}

	result, err := h.invoiceQueryHandler.ListInvoices(c.Request.Context(), query_)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	invoiceList := dto.InvoiceListResponse{
		Items:      dto.FromInvoices(result.Items),
		Total:      result.Total,
		Page:       result.Page,
		Size:       result.Size,
		TotalPages: result.TotalPages,
		HasNext:    result.HasNext,
		HasPrev:    result.HasPrev,
	}

	response.OK(c, "Invoices retrieved successfully", invoiceList)
}

// GetUserInvoiceStats handles getting invoice statistics for the authenticated user
// @Summary Get user's invoice statistics
// @Description Get comprehensive statistics about the user's invoices
// @Tags user,invoices
// @Accept json
// @Produce json
// @Param date_from query string false "Filter from date (RFC3339)"
// @Param date_to query string false "Filter to date (RFC3339)"
// @Success 200 {object} response.StandardResponse{data=dto.InvoiceStatsResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/user/invoices/stats [get]
// @Security BearerAuth
func (h *InvoiceHandler) GetUserInvoiceStats(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	userIDUint, ok := userID.(uint)
	if !ok {
		response.InternalServerError(c, "Invalid user ID format")
		return
	}

	query_ := query.GetInvoiceStatsQuery{
		UserID:   &userIDUint,
		DateFrom: parseTimeFromQuery(c, "date_from"),
		DateTo:   parseTimeFromQuery(c, "date_to"),
	}

	stats, err := h.invoiceQueryHandler.GetInvoiceStats(c.Request.Context(), query_)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	statsResponse := dto.InvoiceStatsResponse{
		TotalInvoices:     stats.TotalInvoices,
		DraftCount:        stats.DraftCount,
		SentCount:         stats.SentCount,
		PaidCount:         stats.PaidCount,
		OverdueCount:      stats.OverdueCount,
		VoidedCount:       stats.VoidedCount,
		TotalAmount:       dto.FromMoney(stats.TotalAmount),
		PaidAmount:        dto.FromMoney(stats.PaidAmount),
		OutstandingAmount: dto.FromMoney(stats.OutstandingAmount),
		OverdueAmount:     dto.FromMoney(stats.OverdueAmount),
	}

	response.OK(c, "Invoice statistics retrieved successfully", statsResponse)
}

// DownloadInvoicePDF handles downloading an invoice PDF for the authenticated user
// @Summary Download user's invoice PDF
// @Description Download the PDF version of an invoice belonging to the authenticated user
// @Tags user,invoices
// @Accept json  
// @Produce application/pdf
// @Param id path string true "Invoice ID"
// @Success 200 {file} binary "PDF file"
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/user/invoices/{id}/pdf [get]
// @Security BearerAuth
func (h *InvoiceHandler) DownloadInvoicePDF(c *gin.Context) {
	invoiceID := c.Param("id")
	if invoiceID == "" {
		response.BadRequest(c, "Invoice ID is required")
		return
	}

	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	userIDUint, ok := userID.(uint)
	if !ok {
		response.InternalServerError(c, "Invalid user ID format")
		return
	}

	query_ := query.GetUserInvoiceQuery{
		InvoiceID: invoiceID,
		UserID:    userIDUint,
	}

	invoice, err := h.invoiceQueryHandler.GetUserInvoice(c.Request.Context(), query_)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	// Check if PDF exists
	if invoice.PDFPath() == "" {
		response.NotFound(c, "Invoice PDF not found")
		return
	}

	// Set headers for PDF download
	c.Header("Content-Description", "Invoice PDF")
	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", `attachment; filename="invoice_`+invoice.InvoiceNumber().String()+`.pdf"`)
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Cache-Control", "must-revalidate")
	c.Header("Pragma", "public")

	// Serve the PDF file
	c.File(invoice.PDFPath())
}

// GetPaymentHistory handles retrieving payment history for a specific invoice
// @Summary Get invoice payment history
// @Description Get the payment history for a specific invoice belonging to the authenticated user
// @Tags user,invoices
// @Accept json
// @Produce json
// @Param id path string true "Invoice ID"
// @Success 200 {object} response.StandardResponse{data=[]dto.PaymentHistoryResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/user/invoices/{id}/payments [get]
// @Security BearerAuth
func (h *InvoiceHandler) GetPaymentHistory(c *gin.Context) {
	invoiceID := c.Param("id")
	if invoiceID == "" {
		response.BadRequest(c, "Invoice ID is required")
		return
	}

	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	userIDUint, ok := userID.(uint)
	if !ok {
		response.InternalServerError(c, "Invalid user ID format")
		return
	}

	query_ := query.GetInvoicePaymentHistoryQuery{
		InvoiceID: invoiceID,
		UserID:    userIDUint,
	}

	history, err := h.invoiceQueryHandler.GetInvoicePaymentHistory(c.Request.Context(), query_)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.OK(c, "Payment history retrieved successfully", history)
}

// Helper functions

func parseTimeFromQuery(c *gin.Context, key string) *time.Time {
	if timeStr := c.Query(key); timeStr != "" {
		if t, err := time.Parse(time.RFC3339, timeStr); err == nil {
			return &t
		}
	}
	return nil
}