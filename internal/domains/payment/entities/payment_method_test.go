package entities

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	"linke/internal/domains/payment/constants"
)

func TestPaymentMethod_IsActive(t *testing.T) {
	tests := []struct {
		name           string
		paymentMethod  *PaymentMethod
		expectedActive bool
	}{
		{
			name: "active payment method",
			paymentMethod: &PaymentMethod{
				Active: true,
				Status: constants.PaymentMethodStatusActive,
			},
			expectedActive: true,
		},
		{
			name: "inactive payment method",
			paymentMethod: &PaymentMethod{
				Active: false,
				Status: constants.PaymentMethodStatusActive,
			},
			expectedActive: false,
		},
		{
			name: "invalid status payment method",
			paymentMethod: &PaymentMethod{
				Active: true,
				Status: constants.PaymentMethodStatusInvalid,
			},
			expectedActive: false,
		},
		{
			name: "expired card",
			paymentMethod: &PaymentMethod{
				Active:      true,
				Status:      constants.PaymentMethodStatusActive,
				ExpiryMonth: intPtr(1),
				ExpiryYear:  intPtr(2020), // Expired year
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

func TestPaymentMethod_IsExpired(t *testing.T) {
	currentYear := time.Now().Year()
	currentMonth := int(time.Now().Month())

	tests := []struct {
		name            string
		paymentMethod   *PaymentMethod
		expectedExpired bool
	}{
		{
			name: "no expiry date",
			paymentMethod: &PaymentMethod{
				ExpiryMonth: nil,
				ExpiryYear:  nil,
			},
			expectedExpired: false,
		},
		{
			name: "future expiry",
			paymentMethod: &PaymentMethod{
				ExpiryMonth: intPtr(12),
				ExpiryYear:  intPtr(currentYear + 1),
			},
			expectedExpired: false,
		},
		{
			name: "current month and year",
			paymentMethod: &PaymentMethod{
				ExpiryMonth: intPtr(currentMonth),
				ExpiryYear:  intPtr(currentYear),
			},
			expectedExpired: false, // Should not be expired until end of month
		},
		{
			name: "past year",
			paymentMethod: &PaymentMethod{
				ExpiryMonth: intPtr(12),
				ExpiryYear:  intPtr(currentYear - 1),
			},
			expectedExpired: true,
		},
		{
			name: "past month current year",
			paymentMethod: &PaymentMethod{
				ExpiryMonth: intPtr(1),
				ExpiryYear:  intPtr(currentYear),
			},
			expectedExpired: currentMonth > 1, // Only expired if current month is after January
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.paymentMethod.IsExpired()
			assert.Equal(t, tt.expectedExpired, result)
		})
	}
}

func TestPaymentMethod_CanBeUsedForPayment(t *testing.T) {
	tests := []struct {
		name           string
		paymentMethod  *PaymentMethod
		expectedUsable bool
	}{
		{
			name: "active payment method",
			paymentMethod: &PaymentMethod{
				Active: true,
				Status: constants.PaymentMethodStatusActive,
			},
			expectedUsable: true,
		},
		{
			name: "inactive payment method",
			paymentMethod: &PaymentMethod{
				Active: false,
				Status: constants.PaymentMethodStatusActive,
			},
			expectedUsable: false,
		},
		{
			name: "soft deleted payment method",
			paymentMethod: &PaymentMethod{
				Active:    true,
				Status:    constants.PaymentMethodStatusActive,
				DeletedAt: gorm.DeletedAt{Valid: true, Time: time.Now()},
			},
			expectedUsable: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.paymentMethod.CanBeUsedForPayment()
			assert.Equal(t, tt.expectedUsable, result)
		})
	}
}

func TestPaymentMethod_GetFailureRate(t *testing.T) {
	tests := []struct {
		name          string
		paymentMethod *PaymentMethod
		expectedRate  float64
	}{
		{
			name: "no usage",
			paymentMethod: &PaymentMethod{
				SuccessfulUses: 0,
				FailedUses:     0,
			},
			expectedRate: 0.0,
		},
		{
			name: "all successful",
			paymentMethod: &PaymentMethod{
				SuccessfulUses: 10,
				FailedUses:     0,
			},
			expectedRate: 0.0,
		},
		{
			name: "all failed",
			paymentMethod: &PaymentMethod{
				SuccessfulUses: 0,
				FailedUses:     5,
			},
			expectedRate: 1.0,
		},
		{
			name: "50% failure rate",
			paymentMethod: &PaymentMethod{
				SuccessfulUses: 5,
				FailedUses:     5,
			},
			expectedRate: 0.5,
		},
		{
			name: "20% failure rate",
			paymentMethod: &PaymentMethod{
				SuccessfulUses: 8,
				FailedUses:     2,
			},
			expectedRate: 0.2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.paymentMethod.GetFailureRate()
			assert.Equal(t, tt.expectedRate, result)
		})
	}
}

func TestPaymentMethod_NeedsRevalidation(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name               string
		paymentMethod      *PaymentMethod
		expectedValidation bool
	}{
		{
			name: "never validated",
			paymentMethod: &PaymentMethod{
				LastValidatedAt: nil,
			},
			expectedValidation: true,
		},
		{
			name: "recently validated",
			paymentMethod: &PaymentMethod{
				LastValidatedAt: &now,
			},
			expectedValidation: false,
		},
		{
			name: "validated 31 days ago",
			paymentMethod: &PaymentMethod{
				LastValidatedAt: timePtr(now.AddDate(0, 0, -31)),
			},
			expectedValidation: true,
		},
		{
			name: "validated 29 days ago",
			paymentMethod: &PaymentMethod{
				LastValidatedAt: timePtr(now.AddDate(0, 0, -29)),
			},
			expectedValidation: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.paymentMethod.NeedsRevalidation()
			assert.Equal(t, tt.expectedValidation, result)
		})
	}
}

func TestPaymentMethod_GenerateValidationHash(t *testing.T) {
	paymentMethod := &PaymentMethod{}

	err := paymentMethod.GenerateValidationHash()

	assert.NoError(t, err)
	assert.NotEmpty(t, paymentMethod.ValidationHash)
	assert.Equal(t, 64, len(paymentMethod.ValidationHash)) // 32 bytes * 2 (hex encoding)
}

func TestPaymentMethod_UpdateLastUsed(t *testing.T) {
	paymentMethod := &PaymentMethod{
		SuccessfulUses: 5,
		FailedUses:     2,
	}

	// Test successful usage
	paymentMethod.UpdateLastUsed(true)
	assert.Equal(t, 6, paymentMethod.SuccessfulUses)
	assert.Equal(t, 2, paymentMethod.FailedUses)
	assert.NotNil(t, paymentMethod.LastUsedAt)

	// Test failed usage
	paymentMethod.UpdateLastUsed(false)
	assert.Equal(t, 6, paymentMethod.SuccessfulUses)
	assert.Equal(t, 3, paymentMethod.FailedUses)
}

// Helper functions
func intPtr(i int) *int {
	return &i
}

func timePtr(t time.Time) *time.Time {
	return &t
}
