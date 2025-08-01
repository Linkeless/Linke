package valueobject

import (
	sharedvo "linke/internal/shared/valueobject"
)

// ConvertToSharedMoney converts payment domain Money to shared Money
func ConvertToSharedMoney(money Money) (sharedvo.Money, error) {
	// Convert currency first
	sharedCurrency, err := ConvertToSharedCurrency(money.Currency())
	if err != nil {
		return sharedvo.Money{}, err
	}
	
	return sharedvo.NewMoney(money.Amount(), sharedCurrency)
}

// ConvertFromSharedMoney converts shared Money to payment domain Money
func ConvertFromSharedMoney(money sharedvo.Money) (Money, error) {
	// Convert currency first
	paymentCurrency, err := ConvertFromSharedCurrency(money.Currency())
	if err != nil {
		return Money{}, err
	}
	
	return NewMoney(money.Amount(), paymentCurrency)
}

// ConvertToSharedCurrency converts payment domain Currency to shared Currency
func ConvertToSharedCurrency(currency Currency) (sharedvo.Currency, error) {
	return sharedvo.NewCurrency(currency.Code())
}

// ConvertFromSharedCurrency converts shared Currency to payment domain Currency
func ConvertFromSharedCurrency(currency sharedvo.Currency) (Currency, error) {
	return NewCurrency(currency.Code())
}