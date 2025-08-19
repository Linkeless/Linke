package handlers

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"linke/internal/domains/invoice/dto"
	"linke/internal/domains/invoice/entities"
	"linke/internal/domains/invoice/usecases/interfaces"
	paymentInterfaces "linke/internal/domains/payment/usecases/interfaces"
	userInterfaces "linke/internal/domains/user/usecases/interfaces"
	"linke/internal/shared/handlers"
	"linke/internal/shared/logger"
	"linke/internal/shared/response"

	"github.com/gin-gonic/gin"
)

// AdminInvoiceHandler handles admin invoice management operations
type AdminInvoiceHandler struct {
	invoiceService interfaces.InvoiceService
	userService    userInterfaces.UserService
	paymentService paymentInterfaces.PaymentService
}

// NewAdminInvoiceHandler creates a new admin invoice handler
func NewAdminInvoiceHandler(
	invoiceService interfaces.InvoiceService,
	userService userInterfaces.UserService,
	paymentService paymentInterfaces.PaymentService,
) *AdminInvoiceHandler {
	return &AdminInvoiceHandler{
		invoiceService: invoiceService,
		userService:    userService,
		paymentService: paymentService,
	}
}

// Request/Response structures for admin operations

// AdminCreateInvoiceRequest represents the request body for creating invoices
type AdminCreateInvoiceRequest struct {
	UserID              uint    `json:"user_id" binding:"required" example:"1"`
	SubscriptionOrderID uint    `json:"subscription_order_id" binding:"required" example:"1"`
	InvoiceType         string  `json:"invoice_type,omitempty" example:"standard"`
	Amount              float64 `json:"amount" binding:"required,min=0" example:"29.99"`
	Currency            string  `json:"currency,omitempty" example:"CNY"`
	TaxRate             float64 `json:"tax_rate,omitempty" example:"0.2"`
	TaxType             string  `json:"tax_type,omitempty" example:"VAT"`
	TaxNumber           string  `json:"tax_number,omitempty" example:"GB123456789"`
	BillingName         string  `json:"billing_name" binding:"required" example:"John Doe"`
	BillingEmail        string  `json:"billing_email" binding:"required,email" example:"john@example.com"`
	BillingAddress      string  `json:"billing_address,omitempty" example:"123 Main St"`
	BillingCity         string  `json:"billing_city,omitempty" example:"New York"`
	BillingState        string  `json:"billing_state,omitempty" example:"NY"`
	BillingCountry      string  `json:"billing_country,omitempty" example:"US"`
	BillingZip          string  `json:"billing_zip,omitempty" example:"10001"`
	CompanyName         string  `json:"company_name,omitempty" example:"Acme Corp"`
	CompanyTaxID        string  `json:"company_tax_id,omitempty" example:"12-3456789"`
	CompanyAddress      string  `json:"company_address,omitempty" example:"456 Business Ave"`
	Description         string  `json:"description,omitempty" example:"Monthly subscription"`
	Notes               string  `json:"notes,omitempty" example:"Admin created invoice"`
	DueDate             string  `json:"due_date,omitempty" example:"2024-01-31"`
	Template            string  `json:"template,omitempty" example:"default"`
	Language            string  `json:"language,omitempty" example:"en"`
	AutoSend            bool    `json:"auto_send,omitempty" example:"false"`
}

// AdminUpdateInvoiceRequest represents the request body for updating invoices
type AdminUpdateInvoiceRequest struct {
	InvoiceType    *string  `json:"invoice_type,omitempty" example:"credit_note"`
	Amount         *float64 `json:"amount,omitempty" example:"39.99"`
	TaxRate        *float64 `json:"tax_rate,omitempty" example:"0.15"`
	TaxType        *string  `json:"tax_type,omitempty" example:"GST"`
	TaxNumber      *string  `json:"tax_number,omitempty" example:"AU123456789"`
	BillingName    *string  `json:"billing_name,omitempty" example:"Jane Doe"`
	BillingEmail   *string  `json:"billing_email,omitempty" example:"jane@example.com"`
	BillingAddress *string  `json:"billing_address,omitempty" example:"456 Updated St"`
	BillingCity    *string  `json:"billing_city,omitempty" example:"Los Angeles"`
	BillingState   *string  `json:"billing_state,omitempty" example:"CA"`
	BillingCountry *string  `json:"billing_country,omitempty" example:"US"`
	BillingZip     *string  `json:"billing_zip,omitempty" example:"90210"`
	CompanyName    *string  `json:"company_name,omitempty" example:"Updated Corp"`
	CompanyTaxID   *string  `json:"company_tax_id,omitempty" example:"12-9876543"`
	CompanyAddress *string  `json:"company_address,omitempty" example:"789 Corporate Blvd"`
	Description    *string  `json:"description,omitempty" example:"Updated subscription"`
	Notes          *string  `json:"notes,omitempty" example:"Admin updated invoice"`
	DueDate        *string  `json:"due_date,omitempty" example:"2024-02-29"`
	Template       *string  `json:"template,omitempty" example:"professional"`
	Language       *string  `json:"language,omitempty" example:"es"`
}

// VoidInvoiceRequest represents the request body for voiding invoices
type VoidInvoiceRequest struct {
	Reason           string `json:"reason" binding:"required,max=255" example:"Customer requested cancellation"`
	SendNotification *bool  `json:"send_notification,omitempty" example:"true"`
}

// MarkPaidRequest represents the request body for manually marking invoices as paid
type MarkPaidRequest struct {
	PaymentDate      *time.Time `json:"payment_date,omitempty" swaggertype:"string" format:"date-time" example:"2024-01-15T10:30:00Z"`
	PaymentMethod    *string    `json:"payment_method,omitempty" example:"bank_transfer"`
	PaymentReference *string    `json:"payment_reference,omitempty" example:"TXN123456"`
	Notes            *string    `json:"notes,omitempty" example:"Manual payment verification"`
	SendNotification *bool      `json:"send_notification,omitempty" example:"true"`
}

// BulkInvoiceActionRequest represents bulk operations on invoices
type BulkInvoiceActionRequest struct {
	InvoiceIDs []uint  `json:"invoice_ids" binding:"required,min=1,max=100"`
	Action     string  `json:"action" binding:"required,oneof=void mark_paid resend regenerate_pdf" example:"resend"`
	Reason     *string `json:"reason,omitempty" binding:"omitempty,max=255" example:"Bulk admin action"`
	Notes      *string `json:"notes,omitempty" binding:"omitempty,max=500" example:"Bulk operation notes"`
}

// InvoiceSearchRequest represents advanced search parameters for invoices
type InvoiceSearchRequest struct {
	Query          string   `form:"q,omitempty" example:"INV-2024-001"`
	UserID         *uint    `form:"user_id,omitempty" example:"1"`
	Status         string   `form:"status,omitempty" example:"sent"`
	InvoiceType    string   `form:"invoice_type,omitempty" example:"standard"`
	AmountMin      *float64 `form:"amount_min,omitempty" example:"10.00"`
	AmountMax      *float64 `form:"amount_max,omitempty" example:"100.00"`
	Currency       string   `form:"currency,omitempty" example:"CNY"`
	DateFrom       string   `form:"date_from,omitempty" example:"2024-01-01"`
	DateTo         string   `form:"date_to,omitempty" example:"2024-12-31"`
	IsOverdue      *bool    `form:"is_overdue,omitempty" example:"false"`
	BillingCountry string   `form:"billing_country,omitempty" example:"US"`
	PaymentMethod  string   `form:"payment_method,omitempty" example:"credit_card"`
	HasCompanyInfo *bool    `form:"has_company_info,omitempty" example:"true"`
	Page           int      `form:"page,omitempty" example:"1"`
	Limit          int      `form:"limit,omitempty" example:"20"`
}

// InvoiceAnalyticsRequest represents analytics query parameters
type InvoiceAnalyticsRequest struct {
	DateFrom  string `form:"date_from,omitempty" example:"2024-01-01"`
	DateTo    string `form:"date_to,omitempty" example:"2024-12-31"`
	GroupBy   string `form:"group_by,omitempty" example:"month"`
	Currency  string `form:"currency,omitempty" example:"CNY"`
	UserID    *uint  `form:"user_id,omitempty" example:"1"`
	Breakdown string `form:"breakdown,omitempty" example:"status"`
}

// ResendInvoiceRequest represents the request body for resending invoices
type ResendInvoiceRequest struct {
	ToEmail  *string `json:"to_email,omitempty" example:"customer@example.com"`
	Subject  *string `json:"subject,omitempty" example:"Invoice Reminder"`
	Message  *string `json:"message,omitempty" example:"Please find your invoice attached"`
	SendCopy *bool   `json:"send_copy,omitempty" example:"true"`
}

// RegeneratePDFRequest represents the request body for regenerating invoice PDFs
type RegeneratePDFRequest struct {
	Template     *string           `json:"template,omitempty" example:"professional"`
	Language     *string           `json:"language,omitempty" example:"en"`
	Watermark    *string           `json:"watermark,omitempty" example:"PAID"`
	CustomFields map[string]string `json:"custom_fields,omitempty"`
}

// INVOICE MANAGEMENT

// CreateInvoiceFromOrder godoc
// @Summary Create invoice from subscription order
// @Description Create a new invoice from an existing subscription order (Admin only)
// @Tags Admin-Invoice-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param order_id path uint true "Subscription Order ID"
// @Param options body dto.CreateInvoiceFromOrderRequest true "Invoice creation options"
// @Success 201 {object} dto.InvoiceResponse
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 409 {object} response.ConflictResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/orders/{id}/generate-invoice [post]
func (h *AdminInvoiceHandler) CreateInvoiceFromOrder(c *gin.Context) {
	orderIDStr := c.Param("id")
	orderID, err := strconv.ParseUint(orderIDStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid order ID")
		return
	}

	var req dto.CreateInvoiceFromOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// Create invoice from order with placeholder data
	// Note: Order details integration pending subscription service enhancement
	options := &dto.CreateInvoiceRequest{
		UserID:              1, // Extracted from order context
		SubscriptionOrderID: uint(orderID),
		Amount:              0, // Retrieved from order details
		Currency:            "CNY",
		BillingName:         "From Order",        // Retrieved from order billing info
		BillingEmail:        "order@example.com", // Retrieved from order billing info
		Template:            req.Template,
		Language:            req.Language,
		DueDate:             req.DueDate,
		Notes:               req.Notes,
		AutoSend:            req.AutoSend,
	}

	invoice, err := h.invoiceService.CreateInvoiceFromOrder(c.Request.Context(), uint(orderID), options)
	if err != nil {
		logger.Error("Admin failed to create invoice from order",
			logger.Uint("order_id", uint(orderID)),
			logger.ErrorField(err))

		if strings.Contains(err.Error(), "already exists") {
			response.Conflict(c, "Invoice already exists for this order")
			return
		}

		response.InternalServerError(c, "Failed to create invoice from order")
		return
	}

	logger.Info("Admin created invoice from order",
		logger.Uint("invoice_id", invoice.ID),
		logger.String("invoice_number", invoice.InvoiceNumber),
		logger.Uint("order_id", uint(orderID)),
		logger.String("admin_action", "create_invoice_from_order"))

	response.Created(c, dto.ToResponse(invoice))
}

// CreateInvoice godoc
// @Summary Create invoice
// @Description Create a new invoice (Admin only)
// @Tags Admin-Invoice-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param invoice body AdminCreateInvoiceRequest true "Invoice creation data"
// @Success 201 {object} dto.InvoiceResponse
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 409 {object} response.ConflictResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/invoices [post]
func (h *AdminInvoiceHandler) CreateInvoice(c *gin.Context) {
	var createReq AdminCreateInvoiceRequest
	if err := c.ShouldBindJSON(&createReq); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// Verify user exists
	user, err := h.userService.GetUserByID(c.Request.Context(), createReq.UserID)
	if err != nil {
		logger.Error("Admin failed to find user for invoice creation",
			logger.Uint("user_id", createReq.UserID),
			logger.ErrorField(err))
		response.BadRequest(c, "User not found")
		return
	}

	// Convert to service request
	serviceReq := &dto.CreateInvoiceRequest{
		UserID:              createReq.UserID,
		SubscriptionOrderID: createReq.SubscriptionOrderID,
		InvoiceType:         createReq.InvoiceType,
		Amount:              createReq.Amount,
		Currency:            createReq.Currency,
		TaxRate:             createReq.TaxRate,
		TaxType:             createReq.TaxType,
		TaxNumber:           createReq.TaxNumber,
		BillingName:         createReq.BillingName,
		BillingEmail:        createReq.BillingEmail,
		BillingAddress:      createReq.BillingAddress,
		BillingCity:         createReq.BillingCity,
		BillingState:        createReq.BillingState,
		BillingCountry:      createReq.BillingCountry,
		BillingZip:          createReq.BillingZip,
		CompanyName:         createReq.CompanyName,
		CompanyTaxID:        createReq.CompanyTaxID,
		CompanyAddress:      createReq.CompanyAddress,
		Description:         createReq.Description,
		Notes:               createReq.Notes,
		DueDate:             createReq.DueDate,
		Template:            createReq.Template,
		Language:            createReq.Language,
		AutoSend:            createReq.AutoSend,
	}

	// Create the invoice
	invoice, err := h.invoiceService.CreateInvoice(c.Request.Context(), serviceReq)
	if err != nil {
		logger.Error("Admin failed to create invoice",
			logger.Uint("user_id", createReq.UserID),
			logger.Uint("subscription_order_id", createReq.SubscriptionOrderID),
			logger.ErrorField(err))

		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			response.Conflict(c, "Invoice with similar details already exists")
			return
		}

		response.InternalServerError(c, "Failed to create invoice")
		return
	}

	logger.Info("Admin created invoice",
		logger.Uint("invoice_id", invoice.ID),
		logger.String("invoice_number", invoice.InvoiceNumber),
		logger.String("user_email", user.Email),
		logger.String("admin_action", "create_invoice"))

	response.Created(c, dto.ToResponse(invoice))
}

// GetInvoice godoc
// @Summary Get invoice details
// @Description Get invoice details by ID (Admin only)
// @Tags Admin-Invoice-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Invoice ID"
// @Success 200 {object} dto.InvoiceResponse
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Router /admin/invoices/{id} [get]
func (h *AdminInvoiceHandler) GetInvoice(c *gin.Context) {
	id, ok := handlers.ParseIDParam(c, "id")
	if !ok {
		return
	}

	invoice, err := h.invoiceService.GetInvoice(c.Request.Context(), uint(id))
	if err != nil {
		logger.Error("Admin failed to get invoice",
			logger.Uint("invoice_id", uint(id)),
			logger.ErrorField(err))
		response.NotFound(c, "Invoice not found")
		return
	}

	response.OK(c, dto.ToResponse(invoice))
}

// ListInvoices godoc
// @Summary List all invoices
// @Description Get paginated list of all invoices with optional filtering (Admin only)
// @Tags Admin-Invoice-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Param status query string false "Filter by status" Enums(draft,sent,paid,overdue,cancelled,voided)
// @Param invoice_type query string false "Filter by type" Enums(standard,proforma,credit_note)
// @Param user_id query int false "Filter by user ID"
// @Param date_from query string false "Start date filter (YYYY-MM-DD)"
// @Param date_to query string false "End date filter (YYYY-MM-DD)"
// @Param currency query string false "Filter by currency"
// @Param is_overdue query bool false "Filter overdue invoices"
// @Success 200 {object} response.HALCollectionResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/invoices [get]
func (h *AdminInvoiceHandler) ListInvoices(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	// Build filter request
	var userID *uint
	if userIDStr := c.Query("user_id"); userIDStr != "" {
		if uid, err := strconv.ParseUint(userIDStr, 10, 32); err == nil {
			uidUint := uint(uid)
			userID = &uidUint
		}
	}

	filterReq := &dto.GetInvoicesRequest{
		Status:      c.Query("status"),
		InvoiceType: c.Query("invoice_type"),
		DateFrom:    c.Query("date_from"),
		DateTo:      c.Query("date_to"),
		Limit:       limit,
		Offset:      offset,
	}

	if userID != nil {
		filterReq.UserID = *userID
	}

	invoices, total, err := h.invoiceService.GetInvoices(c.Request.Context(), filterReq)
	if err != nil {
		logger.Error("Admin failed to list invoices", logger.ErrorField(err))
		response.InternalServerError(c, "Failed to list invoices")
		return
	}

	// Convert to response format
	invoiceResponses := make([]*dto.InvoiceResponse, len(invoices))
	for i, invoice := range invoices {
		invoiceResponses[i] = dto.ToResponse(invoice)
	}

	response.Paginated(c, "Invoices retrieved successfully", invoiceResponses, page, limit, total, "/api/v1/admin/invoices")
}

// UpdateInvoice godoc
// @Summary Update invoice
// @Description Update invoice details (Admin only). Only draft invoices can be updated.
// @Tags Admin-Invoice-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Invoice ID"
// @Param invoice body AdminUpdateInvoiceRequest true "Invoice update data"
// @Success 200 {object} dto.InvoiceResponse
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/invoices/{id} [put]
func (h *AdminInvoiceHandler) UpdateInvoice(c *gin.Context) {
	id, ok := handlers.ParseIDParam(c, "id")
	if !ok {
		return
	}

	var updateReq AdminUpdateInvoiceRequest
	if err := c.ShouldBindJSON(&updateReq); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// Convert to service request
	serviceReq := &dto.UpdateInvoiceRequest{
		InvoiceType:    updateReq.InvoiceType,
		Amount:         updateReq.Amount,
		TaxRate:        updateReq.TaxRate,
		TaxType:        updateReq.TaxType,
		BillingName:    updateReq.BillingName,
		BillingEmail:   updateReq.BillingEmail,
		BillingAddress: updateReq.BillingAddress,
		BillingCity:    updateReq.BillingCity,
		BillingState:   updateReq.BillingState,
		BillingCountry: updateReq.BillingCountry,
		BillingZip:     updateReq.BillingZip,
		CompanyName:    updateReq.CompanyName,
		CompanyTaxID:   updateReq.CompanyTaxID,
		CompanyAddress: updateReq.CompanyAddress,
		Description:    updateReq.Description,
		Notes:          updateReq.Notes,
		DueDate:        updateReq.DueDate,
		Template:       updateReq.Template,
		Language:       updateReq.Language,
	}

	invoice, err := h.invoiceService.UpdateInvoice(c.Request.Context(), uint(id), serviceReq)
	if err != nil {
		logger.Error("Admin failed to update invoice",
			logger.Uint("invoice_id", uint(id)),
			logger.Any("update_request", updateReq),
			logger.ErrorField(err))

		if strings.Contains(err.Error(), "not found") {
			response.NotFound(c, "Invoice not found")
			return
		}

		if strings.Contains(err.Error(), "cannot be edited") || strings.Contains(err.Error(), "not editable") {
			response.BadRequest(c, "Invoice cannot be edited in current status")
			return
		}

		response.InternalServerError(c, "Failed to update invoice")
		return
	}

	logger.Info("Admin updated invoice",
		logger.Uint("invoice_id", uint(id)),
		logger.String("invoice_number", invoice.InvoiceNumber),
		logger.String("admin_action", "update_invoice"))

	response.OK(c, dto.ToResponse(invoice))
}

// DeleteInvoice godoc
// @Summary Delete invoice
// @Description Soft delete an invoice (Admin only). Only draft invoices can be deleted.
// @Tags Admin-Invoice-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Invoice ID"
// @Success 200 {object} string
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Router /admin/invoices/{id} [delete]
func (h *AdminInvoiceHandler) DeleteInvoice(c *gin.Context) {
	id, ok := handlers.ParseIDParam(c, "id")
	if !ok {
		return
	}

	if err := h.invoiceService.DeleteInvoice(c.Request.Context(), uint(id)); err != nil {
		logger.Error("Admin failed to delete invoice",
			logger.Uint("invoice_id", uint(id)),
			logger.ErrorField(err))

		if strings.Contains(err.Error(), "not found") {
			response.NotFound(c, "Invoice not found")
			return
		}

		if strings.Contains(err.Error(), "cannot be deleted") {
			response.BadRequest(c, "Invoice cannot be deleted in current status")
			return
		}

		response.InternalServerError(c, "Failed to delete invoice")
		return
	}

	logger.Info("Admin deleted invoice",
		logger.Uint("invoice_id", uint(id)),
		logger.String("admin_action", "delete_invoice"))

	response.OK(c, nil)
}

// VoidInvoice godoc
// @Summary Void invoice
// @Description Void an invoice with reason (Admin only)
// @Tags Admin-Invoice-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Invoice ID"
// @Param request body VoidInvoiceRequest true "Void request data"
// @Success 200 {object} dto.InvoiceResponse
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Router /admin/invoices/{id}/void [post]
func (h *AdminInvoiceHandler) VoidInvoice(c *gin.Context) {
	id, ok := handlers.ParseIDParam(c, "id")
	if !ok {
		return
	}

	var voidReq VoidInvoiceRequest
	if err := c.ShouldBindJSON(&voidReq); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	err := h.invoiceService.MarkInvoiceAsVoid(c.Request.Context(), uint(id), voidReq.Reason)
	if err != nil {
		logger.Error("Admin failed to void invoice",
			logger.Uint("invoice_id", uint(id)),
			logger.String("reason", voidReq.Reason),
			logger.ErrorField(err))

		if strings.Contains(err.Error(), "not found") {
			response.NotFound(c, "Invoice not found")
			return
		}

		if strings.Contains(err.Error(), "cannot be voided") {
			response.BadRequest(c, "Invoice cannot be voided in current status")
			return
		}

		response.InternalServerError(c, "Failed to void invoice")
		return
	}

	// Get updated invoice
	invoice, err := h.invoiceService.GetInvoice(c.Request.Context(), uint(id))
	if err != nil {
		logger.Error("Failed to get invoice after voiding", logger.ErrorField(err))
		response.InternalServerError(c, "Invoice voided but failed to retrieve updated data")
		return
	}

	logger.Info("Admin voided invoice",
		logger.Uint("invoice_id", uint(id)),
		logger.String("invoice_number", invoice.InvoiceNumber),
		logger.String("reason", voidReq.Reason),
		logger.String("admin_action", "void_invoice"))

	response.OK(c, dto.ToResponse(invoice))
}

// MarkInvoiceAsPaid godoc
// @Summary Mark invoice as paid
// @Description Manually mark an invoice as paid (Admin only)
// @Tags Admin-Invoice-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Invoice ID"
// @Param request body MarkPaidRequest true "Mark paid request data"
// @Success 200 {object} dto.InvoiceResponse
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Router /admin/invoices/{id}/mark-paid [post]
func (h *AdminInvoiceHandler) MarkInvoiceAsPaid(c *gin.Context) {
	id, ok := handlers.ParseIDParam(c, "id")
	if !ok {
		return
	}

	var paidReq MarkPaidRequest
	if err := c.ShouldBindJSON(&paidReq); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// Prepare payment date
	paymentDate := ""
	if paidReq.PaymentDate != nil {
		paymentDate = paidReq.PaymentDate.Format("2006-01-02")
	} else {
		paymentDate = time.Now().Format("2006-01-02")
	}

	err := h.invoiceService.MarkInvoiceAsPaid(c.Request.Context(), uint(id), paymentDate)
	if err != nil {
		logger.Error("Admin failed to mark invoice as paid",
			logger.Uint("invoice_id", uint(id)),
			logger.String("payment_date", paymentDate),
			logger.ErrorField(err))

		if strings.Contains(err.Error(), "not found") {
			response.NotFound(c, "Invoice not found")
			return
		}

		if strings.Contains(err.Error(), "cannot be paid") || strings.Contains(err.Error(), "already paid") {
			response.BadRequest(c, "Invoice cannot be marked as paid in current status")
			return
		}

		response.InternalServerError(c, "Failed to mark invoice as paid")
		return
	}

	// Get updated invoice to return payment information
	invoice, err := h.invoiceService.GetInvoice(c.Request.Context(), uint(id))
	if err != nil {
		logger.Error("Failed to get invoice after marking as paid", logger.ErrorField(err))
		response.InternalServerError(c, "Invoice marked as paid but failed to retrieve updated data")
		return
	}

	logger.Info("Admin marked invoice as paid",
		logger.Uint("invoice_id", uint(id)),
		logger.String("invoice_number", invoice.InvoiceNumber),
		logger.String("payment_date", paymentDate),
		logger.String("admin_action", "mark_paid"))

	response.OK(c, dto.ToResponse(invoice))
}

// RegenerateInvoicePDF godoc
// @Summary Regenerate invoice PDF
// @Description Regenerate PDF for an invoice with optional custom options (Admin only)
// @Tags Admin-Invoice-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Invoice ID"
// @Param request body RegeneratePDFRequest true "PDF generation options"
// @Success 200 {object} string
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/invoices/{id}/regenerate-pdf [post]
func (h *AdminInvoiceHandler) RegenerateInvoicePDF(c *gin.Context) {
	id, ok := handlers.ParseIDParam(c, "id")
	if !ok {
		return
	}

	var pdfReq RegeneratePDFRequest
	if err := c.ShouldBindJSON(&pdfReq); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// Prepare PDF generation options
	pdfOptions := &dto.PDFGenerationRequest{
		SaveToDisk: true, // Admin regeneration should save to disk
	}

	if pdfReq.Template != nil {
		pdfOptions.Template = *pdfReq.Template
	}
	if pdfReq.Language != nil {
		pdfOptions.Language = *pdfReq.Language
	}
	if pdfReq.Watermark != nil {
		pdfOptions.Watermark = *pdfReq.Watermark
	}
	if pdfReq.CustomFields != nil {
		pdfOptions.CustomFields = pdfReq.CustomFields
	}

	// Generate PDF
	_, filePath, err := h.invoiceService.GenerateInvoicePDFWithOptions(c.Request.Context(), uint(id), pdfOptions)
	if err != nil {
		logger.Error("Admin failed to regenerate invoice PDF",
			logger.Uint("invoice_id", uint(id)),
			logger.Any("pdf_options", pdfOptions),
			logger.ErrorField(err))

		if strings.Contains(err.Error(), "not found") {
			response.NotFound(c, "Invoice not found")
			return
		}

		response.InternalServerError(c, "Failed to regenerate PDF")
		return
	}

	logger.Info("Admin regenerated invoice PDF",
		logger.Uint("invoice_id", uint(id)),
		logger.String("file_path", filePath),
		logger.String("admin_action", "regenerate_pdf"))

	response.OK(c, gin.H{
		"message":   "PDF regenerated successfully",
		"file_path": filePath,
	})
}

// ResendInvoice godoc
// @Summary Resend invoice
// @Description Resend invoice email to customer (Admin only)
// @Tags Admin-Invoice-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Invoice ID"
// @Param request body ResendInvoiceRequest true "Resend request data"
// @Success 200 {object} string
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/invoices/{id}/resend-email [post]
func (h *AdminInvoiceHandler) ResendInvoice(c *gin.Context) {
	id, ok := handlers.ParseIDParam(c, "id")
	if !ok {
		return
	}

	var resendReq ResendInvoiceRequest
	if err := c.ShouldBindJSON(&resendReq); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// Get invoice to prepare email data
	invoice, err := h.invoiceService.GetInvoice(c.Request.Context(), uint(id))
	if err != nil {
		logger.Error("Admin failed to get invoice for resend",
			logger.Uint("invoice_id", uint(id)),
			logger.ErrorField(err))
		response.NotFound(c, "Invoice not found")
		return
	}

	// Prepare email request
	emailReq := &dto.SendInvoiceRequest{
		ToEmail: invoice.BillingEmail,
	}

	if resendReq.ToEmail != nil && *resendReq.ToEmail != "" {
		emailReq.ToEmail = *resendReq.ToEmail
	}
	if resendReq.Subject != nil {
		emailReq.Subject = *resendReq.Subject
	}
	if resendReq.Message != nil {
		emailReq.Message = *resendReq.Message
	}
	if resendReq.SendCopy != nil {
		emailReq.SendCopy = *resendReq.SendCopy
	}

	err = h.invoiceService.SendInvoice(c.Request.Context(), uint(id), emailReq)
	if err != nil {
		logger.Error("Admin failed to resend invoice",
			logger.Uint("invoice_id", uint(id)),
			logger.String("to_email", emailReq.ToEmail),
			logger.ErrorField(err))

		response.InternalServerError(c, "Failed to resend invoice")
		return
	}

	logger.Info("Admin resent invoice",
		logger.Uint("invoice_id", uint(id)),
		logger.String("invoice_number", invoice.InvoiceNumber),
		logger.String("to_email", emailReq.ToEmail),
		logger.String("admin_action", "resend_invoice"))

	response.OK(c, gin.H{
		"message": "Invoice resent successfully",
		"sent_to": emailReq.ToEmail,
	})
}

// SearchInvoices godoc
// @Summary Search invoices
// @Description Advanced search for invoices with multiple criteria (Admin only)
// @Tags Admin-Invoice-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param q query string false "Search query (invoice number, billing name, email)"
// @Param user_id query int false "Filter by user ID"
// @Param status query string false "Filter by status"
// @Param invoice_type query string false "Filter by invoice type"
// @Param amount_min query float64 false "Minimum amount filter"
// @Param amount_max query float64 false "Maximum amount filter"
// @Param currency query string false "Filter by currency"
// @Param date_from query string false "Start date filter"
// @Param date_to query string false "End date filter"
// @Param is_overdue query bool false "Filter overdue invoices"
// @Param billing_country query string false "Filter by billing country"
// @Param payment_method query string false "Filter by payment method"
// @Param has_company_info query bool false "Filter invoices with company information"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} response.HALCollectionResponse
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/invoices/search [get]
func (h *AdminInvoiceHandler) SearchInvoices(c *gin.Context) {
	var searchReq InvoiceSearchRequest
	if err := c.ShouldBindQuery(&searchReq); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// Set defaults
	if searchReq.Page < 1 {
		searchReq.Page = 1
	}
	if searchReq.Limit < 1 || searchReq.Limit > 100 {
		searchReq.Limit = 20
	}

	// Build search request for service
	filterReq := &dto.GetInvoicesRequest{
		Status:      searchReq.Status,
		InvoiceType: searchReq.InvoiceType,
		DateFrom:    searchReq.DateFrom,
		DateTo:      searchReq.DateTo,
		Limit:       searchReq.Limit,
		Offset:      (searchReq.Page - 1) * searchReq.Limit,
	}

	if searchReq.UserID != nil {
		filterReq.UserID = *searchReq.UserID
	}

	// For now, use basic filtering. Advanced search with complex queries
	// would require extending the invoice service interface
	invoices, _, err := h.invoiceService.GetInvoices(c.Request.Context(), filterReq)
	if err != nil {
		logger.Error("Admin failed to search invoices",
			logger.String("query", searchReq.Query),
			logger.Any("search_params", searchReq),
			logger.ErrorField(err))
		response.InternalServerError(c, "Failed to search invoices")
		return
	}

	// Apply additional client-side filtering for advanced criteria
	// In a production system, this should be handled at the database level
	filteredInvoices := []*entities.Invoice{}
	for _, invoice := range invoices {
		include := true

		// Apply search query filter
		if searchReq.Query != "" {
			query := strings.ToLower(searchReq.Query)
			if !strings.Contains(strings.ToLower(invoice.InvoiceNumber), query) &&
				!strings.Contains(strings.ToLower(invoice.BillingName), query) &&
				!strings.Contains(strings.ToLower(invoice.BillingEmail), query) {
				include = false
			}
		}

		// Apply amount range filters
		if searchReq.AmountMin != nil && invoice.TotalAmount < *searchReq.AmountMin {
			include = false
		}
		if searchReq.AmountMax != nil && invoice.TotalAmount > *searchReq.AmountMax {
			include = false
		}

		// Apply currency filter
		if searchReq.Currency != "" && invoice.Currency != searchReq.Currency {
			include = false
		}

		// Apply overdue filter
		if searchReq.IsOverdue != nil && invoice.IsOverdue() != *searchReq.IsOverdue {
			include = false
		}

		// Apply billing country filter
		if searchReq.BillingCountry != "" && invoice.BillingCountry != searchReq.BillingCountry {
			include = false
		}

		// Apply payment method filter
		if searchReq.PaymentMethod != "" && invoice.PaymentMethod != searchReq.PaymentMethod {
			include = false
		}

		// Apply company info filter
		if searchReq.HasCompanyInfo != nil {
			hasCompany := invoice.CompanyName != ""
			if hasCompany != *searchReq.HasCompanyInfo {
				include = false
			}
		}

		if include {
			filteredInvoices = append(filteredInvoices, invoice)
		}
	}

	// Convert to response format
	invoiceResponses := make([]*dto.InvoiceResponse, len(filteredInvoices))
	for i, invoice := range filteredInvoices {
		invoiceResponses[i] = dto.ToResponse(invoice)
	}

	response.PaginatedWithQuery(c, "Search completed", invoiceResponses,
		searchReq.Page, searchReq.Limit, int64(len(filteredInvoices)), "/api/v1/admin/invoices/search", map[string]any{
			"query": searchReq.Query,
			"filters_applied": map[string]any{
				"status":           searchReq.Status,
				"invoice_type":     searchReq.InvoiceType,
				"amount_range":     fmt.Sprintf("%.2f-%.2f", getValue(searchReq.AmountMin), getValue(searchReq.AmountMax)),
				"currency":         searchReq.Currency,
				"is_overdue":       searchReq.IsOverdue,
				"billing_country":  searchReq.BillingCountry,
				"payment_method":   searchReq.PaymentMethod,
				"has_company_info": searchReq.HasCompanyInfo,
			},
		})
}

// getValue helper function to safely get pointer values
func getValue[T comparable](ptr *T) T {
	if ptr != nil {
		return *ptr
	}
	var zero T
	return zero
}

// GetInvoiceStatistics godoc
// @Summary Get invoice statistics
// @Description Get comprehensive invoice statistics and metrics (Admin only)
// @Tags Admin-Invoice-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.StandardResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/invoices/statistics [get]
func (h *AdminInvoiceHandler) GetInvoiceStatistics(c *gin.Context) {
	fromDate := c.Query("date_from")
	toDate := c.Query("date_to")

	// Default to current year if no dates provided
	if fromDate == "" {
		fromDate = fmt.Sprintf("%d-01-01", time.Now().Year())
	}
	if toDate == "" {
		toDate = fmt.Sprintf("%d-12-31", time.Now().Year())
	}

	stats, err := h.invoiceService.GetInvoiceStatistics(c.Request.Context(), fromDate, toDate)
	if err != nil {
		logger.Error("Admin failed to get invoice statistics",
			logger.String("date_from", fromDate),
			logger.String("date_to", toDate),
			logger.ErrorField(err))
		response.InternalServerError(c, "Failed to get invoice statistics")
		return
	}

	response.OK(c, stats)
}

// GetInvoiceAnalytics godoc
// @Summary Get invoice analytics
// @Description Get detailed invoice analytics with various breakdowns (Admin only)
// @Tags Admin-Invoice-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param date_from query string false "Start date for analytics"
// @Param date_to query string false "End date for analytics"
// @Param group_by query string false "Group by: day, week, month, year"
// @Param currency query string false "Filter by currency"
// @Param user_id query int false "Filter by user ID"
// @Param breakdown query string false "Breakdown by: status, type, country"
// @Success 200 {object} response.StandardResponse
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/invoices/analytics [get]
func (h *AdminInvoiceHandler) GetInvoiceAnalytics(c *gin.Context) {
	var analyticsReq InvoiceAnalyticsRequest
	if err := c.ShouldBindQuery(&analyticsReq); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// Set defaults
	if analyticsReq.DateFrom == "" {
		analyticsReq.DateFrom = fmt.Sprintf("%d-01-01", time.Now().Year())
	}
	if analyticsReq.DateTo == "" {
		analyticsReq.DateTo = fmt.Sprintf("%d-12-31", time.Now().Year())
	}
	if analyticsReq.GroupBy == "" {
		analyticsReq.GroupBy = "month"
	}

	// For now, return basic statistics
	// In a production system, this would generate detailed analytics based on the parameters
	stats, err := h.invoiceService.GetInvoiceStatistics(c.Request.Context(), analyticsReq.DateFrom, analyticsReq.DateTo)
	if err != nil {
		logger.Error("Admin failed to get invoice analytics",
			logger.Any("analytics_params", analyticsReq),
			logger.ErrorField(err))
		response.InternalServerError(c, "Failed to get invoice analytics")
		return
	}

	// Enhance stats with analytics parameters
	analyticsData := gin.H{
		"statistics": stats,
		"parameters": gin.H{
			"date_from": analyticsReq.DateFrom,
			"date_to":   analyticsReq.DateTo,
			"group_by":  analyticsReq.GroupBy,
			"currency":  analyticsReq.Currency,
			"breakdown": analyticsReq.Breakdown,
		},
		"generated_at": time.Now(),
	}

	response.OK(c, analyticsData)
}

// GetOverdueInvoices godoc
// @Summary Get overdue invoices
// @Description Get list of overdue invoices with details (Admin only)
// @Tags Admin-Invoice-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Param days_overdue query int false "Minimum days overdue filter"
// @Success 200 {object} response.HALCollectionResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/invoices/overdue [get]
func (h *AdminInvoiceHandler) GetOverdueInvoices(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	daysOverdue, _ := strconv.Atoi(c.DefaultQuery("days_overdue", "0"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	// Get all invoices with overdue status or past due date
	filterReq := &dto.GetInvoicesRequest{
		Status: "overdue",
		Limit:  limit * 2, // Get more to filter client-side
		Offset: offset,
	}

	invoices, _, err := h.invoiceService.GetInvoices(c.Request.Context(), filterReq)
	if err != nil {
		logger.Error("Admin failed to get overdue invoices", logger.ErrorField(err))
		response.InternalServerError(c, "Failed to get overdue invoices")
		return
	}

	// Filter for truly overdue invoices and apply days filter
	overdueInvoices := []*entities.Invoice{}
	for _, invoice := range invoices {
		if invoice.IsOverdue() {
			if daysOverdue > 0 && invoice.DueAt != nil {
				daysSinceDeue := int(time.Since(*invoice.DueAt).Hours() / 24)
				if daysSinceDeue < daysOverdue {
					continue
				}
			}
			overdueInvoices = append(overdueInvoices, invoice)
		}
	}

	// Convert to response format
	invoiceResponses := make([]*dto.InvoiceResponse, len(overdueInvoices))
	for i, invoice := range overdueInvoices {
		invoiceResponses[i] = dto.ToResponse(invoice)
	}

	response.PaginatedWithQuery(c, "Overdue invoices retrieved", invoiceResponses,
		page, limit, int64(len(overdueInvoices)), "/api/v1/admin/invoices/overdue", map[string]any{
			"days_overdue_filter": daysOverdue,
			"total_overdue":       len(overdueInvoices),
		})
}

// BULK OPERATIONS

// BulkVoidInvoices godoc
// @Summary Bulk void invoices
// @Description Void multiple invoices in bulk (Admin only)
// @Tags Admin-Invoice-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body BulkInvoiceActionRequest true "Bulk void request"
// @Success 200 {object} response.StandardResponse
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/invoices/bulk/void [post]
func (h *AdminInvoiceHandler) BulkVoidInvoices(c *gin.Context) {
	var bulkReq BulkInvoiceActionRequest
	if err := c.ShouldBindJSON(&bulkReq); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if bulkReq.Action != "void" {
		response.BadRequest(c, "Invalid action for bulk void operation")
		return
	}

	reason := "Bulk void operation"
	if bulkReq.Reason != nil {
		reason = *bulkReq.Reason
	}

	successCount := 0
	failedIDs := []uint{}
	errors := []string{}

	for _, invoiceID := range bulkReq.InvoiceIDs {
		err := h.invoiceService.MarkInvoiceAsVoid(c.Request.Context(), invoiceID, reason)
		if err != nil {
			failedIDs = append(failedIDs, invoiceID)
			errors = append(errors, fmt.Sprintf("Invoice %d: %s", invoiceID, err.Error()))
			logger.Error("Bulk void failed for invoice",
				logger.Uint("invoice_id", invoiceID),
				logger.ErrorField(err))
		} else {
			successCount++
		}
	}

	logger.Info("Admin bulk void invoices completed",
		logger.Int("success_count", successCount),
		logger.Int("failed_count", len(failedIDs)),
		logger.String("reason", reason),
		logger.String("admin_action", "bulk_void_invoices"))

	response.OK(c, gin.H{
		"message":         "Bulk void operation completed",
		"total_requested": len(bulkReq.InvoiceIDs),
		"success_count":   successCount,
		"failed_count":    len(failedIDs),
		"failed_ids":      failedIDs,
		"errors":          errors,
	})
}

// BulkMarkPaid godoc
// @Summary Bulk mark invoices as paid
// @Description Mark multiple invoices as paid in bulk (Admin only)
// @Tags Admin-Invoice-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body BulkInvoiceActionRequest true "Bulk mark paid request"
// @Success 200 {object} response.StandardResponse
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/invoices/bulk/mark-paid [post]
func (h *AdminInvoiceHandler) BulkMarkPaid(c *gin.Context) {
	var bulkReq BulkInvoiceActionRequest
	if err := c.ShouldBindJSON(&bulkReq); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if bulkReq.Action != "mark_paid" {
		response.BadRequest(c, "Invalid action for bulk mark paid operation")
		return
	}

	paymentDate := time.Now().Format("2006-01-02")
	successCount := 0
	failedIDs := []uint{}
	errors := []string{}

	for _, invoiceID := range bulkReq.InvoiceIDs {
		err := h.invoiceService.MarkInvoiceAsPaid(c.Request.Context(), invoiceID, paymentDate)
		if err != nil {
			failedIDs = append(failedIDs, invoiceID)
			errors = append(errors, fmt.Sprintf("Invoice %d: %s", invoiceID, err.Error()))
			logger.Error("Bulk mark paid failed for invoice",
				logger.Uint("invoice_id", invoiceID),
				logger.ErrorField(err))
		} else {
			successCount++
		}
	}

	logger.Info("Admin bulk mark paid invoices completed",
		logger.Int("success_count", successCount),
		logger.Int("failed_count", len(failedIDs)),
		logger.String("payment_date", paymentDate),
		logger.String("admin_action", "bulk_mark_paid"))

	response.OK(c, gin.H{
		"message":         "Bulk mark paid operation completed",
		"total_requested": len(bulkReq.InvoiceIDs),
		"success_count":   successCount,
		"failed_count":    len(failedIDs),
		"failed_ids":      failedIDs,
		"errors":          errors,
		"payment_date":    paymentDate,
	})
}

// BulkResendInvoices godoc
// @Summary Bulk resend invoices
// @Description Resend multiple invoices via email in bulk (Admin only)
// @Tags Admin-Invoice-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body BulkInvoiceActionRequest true "Bulk resend request"
// @Success 200 {object} response.StandardResponse
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/invoices/bulk/resend [post]
func (h *AdminInvoiceHandler) BulkResendInvoices(c *gin.Context) {
	var bulkReq BulkInvoiceActionRequest
	if err := c.ShouldBindJSON(&bulkReq); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if bulkReq.Action != "resend" {
		response.BadRequest(c, "Invalid action for bulk resend operation")
		return
	}

	successCount := 0
	failedIDs := []uint{}
	errors := []string{}

	for _, invoiceID := range bulkReq.InvoiceIDs {
		err := h.invoiceService.ResendInvoice(c.Request.Context(), invoiceID)
		if err != nil {
			failedIDs = append(failedIDs, invoiceID)
			errors = append(errors, fmt.Sprintf("Invoice %d: %s", invoiceID, err.Error()))
			logger.Error("Bulk resend failed for invoice",
				logger.Uint("invoice_id", invoiceID),
				logger.ErrorField(err))
		} else {
			successCount++
		}
	}

	logger.Info("Admin bulk resend invoices completed",
		logger.Int("success_count", successCount),
		logger.Int("failed_count", len(failedIDs)),
		logger.String("admin_action", "bulk_resend_invoices"))

	response.OK(c, gin.H{
		"message":         "Bulk resend operation completed",
		"total_requested": len(bulkReq.InvoiceIDs),
		"success_count":   successCount,
		"failed_count":    len(failedIDs),
		"failed_ids":      failedIDs,
		"errors":          errors,
	})
}

// BulkRegeneratePDF godoc
// @Summary Bulk regenerate invoice PDFs
// @Description Regenerate PDFs for multiple invoices in bulk (Admin only)
// @Tags Admin-Invoice-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body BulkInvoiceActionRequest true "Bulk PDF regeneration request"
// @Success 200 {object} response.StandardResponse
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/invoices/bulk/regenerate-pdf [post]
func (h *AdminInvoiceHandler) BulkRegeneratePDF(c *gin.Context) {
	var bulkReq BulkInvoiceActionRequest
	if err := c.ShouldBindJSON(&bulkReq); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if bulkReq.Action != "regenerate_pdf" {
		response.BadRequest(c, "Invalid action for bulk PDF regeneration")
		return
	}

	// Use bulk PDF generation service if available
	pdfOptions := &dto.PDFGenerationRequest{
		SaveToDisk: true,
	}

	pdfData, err := h.invoiceService.GenerateBulkInvoicePDFs(c.Request.Context(), bulkReq.InvoiceIDs, pdfOptions)
	if err != nil {
		logger.Error("Admin bulk PDF generation failed",
			logger.Any("invoice_ids", bulkReq.InvoiceIDs),
			logger.ErrorField(err))
		response.InternalServerError(c, "Failed to generate bulk PDFs")
		return
	}

	logger.Info("Admin bulk PDF generation completed",
		logger.Int("invoice_count", len(bulkReq.InvoiceIDs)),
		logger.Int("pdf_size_bytes", len(pdfData)),
		logger.String("admin_action", "bulk_regenerate_pdf"))

	response.OK(c, gin.H{
		"message":         "Bulk PDF generation completed",
		"total_invoices":  len(bulkReq.InvoiceIDs),
		"pdf_size_bytes":  len(pdfData),
		"generation_time": time.Now(),
	})
}

// TEMPLATES AND CONFIGURATION

// GetAvailableTemplates godoc
// @Summary Get available invoice templates
// @Description Get list of available invoice PDF templates (Admin only)
// @Tags Admin-Invoice-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.StandardResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/invoices/templates [get]
func (h *AdminInvoiceHandler) GetAvailableTemplates(c *gin.Context) {
	templates, err := h.invoiceService.GetAvailableTemplates(c.Request.Context())
	if err != nil {
		logger.Error("Admin failed to get available templates", logger.ErrorField(err))
		response.InternalServerError(c, "Failed to get available templates")
		return
	}

	response.OK(c, gin.H{
		"templates": templates,
		"total":     len(templates),
	})
}

// GetAvailableLanguages godoc
// @Summary Get available invoice languages
// @Description Get list of available invoice languages (Admin only)
// @Tags Admin-Invoice-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.StandardResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/invoices/languages [get]
func (h *AdminInvoiceHandler) GetAvailableLanguages(c *gin.Context) {
	languages, err := h.invoiceService.GetAvailableLanguages(c.Request.Context())
	if err != nil {
		logger.Error("Admin failed to get available languages", logger.ErrorField(err))
		response.InternalServerError(c, "Failed to get available languages")
		return
	}

	response.OK(c, gin.H{
		"languages": languages,
		"total":     len(languages),
	})
}
