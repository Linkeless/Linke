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

// NewMoney creates a new Money value with validation
func NewMoney(amount float64, currency Currency) (Money, error) {
	if amount < 0 {
		return Money{}, fmt.Errorf("money amount cannot be negative: %f", amount)
	}

	if math.IsNaN(amount) || math.IsInf(amount, 0) {
		return Money{}, fmt.Errorf("money amount must be a valid number: %f", amount)
	}

	if currency.IsEmpty() {
		return Money{}, fmt.Errorf("currency cannot be empty")
	}

	// Round to appropriate decimal places based on currency
	decimalPlaces := currency.GetDecimalPlaces()
	multiplier := math.Pow(10, float64(decimalPlaces))
	roundedAmount := math.Round(amount*multiplier) / multiplier

	return Money{
		amount:   roundedAmount,
		currency: currency,
	}, nil
}

// NewMoneyFromString creates Money from amount string and currency code
func NewMoneyFromString(amount string, currencyCode string) (Money, error) {
	currency, err := NewCurrency(currencyCode)
	if err != nil {
		return Money{}, fmt.Errorf("invalid currency: %w", err)
	}

	var amountFloat float64
	if _, err := fmt.Sscanf(amount, "%f", &amountFloat); err != nil {
		return Money{}, fmt.Errorf("invalid amount format: %w", err)
	}

	return NewMoney(amountFloat, currency)
}

// Zero creates a zero money value with the given currency
func Zero(currency Currency) Money {
	// This should never fail since we use 0 amount and validated currency
	money, _ := NewMoney(0, currency)
	return money
}

// NewZeroMoney creates a zero money value with specified currency (alias for Zero)
func NewZeroMoney(currency Currency) Money {
	return Zero(currency)
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
	decimalPlaces := m.currency.GetDecimalPlaces()
	tolerance := 1.0 / math.Pow(10, float64(decimalPlaces+2))
	return math.Abs(m.amount) < tolerance
}

// IsPositive checks if the money amount is positive
func (m Money) IsPositive() bool {
	decimalPlaces := m.currency.GetDecimalPlaces()
	tolerance := 1.0 / math.Pow(10, float64(decimalPlaces+2))
	return m.amount > tolerance
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
		return Money{}, fmt.Errorf("subtraction would result in negative amount: %.2f - %.2f = %.2f", 
			m.amount, other.amount, result)
	}

	return NewMoney(result, m.currency)
}

// Multiply multiplies money by a factor
func (m Money) Multiply(factor float64) (Money, error) {
	if factor < 0 {
		return Money{}, fmt.Errorf("multiplication factor cannot be negative: %f", factor)
	}

	if math.IsNaN(factor) || math.IsInf(factor, 0) {
		return Money{}, fmt.Errorf("multiplication factor must be a valid number: %f", factor)
	}

	return NewMoney(m.amount*factor, m.currency)
}

// MultiplyByPercentage multiplies the amount by a percentage (0-100)
func (m Money) MultiplyByPercentage(percentage float64) (Money, error) {
	if percentage < 0 || percentage > 100 {
		return Money{}, fmt.Errorf("percentage must be between 0 and 100: %.2f", percentage)
	}
	
	return m.Multiply(percentage / 100)
}

// Divide divides the amount by a divisor
func (m Money) Divide(divisor float64) (Money, error) {
	if divisor <= 0 {
		return Money{}, fmt.Errorf("cannot divide money by zero or negative divisor: %f", divisor)
	}

	if math.IsNaN(divisor) || math.IsInf(divisor, 0) {
		return Money{}, fmt.Errorf("divisor must be a valid number: %f", divisor)
	}

	return NewMoney(m.amount/divisor, m.currency)
}

// GreaterThan checks if this money is greater than other
func (m Money) GreaterThan(other Money) (bool, error) {
	if !m.currency.Equals(other.currency) {
		return false, fmt.Errorf("cannot compare money with different currencies: %s and %s", 
			m.currency.Code(), other.currency.Code())
	}

	return m.amount > other.amount, nil
}

// IsGreaterThan checks if this money is greater than other (alias for GreaterThan)
func (m Money) IsGreaterThan(other Money) (bool, error) {
	return m.GreaterThan(other)
}

// GreaterThanOrEqual checks if this money is greater than or equal to other
func (m Money) GreaterThanOrEqual(other Money) (bool, error) {
	if !m.currency.Equals(other.currency) {
		return false, fmt.Errorf("cannot compare money with different currencies: %s and %s", 
			m.currency.Code(), other.currency.Code())
	}

	decimalPlaces := m.currency.GetDecimalPlaces()
	tolerance := 1.0 / math.Pow(10, float64(decimalPlaces+2))
	return m.amount >= other.amount-tolerance, nil
}

// IsGreaterThanOrEqual checks if this money is greater than or equal to other (alias)
func (m Money) IsGreaterThanOrEqual(other Money) (bool, error) {
	return m.GreaterThanOrEqual(other)
}

// LessThan checks if this money is less than other
func (m Money) LessThan(other Money) (bool, error) {
	if !m.currency.Equals(other.currency) {
		return false, fmt.Errorf("cannot compare money with different currencies: %s and %s", 
			m.currency.Code(), other.currency.Code())
	}

	return m.amount < other.amount, nil
}

// IsLessThan checks if this money is less than other (alias for LessThan)
func (m Money) IsLessThan(other Money) (bool, error) {
	return m.LessThan(other)
}

// LessThanOrEqual checks if this money is less than or equal to other
func (m Money) LessThanOrEqual(other Money) (bool, error) {
	if !m.currency.Equals(other.currency) {
		return false, fmt.Errorf("cannot compare money with different currencies: %s and %s", 
			m.currency.Code(), other.currency.Code())
	}

	decimalPlaces := m.currency.GetDecimalPlaces()
	tolerance := 1.0 / math.Pow(10, float64(decimalPlaces+2))
	return m.amount <= other.amount+tolerance, nil
}

// IsLessThanOrEqual checks if this money is less than or equal to other (alias)
func (m Money) IsLessThanOrEqual(other Money) (bool, error) {
	return m.LessThanOrEqual(other)
}

// Min returns the smaller of two money values (must have same currency)
func (m Money) Min(other Money) (Money, error) {
	if !m.currency.Equals(other.currency) {
		return Money{}, fmt.Errorf("cannot compare different currencies: %s and %s", 
			m.currency.Code(), other.currency.Code())
	}
	
	if m.amount <= other.amount {
		return m, nil
	}
	return other, nil
}

// Max returns the larger of two money values (must have same currency)
func (m Money) Max(other Money) (Money, error) {
	if !m.currency.Equals(other.currency) {
		return Money{}, fmt.Errorf("cannot compare different currencies: %s and %s", 
			m.currency.Code(), other.currency.Code())
	}
	
	if m.amount >= other.amount {
		return m, nil
	}
	return other, nil
}

// Equals checks if two money values are equal
func (m Money) Equals(other Money) bool {
	if !m.currency.Equals(other.currency) {
		return false
	}
	
	decimalPlaces := m.currency.GetDecimalPlaces()
	tolerance := 1.0 / math.Pow(10, float64(decimalPlaces+2))
	return math.Abs(m.amount-other.amount) < tolerance
}

// String returns string representation of the money
func (m Money) String() string {
	decimalPlaces := m.currency.GetDecimalPlaces()
	format := fmt.Sprintf("%%.%df %%s", decimalPlaces)
	return fmt.Sprintf(format, m.amount, m.currency.Code())
}

// DisplayValue returns a formatted display value for the money
func (m Money) DisplayValue() string {
	decimalPlaces := m.currency.GetDecimalPlaces()
	symbol := m.currency.GetSymbol()
	
	// Special formatting for different currencies
	switch m.currency.Code() {
	case CurrencyJPY:
		return fmt.Sprintf("%s%.0f", symbol, m.amount)
	case CurrencyBTC, CurrencyETH:
		return fmt.Sprintf("%.8f %s", m.amount, symbol)
	default:
		format := fmt.Sprintf("%%s%%.%df", decimalPlaces)
		return fmt.Sprintf(format, symbol, m.amount)
	}
}

// ConvertTo converts to another currency using exchange rate
func (m Money) ConvertTo(targetCurrency Currency, exchangeRate float64) (Money, error) {
	if exchangeRate <= 0 {
		return Money{}, fmt.Errorf("exchange rate must be positive: %f", exchangeRate)
	}
	
	if math.IsNaN(exchangeRate) || math.IsInf(exchangeRate, 0) {
		return Money{}, fmt.Errorf("exchange rate must be a valid number: %f", exchangeRate)
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
			decimalPlaces := m.currency.GetDecimalPlaces()
			multiplier := math.Pow(10, float64(decimalPlaces))
			amount = math.Round((m.amount*ratio/totalRatio)*multiplier) / multiplier
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

// MarshalJSON implements json.Marshaler
func (m Money) MarshalJSON() ([]byte, error) {
	decimalPlaces := m.currency.GetDecimalPlaces()
	format := fmt.Sprintf(`{"amount":%%.%df,"currency":"%%s"}`, decimalPlaces)
	return []byte(fmt.Sprintf(format, m.amount, m.currency.Code())), nil
}

// UnmarshalJSON implements json.Unmarshaler
func (m *Money) UnmarshalJSON(data []byte) error {
	var temp struct {
		Amount   float64 `json:"amount"`
		Currency string  `json:"currency"`
	}
	
	n, err := fmt.Sscanf(string(data), `{"amount":%f,"currency":"%s"}`, &temp.Amount, &temp.Currency)
	if err != nil || n != 2 {
		return fmt.Errorf("invalid money JSON format: %w", err)
	}
	
	currency, err := NewCurrency(temp.Currency)
	if err != nil {
		return fmt.Errorf("invalid currency in JSON: %w", err)
	}
	
	money, err := NewMoney(temp.Amount, currency)
	if err != nil {
		return fmt.Errorf("invalid money value in JSON: %w", err)
	}
	
	*m = money
	return nil
}