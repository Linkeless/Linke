package valueobject

import (
	"errors"
	"fmt"
)

type Price struct {
	amount   float64
	currency Currency
}

func NewPrice(amount float64, currency Currency) (*Price, error) {
	if amount < 0 {
		return nil, errors.New("price amount cannot be negative")
	}
	if !currency.IsValid() {
		return nil, errors.New("invalid currency")
	}
	return &Price{amount: amount, currency: currency}, nil
}

func (p Price) Amount() float64 {
	return p.amount
}

func (p Price) Currency() Currency {
	return p.currency
}

func (p Price) String() string {
	return fmt.Sprintf("%.2f %s", p.amount, p.currency.Symbol())
}

func (p Price) Equals(other Price) bool {
	return p.amount == other.amount && p.currency == other.currency
}

func (p Price) IsZero() bool {
	return p.amount == 0
}

func (p Price) Add(other Price) (*Price, error) {
	if p.currency != other.currency {
		return nil, errors.New("cannot add prices with different currencies")
	}
	return NewPrice(p.amount+other.amount, p.currency)
}

func (p Price) Multiply(factor float64) (*Price, error) {
	if factor < 0 {
		return nil, errors.New("multiplication factor cannot be negative")
	}
	return NewPrice(p.amount*factor, p.currency)
}