package valueobject

import (
	sharedvo "linke/internal/shared/valueobject"
)

// ConvertToSharedMoney converts invoice domain Money to shared Money
func ConvertToSharedMoney(money Money) (sharedvo.Money, error) {
	// Convert currency first
	sharedCurrency, err := sharedvo.NewCurrency(money.Currency().Code())
	if err != nil {
		return sharedvo.Money{}, err
	}
	
	return sharedvo.NewMoney(money.Amount(), sharedCurrency)
}

// ConvertFromSharedMoney converts shared Money to invoice domain Money  
func ConvertFromSharedMoney(money sharedvo.Money) (Money, error) {
	domainCurrency, err := NewCurrency(money.Currency().Code())
	if err != nil {
		return Money{}, err
	}
	
	return NewMoney(money.Amount(), domainCurrency)
}

// ConvertToSharedCurrency converts invoice domain Currency to shared Currency
func ConvertToSharedCurrency(currency Currency) (sharedvo.Currency, error) {
	return sharedvo.NewCurrency(currency.Code())
}

// ConvertFromSharedCurrency converts shared Currency to invoice domain Currency
func ConvertFromSharedCurrency(currency sharedvo.Currency) (Currency, error) {
	return NewCurrency(currency.Code())
}

// ConvertToSharedInvoiceID converts invoice domain InvoiceID to shared InvoiceID
func ConvertToSharedInvoiceID(id InvoiceID) (sharedvo.InvoiceID, error) {
	// Invoice domain doesn't validate zero values, but shared does
	if id.Value() == 0 {
		return sharedvo.GenerateInvoiceID(), nil // Return placeholder for zero values
	}
	return sharedvo.NewInvoiceID(id.Value())
}

// ConvertFromSharedInvoiceID converts shared InvoiceID to invoice domain InvoiceID
func ConvertFromSharedInvoiceID(id sharedvo.InvoiceID) InvoiceID {
	return NewInvoiceID(id.Value())
}

// ConvertToSharedUserID converts uint UserID to shared UserID
func ConvertToSharedUserID(userID uint) (sharedvo.UserID, error) {
	if userID == 0 {
		return sharedvo.UserID{}, nil // Allow zero for compatibility
	}
	return sharedvo.NewUserIDFromUint(userID)
}

// ConvertFromSharedUserID converts shared UserID to uint UserID
func ConvertFromSharedUserID(id sharedvo.UserID) uint {
	if id.IsZero() {
		return 0
	}
	return id.ToUint()
}