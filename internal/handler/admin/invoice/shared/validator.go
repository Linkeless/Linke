package shared

import (
	"errors"
	"strconv"
	"strings"

	"linke/internal/response"

	"github.com/gin-gonic/gin"
)

// InvoiceValidator provides validation utilities for invoice operations
type InvoiceValidator struct{}

// NewInvoiceValidator creates a new invoice validator
func NewInvoiceValidator() *InvoiceValidator {
	return &InvoiceValidator{}
}

// ValidateInvoiceID validates and parses invoice ID from URL parameter
func (v *InvoiceValidator) ValidateInvoiceID(c *gin.Context) (uint, error) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid invoice ID", "Invoice ID must be a valid number")
		return 0, errors.New("invalid invoice ID")
	}
	return uint(id), nil
}

// ValidateInvoiceStatus validates invoice status value
func (v *InvoiceValidator) ValidateInvoiceStatus(status string) error {
	validStatuses := map[string]bool{
		"draft":    true,
		"sent":     true,
		"paid":     true,
		"overdue":  true,
		"cancelled": true,
		"voided":   true,
	}

	if !validStatuses[status] {
		return errors.New("invalid status value, must be one of: draft, sent, paid, overdue, cancelled, voided")
	}
	return nil
}

// ValidateInvoiceType validates invoice type value
func (v *InvoiceValidator) ValidateInvoiceType(invoiceType string) error {
	validTypes := map[string]bool{
		"standard":     true,
		"proforma":     true,
		"credit_note":  true,
	}

	if !validTypes[invoiceType] {
		return errors.New("invalid invoice type, must be one of: standard, proforma, credit_note")
	}
	return nil
}

// ValidateCurrency validates currency code
func (v *InvoiceValidator) ValidateCurrency(currency string) error {
	if currency == "" {
		return errors.New("currency is required")
	}
	
	// Basic currency code validation (3-letter codes)
	if len(currency) != 3 {
		return errors.New("currency must be a 3-letter code")
	}
	
	currency = strings.ToUpper(currency)
	validCurrencies := map[string]bool{
		"USD": true,
		"EUR": true,
		"GBP": true,
		"JPY": true,
		"CNY": true,
		"KRW": true,
		// Add more currencies as needed
	}
	
	if !validCurrencies[currency] {
		return errors.New("unsupported currency")
	}
	
	return nil
}

// ValidatePaymentMethod validates payment method
func (v *InvoiceValidator) ValidatePaymentMethod(method string) error {
	if method == "" {
		return errors.New("payment method is required")
	}
	
	validMethods := map[string]bool{
		"credit_card":    true,
		"bank_transfer":  true,
		"paypal":         true,
		"stripe":         true,
		"alipay":         true,
		"wechat_pay":     true,
		"cash":           true,
		"check":          true,
		"other":          true,
	}
	
	if !validMethods[method] {
		return errors.New("invalid payment method")
	}
	
	return nil
}

// ValidatePaginationParams validates and returns pagination parameters
func (v *InvoiceValidator) ValidatePaginationParams(c *gin.Context) (int, int, int) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	offset := (page - 1) * limit
	return page, limit, offset
}

// ValidateSearchQuery validates search query parameter
func (v *InvoiceValidator) ValidateSearchQuery(c *gin.Context) (string, error) {
	query := strings.TrimSpace(c.Query("search"))
	if query == "" {
		response.BadRequest(c, "Search query is required")
		return "", errors.New("search query is required")
	}
	return query, nil
}

// ValidateSortParams validates sort parameters
func (v *InvoiceValidator) ValidateSortParams(c *gin.Context) (string, string) {
	sortBy := c.DefaultQuery("sort_by", "created_at")
	sortOrder := c.DefaultQuery("sort_order", "desc")
	
	validSortFields := map[string]bool{
		"created_at":    true,
		"issued_at":     true,
		"due_at":        true,
		"paid_at":       true,
		"total_amount":  true,
		"status":        true,
	}
	
	if !validSortFields[sortBy] {
		sortBy = "created_at"
	}
	
	if sortOrder != "asc" && sortOrder != "desc" {
		sortOrder = "desc"
	}
	
	return sortBy, sortOrder
}

// ValidateUserID validates user ID parameter
func (v *InvoiceValidator) ValidateUserID(userIDStr string) (uint, error) {
	if userIDStr == "" {
		return 0, nil // Optional parameter
	}
	
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		return 0, errors.New("invalid user ID")
	}
	
	return uint(userID), nil
}

// ValidateOrderID validates order ID parameter
func (v *InvoiceValidator) ValidateOrderID(orderID uint) error {
	if orderID == 0 {
		return errors.New("order ID is required")
	}
	return nil
}

// ValidateVoidReason validates void reason
func (v *InvoiceValidator) ValidateVoidReason(reason string) error {
	if strings.TrimSpace(reason) == "" {
		return errors.New("void reason is required")
	}
	if len(reason) > 500 {
		return errors.New("void reason must be less than 500 characters")
	}
	return nil
}