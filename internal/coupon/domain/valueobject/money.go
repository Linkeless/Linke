package valueobject

import (
	"fmt"
	"strings"
)

// Money represents a monetary amount with currency
type Money struct {
	amount   float64
	currency Currency
}

// Currency represents a currency code
type Currency string

const (
	CurrencyUSD Currency = "USD"
	CurrencyEUR Currency = "EUR"
	CurrencyGBP Currency = "GBP"
	CurrencyJPY Currency = "JPY"
	CurrencyCNY Currency = "CNY"
)

// ValidCurrencies contains all valid currencies
var ValidCurrencies = []Currency{
	CurrencyUSD,
	CurrencyEUR,
	CurrencyGBP,
	CurrencyJPY,
	CurrencyCNY,
}

// NewMoney creates a new money value object
func NewMoney(amount float64, currency string) (Money, error) {
	if amount < 0 {
		return Money{}, fmt.Errorf("amount cannot be negative: %.2f", amount)
	}
	
	curr, err := NewCurrency(currency)
	if err != nil {
		return Money{}, err
	}
	
	return Money{
		amount:   amount,
		currency: curr,
	}, nil
}

// MustNewMoney creates a new money value object and panics on error
func MustNewMoney(amount float64, currency string) Money {
	m, err := NewMoney(amount, currency)
	if err != nil {
		panic(err)
	}
	return m
}

// NewCurrency creates and validates a currency
func NewCurrency(code string) (Currency, error) {
	normalized := Currency(strings.ToUpper(strings.TrimSpace(code)))
	
	for _, validCurrency := range ValidCurrencies {
		if normalized == validCurrency {
			return normalized, nil
		}
	}
	
	return "", fmt.Errorf("invalid currency code: %s", code)
}

// Amount returns the monetary amount
func (m Money) Amount() float64 {
	return m.amount
}

// Currency returns the currency
func (m Money) Currency() Currency {
	return m.currency
}

// String returns string representation
func (m Money) String() string {
	return fmt.Sprintf("%.2f %s", m.amount, m.currency)
}

// IsZero checks if the amount is zero
func (m Money) IsZero() bool {
	return m.amount == 0
}

// IsPositive checks if the amount is positive
func (m Money) IsPositive() bool {
	return m.amount > 0
}

// Equals checks if two money values are equal
func (m Money) Equals(other Money) bool {
	return m.amount == other.amount && m.currency == other.currency
}

// Add adds two money values (must have same currency)
func (m Money) Add(other Money) (Money, error) {
	if m.currency != other.currency {
		return Money{}, fmt.Errorf("cannot add different currencies: %s and %s", m.currency, other.currency)
	}
	
	return Money{
		amount:   m.amount + other.amount,
		currency: m.currency,
	}, nil
}

// Subtract subtracts two money values (must have same currency)
func (m Money) Subtract(other Money) (Money, error) {
	if m.currency != other.currency {
		return Money{}, fmt.Errorf("cannot subtract different currencies: %s and %s", m.currency, other.currency)
	}
	
	result := m.amount - other.amount
	if result < 0 {
		return Money{}, fmt.Errorf("result cannot be negative: %.2f - %.2f = %.2f", m.amount, other.amount, result)
	}
	
	return Money{
		amount:   result,
		currency: m.currency,
	}, nil
}

// MultiplyByPercentage multiplies the amount by a percentage (0-100)
func (m Money) MultiplyByPercentage(percentage float64) (Money, error) {
	if percentage < 0 || percentage > 100 {
		return Money{}, fmt.Errorf("percentage must be between 0 and 100: %.2f", percentage)
	}
	
	return Money{
		amount:   m.amount * (percentage / 100),
		currency: m.currency,
	}, nil
}

// Min returns the smaller of two money values (must have same currency)
func (m Money) Min(other Money) (Money, error) {
	if m.currency != other.currency {
		return Money{}, fmt.Errorf("cannot compare different currencies: %s and %s", m.currency, other.currency)
	}
	
	if m.amount <= other.amount {
		return m, nil
	}
	return other, nil
}

// GreaterThanOrEqual checks if this money is >= other money
func (m Money) GreaterThanOrEqual(other Money) (bool, error) {
	if m.currency != other.currency {
		return false, fmt.Errorf("cannot compare different currencies: %s and %s", m.currency, other.currency)
	}
	
	return m.amount >= other.amount, nil
}

// MarshalJSON implements json.Marshaler
func (m Money) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`{"amount":%.2f,"currency":"%s"}`, m.amount, m.currency)), nil
}

// String returns string representation for Currency
func (c Currency) String() string {
	return string(c)
}

// IsValid checks if the currency is valid
func (c Currency) IsValid() bool {
	for _, validCurrency := range ValidCurrencies {
		if c == validCurrency {
			return true
		}
	}
	return false
}