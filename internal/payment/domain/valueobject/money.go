package valueobject

import (
	"fmt"
	"math"
)

// Money represents a monetary amount with currency
type Money struct {
	amount   float64
	currency Currency
}

// NewMoney creates a new Money value object
func NewMoney(amount float64, currency Currency) (Money, error) {
	if amount < 0 {
		return Money{}, fmt.Errorf("money amount cannot be negative: %f", amount)
	}
	
	if math.IsNaN(amount) || math.IsInf(amount, 0) {
		return Money{}, fmt.Errorf("money amount must be a valid number: %f", amount)
	}
	
	// Round to 2 decimal places to avoid floating point precision issues
	roundedAmount := math.Round(amount*100) / 100
	
	return Money{
		amount:   roundedAmount,
		currency: currency,
	}, nil
}

// NewZeroMoney creates a zero money value with specified currency
func NewZeroMoney(currency Currency) Money {
	// This should never fail since we use 0 amount
	money, _ := NewMoney(0, currency)
	return money
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
	return fmt.Sprintf("%.2f %s", m.amount, m.currency.String())
}

// Equals checks if two Money values are equal
func (m Money) Equals(other Money) bool {
	return math.Abs(m.amount-other.amount) < 0.01 && m.currency.Equals(other.currency)
}

// IsZero checks if the amount is zero
func (m Money) IsZero() bool {
	return math.Abs(m.amount) < 0.01
}

// IsPositive checks if the amount is positive
func (m Money) IsPositive() bool {
	return m.amount > 0.01
}

// Add adds another Money value (must be same currency)
func (m Money) Add(other Money) (Money, error) {
	if !m.currency.Equals(other.currency) {
		return Money{}, fmt.Errorf("cannot add different currencies: %s and %s", 
			m.currency.String(), other.currency.String())
	}
	
	return NewMoney(m.amount+other.amount, m.currency)
}

// Subtract subtracts another Money value (must be same currency)
func (m Money) Subtract(other Money) (Money, error) {
	if !m.currency.Equals(other.currency) {
		return Money{}, fmt.Errorf("cannot subtract different currencies: %s and %s", 
			m.currency.String(), other.currency.String())
	}
	
	return NewMoney(m.amount-other.amount, m.currency)
}

// Multiply multiplies the amount by a factor
func (m Money) Multiply(factor float64) (Money, error) {
	if factor < 0 {
		return Money{}, fmt.Errorf("cannot multiply money by negative factor: %f", factor)
	}
	
	return NewMoney(m.amount*factor, m.currency)
}

// Divide divides the amount by a divisor
func (m Money) Divide(divisor float64) (Money, error) {
	if divisor <= 0 {
		return Money{}, fmt.Errorf("cannot divide money by zero or negative divisor: %f", divisor)
	}
	
	return NewMoney(m.amount/divisor, m.currency)
}

// IsGreaterThan checks if this money is greater than other
func (m Money) IsGreaterThan(other Money) (bool, error) {
	if !m.currency.Equals(other.currency) {
		return false, fmt.Errorf("cannot compare different currencies: %s and %s", 
			m.currency.String(), other.currency.String())
	}
	
	return m.amount > other.amount, nil
}

// IsLessThan checks if this money is less than other
func (m Money) IsLessThan(other Money) (bool, error) {
	if !m.currency.Equals(other.currency) {
		return false, fmt.Errorf("cannot compare different currencies: %s and %s", 
			m.currency.String(), other.currency.String())
	}
	
	return m.amount < other.amount, nil
}

// IsGreaterThanOrEqual checks if this money is greater than or equal to other
func (m Money) IsGreaterThanOrEqual(other Money) (bool, error) {
	if !m.currency.Equals(other.currency) {
		return false, fmt.Errorf("cannot compare different currencies: %s and %s", 
			m.currency.String(), other.currency.String())
	}
	
	return m.amount >= other.amount-0.01, nil
}

// IsLessThanOrEqual checks if this money is less than or equal to other
func (m Money) IsLessThanOrEqual(other Money) (bool, error) {
	if !m.currency.Equals(other.currency) {
		return false, fmt.Errorf("cannot compare different currencies: %s and %s", 
			m.currency.String(), other.currency.String())
	}
	
	return m.amount <= other.amount+0.01, nil
}

// ConvertTo converts to another currency using exchange rate
func (m Money) ConvertTo(targetCurrency Currency, exchangeRate float64) (Money, error) {
	if exchangeRate <= 0 {
		return Money{}, fmt.Errorf("exchange rate must be positive: %f", exchangeRate)
	}
	
	convertedAmount := m.amount * exchangeRate
	return NewMoney(convertedAmount, targetCurrency)
}

// AllocateProportionally allocates this money amount proportionally based on ratios
func (m Money) AllocateProportionally(ratios []float64) ([]Money, error) {
	if len(ratios) == 0 {
		return nil, fmt.Errorf("ratios cannot be empty")
	}
	
	// Calculate total ratio
	totalRatio := 0.0
	for _, ratio := range ratios {
		if ratio < 0 {
			return nil, fmt.Errorf("ratio cannot be negative: %f", ratio)
		}
		totalRatio += ratio
	}
	
	if totalRatio == 0 {
		return nil, fmt.Errorf("total ratio cannot be zero")
	}
	
	// Allocate amounts
	var allocated []Money
	remaining := m.amount
	
	for i, ratio := range ratios {
		var amount float64
		if i == len(ratios)-1 {
			// Last allocation gets the remaining amount to avoid rounding errors
			amount = remaining
		} else {
			amount = math.Round((m.amount*ratio/totalRatio)*100) / 100
			remaining -= amount
		}
		
		money, err := NewMoney(amount, m.currency)
		if err != nil {
			return nil, fmt.Errorf("failed to create allocated money: %w", err)
		}
		
		allocated = append(allocated, money)
	}
	
	return allocated, nil
}