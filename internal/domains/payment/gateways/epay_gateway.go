package gateways

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
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

// EpayResponse represents the response structure from epay gateway
type EpayResponse struct {
	Code      int    `json:"code"`
	Message   string `json:"msg,omitempty"`
	TradeNo   string `json:"trade_no,omitempty"`
	PayURL    string `json:"payurl,omitempty"`
	QRCode    string `json:"qrcode,omitempty"`
	URLScheme string `json:"urlscheme,omitempty"`
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

	// Generate signature (before adding sign)
	signature := e.generateSignature(params)
	
	// Add sign after signature generation (不添加sign_type)
	params["sign"] = signature

	// Submit payment order to epay and get payment URL
	paymentURL, err := e.submitPaymentOrder(params)
	if err != nil {
		logger.Error("Failed to submit payment order to epay", logger.ErrorField(err))
		return nil, fmt.Errorf("failed to create payment order: %w", err)
	}

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
	// Determine notify URL - use request value, config default, or generate one
	notifyURL := req.NotifyURL
	if notifyURL == "" && e.config.NotifyURL != "" {
		notifyURL = e.config.NotifyURL
	}
	if notifyURL == "" {
		// Generate a default notify URL - this should be configured properly in production
		notifyURL = "https://linkeless.com/api/v1/payments/notify"
	}

	params := map[string]string{
		"pid":          e.config.PID,
		"type":         e.mapPaymentMethodToEpayType(req.PaymentMethod),
		"out_trade_no": outTradeNo,
		"notify_url":   notifyURL,
		"name":         req.Subject,
		"money":        fmt.Sprintf("%.2f", req.Amount),
	}

	// Add optional parameters
	if req.ReturnURL != "" {
		params["return_url"] = req.ReturnURL
	}

	// Use correct field name for client IP (clientip not client_ip)
	if req.ClientIP != "" {
		params["clientip"] = req.ClientIP
	}

	// 移除device参数，不发送

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
// 按照标准算法：1. 参数按ASCII码排序 2. sign、sign_type、空值不参与签名 3. 拼接成a=b&c=d格式 4. 加上KEY做MD5
func (e *EpayGateway) generateSignature(params map[string]string) string {
	// Log current config for debugging - 明文输出完整key
	logger.Debug("Current epay config",
		logger.String("url", e.config.URL),
		logger.String("pid", e.config.PID),
		logger.String("key_length", fmt.Sprintf("%d", len(e.config.Key))),
		logger.String("key", e.config.Key))

	// 1. 按照参数名ASCII码从小到大排序（a-z），sign、sign_type、和空值不参与签名
	keys := make([]string, 0, len(params))
	for k, v := range params {
		if k != "sign" && k != "sign_type" && v != "" { // 排除sign、sign_type和空值
			keys = append(keys, k)
		}
	}
	sort.Strings(keys) // ASCII排序

	// 2. 拼接成URL键值对格式：a=b&c=d&e=f，参数值不进行url编码
	var parts []string
	for _, key := range keys {
		parts = append(parts, key+"="+params[key])
	}
	queryString := strings.Join(parts, "&")
	
	// 3. 拼接商户密钥KEY进行MD5加密：sign = md5(a=b&c=d&e=f + KEY)
	signString := queryString + e.config.Key

	// Log signature details for debugging
	logger.Debug("Standard epay signature generation",
		logger.String("query_string", queryString),
		logger.String("sign_string", signString))

	// 4. MD5加密，结果为小写
	hash := md5.Sum([]byte(signString))
	signature := hex.EncodeToString(hash[:]) // hex.EncodeToString已经返回小写
	
	logger.Debug("Generated signature", 
		logger.String("signature", signature))
	
	return signature
}

// submitPaymentOrder submits payment order to epay via POST request and returns payment URL
func (e *EpayGateway) submitPaymentOrder(params map[string]string) (string, error) {
	// Build form data manually without URL encoding parameter values
	// 重要：参数值不要进行URL编码，保持与签名时一致
	var formParts []string
	for k, v := range params {
		formParts = append(formParts, k+"="+v)
	}
	formData := strings.Join(formParts, "&")

	// Use the configured URL directly, don't auto-append mapi.php
	apiURL := e.config.URL

	// Create POST request
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Log request details for debugging  
	logger.Info("Sending epay payment request",
		logger.String("url", apiURL),
		logger.String("pid", e.config.PID),
		logger.String("post_data", formData))

	// Log equivalent curl command for debugging
	curlCmd := fmt.Sprintf(`curl -X POST "%s" \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "%s"`, apiURL, formData)
	logger.Debug("Equivalent curl command", logger.String("curl", curlCmd))

	req, err := http.NewRequest("POST", apiURL, strings.NewReader(formData))
	if err != nil {
		return "", fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "Linke-Payment-Gateway/1.0")

	// Log all request headers
	logger.Debug("Request headers",
		logger.String("content_type", req.Header.Get("Content-Type")),
		logger.String("user_agent", req.Header.Get("User-Agent")),
		logger.String("content_length", fmt.Sprintf("%d", len(formData))))

	// Send request
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send POST request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body first to log actual content
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	// Log actual response for debugging
	logger.Debug("Epay response received",
		logger.Int("status_code", resp.StatusCode),
		logger.String("response_body", string(body)),
		logger.String("content_type", resp.Header.Get("Content-Type")))

	// Check status code but don't fail immediately on non-200
	if resp.StatusCode != http.StatusOK {
		logger.Warn("Epay returned non-200 status code",
			logger.Int("status_code", resp.StatusCode),
			logger.String("response", string(body)))
		// For redirect responses, still try to extract payment URL
	}

	// Try to parse as JSON first
	var epayResp EpayResponse
	if err := json.Unmarshal(body, &epayResp); err != nil {
		// If JSON parsing fails, treat response as HTML redirect page
		logger.Debug("Response is not JSON, treating as HTML redirect page",
			logger.String("response_preview", string(body)[:min(500, len(body))]))
		
		// For HTML responses, extract payment URL from meta refresh or form action
		paymentURL := e.extractPaymentURLFromHTML(string(body))
		if paymentURL != "" {
			logger.Debug("Extracted payment URL from HTML", logger.String("payment_url", paymentURL))
			return paymentURL, nil
		}
		
		// If we can't extract URL, return the original URL for redirect
		// This means the user should be redirected to the epay page directly
		return apiURL + "?" + e.buildPaymentQueryString(params), nil
	}

	// Check if request was successful
	if epayResp.Code != 1 {
		return "", fmt.Errorf("epay request failed: code=%d, message=%s", epayResp.Code, epayResp.Message)
	}

	// Extract payment URL from response (priority: payurl > qrcode > urlscheme)
	var paymentURL string
	if epayResp.PayURL != "" {
		paymentURL = epayResp.PayURL
	} else if epayResp.QRCode != "" {
		paymentURL = epayResp.QRCode
	} else if epayResp.URLScheme != "" {
		paymentURL = epayResp.URLScheme
	} else {
		return "", fmt.Errorf("no payment URL found in epay response")
	}

	logger.Info("Epay response parsed successfully",
		logger.String("trade_no", epayResp.TradeNo),
		logger.String("payment_url", paymentURL))

	return paymentURL, nil
}

// extractPaymentURLFromHTML tries to extract payment URL from HTML response
func (e *EpayGateway) extractPaymentURLFromHTML(html string) string {
	// Try to find meta refresh URL
	metaRefreshPattern := regexp.MustCompile(`<meta[^>]*http-equiv=["']refresh["'][^>]*content=["'][^"']*url=([^"']+)["'][^>]*>`)
	if matches := metaRefreshPattern.FindStringSubmatch(html); len(matches) > 1 {
		return matches[1]
	}
	
	// Try to find form action URL
	formActionPattern := regexp.MustCompile(`<form[^>]*action=["']([^"']+)["'][^>]*>`)
	if matches := formActionPattern.FindStringSubmatch(html); len(matches) > 1 {
		return matches[1]
	}
	
	// Try to find direct redirect URL in JavaScript
	jsRedirectPattern := regexp.MustCompile(`location\.href\s*=\s*["']([^"']+)["']`)
	if matches := jsRedirectPattern.FindStringSubmatch(html); len(matches) > 1 {
		return matches[1]
	}
	
	return ""
}

// buildPaymentQueryString builds query string from parameters
func (e *EpayGateway) buildPaymentQueryString(params map[string]string) string {
	values := url.Values{}
	for k, v := range params {
		values.Add(k, v)
	}
	return values.Encode()
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
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