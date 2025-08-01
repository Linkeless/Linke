package valueobject

import "errors"

type Currency string

const (
	USD Currency = "USD"
	EUR Currency = "EUR"
	CNY Currency = "CNY"
	JPY Currency = "JPY"
)

func NewCurrency(value string) (*Currency, error) {
	currency := Currency(value)
	if !currency.IsValid() {
		return nil, errors.New("invalid currency code")
	}
	return &currency, nil
}

func (c Currency) IsValid() bool {
	switch c {
	case USD, EUR, CNY, JPY:
		return true
	default:
		return false
	}
}

func (c Currency) String() string {
	return string(c)
}

func (c Currency) Symbol() string {
	switch c {
	case USD:
		return "$"
	case EUR:
		return "€"
	case CNY:
		return "¥"
	case JPY:
		return "¥"
	default:
		return string(c)
	}
}