package implementations

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"linke/internal/domains/payment/constants"
	"linke/internal/domains/payment/dto"
	"linke/internal/domains/payment/entities"
)

// Test basic payment method service configuration constants
func TestPaymentMethodServiceConstants(t *testing.T) {
	assert.Equal(t, 10, MaxPaymentMethodsPerUser)
	assert.Equal(t, 0.5, HighFailureRateThreshold)
}

// Test payment method entity validation functions
func TestPaymentMethodValidation(t *testing.T) {
	tests := []struct {
		name           string
		paymentMethod  *entities.PaymentMethod
		expectedActive bool
	}{
		{
			name: "active payment method",
			paymentMethod: &entities.PaymentMethod{
				Active: true,
				Status: constants.PaymentMethodStatusActive,
			},
			expectedActive: true,
		},
		{
			name: "inactive payment method",
			paymentMethod: &entities.PaymentMethod{
				Active: false,
				Status: constants.PaymentMethodStatusActive,
			},
			expectedActive: false,
		},
		{
			name: "invalid status payment method",
			paymentMethod: &entities.PaymentMethod{
				Active: true,
				Status: constants.PaymentMethodStatusInvalid,
			},
			expectedActive: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.paymentMethod.IsActive()
			assert.Equal(t, tt.expectedActive, result)
		})
	}
}

// Test payment method request validation
func TestCreatePaymentMethodRequest(t *testing.T) {
	req := &dto.CreatePaymentMethodRequest{
		Type:            constants.PaymentMethodTypeCard,
		Gateway:         constants.PaymentGatewayEpay,
		Method:          constants.PaymentMethodAlipay,
		DisplayName:     "My Alipay Account",
		PaymentToken:    "epay_test_token_12345",
		MaskedInfo:      "ali***@example.com",
		Brand:           "Alipay",
		BillingCountry:  "CN",
		BillingPostcode: "100000",
		SetAsDefault:    false,
	}

	assert.Equal(t, constants.PaymentMethodTypeCard, req.Type)
	assert.Equal(t, constants.PaymentGatewayEpay, req.Gateway)
	assert.Equal(t, constants.PaymentMethodAlipay, req.Method)
	assert.Equal(t, "My Alipay Account", req.DisplayName)
	assert.Equal(t, "epay_test_token_12345", req.PaymentToken)
	assert.False(t, req.SetAsDefault)
}

// Test payment method list response
func TestPaymentMethodListResponse(t *testing.T) {
	methods := []dto.PaymentMethodResponse{
		{
			ID:          1,
			UserID:      100,
			Type:        constants.PaymentMethodTypeCard,
			Gateway:     constants.PaymentGatewayEpay,
			Method:      constants.PaymentMethodAlipay,
			DisplayName: "Method 1",
			IsDefault:   true,
		},
		{
			ID:          2,
			UserID:      100,
			Type:        constants.PaymentMethodTypeCard,
			Gateway:     constants.PaymentGatewayEpay,
			Method:      constants.PaymentMethodWechat,
			DisplayName: "Method 2",
			IsDefault:   false,
		},
	}

	listResponse := &dto.PaymentMethodListResponse{
		PaymentMethods: methods,
		Total:          len(methods),
		DefaultMethod:  &methods[0],
	}

	assert.Equal(t, 2, listResponse.Total)
	assert.NotNil(t, listResponse.DefaultMethod)
	assert.Equal(t, uint(1), listResponse.DefaultMethod.ID)
	assert.True(t, listResponse.DefaultMethod.IsDefault)
}
