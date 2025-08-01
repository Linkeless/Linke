package valueobject

import (
	"testing"
)

func TestUserID(t *testing.T) {
	// Test NewUserID
	userID := NewUserID()
	if userID.IsEmpty() {
		t.Error("NewUserID should not be empty")
	}
	if !userID.IsUUID() {
		t.Error("NewUserID should generate UUID format")
	}

	// Test NewUserIDFromString
	id, err := NewUserIDFromString("123")
	if err != nil {
		t.Errorf("NewUserIDFromString failed: %v", err)
	}
	if id.String() != "123" {
		t.Error("UserID string mismatch")
	}

	// Test NewUserIDFromUint
	id2, err := NewUserIDFromUint(456)
	if err != nil {
		t.Errorf("NewUserIDFromUint failed: %v", err)
	}
	if id2.ToUint() != 456 {
		t.Error("UserID uint conversion failed")
	}

	// Test zero validation
	_, err = NewUserIDFromUint(0)
	if err == nil {
		t.Error("Should reject zero UserID")
	}
}

func TestCurrency(t *testing.T) {
	// Test valid currencies
	usd, err := NewCurrency("USD")
	if err != nil {
		t.Errorf("USD should be valid: %v", err)
	}
	if !usd.IsFiat() {
		t.Error("USD should be fiat currency")
	}

	btc, err := NewCurrency("BTC")
	if err != nil {
		t.Errorf("BTC should be valid: %v", err)
	}
	if !btc.IsCrypto() {
		t.Error("BTC should be crypto currency")
	}

	// Test invalid currency
	_, err = NewCurrency("INVALID")
	if err == nil {
		t.Error("Should reject invalid currency")
	}

	// Test predefined currencies
	if !USD.Equals(usd) {
		t.Error("Predefined USD should equal new USD")
	}
}

func TestMoney(t *testing.T) {
	// Test valid money creation
	money, err := NewMoney(100.50, USD)
	if err != nil {
		t.Errorf("Valid money creation failed: %v", err)
	}
	if money.Amount() != 100.50 {
		t.Error("Money amount mismatch")
	}

	// Test negative amount
	_, err = NewMoney(-10, USD)
	if err == nil {
		t.Error("Should reject negative amount")
	}

	// Test operations
	money2, err := NewMoney(50.25, USD)
	if err != nil {
		t.Fatal(err)
	}

	sum, err := money.Add(money2)
	if err != nil {
		t.Errorf("Money addition failed: %v", err)
	}
	if sum.Amount() != 150.75 {
		t.Errorf("Expected 150.75, got %f", sum.Amount())
	}

	// Test different currency addition
	eur, _ := NewCurrency("EUR")
	eurMoney, _ := NewMoney(100, eur)
	_, err = money.Add(eurMoney)
	if err == nil {
		t.Error("Should reject addition of different currencies")
	}
}

func TestInvoiceID(t *testing.T) {
	// Test valid invoice ID
	id, err := NewInvoiceID(123)
	if err != nil {
		t.Errorf("Valid invoice ID creation failed: %v", err)
	}
	if id.Value() != 123 {
		t.Error("Invoice ID value mismatch")
	}

	// Test zero validation
	_, err = NewInvoiceID(0)
	if err == nil {
		t.Error("Should reject zero invoice ID")
	}

	// Test string parsing
	id2, err := ParseInvoiceID("456")
	if err != nil {
		t.Errorf("Invoice ID parsing failed: %v", err)
	}
	if id2.Value() != 456 {
		t.Error("Parsed invoice ID value mismatch")
	}

	// Test invalid string parsing
	_, err = ParseInvoiceID("invalid")
	if err == nil {
		t.Error("Should reject invalid invoice ID string")
	}
}