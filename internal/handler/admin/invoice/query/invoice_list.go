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

// InvoiceListHandler handles invoice listing operations
type InvoiceListHandler struct {
	*shared.BaseHandler
}

// NewInvoiceListHandler creates a new invoice list handler
func NewInvoiceListHandler(invoiceService *service.InvoiceService) *InvoiceListHandler {
	return &InvoiceListHandler{
		BaseHandler: shared.NewBaseHandler(invoiceService),
	}
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
func (h *InvoiceListHandler) ListInvoices(c *gin.Context) {
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
	var req service.InvoiceFilters
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

	// Validate parameters
	if req.Status != "" {
		if err := h.Validator.ValidateInvoiceStatus(req.Status); err != nil {
			response.BadRequest(c, "Invalid status parameter", err.Error())
			return
		}
	}

	if req.InvoiceType != "" {
		if err := h.Validator.ValidateInvoiceType(req.InvoiceType); err != nil {
			response.BadRequest(c, "Invalid invoice type parameter", err.Error())
			return
		}
	}

	if req.Currency != "" {
		if err := h.Validator.ValidateCurrency(req.Currency); err != nil {
			response.BadRequest(c, "Invalid currency parameter", err.Error())
			return
		}
	}

	// Get invoices
	invoices, totalCount, err := h.InvoiceService.ListInvoices(c.Request.Context(), &req)
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