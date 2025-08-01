package admin

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"linke/internal/invoice/domain/valueobject"
	"linke/internal/invoice/handler/dto"
	"linke/internal/invoice/service/command"
	"linke/internal/invoice/service/query"
	"linke/internal/response"
)

// InvoiceAdminHandler handles admin operations for invoices
type InvoiceAdminHandler struct {
	invoiceCommandHandler *command.InvoiceCommandHandler
	invoiceQueryHandler   *query.InvoiceQueryHandler
}

// NewInvoiceAdminHandler creates a new invoice admin handler
func NewInvoiceAdminHandler(
	invoiceCommandHandler *command.InvoiceCommandHandler,
	invoiceQueryHandler *query.InvoiceQueryHandler,
) *InvoiceAdminHandler {
	return &InvoiceAdminHandler{
		invoiceCommandHandler: invoiceCommandHandler,
		invoiceQueryHandler:   invoiceQueryHandler,
	}
}

// CreateInvoice handles creating a new invoice
// @Summary Create a new invoice
// @Description Create a new invoice for an order
// @Tags admin,invoices
// @Accept json
// @Produce json
// @Param request body dto.CreateInvoiceRequest true "Create invoice request"
// @Success 201 {object} response.StandardResponse{data=dto.InvoiceResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/admin/invoices [post]
// @Security BearerAuth
func (h *InvoiceAdminHandler) CreateInvoice(c *gin.Context) {
	var req dto.CreateInvoiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request format", err.Error())
		return
	}

	// Convert DTO to command
	cmd, err := h.createInvoiceCommand(req)
	if err != nil {
		response.BadRequest(c, "Invalid request data", err.Error())
		return
	}

	// Execute command
	invoice, err := h.invoiceCommandHandler.CreateInvoice(c.Request.Context(), cmd)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	// Convert to response
	invoiceResponse := dto.FromInvoice(invoice)
	response.CreatedWithMessage(c, "Invoice created successfully", invoiceResponse)
}

// GetInvoice handles retrieving a specific invoice
// @Summary Get invoice by ID
// @Description Get detailed information about a specific invoice
// @Tags admin,invoices
// @Accept json
// @Produce json
// @Param id path string true "Invoice ID"
// @Success 200 {object} response.StandardResponse{data=dto.InvoiceResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/admin/invoices/{id} [get]
// @Security BearerAuth
func (h *InvoiceAdminHandler) GetInvoice(c *gin.Context) {
	invoiceID := c.Param("id")
	if invoiceID == "" {
		response.BadRequest(c, "Invoice ID is required")
		return
	}

	query_ := query.GetInvoiceQuery{InvoiceID: invoiceID}
	invoice, err := h.invoiceQueryHandler.GetInvoice(c.Request.Context(), query_)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	invoiceResponse := dto.FromInvoice(invoice)
	response.OK(c, "Invoice retrieved successfully", invoiceResponse)
}

// ListInvoices handles listing invoices with filtering and pagination
// @Summary List invoices
// @Description Get a paginated list of invoices with optional filtering
// @Tags admin,invoices
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param size query int false "Page size" default(20)
// @Param user_id query int false "Filter by user ID"
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
// @Failure 403 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/admin/invoices [get]
// @Security BearerAuth
func (h *InvoiceAdminHandler) ListInvoices(c *gin.Context) {
	var req dto.InvoiceListQuery
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "Invalid query parameters", err.Error())
		return
	}

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

// UpdateInvoice handles updating an existing invoice
// @Summary Update an invoice
// @Description Update an existing invoice (only in draft status)
// @Tags admin,invoices
// @Accept json
// @Produce json
// @Param id path string true "Invoice ID"
// @Param request body dto.UpdateInvoiceRequest true "Update invoice request"
// @Success 200 {object} response.StandardResponse{data=dto.InvoiceResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/admin/invoices/{id} [put]
// @Security BearerAuth
func (h *InvoiceAdminHandler) UpdateInvoice(c *gin.Context) {
	invoiceID := c.Param("id")
	if invoiceID == "" {
		response.BadRequest(c, "Invoice ID is required")
		return
	}

	var req dto.UpdateInvoiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request format", err.Error())
		return
	}

	// Convert DTO to command
	cmd, err := h.updateInvoiceCommand(invoiceID, req)
	if err != nil {
		response.BadRequest(c, "Invalid request data", err.Error())
		return
	}

	// Execute command
	invoice, err := h.invoiceCommandHandler.UpdateInvoice(c.Request.Context(), cmd)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	// Convert to response
	invoiceResponse := dto.FromInvoice(invoice)
	response.OK(c, "Invoice updated successfully", invoiceResponse)
}

// SendInvoice handles sending an invoice to the customer
// @Summary Send an invoice
// @Description Send an invoice to the customer via email
// @Tags admin,invoices
// @Accept json
// @Produce json
// @Param id path string true "Invoice ID"
// @Param request body dto.SendInvoiceRequest true "Send invoice request"
// @Success 200 {object} response.StandardResponse{data=dto.InvoiceResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/admin/invoices/{id}/send [post]
// @Security BearerAuth
func (h *InvoiceAdminHandler) SendInvoice(c *gin.Context) {
	invoiceID := c.Param("id")
	if invoiceID == "" {
		response.BadRequest(c, "Invoice ID is required")
		return
	}

	var req dto.SendInvoiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request format", err.Error())
		return
	}

	cmd := command.SendInvoiceCommand{
		InvoiceID: invoiceID,
		Email:     req.Email,
		Subject:   req.Subject,
		Message:   req.Message,
		CCEmails:  req.CCEmails,
	}

	invoice, err := h.invoiceCommandHandler.SendInvoice(c.Request.Context(), cmd)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	invoiceResponse := dto.FromInvoice(invoice)
	response.OK(c, "Invoice sent successfully", invoiceResponse)
}

// MarkAsPaid handles marking an invoice as paid
// @Summary Mark invoice as paid
// @Description Mark an invoice as paid with payment details
// @Tags admin,invoices
// @Accept json
// @Produce json
// @Param id path string true "Invoice ID"
// @Param request body dto.PayInvoiceRequest true "Payment request"
// @Success 200 {object} response.StandardResponse{data=dto.InvoiceResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/admin/invoices/{id}/pay [post]
// @Security BearerAuth
func (h *InvoiceAdminHandler) MarkAsPaid(c *gin.Context) {
	invoiceID := c.Param("id")
	if invoiceID == "" {
		response.BadRequest(c, "Invoice ID is required")
		return
	}

	var req dto.PayInvoiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request format", err.Error())
		return
	}

	// Convert DTO to domain money
	amount, err := req.Amount.ToMoney()
	if err != nil {
		response.BadRequest(c, "Invalid amount", err.Error())
		return
	}

	cmd := command.PayInvoiceCommand{
		InvoiceID:  invoiceID,
		Amount:     amount,
		PaymentRef: req.PaymentRef,
		Notes:      req.Notes,
	}

	invoice, err := h.invoiceCommandHandler.PayInvoice(c.Request.Context(), cmd)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	invoiceResponse := dto.FromInvoice(invoice)
	response.OK(c, "Invoice marked as paid successfully", invoiceResponse)
}

// VoidInvoice handles voiding an invoice
// @Summary Void an invoice
// @Description Void an invoice with a reason
// @Tags admin,invoices
// @Accept json
// @Produce json
// @Param id path string true "Invoice ID"
// @Param request body dto.VoidInvoiceRequest true "Void invoice request"
// @Success 200 {object} response.StandardResponse{data=dto.InvoiceResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/admin/invoices/{id}/void [post]
// @Security BearerAuth
func (h *InvoiceAdminHandler) VoidInvoice(c *gin.Context) {
	invoiceID := c.Param("id")
	if invoiceID == "" {
		response.BadRequest(c, "Invoice ID is required")
		return
	}

	var req dto.VoidInvoiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request format", err.Error())
		return
	}

	cmd := command.VoidInvoiceCommand{
		InvoiceID: invoiceID,
		Reason:    req.Reason,
	}

	invoice, err := h.invoiceCommandHandler.VoidInvoice(c.Request.Context(), cmd)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	invoiceResponse := dto.FromInvoice(invoice)
	response.OK(c, "Invoice voided successfully", invoiceResponse)
}

// MarkAsOverdue handles marking overdue invoices
// @Summary Mark invoice as overdue
// @Description Mark an invoice as overdue (typically called by background job)
// @Tags admin,invoices
// @Accept json
// @Produce json
// @Param id path string true "Invoice ID"
// @Success 200 {object} response.StandardResponse{data=dto.InvoiceResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/admin/invoices/{id}/overdue [post]
// @Security BearerAuth
func (h *InvoiceAdminHandler) MarkAsOverdue(c *gin.Context) {
	invoiceID := c.Param("id")
	if invoiceID == "" {
		response.BadRequest(c, "Invoice ID is required")
		return
	}

	cmd := command.MarkOverdueCommand{
		InvoiceID: invoiceID,
	}

	invoice, err := h.invoiceCommandHandler.MarkAsOverdue(c.Request.Context(), cmd)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	invoiceResponse := dto.FromInvoice(invoice)
	response.OK(c, "Invoice marked as overdue", invoiceResponse)
}

// GetInvoiceStats handles getting invoice statistics
// @Summary Get invoice statistics
// @Description Get comprehensive statistics about invoices
// @Tags admin,invoices
// @Accept json
// @Produce json
// @Param user_id query int false "Filter by user ID"
// @Param date_from query string false "Filter from date (RFC3339)"
// @Param date_to query string false "Filter to date (RFC3339)"
// @Success 200 {object} response.StandardResponse{data=dto.InvoiceStatsResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/admin/invoices/stats [get]
// @Security BearerAuth
func (h *InvoiceAdminHandler) GetInvoiceStats(c *gin.Context) {
	var userID *uint
	if userIDStr := c.Query("user_id"); userIDStr != "" {
		if id, err := strconv.ParseUint(userIDStr, 10, 32); err == nil {
			uid := uint(id)
			userID = &uid
		}
	}

	query_ := query.GetInvoiceStatsQuery{
		UserID:   userID,
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

// DeleteInvoice handles soft deleting an invoice
// @Summary Delete an invoice
// @Description Soft delete an invoice (only in draft status)
// @Tags admin,invoices
// @Accept json
// @Produce json
// @Param id path string true "Invoice ID"
// @Success 204 "Invoice deleted successfully"
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/admin/invoices/{id} [delete]
// @Security BearerAuth
func (h *InvoiceAdminHandler) DeleteInvoice(c *gin.Context) {
	invoiceID := c.Param("id")
	if invoiceID == "" {
		response.BadRequest(c, "Invoice ID is required")
		return
	}

	cmd := command.DeleteInvoiceCommand{
		InvoiceID: invoiceID,
	}

	err := h.invoiceCommandHandler.DeleteInvoice(c.Request.Context(), cmd)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// Helper methods

func (h *InvoiceAdminHandler) createInvoiceCommand(req dto.CreateInvoiceRequest) (command.CreateInvoiceCommand, error) {
	subtotal, err := req.Subtotal.ToMoney()
	if err != nil {
		return command.CreateInvoiceCommand{}, err
	}

	taxInfo, err := req.TaxInfo.ToTaxInfo(subtotal)
	if err != nil {
		return command.CreateInvoiceCommand{}, err
	}

	billingInfo, err := req.BillingInfo.ToBillingAddress()
	if err != nil {
		return command.CreateInvoiceCommand{}, err
	}

	var companyInfo *valueobject.CompanyInfo
	if req.CompanyInfo.Name != "" {
		ci, err := req.CompanyInfo.ToCompanyInfo()
		if err != nil {
			return command.CreateInvoiceCommand{}, err
		}
		companyInfo = &ci
	}

	return command.CreateInvoiceCommand{
		OrderID:         req.OrderID,
		UserID:          req.UserID,
		Type:            req.Type,
		Subtotal:        subtotal,
		TaxInfo:         taxInfo,
		BillingInfo:     billingInfo,
		CompanyInfo:     companyInfo,
		Description:     req.Description,
		Notes:           req.Notes,
		DueDate:         req.DueDate,
		PaymentTerms:    req.PaymentTerms,
		Template:        req.Template,
		Language:        req.Language,
	}, nil
}

func (h *InvoiceAdminHandler) updateInvoiceCommand(invoiceID string, req dto.UpdateInvoiceRequest) (command.UpdateInvoiceCommand, error) {
	cmd := command.UpdateInvoiceCommand{
		InvoiceID:     invoiceID,
		Description:   req.Description,
		Notes:         req.Notes,
		DueDate:       req.DueDate,
		PaymentTerms:  req.PaymentTerms,
		InternalNotes: req.InternalNotes,
	}

	if req.BillingInfo != nil {
		billingInfo, err := req.BillingInfo.ToBillingAddress()
		if err != nil {
			return command.UpdateInvoiceCommand{}, err
		}
		cmd.BillingInfo = &billingInfo
	}

	if req.CompanyInfo != nil {
		companyInfo, err := req.CompanyInfo.ToCompanyInfo()
		if err != nil {
			return command.UpdateInvoiceCommand{}, err
		}
		cmd.CompanyInfo = &companyInfo
	}

	if req.TaxInfo != nil {
		// Note: We need the subtotal to create TaxInfo, which requires fetching the invoice first
		// This is a limitation of the current design - consider refactoring
		// For now, we'll handle this in the command handler
		cmd.TaxInfo = req.TaxInfo
	}

	return cmd, nil
}

func parseTimeFromQuery(c *gin.Context, key string) *time.Time {
	if timeStr := c.Query(key); timeStr != "" {
		if t, err := time.Parse(time.RFC3339, timeStr); err == nil {
			return &t
		}
	}
	return nil
}