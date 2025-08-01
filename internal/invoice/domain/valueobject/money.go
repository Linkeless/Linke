package valueobject

import (
	"fmt"
	"math"
	"strings"
)

// Money represents a monetary amount with currency
type Money struct {
	amount   float64
	currency Currency
}

// Currency represents a currency code
type Currency struct {
	code string
}

// Common currencies
var (
	USD = Currency{code: "USD"}
	EUR = Currency{code: "EUR"}
	GBP = Currency{code: "GBP"}
	JPY = Currency{code: "JPY"}
	CNY = Currency{code: "CNY"}
)

// NewCurrency creates a new Currency
func NewCurrency(code string) (Currency, error) {
	code = strings.ToUpper(strings.TrimSpace(code))
	
	if code == "" {
		return Currency{}, fmt.Errorf("currency code cannot be empty")
	}

	if len(code) != 3 {
		return Currency{}, fmt.Errorf("currency code must be exactly 3 characters")
	}

	// Basic validation for alphabetic characters
	for _, char := range code {
		if char < 'A' || char > 'Z' {
			return Currency{}, fmt.Errorf("currency code must contain only uppercase letters")
		}
	}

	return Currency{code: code}, nil
}

// Code returns the currency code
func (c Currency) Code() string {
	return c.code
}

// String returns string representation of the currency
func (c Currency) String() string {
	return c.code
}

// Equals checks if two currencies are equal
func (c Currency) Equals(other Currency) bool {
	return c.code == other.code
}

// NewMoney creates a new Money value
func NewMoney(amount float64, currency Currency) (Money, error) {
	if amount < 0 {
		return Money{}, fmt.Errorf("money amount cannot be negative")
	}

	if math.IsNaN(amount) || math.IsInf(amount, 0) {
		return Money{}, fmt.Errorf("money amount must be a valid number")
	}

	// Round to 2 decimal places for monetary precision
	roundedAmount := math.Round(amount*100) / 100

	return Money{
		amount:   roundedAmount,
		currency: currency,
	}, nil
}

// Zero creates a zero money value with the given currency
func Zero(currency Currency) Money {
	return Money{
		amount:   0,
		currency: currency,
	}
}

// Amount returns the monetary amount
func (m Money) Amount() float64 {
	return m.amount
}

// Currency returns the currency
func (m Money) Currency() Currency {
	return m.currency
}

// IsZero checks if the money amount is zero
func (m Money) IsZero() bool {
	return m.amount == 0
}

// IsPositive checks if the money amount is positive
func (m Money) IsPositive() bool {
	return m.amount > 0
}

// Add adds two money values (must have same currency)
func (m Money) Add(other Money) (Money, error) {
	if !m.currency.Equals(other.currency) {
		return Money{}, fmt.Errorf("cannot add money with different currencies: %s and %s", 
			m.currency.Code(), other.currency.Code())
	}

	return NewMoney(m.amount+other.amount, m.currency)
}

// Subtract subtracts two money values (must have same currency)
func (m Money) Subtract(other Money) (Money, error) {
	if !m.currency.Equals(other.currency) {
		return Money{}, fmt.Errorf("cannot subtract money with different currencies: %s and %s", 
			m.currency.Code(), other.currency.Code())
	}

	result := m.amount - other.amount
	if result < 0 {
		return Money{}, fmt.Errorf("subtraction would result in negative amount")
	}

	return NewMoney(result, m.currency)
}

// Multiply multiplies money by a factor
func (m Money) Multiply(factor float64) (Money, error) {
	if factor < 0 {
		return Money{}, fmt.Errorf("multiplication factor cannot be negative")
	}

	return NewMoney(m.amount*factor, m.currency)
}

// GreaterThan checks if this money is greater than other
func (m Money) GreaterThan(other Money) (bool, error) {
	if !m.currency.Equals(other.currency) {
		return false, fmt.Errorf("cannot compare money with different currencies: %s and %s", 
			m.currency.Code(), other.currency.Code())
	}

	return m.amount > other.amount, nil
}

// GreaterThanOrEqual checks if this money is greater than or equal to other
func (m Money) GreaterThanOrEqual(other Money) (bool, error) {
	if !m.currency.Equals(other.currency) {
		return false, fmt.Errorf("cannot compare money with different currencies: %s and %s", 
			m.currency.Code(), other.currency.Code())
	}

	return m.amount >= other.amount, nil
}

// Equals checks if two money values are equal
func (m Money) Equals(other Money) bool {
	return m.currency.Equals(other.currency) && m.amount == other.amount
}

// String returns string representation of the money
func (m Money) String() string {
	return fmt.Sprintf("%.2f %s", m.amount, m.currency.Code())
}

// DisplayValue returns a formatted display value for the money
func (m Money) DisplayValue() string {
	// Format based on currency for better user experience
	switch m.currency.Code() {
	case "USD":
		return fmt.Sprintf("$%.2f", m.amount)
	case "EUR":
		return fmt.Sprintf("€%.2f", m.amount)
	case "GBP":
		return fmt.Sprintf("£%.2f", m.amount)
	case "JPY":
		return fmt.Sprintf("¥%.0f", m.amount) // JPY typically doesn't use decimals
	case "CNY":
		return fmt.Sprintf("¥%.2f", m.amount)
	default:
		return fmt.Sprintf("%.2f %s", m.amount, m.currency.Code())
	}
}