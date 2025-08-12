package gateways

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"linke/internal/domains/payment/constants"
	"linke/internal/domains/payment/dto"
	"linke/internal/domains/payment/entities"
	"linke/internal/shared/logger"
)

// EpayGateway implements the PaymentGateway interface for epay payment system
type EpayGateway struct {
	config *entities.PaymentConfig
}

// NewEpayGateway creates a new epay gateway instance
func NewEpayGateway(config *entities.PaymentConfig) *EpayGateway {
	return &EpayGateway{
		config: config,
	}
}

// CreatePaymentOrder creates a payment order through epay gateway
func (e *EpayGateway) CreatePaymentOrder(req *dto.CreatePaymentOrderRequest) (*dto.CreatePaymentOrderResponse, error) {
	logger.Info("Creating epay payment order",
		logger.String("gateway", req.Gateway),
		logger.String("method", req.PaymentMethod),
		logger.Float64("amount", req.Amount),
		logger.Uint("user_id", req.UserID))

	// Validate payment method
	if !e.isValidPaymentMethod(req.PaymentMethod) {
		return nil, fmt.Errorf("unsupported payment method: %s", req.PaymentMethod)
	}

	// Validate amount
	if !e.config.IsAmountValid(req.Amount) {
		return nil, fmt.Errorf("amount %.2f is outside valid range [%.2f, %.2f]", 
			req.Amount, e.config.MinAmount, e.config.MaxAmount)
	}

	// Generate out trade number
	outTradeNo := e.generateOutTradeNo()

	// Calculate expiration time
	expiredAt := time.Now().Add(30 * time.Minute) // Default 30 minutes
	if req.ExpiredMinutes > 0 {
		expiredAt = time.Now().Add(time.Duration(req.ExpiredMinutes) * time.Minute)
	}

	// Prepare epay parameters
	params := e.buildEpayParams(req, outTradeNo)

	// Generate signature
	signature := e.generateSignature(params)
	params["sign"] = signature

	// Create payment URL
	paymentURL := e.buildPaymentURL(params)

	// Generate QR code URL for mobile payments
	qrCodeURL := e.generateQRCodeURL(paymentURL)

	logger.Info("Epay payment order created successfully",
		logger.String("out_trade_no", outTradeNo),
		logger.String("payment_method", req.PaymentMethod),
		logger.String("payment_url", paymentURL))

	return &dto.CreatePaymentOrderResponse{
		PaymentNo:  outTradeNo, // Use out trade number as payment number
		PaymentURL: paymentURL,
		QRCodeURL:  qrCodeURL,
		Amount:     req.Amount,
		Currency:   req.Currency,
		ExpiredAt:  expiredAt,
		GatewayData: e.serializeGatewayData(params),
	}, nil
}

// QueryPaymentOrder queries payment order status from epay
func (e *EpayGateway) QueryPaymentOrder(outTradeNo string) (*dto.QueryPaymentOrderResponse, error) {
	logger.Info("Querying epay payment order status", logger.String("out_trade_no", outTradeNo))

	// Prepare query parameters
	params := map[string]string{
		"pid":          e.config.PID,
		"out_trade_no": outTradeNo,
	}

	// Generate signature for query
	signature := e.generateSignature(params)
	params["sign"] = signature

	// Build query URL
	queryURL := e.buildQueryURL(params)

	// Make HTTP request to epay
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Get(queryURL)
	if err != nil {
		logger.Error("Failed to query epay order status", logger.ErrorField(err))
		return nil, fmt.Errorf("failed to query payment status: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("epay query returned status code: %d", resp.StatusCode)
	}

	// Parse response (simplified for this implementation)
	// In real implementation, you would parse the actual epay response format
	return &dto.QueryPaymentOrderResponse{
		PaymentNo:     outTradeNo,
		Status:        "pending", // This would be parsed from actual response
		PaidAmount:    "",
		TransactionID: "",
		PaidAt:        "",
	}, nil
}

// VerifyPaymentNotify verifies payment notification from epay
func (e *EpayGateway) VerifyPaymentNotify(data map[string]any) (bool, *dto.NotifyData) {
	logger.Info("Verifying epay payment notification", logger.Any("notification_data", data))

	// Convert data to string map for processing
	stringData := make(map[string]string)
	for k, v := range data {
		if str, ok := v.(string); ok {
			stringData[k] = str
		} else {
			stringData[k] = fmt.Sprintf("%v", v)
		}
	}

	// Extract and verify signature
	receivedSign, exists := stringData["sign"]
	if !exists {
		logger.Warn("No signature found in notification")
		return false, nil
	}

	// Remove sign from data for verification
	delete(stringData, "sign")

	// Generate expected signature
	expectedSign := e.generateSignature(stringData)

	// Verify signature
	if receivedSign != expectedSign {
		logger.Warn("Signature verification failed",
			logger.String("received", receivedSign),
			logger.String("expected", expectedSign))
		return false, nil
	}

	// Extract payment information
	outTradeNo, _ := stringData["out_trade_no"]
	transactionID, _ := stringData["trade_no"]
	status, _ := stringData["trade_status"]
	amountStr, _ := stringData["money"]
	paidAt, _ := stringData["pay_time"]

	// Parse amount
	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		logger.Error("Failed to parse amount from notification", logger.ErrorField(err))
		amount = 0
	}

	notifyData := &dto.NotifyData{
		PaymentNo:     outTradeNo,
		OutTradeNo:    outTradeNo,
		TransactionID: transactionID,
		Amount:        amount,
		Status:        status,
		PaidAt:        paidAt,
	}

	logger.Info("Epay notification verified successfully",
		logger.String("out_trade_no", outTradeNo),
		logger.String("status", status),
		logger.String("transaction_id", transactionID))

	return true, notifyData
}

// IsPaymentCompleted checks if payment is completed based on epay status
func (e *EpayGateway) IsPaymentCompleted(status string) bool {
	completedStatuses := []string{
		"TRADE_SUCCESS",  // Epay payment successful
		"TRADE_FINISHED", // Epay payment finished
		"success",        // Generic success status
		"1",              // Numeric success status
	}

	for _, completedStatus := range completedStatuses {
		if status == completedStatus {
			return true
		}
	}

	return false
}

// GetSupportedPaymentMethods returns list of supported payment methods
func (e *EpayGateway) GetSupportedPaymentMethods() []string {
	return []string{
		constants.PaymentMethodAlipay,
		constants.PaymentMethodWechat,
		constants.PaymentMethodQQ,
	}
}

// GetPaymentMethodName returns human-readable name for payment method
func (e *EpayGateway) GetPaymentMethodName(method string) string {
	names := map[string]string{
		constants.PaymentMethodAlipay: "支付宝",
		constants.PaymentMethodWechat: "微信支付",
		constants.PaymentMethodQQ:     "QQ钱包",
	}

	if name, exists := names[method]; exists {
		return name
	}

	return method
}

// ValidateConfig validates the epay configuration
func (e *EpayGateway) ValidateConfig() error {
	if e.config == nil {
		return fmt.Errorf("epay config is nil")
	}

	if e.config.URL == "" {
		return fmt.Errorf("epay URL is required")
	}

	if e.config.PID == "" {
		return fmt.Errorf("epay PID is required")
	}

	if e.config.Key == "" {
		return fmt.Errorf("epay Key is required")
	}

	// Validate URL format
	if !strings.HasPrefix(e.config.URL, "http://") && !strings.HasPrefix(e.config.URL, "https://") {
		return fmt.Errorf("invalid epay URL format")
	}

	// Validate amount limits
	if e.config.MinAmount < 0 {
		return fmt.Errorf("minimum amount cannot be negative")
	}

	if e.config.MaxAmount <= 0 {
		return fmt.Errorf("maximum amount must be positive")
	}

	if e.config.MinAmount >= e.config.MaxAmount {
		return fmt.Errorf("minimum amount must be less than maximum amount")
	}

	return nil
}

// TestConnection tests the connection to epay gateway
func (e *EpayGateway) TestConnection() error {
	logger.Info("Testing epay gateway connection", logger.String("url", e.config.URL))

	// Create a test request to check if the gateway is reachable
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Head(e.config.URL)
	if err != nil {
		logger.Error("Epay gateway connection test failed", logger.ErrorField(err))
		return fmt.Errorf("failed to connect to epay gateway: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("epay gateway returned error status: %d", resp.StatusCode)
	}

	logger.Info("Epay gateway connection test successful")
	return nil
}

// Private helper methods

// isValidPaymentMethod checks if the payment method is supported
func (e *EpayGateway) isValidPaymentMethod(method string) bool {
	supportedMethods := e.GetSupportedPaymentMethods()
	for _, supportedMethod := range supportedMethods {
		if method == supportedMethod {
			return true
		}
	}
	return false
}

// generateOutTradeNo generates a unique out trade number
func (e *EpayGateway) generateOutTradeNo() string {
	// Format: EPAY + timestamp + random suffix
	timestamp := time.Now().Unix()
	return fmt.Sprintf("EPAY%d%04d", timestamp, time.Now().Nanosecond()%10000)
}

// buildEpayParams builds epay request parameters according to new API specification
func (e *EpayGateway) buildEpayParams(req *dto.CreatePaymentOrderRequest, outTradeNo string) map[string]string {
	params := map[string]string{
		"pid":          e.config.PID,
		"type":         e.mapPaymentMethodToEpayType(req.PaymentMethod),
		"out_trade_no": outTradeNo,
		"notify_url":   req.NotifyURL,
		"name":         req.Subject,
		"money":        fmt.Sprintf("%.2f", req.Amount),
		"sign_type":    "MD5", // Required by new API spec
	}

	// Add optional parameters
	if req.ReturnURL != "" {
		params["return_url"] = req.ReturnURL
	}

	// Use correct field name for client IP (clientip not client_ip)
	if req.ClientIP != "" {
		params["clientip"] = req.ClientIP
	}

	// Add device type based on user agent or default to pc
	device := "pc" // Default device type
	params["device"] = device

	// Add business extension parameter (param) - can be used for metadata
	if req.Metadata != "" {
		params["param"] = req.Metadata
	}

	return params
}

// mapPaymentMethodToEpayType maps internal payment method to epay type
func (e *EpayGateway) mapPaymentMethodToEpayType(method string) string {
	mapping := map[string]string{
		constants.PaymentMethodAlipay: "alipay",
		constants.PaymentMethodWechat: "wxpay",
		constants.PaymentMethodQQ:     "qqpay",
	}

	if epayType, exists := mapping[method]; exists {
		return epayType
	}

	return "alipay" // Default to alipay
}

// generateSignature generates MD5 signature for epay parameters
func (e *EpayGateway) generateSignature(params map[string]string) string {
	// Sort parameters by key
	keys := make([]string, 0, len(params))
	for k := range params {
		if k != "sign" { // Exclude sign parameter
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)

	// Build query string
	var parts []string
	for _, key := range keys {
		if params[key] != "" { // Only include non-empty values
			parts = append(parts, key+"="+params[key])
		}
	}

	// Add key
	queryString := strings.Join(parts, "&") + e.config.Key

	// Generate MD5 hash
	hash := md5.Sum([]byte(queryString))
	return hex.EncodeToString(hash[:])
}

// buildPaymentURL builds the complete payment URL using POST method
// According to new API spec, this should be a POST request to mapi.php
func (e *EpayGateway) buildPaymentURL(params map[string]string) string {
	values := url.Values{}
	for k, v := range params {
		values.Add(k, v)
	}

	return e.config.URL + "?" + values.Encode()
}

// buildQueryURL builds the query URL for payment status
func (e *EpayGateway) buildQueryURL(params map[string]string) string {
	values := url.Values{}
	for k, v := range params {
		values.Add(k, v)
	}

	// Assume query endpoint is at /query.php
	baseURL := strings.TrimSuffix(e.config.URL, "/submit.php")
	return baseURL + "/query.php?" + values.Encode()
}

// generateQRCodeURL generates QR code URL for mobile payments
func (e *EpayGateway) generateQRCodeURL(paymentURL string) string {
	// For QR code payments, we might use a different endpoint or parameter
	// This is a simplified implementation
	return paymentURL + "&qr=1"
}

// serializeGatewayData serializes gateway parameters to string
func (e *EpayGateway) serializeGatewayData(params map[string]string) string {
	var parts []string
	for k, v := range params {
		parts = append(parts, k+"="+v)
	}
	return strings.Join(parts, "&")
}