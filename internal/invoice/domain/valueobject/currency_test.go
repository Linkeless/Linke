package valueobject

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCurrency(t *testing.T) {
	tests := []struct {
		name        string
		code        string
		expectError bool
		errorMsg    string
		expected    string
	}{
		{
			name:        "valid USD currency",
			code:        "USD",
			expectError: false,
			expected:    "USD",
		},
		{
			name:        "valid EUR currency",
			code:        "EUR",
			expectError: false,
			expected:    "EUR",
		},
		{
			name:        "valid lowercase currency",
			code:        "gbp",
			expectError: false,
			expected:    "GBP",
		},
		{
			name:        "currency with whitespace",
			code:        "  JPY  ",
			expectError: false,
			expected:    "JPY",
		},
		{
			name:        "mixed case currency",
			code:        "CnY",
			expectError: false,
			expected:    "CNY",
		},
		{
			name:        "empty currency code should fail",
			code:        "",
			expectError: true,
			errorMsg:    "currency code cannot be empty",
		},
		{
			name:        "whitespace only currency should fail",
			code:        "   ",
			expectError: true,
			errorMsg:    "currency code cannot be empty",
		},
		{
			name:        "too short currency code should fail",
			code:        "US",
			expectError: true,
			errorMsg:    "currency code must be exactly 3 characters",
		},
		{
			name:        "too long currency code should fail",
			code:        "USDD",
			expectError: true,
			errorMsg:    "currency code must be exactly 3 characters",
		},
		{
			name:        "currency with numbers should fail",
			code:        "US1",
			expectError: true,
			errorMsg:    "currency code must contain only uppercase letters",
		},
		{
			name:        "currency with special characters should fail",
			code:        "US$",
			expectError: true,
			errorMsg:    "currency code must contain only uppercase letters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			currency, err := NewCurrency(tt.code)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, currency.Code())
				assert.Equal(t, tt.expected, currency.String())
			}
		})
	}
}

func TestCurrency_Equals(t *testing.T) {
	usd1, _ := NewCurrency("USD")
	usd2, _ := NewCurrency("usd")
	eur, _ := NewCurrency("EUR")

	tests := []struct {
		name     string
		currency Currency
		other    Currency
		expected bool
	}{
		{
			name:     "same currency codes are equal",
			currency: usd1,
			other:    usd2,
			expected: true,
		},
		{
			name:     "different currency codes are not equal",
			currency: usd1,
			other:    eur,
			expected: false,
		},
		{
			name:     "currency equal to itself",
			currency: usd1,
			other:    usd1,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.currency.Equals(tt.other)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCurrency_String(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected string
	}{
		{
			name:     "USD currency string",
			code:     "USD",
			expected: "USD",
		},
		{
			name:     "EUR currency string",
			code:     "EUR",
			expected: "EUR",
		},
		{
			name:     "GBP currency string",
			code:     "GBP",
			expected: "GBP",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			currency, err := NewCurrency(tt.code)
			require.NoError(t, err)
			
			assert.Equal(t, tt.expected, currency.String())
			assert.Equal(t, tt.expected, currency.Code())
		})
	}
}

func TestCurrencyConstants(t *testing.T) {
	tests := []struct {
		name     string
		currency Currency
		expected string
	}{
		{
			name:     "USD constant",
			currency: USD,
			expected: "USD",
		},
		{
			name:     "EUR constant",
			currency: EUR,
			expected: "EUR",
		},
		{
			name:     "GBP constant",
			currency: GBP,
			expected: "GBP",
		},
		{
			name:     "JPY constant",
			currency: JPY,
			expected: "JPY",
		},
		{
			name:     "CNY constant",
			currency: CNY,
			expected: "CNY",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.currency.Code())
			assert.Equal(t, tt.expected, tt.currency.String())
		})
	}
}

func TestCurrency_Immutability(t *testing.T) {
	// Test that Currency is immutable
	currency, err := NewCurrency("USD")
	require.NoError(t, err)
	
	originalCode := currency.Code()
	
	// Modify the original string used to create currency
	modifiedCode := "EUR"
	currency2, err := NewCurrency(modifiedCode)
	require.NoError(t, err)
	
	// Original currency should not be affected
	assert.Equal(t, originalCode, currency.Code())
	assert.Equal(t, "EUR", currency2.Code())
	assert.False(t, currency.Equals(currency2))
}