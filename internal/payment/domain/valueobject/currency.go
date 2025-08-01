package valueobject

import "fmt"

// Currency represents a monetary currency
type Currency struct {
	code string
}

// Currency constants
const (
	CurrencyUSD  = "USD"
	CurrencyCNY  = "CNY"
	CurrencyEUR  = "EUR"
	CurrencyGBP  = "GBP"
	CurrencyJPY  = "JPY"
	CurrencyUSDT = "USDT"
	CurrencyBTC  = "BTC"
	CurrencyETH  = "ETH"
)

var validCurrencies = map[string]bool{
	CurrencyUSD:  true,
	CurrencyCNY:  true,
	CurrencyEUR:  true,
	CurrencyGBP:  true,
	CurrencyJPY:  true,
	CurrencyUSDT: true,
	CurrencyBTC:  true,
	CurrencyETH:  true,
}

// NewCurrency creates a new Currency with validation
func NewCurrency(code string) (Currency, error) {
	if code == "" {
		return Currency{}, fmt.Errorf("currency code cannot be empty")
	}
	
	if !validCurrencies[code] {
		return Currency{}, fmt.Errorf("invalid currency code: %s", code)
	}
	
	return Currency{code: code}, nil
}

// NewUSDCurrency creates a USD currency
func NewUSDCurrency() Currency {
	currency, _ := NewCurrency(CurrencyUSD)
	return currency
}

// NewCNYCurrency creates a CNY currency
func NewCNYCurrency() Currency {
	currency, _ := NewCurrency(CurrencyCNY)
	return currency
}

// NewEURCurrency creates a EUR currency
func NewEURCurrency() Currency {
	currency, _ := NewCurrency(CurrencyEUR)
	return currency
}

// NewUSDTCurrency creates a USDT currency
func NewUSDTCurrency() Currency {
	currency, _ := NewCurrency(CurrencyUSDT)
	return currency
}

// Code returns the currency code
func (c Currency) Code() string {
	return c.code
}

// String returns string representation
func (c Currency) String() string {
	return c.code
}

// Equals checks if two currencies are equal
func (c Currency) Equals(other Currency) bool {
	return c.code == other.code
}

// IsEmpty checks if the currency is empty
func (c Currency) IsEmpty() bool {
	return c.code == ""
}

// IsFiat checks if the currency is a fiat currency
func (c Currency) IsFiat() bool {
	fiatCurrencies := map[string]bool{
		CurrencyUSD: true,
		CurrencyCNY: true,
		CurrencyEUR: true,
		CurrencyGBP: true,
		CurrencyJPY: true,
	}
	
	return fiatCurrencies[c.code]
}

// IsCrypto checks if the currency is a cryptocurrency
func (c Currency) IsCrypto() bool {
	cryptoCurrencies := map[string]bool{
		CurrencyUSDT: true,
		CurrencyBTC:  true,
		CurrencyETH:  true,
	}
	
	return cryptoCurrencies[c.code]
}

// GetSymbol returns the currency symbol
func (c Currency) GetSymbol() string {
	symbols := map[string]string{
		CurrencyUSD:  "$",
		CurrencyCNY:  "¥",
		CurrencyEUR:  "€",
		CurrencyGBP:  "£",
		CurrencyJPY:  "¥",
		CurrencyUSDT: "USDT",
		CurrencyBTC:  "₿",
		CurrencyETH:  "Ξ",
	}
	
	if symbol, exists := symbols[c.code]; exists {
		return symbol
	}
	
	return c.code
}

// GetDecimalPlaces returns the number of decimal places for the currency
func (c Currency) GetDecimalPlaces() int {
	decimalPlaces := map[string]int{
		CurrencyUSD:  2,
		CurrencyCNY:  2,
		CurrencyEUR:  2,
		CurrencyGBP:  2,
		CurrencyJPY:  0, // Japanese Yen has no fractional unit
		CurrencyUSDT: 2,
		CurrencyBTC:  8,
		CurrencyETH:  8,
	}
	
	if places, exists := decimalPlaces[c.code]; exists {
		return places
	}
	
	return 2 // Default to 2 decimal places
}