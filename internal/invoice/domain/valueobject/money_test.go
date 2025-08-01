package valueobject

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMoney(t *testing.T) {
	tests := []struct {
		name        string
		amount      float64
		currency    Currency
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid positive amount",
			amount:      100.50,
			currency:    Currency{code: "USD"},
			expectError: false,
		},
		{
			name:        "valid zero amount",
			amount:      0,
			currency:    Currency{code: "USD"},
			expectError: false,
		},
		{
			name:        "negative amount should fail",
			amount:      -10.50,
			currency:    Currency{code: "USD"},
			expectError: true,
			errorMsg:    "amount cannot be negative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			money, err := NewMoney(tt.amount, tt.currency)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.amount, money.Amount())
				assert.Equal(t, tt.currency.Code(), money.Currency().Code())
			}
		})
	}
}

func TestMoney_Add(t *testing.T) {
	usd, _ := NewCurrency("USD")
	eur, _ := NewCurrency("EUR")

	tests := []struct {
		name        string
		money1      Money
		money2      Money
		expected    float64
		expectError bool
		errorMsg    string
	}{
		{
			name:        "add same currency",
			money1:      mustCreateMoney(100.50, usd),
			money2:      mustCreateMoney(50.25, usd),
			expected:    150.75,
			expectError: false,
		},
		{
			name:        "add zero",
			money1:      mustCreateMoney(100.50, usd),
			money2:      mustCreateMoney(0, usd),
			expected:    100.50,
			expectError: false,
		},
		{
			name:        "different currencies should fail",
			money1:      mustCreateMoney(100.50, usd),
			money2:      mustCreateMoney(50.25, eur),
			expectError: true,
			errorMsg:    "cannot add money with different currencies",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.money1.Add(tt.money2)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result.Amount())
				assert.Equal(t, tt.money1.Currency().Code(), result.Currency().Code())
			}
		})
	}
}

func TestMoney_Subtract(t *testing.T) {
	usd, _ := NewCurrency("USD")

	tests := []struct {
		name        string
		money1      Money
		money2      Money
		expected    float64
		expectError bool
		errorMsg    string
	}{
		{
			name:        "subtract smaller amount",
			money1:      mustCreateMoney(100.50, usd),
			money2:      mustCreateMoney(50.25, usd),
			expected:    50.25,
			expectError: false,
		},
		{
			name:        "subtract equal amount",
			money1:      mustCreateMoney(100.50, usd),
			money2:      mustCreateMoney(100.50, usd),
			expected:    0,
			expectError: false,
		},
		{
			name:        "subtract larger amount should fail",
			money1:      mustCreateMoney(50.25, usd),
			money2:      mustCreateMoney(100.50, usd),
			expectError: true,
			errorMsg:    "subtraction would result in negative amount",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.money1.Subtract(tt.money2)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result.Amount())
			}
		})
	}
}

func TestMoney_IsZero(t *testing.T) {
	usd, _ := NewCurrency("USD")

	tests := []struct {
		name     string
		money    Money
		expected bool
	}{
		{
			name:     "zero money",
			money:    mustCreateMoney(0, usd),
			expected: true,
		},
		{
			name:     "non-zero money",
			money:    mustCreateMoney(100.50, usd),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.money.IsZero())
		})
	}
}

func TestMoney_DisplayValue(t *testing.T) {
	tests := []struct {
		name     string
		amount   float64
		currency string
		expected string
	}{
		{
			name:     "USD currency",
			amount:   100.50,
			currency: "USD",
			expected: "$100.50",
		},
		{
			name:     "EUR currency",
			amount:   75.25,
			currency: "EUR",
			expected: "€75.25",
		},
		{
			name:     "GBP currency",
			amount:   50.00,
			currency: "GBP",
			expected: "£50.00",
		},
		{
			name:     "unknown currency",
			amount:   100.00,
			currency: "XYZ",
			expected: "100.00 XYZ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			currency, err := NewCurrency(tt.currency)
			require.NoError(t, err)
			
			money, err := NewMoney(tt.amount, currency)
			require.NoError(t, err)
			
			assert.Equal(t, tt.expected, money.DisplayValue())
		})
	}
}

// Helper function for creating money without error handling in tests
func mustCreateMoney(amount float64, currency Currency) Money {
	money, err := NewMoney(amount, currency)
	if err != nil {
		panic(err)
	}
	return money
}