package shared

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"linke/internal/response"

	"github.com/gin-gonic/gin"
)

// OrderValidator provides validation utilities for order operations
type OrderValidator struct{}

// NewOrderValidator creates a new order validator
func NewOrderValidator() *OrderValidator {
	return &OrderValidator{}
}

// ValidateOrderID validates and parses order ID from URL parameter
func (v *OrderValidator) ValidateOrderID(c *gin.Context) (uint, error) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid order ID", "Order ID must be a valid number")
		return 0, errors.New("invalid order ID")
	}
	return uint(id), nil
}

// ValidateOrderStatus validates order status value
func (v *OrderValidator) ValidateOrderStatus(status string) error {
	validStatuses := map[string]bool{
		"pending":   true,
		"paid":      true,
		"failed":    true,
		"cancelled": true,
		"refunded":  true,
	}

	if !validStatuses[status] {
		return errors.New("invalid status value, must be one of: pending, paid, failed, cancelled, refunded")
	}
	return nil
}

// ValidateOrderType validates order type value
func (v *OrderValidator) ValidateOrderType(orderType string) error {
	validTypes := map[string]bool{
		"new":       true,
		"renewal":   true,
		"upgrade":   true,
		"downgrade": true,
	}

	if !validTypes[orderType] {
		return errors.New("invalid order type, must be one of: new, renewal, upgrade, downgrade")
	}
	return nil
}

// ValidatePaymentMethod validates payment method
func (v *OrderValidator) ValidatePaymentMethod(paymentMethod string) error {
	if paymentMethod == "" {
		return nil // Optional parameter
	}

	validMethods := map[string]bool{
		"alipay":      true,
		"wechat":      true,
		"paypal":      true,
		"stripe":      true,
		"coinbase":    true,
		"usdt":        true,
		"bank_card":   true,
		"crypto":      true,
		"manual":      true,
	}

	if !validMethods[paymentMethod] {
		return errors.New("invalid payment method")
	}
	return nil
}

// ValidatePaymentGateway validates payment gateway
func (v *OrderValidator) ValidatePaymentGateway(gateway string) error {
	if gateway == "" {
		return nil // Optional parameter
	}

	validGateways := map[string]bool{
		"epay":     true,
		"stripe":   true,
		"paypal":   true,
		"coinbase": true,
		"manual":   true,
	}

	if !validGateways[gateway] {
		return errors.New("invalid payment gateway")
	}
	return nil
}

// ValidateAmount validates monetary amount
func (v *OrderValidator) ValidateAmount(amount float64) error {
	if amount < 0 {
		return errors.New("amount cannot be negative")
	}
	if amount > 100000 { // Reasonable upper limit
		return errors.New("amount exceeds maximum allowed value")
	}
	return nil
}

// ValidateDateRange validates date range parameters
func (v *OrderValidator) ValidateDateRange(startDate, endDate string) error {
	if startDate != "" {
		if _, err := time.Parse("2006-01-02", startDate); err != nil {
			return errors.New("invalid start date format, use YYYY-MM-DD")
		}
	}

	if endDate != "" {
		if _, err := time.Parse("2006-01-02", endDate); err != nil {
			return errors.New("invalid end date format, use YYYY-MM-DD")
		}
	}

	if startDate != "" && endDate != "" {
		start, _ := time.Parse("2006-01-02", startDate)
		end, _ := time.Parse("2006-01-02", endDate)
		if start.After(end) {
			return errors.New("start date cannot be after end date")
		}
	}

	return nil
}

// ValidatePaginationParams validates and returns pagination parameters
func (v *OrderValidator) ValidatePaginationParams(c *gin.Context) (int, int, int) {
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
func (v *OrderValidator) ValidateSearchQuery(search string) error {
	if len(search) > 500 {
		return errors.New("search query must be less than 500 characters")
	}
	return nil
}

// ValidateSortParams validates sort parameters
func (v *OrderValidator) ValidateSortParams(c *gin.Context) (string, string) {
	sortBy := c.DefaultQuery("sort_by", "created_at")
	sortOrder := c.DefaultQuery("sort_order", "desc")

	validSortFields := map[string]bool{
		"created_at":    true,
		"paid_at":       true,
		"amount":        true,
		"total_amount":  true,
		"status":        true,
		"order_type":    true,
		"user_id":       true,
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
func (v *OrderValidator) ValidateUserID(userIDStr string) (uint, error) {
	if userIDStr == "" {
		return 0, nil // Optional parameter
	}

	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		return 0, errors.New("invalid user ID")
	}

	return uint(userID), nil
}

// ValidateRefundAmount validates refund amount
func (v *OrderValidator) ValidateRefundAmount(amount float64) error {
	if amount <= 0 {
		return errors.New("refund amount must be greater than 0")
	}
	if amount > 100000 {
		return errors.New("refund amount exceeds maximum allowed value")
	}
	return nil
}

// ValidateRefundReason validates refund reason
func (v *OrderValidator) ValidateRefundReason(reason string) error {
	if strings.TrimSpace(reason) == "" {
		return errors.New("refund reason is required")
	}
	if len(reason) > 500 {
		return errors.New("refund reason must be less than 500 characters")
	}
	return nil
}

// ValidateAdminConfirmation validates admin confirmation requirement
func (v *OrderValidator) ValidateAdminConfirmation(confirmed bool, operation string) error {
	criticalOperations := map[string]bool{
		"paid":      true,
		"refunded":  true,
		"cancelled": true,
		"refund":    true,
		"bulk":      true,
	}

	if criticalOperations[operation] && !confirmed {
		return errors.New("admin confirmation is required for this operation")
	}
	return nil
}

// ValidateBulkOrderIDs validates bulk operation order IDs
func (v *OrderValidator) ValidateBulkOrderIDs(orderIDs []uint) error {
	if len(orderIDs) == 0 {
		return errors.New("at least one order ID is required")
	}
	if len(orderIDs) > 100 {
		return errors.New("maximum 100 orders can be processed at once")
	}
	return nil
}

// ValidateBulkOperation validates bulk operation type
func (v *OrderValidator) ValidateBulkOperation(operation string) error {
	validOperations := map[string]bool{
		"cancel": true,
		"refund": true,
		"export": true,
	}

	if !validOperations[operation] {
		return errors.New("invalid bulk operation, must be one of: cancel, refund, export")
	}
	return nil
}

// ValidateStatsPeriod validates statistics period parameter
func (v *OrderValidator) ValidateStatsPeriod(period string) error {
	validPeriods := map[string]bool{
		"today":   true,
		"week":    true,
		"month":   true,
		"quarter": true,
		"year":    true,
		"all":     true,
	}

	if period != "" && !validPeriods[period] {
		return errors.New("invalid period, must be one of: today, week, month, quarter, year, all")
	}
	return nil
}

// ValidatePaymentEvidence validates payment evidence data
func (v *OrderValidator) ValidatePaymentEvidence(evidence *PaymentEvidenceRequest) error {
	if evidence == nil {
		return errors.New("payment evidence is required")
	}

	if strings.TrimSpace(evidence.TransactionID) == "" {
		return errors.New("transaction ID is required")
	}

	if strings.TrimSpace(evidence.PaymentGateway) == "" {
		return errors.New("payment gateway is required")
	}

	if strings.TrimSpace(evidence.PaymentMethod) == "" {
		return errors.New("payment method is required")
	}

	if evidence.AmountReceived <= 0 {
		return errors.New("amount received must be greater than 0")
	}

	if strings.TrimSpace(evidence.Currency) == "" {
		return errors.New("currency is required")
	}

	if strings.TrimSpace(evidence.PaymentTime) == "" {
		return errors.New("payment time is required")
	}

	// Validate payment time format
	if _, err := time.Parse(time.RFC3339, evidence.PaymentTime); err != nil {
		return errors.New("invalid payment time format, use RFC3339 format")
	}

	return nil
}

// PaymentEvidenceRequest represents payment verification data
type PaymentEvidenceRequest struct {
	TransactionID     string  `json:"transaction_id" binding:"required" example:"txn_123456789"`
	PaymentGateway    string  `json:"payment_gateway" binding:"required" example:"alipay"`
	PaymentMethod     string  `json:"payment_method" binding:"required" example:"alipay"`
	AmountReceived    float64 `json:"amount_received" binding:"required,min=0.01" example:"29.99"`
	Currency          string  `json:"currency" binding:"required" example:"USD"`
	PaymentTime       string  `json:"payment_time" binding:"required" example:"2024-01-01T10:30:00Z"`
	GatewayResponse   string  `json:"gateway_response,omitempty" example:"SUCCESS"`
	VerificationNotes string  `json:"verification_notes,omitempty" example:"Payment verified manually"`
}