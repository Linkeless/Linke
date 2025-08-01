package valueobject

import (
	sharedvo "linke/internal/shared/valueobject"
)

// ConvertToSharedMoney converts coupon domain Money to shared Money
func ConvertToSharedMoney(money Money) (sharedvo.Money, error) {
	// Convert currency first
	sharedCurrency, err := sharedvo.NewCurrency(string(money.Currency()))
	if err != nil {
		return sharedvo.Money{}, err
	}
	
	return sharedvo.NewMoney(money.Amount(), sharedCurrency)
}

// ConvertFromSharedMoney converts shared Money to coupon domain Money  
func ConvertFromSharedMoney(money sharedvo.Money) (Money, error) {
	return NewMoney(money.Amount(), money.Currency().Code())
}

// ConvertToSharedCurrency converts coupon domain Currency to shared Currency
func ConvertToSharedCurrency(currency Currency) (sharedvo.Currency, error) {
	return sharedvo.NewCurrency(string(currency))
}

// ConvertFromSharedCurrency converts shared Currency to coupon domain Currency
func ConvertFromSharedCurrency(currency sharedvo.Currency) (Currency, error) {
	return NewCurrency(currency.Code())
}