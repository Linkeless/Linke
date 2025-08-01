package valueobject

import (
	sharedvo "linke/internal/shared/valueobject"
)

// UserID conversions

// ConvertToSharedUserID converts subscription domain UserID to shared UserID
func ConvertToSharedUserID(id *UserID) (sharedvo.UserID, error) {
	if id == nil {
		return sharedvo.UserID{}, nil // Allow nil for compatibility
	}
	// Subscription domain UserID is uint-based, convert to shared
	return sharedvo.NewUserIDFromUint(id.Value())
}

// ConvertFromSharedUserID converts shared UserID to subscription domain UserID
func ConvertFromSharedUserID(id sharedvo.UserID) (*UserID, error) {
	if id.IsZero() {
		return nil, nil // Return nil for zero values
	}
	// Convert shared UserID to uint for subscription domain compatibility
	return NewUserID(id.ToUint())
}

// Currency conversions

// ConvertToSharedCurrency converts subscription domain Currency to shared Currency
func ConvertToSharedCurrency(currency *Currency) (sharedvo.Currency, error) {
	if currency == nil {
		return sharedvo.Currency{}, nil // Allow nil for compatibility
	}
	// Subscription domain Currency is string-based, convert to shared
	return sharedvo.NewCurrency(currency.String())
}

// ConvertFromSharedCurrency converts shared Currency to subscription domain Currency
func ConvertFromSharedCurrency(currency sharedvo.Currency) (*Currency, error) {
	if currency.IsEmpty() {
		return nil, nil // Return nil for zero values
	}
	// Convert shared Currency to string for subscription domain compatibility
	return NewCurrency(currency.Code())
}

// Price/Money conversions

// ConvertToSharedMoney converts subscription domain Price to shared Money
func ConvertToSharedMoney(price *Price) (sharedvo.Money, error) {
	if price == nil {
		return sharedvo.Money{}, nil // Allow nil for compatibility
	}
	
	// Convert currency first - Price.Currency() returns a value, not pointer
	currency := price.Currency()
	sharedCurrency, err := ConvertToSharedCurrency(&currency)
	if err != nil {
		return sharedvo.Money{}, err
	}
	
	// Convert price to money
	return sharedvo.NewMoney(price.Amount(), sharedCurrency)
}

// ConvertFromSharedMoney converts shared Money to subscription domain Price
func ConvertFromSharedMoney(money sharedvo.Money) (*Price, error) {
	if money.IsZero() {
		return nil, nil // Return nil for zero values
	}
	
	// Convert currency first
	domainCurrency, err := ConvertFromSharedCurrency(money.Currency())
	if err != nil {
		return nil, err
	}
	
	// Convert money to price
	return NewPrice(money.Amount(), *domainCurrency)
}

// Utility functions for legacy support

// ConvertToSharedUserIDFromUint converts uint to shared UserID (for legacy support)
func ConvertToSharedUserIDFromUint(id uint) (sharedvo.UserID, error) {
	if id == 0 {
		return sharedvo.UserID{}, nil // Allow zero for compatibility
	}
	return sharedvo.NewUserIDFromUint(id)
}

// ConvertFromSharedUserIDToUint converts shared UserID to uint (for legacy support)
func ConvertFromSharedUserIDToUint(id sharedvo.UserID) uint {
	if id.IsZero() {
		return 0
	}
	return id.ToUint()
}