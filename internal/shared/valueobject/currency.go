package valueobject

import (
	"fmt"
	"strings"
)

// Currency represents a monetary currency code
type Currency struct {
	code string
}

// Currency constants - Fiat currencies
const (
	CurrencyUSD  = "USD"
	CurrencyCNY  = "CNY"
	CurrencyEUR  = "EUR"
	CurrencyGBP  = "GBP"
	CurrencyJPY  = "JPY"
)

// Currency constants - Cryptocurrencies
const (
	CurrencyUSDT = "USDT"
	CurrencyBTC  = "BTC"
	CurrencyETH  = "ETH"
)

// Predefined currency instances
var (
	USD  = Currency{code: CurrencyUSD}
	CNY  = Currency{code: CurrencyCNY}
	EUR  = Currency{code: CurrencyEUR}
	GBP  = Currency{code: CurrencyGBP}
	JPY  = Currency{code: CurrencyJPY}
	USDT = Currency{code: CurrencyUSDT}
	BTC  = Currency{code: CurrencyBTC}
	ETH  = Currency{code: CurrencyETH}
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
	code = strings.ToUpper(strings.TrimSpace(code))
	
	if code == "" {
		return Currency{}, fmt.Errorf("currency code cannot be empty")
	}

	if len(code) < 3 || len(code) > 4 {
		return Currency{}, fmt.Errorf("currency code must be 3-4 characters long")
	}

	// Basic validation for known currencies
	if !validCurrencies[code] {
		return Currency{}, fmt.Errorf("unsupported currency code: %s", code)
	}

	return Currency{code: code}, nil
}

// NewUSDCurrency creates a USD currency
func NewUSDCurrency() Currency {
	return USD
}

// NewCNYCurrency creates a CNY currency
func NewCNYCurrency() Currency {
	return CNY
}

// NewEURCurrency creates a EUR currency
func NewEURCurrency() Currency {
	return EUR
}

// NewGBPCurrency creates a GBP currency
func NewGBPCurrency() Currency {
	return GBP
}

// NewJPYCurrency creates a JPY currency
func NewJPYCurrency() Currency {
	return JPY
}

// NewUSDTCurrency creates a USDT currency
func NewUSDTCurrency() Currency {
	return USDT
}

// NewBTCCurrency creates a BTC currency
func NewBTCCurrency() Currency {
	return BTC
}

// NewETHCurrency creates an ETH currency
func NewETHCurrency() Currency {
	return ETH
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

// IsValid checks if the currency code is valid
func (c Currency) IsValid() bool {
	return validCurrencies[c.code]
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

// Symbol returns the currency symbol (alias for GetSymbol for compatibility)
func (c Currency) Symbol() string {
	return c.GetSymbol()
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

// MarshalJSON implements json.Marshaler
func (c Currency) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, c.code)), nil
}

// UnmarshalJSON implements json.Unmarshaler
func (c *Currency) UnmarshalJSON(data []byte) error {
	str := string(data)
	if str == "null" {
		c.code = ""
		return nil
	}
	
	// Remove quotes if present
	if len(str) >= 2 && str[0] == '"' && str[len(str)-1] == '"' {
		str = str[1 : len(str)-1]
	}
	
	currency, err := NewCurrency(str)
	if err != nil {
		return err
	}
	
	*c = currency
	return nil
}