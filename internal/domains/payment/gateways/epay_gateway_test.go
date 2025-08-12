package gateways

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"linke/internal/domains/payment/constants"
	"linke/internal/domains/payment/dto"
	"linke/internal/domains/payment/entities"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEpayGateway_NewEpayGateway tests the creation of a new epay gateway
func TestEpayGateway_NewEpayGateway(t *testing.T) {
	config := &entities.PaymentConfig{
		Method: constants.PaymentGatewayEpay,
		URL:    "https://api.epay.test.com/submit.php",
		PID:    "test_pid",
		Key:    "test_key",
	}

	gateway := NewEpayGateway(config)
	assert.NotNil(t, gateway)
	assert.Equal(t, config, gateway.config)
}

// TestEpayGateway_ValidateConfig tests configuration validation
func TestEpayGateway_ValidateConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  *entities.PaymentConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config",
			config: &entities.PaymentConfig{
				URL:       "https://api.epay.test.com/submit.php",
				PID:       "test_pid",
				Key:       "test_key",
				MinAmount: 0.01,
				MaxAmount: 99999.99,
			},
			wantErr: false,
		},
		{
			name:    "nil config",
			config:  nil,
			wantErr: true,
			errMsg:  "epay config is nil",
		},
		{
			name: "empty URL",
			config: &entities.PaymentConfig{
				URL: "",
				PID: "test_pid",
				Key: "test_key",
			},
			wantErr: true,
			errMsg:  "epay URL is required",
		},
		{
			name: "empty PID",
			config: &entities.PaymentConfig{
				URL: "https://api.epay.test.com/submit.php",
				PID: "",
				Key: "test_key",
			},
			wantErr: true,
			errMsg:  "epay PID is required",
		},
		{
			name: "empty Key",
			config: &entities.PaymentConfig{
				URL: "https://api.epay.test.com/submit.php",
				PID: "test_pid",
				Key: "",
			},
			wantErr: true,
			errMsg:  "epay Key is required",
		},
		{
			name: "invalid URL format",
			config: &entities.PaymentConfig{
				URL: "invalid-url",
				PID: "test_pid",
				Key: "test_key",
			},
			wantErr: true,
			errMsg:  "invalid epay URL format",
		},
		{
			name: "negative min amount",
			config: &entities.PaymentConfig{
				URL:       "https://api.epay.test.com/submit.php",
				PID:       "test_pid",
				Key:       "test_key",
				MinAmount: -1.0,
				MaxAmount: 100.0,
			},
			wantErr: true,
			errMsg:  "minimum amount cannot be negative",
		},
		{
			name: "invalid amount range",
			config: &entities.PaymentConfig{
				URL:       "https://api.epay.test.com/submit.php",
				PID:       "test_pid",
				Key:       "test_key",
				MinAmount: 100.0,
				MaxAmount: 50.0,
			},
			wantErr: true,
			errMsg:  "minimum amount must be less than maximum amount",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gateway := NewEpayGateway(tt.config)
			err := gateway.ValidateConfig()

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestEpayGateway_TestConnection tests the connection testing functionality
func TestEpayGateway_TestConnection(t *testing.T) {
	tests := []struct {
		name       string
		setupMock  func() *httptest.Server
		wantErr    bool
		errPattern string
	}{
		{
			name: "successful connection",
			setupMock: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
				}))
			},
			wantErr: false,
		},
		{
			name: "connection error",
			setupMock: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusInternalServerError)
				}))
			},
			wantErr:    true,
			errPattern: "epay gateway returned error status",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := tt.setupMock()
			defer server.Close()

			config := &entities.PaymentConfig{
				URL:       server.URL,
				PID:       "test_pid",
				Key:       "test_key",
				MinAmount: 0.01,
				MaxAmount: 99999.99,
			}

			gateway := NewEpayGateway(config)
			err := gateway.TestConnection()

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errPattern != "" {
					assert.Contains(t, err.Error(), tt.errPattern)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestEpayGateway_GetSupportedPaymentMethods tests supported payment methods
func TestEpayGateway_GetSupportedPaymentMethods(t *testing.T) {
	config := &entities.PaymentConfig{
		URL: "https://api.epay.test.com/submit.php",
		PID: "test_pid",
		Key: "test_key",
	}

	gateway := NewEpayGateway(config)
	methods := gateway.GetSupportedPaymentMethods()

	expectedMethods := []string{
		constants.PaymentMethodAlipay,
		constants.PaymentMethodWechat,
		constants.PaymentMethodQQ,
	}

	assert.Equal(t, expectedMethods, methods)
}

// TestEpayGateway_GetPaymentMethodName tests payment method name mapping
func TestEpayGateway_GetPaymentMethodName(t *testing.T) {
	config := &entities.PaymentConfig{
		URL: "https://api.epay.test.com/submit.php",
		PID: "test_pid",
		Key: "test_key",
	}

	gateway := NewEpayGateway(config)

	tests := []struct {
		method   string
		expected string
	}{
		{constants.PaymentMethodAlipay, "支付宝"},
		{constants.PaymentMethodWechat, "微信支付"},
		{constants.PaymentMethodQQ, "QQ钱包"},
		{"unknown", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.method, func(t *testing.T) {
			name := gateway.GetPaymentMethodName(tt.method)
			assert.Equal(t, tt.expected, name)
		})
	}
}

// TestEpayGateway_IsPaymentCompleted tests payment completion status checking
func TestEpayGateway_IsPaymentCompleted(t *testing.T) {
	config := &entities.PaymentConfig{
		URL: "https://api.epay.test.com/submit.php",
		PID: "test_pid",
		Key: "test_key",
	}

	gateway := NewEpayGateway(config)

	tests := []struct {
		status   string
		expected bool
	}{
		{"TRADE_SUCCESS", true},
		{"TRADE_FINISHED", true},
		{"success", true},
		{"1", true},
		{"TRADE_PENDING", false},
		{"TRADE_FAILED", false},
		{"TRADE_CLOSED", false},
		{"0", false},
		{"pending", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			result := gateway.IsPaymentCompleted(tt.status)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestEpayGateway_CreatePaymentOrder tests payment order creation
func TestEpayGateway_CreatePaymentOrder(t *testing.T) {
	config := &entities.PaymentConfig{
		URL:       "https://api.epay.test.com/submit.php",
		PID:       "test_pid",
		Key:       "test_secret_key",
		MinAmount: 0.01,
		MaxAmount: 99999.99,
	}

	gateway := NewEpayGateway(config)

	tests := []struct {
		name    string
		request *dto.CreatePaymentOrderRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid alipay order",
			request: &dto.CreatePaymentOrderRequest{
				UserID:        1,
				Gateway:       constants.PaymentGatewayEpay,
				PaymentMethod: constants.PaymentMethodAlipay,
				Amount:        29.99,
				Currency:      constants.CurrencyCNY,
				Subject:       "Test Payment",
				Body:          "Test payment description",
				NotifyURL:     "https://example.com/notify",
				ReturnURL:     "https://example.com/return",
			},
			wantErr: false,
		},
		{
			name: "valid wechat order",
			request: &dto.CreatePaymentOrderRequest{
				UserID:        1,
				Gateway:       constants.PaymentGatewayEpay,
				PaymentMethod: constants.PaymentMethodWechat,
				Amount:        50.00,
				Currency:      constants.CurrencyCNY,
				Subject:       "WeChat Payment",
				NotifyURL:     "https://example.com/notify",
				ReturnURL:     "https://example.com/return",
			},
			wantErr: false,
		},
		{
			name: "unsupported payment method",
			request: &dto.CreatePaymentOrderRequest{
				UserID:        1,
				Gateway:       constants.PaymentGatewayEpay,
				PaymentMethod: "unsupported",
				Amount:        29.99,
				Currency:      constants.CurrencyCNY,
				Subject:       "Test Payment",
				NotifyURL:     "https://example.com/notify",
				ReturnURL:     "https://example.com/return",
			},
			wantErr: true,
			errMsg:  "unsupported payment method",
		},
		{
			name: "amount too low",
			request: &dto.CreatePaymentOrderRequest{
				UserID:        1,
				Gateway:       constants.PaymentGatewayEpay,
				PaymentMethod: constants.PaymentMethodAlipay,
				Amount:        0.001,
				Currency:      constants.CurrencyCNY,
				Subject:       "Test Payment",
				NotifyURL:     "https://example.com/notify",
				ReturnURL:     "https://example.com/return",
			},
			wantErr: true,
			errMsg:  "amount 0.00 is outside valid range",
		},
		{
			name: "amount too high",
			request: &dto.CreatePaymentOrderRequest{
				UserID:        1,
				Gateway:       constants.PaymentGatewayEpay,
				PaymentMethod: constants.PaymentMethodAlipay,
				Amount:        100000.00,
				Currency:      constants.CurrencyCNY,
				Subject:       "Test Payment",
				NotifyURL:     "https://example.com/notify",
				ReturnURL:     "https://example.com/return",
			},
			wantErr: true,
			errMsg:  "amount 100000.00 is outside valid range",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response, err := gateway.CreatePaymentOrder(tt.request)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				assert.Nil(t, response)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, response)
				
				// Verify response fields
				assert.NotEmpty(t, response.PaymentNo)
				assert.NotEmpty(t, response.PaymentURL)
				assert.NotEmpty(t, response.QRCodeURL)
				assert.Equal(t, tt.request.Amount, response.Amount)
				assert.Equal(t, tt.request.Currency, response.Currency)
				assert.WithinDuration(t, time.Now().Add(30*time.Minute), response.ExpiredAt, time.Minute)
				
				// Verify payment URL contains required parameters
				parsedURL, err := url.Parse(response.PaymentURL)
				require.NoError(t, err)
				assert.Equal(t, config.URL, parsedURL.Scheme+"://"+parsedURL.Host+parsedURL.Path)
				
				query := parsedURL.Query()
				assert.Equal(t, config.PID, query.Get("pid"))
				assert.Equal(t, tt.request.Subject, query.Get("name"))
				expectedAmount := fmt.Sprintf("%.2f", tt.request.Amount)
				assert.Equal(t, expectedAmount, query.Get("money"))
				assert.NotEmpty(t, query.Get("sign"))
			}
		})
	}
}

// TestEpayGateway_VerifyPaymentNotify tests payment notification verification
func TestEpayGateway_VerifyPaymentNotify(t *testing.T) {
	config := &entities.PaymentConfig{
		URL: "https://api.epay.test.com/submit.php",
		PID: "test_pid",
		Key: "test_secret_key",
	}

	gateway := NewEpayGateway(config)

	// Create valid notification data
	validData := map[string]any{
		"pid":          "test_pid",
		"out_trade_no": "EPAY123456789",
		"trade_no":     "TXN987654321",
		"trade_status": "TRADE_SUCCESS",
		"money":        "29.99",
		"pay_time":     "2024-01-01 12:00:00",
	}

	// Calculate valid signature
	params := map[string]string{
		"pid":          "test_pid",
		"out_trade_no": "EPAY123456789",
		"trade_no":     "TXN987654321",
		"trade_status": "TRADE_SUCCESS",
		"money":        "29.99",
		"pay_time":     "2024-01-01 12:00:00",
	}
	validSign := gateway.generateSignature(params)
	validData["sign"] = validSign

	tests := []struct {
		name        string
		data        map[string]any
		expectValid bool
		expectData  bool
	}{
		{
			name:        "valid notification",
			data:        validData,
			expectValid: true,
			expectData:  true,
		},
		{
			name: "missing signature",
			data: map[string]any{
				"pid":          "test_pid",
				"out_trade_no": "EPAY123456789",
				"trade_status": "TRADE_SUCCESS",
				"money":        "29.99",
			},
			expectValid: false,
			expectData:  false,
		},
		{
			name: "invalid signature",
			data: map[string]any{
				"pid":          "test_pid",
				"out_trade_no": "EPAY123456789",
				"trade_status": "TRADE_SUCCESS",
				"money":        "29.99",
				"sign":         "invalid_signature",
			},
			expectValid: false,
			expectData:  false,
		},
		{
			name: "tampered amount",
			data: map[string]any{
				"pid":          "test_pid",
				"out_trade_no": "EPAY123456789",
				"trade_status": "TRADE_SUCCESS",
				"money":        "99.99", // Changed amount
				"sign":         validSign,
			},
			expectValid: false,
			expectData:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid, notifyData := gateway.VerifyPaymentNotify(tt.data)

			assert.Equal(t, tt.expectValid, isValid)

			if tt.expectData {
				assert.NotNil(t, notifyData)
				assert.Equal(t, "EPAY123456789", notifyData.OutTradeNo)
				assert.Equal(t, "TXN987654321", notifyData.TransactionID)
				assert.Equal(t, "TRADE_SUCCESS", notifyData.Status)
				assert.Equal(t, 29.99, notifyData.Amount)
			} else {
				assert.Nil(t, notifyData)
			}
		})
	}
}

// TestEpayGateway_QueryPaymentOrder tests payment order query functionality
func TestEpayGateway_QueryPaymentOrder(t *testing.T) {
	// Create a mock server for testing query endpoint
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify query parameters
		outTradeNo := r.URL.Query().Get("out_trade_no")
		assert.Equal(t, "EPAY123456789", outTradeNo)
		
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	}))
	defer server.Close()

	// Use the server URL directly without modification
	config := &entities.PaymentConfig{
		URL: server.URL + "/submit.php",
		PID: "test_pid",
		Key: "test_secret_key",
	}

	gateway := NewEpayGateway(config)

	response, err := gateway.QueryPaymentOrder("EPAY123456789")
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, "EPAY123456789", response.PaymentNo)
	assert.Equal(t, "pending", response.Status) // Default status in simplified implementation
}

// Helper test for signature generation
func TestEpayGateway_GenerateSignature(t *testing.T) {
	config := &entities.PaymentConfig{
		Key: "test_secret_key",
	}

	gateway := NewEpayGateway(config)

	params := map[string]string{
		"pid":          "test_pid",
		"type":         "alipay",
		"out_trade_no": "EPAY123456789",
		"name":         "Test Payment",
		"money":        "29.99",
	}

	signature1 := gateway.generateSignature(params)
	signature2 := gateway.generateSignature(params)

	// Same parameters should generate same signature
	assert.Equal(t, signature1, signature2)
	assert.NotEmpty(t, signature1)
	assert.Len(t, signature1, 32) // MD5 hash length

	// Different parameters should generate different signatures
	params["money"] = "30.00"
	signature3 := gateway.generateSignature(params)
	assert.NotEqual(t, signature1, signature3)
}

// Benchmark tests for performance critical operations
func BenchmarkEpayGateway_CreatePaymentOrder(b *testing.B) {
	config := &entities.PaymentConfig{
		URL:       "https://api.epay.test.com/submit.php",
		PID:       "test_pid",
		Key:       "test_secret_key",
		MinAmount: 0.01,
		MaxAmount: 99999.99,
	}

	gateway := NewEpayGateway(config)

	request := &dto.CreatePaymentOrderRequest{
		UserID:        1,
		Gateway:       constants.PaymentGatewayEpay,
		PaymentMethod: constants.PaymentMethodAlipay,
		Amount:        29.99,
		Currency:      constants.CurrencyCNY,
		Subject:       "Test Payment",
		NotifyURL:     "https://example.com/notify",
		ReturnURL:     "https://example.com/return",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := gateway.CreatePaymentOrder(request)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkEpayGateway_GenerateSignature(b *testing.B) {
	config := &entities.PaymentConfig{
		Key: "test_secret_key",
	}

	gateway := NewEpayGateway(config)

	params := map[string]string{
		"pid":          "test_pid",
		"type":         "alipay",
		"out_trade_no": "EPAY123456789",
		"name":         "Test Payment",
		"money":        "29.99",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		gateway.generateSignature(params)
	}
}