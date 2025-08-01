package valueobject

import (
	"fmt"
	"strings"
)

// TaxInfo represents tax information for an invoice
type TaxInfo struct {
	rate       float64
	taxType    string
	taxNumber  string
	taxAmount  Money
}

// NewTaxInfo creates a new TaxInfo
func NewTaxInfo(rate float64, taxType, taxNumber string, taxAmount Money) (TaxInfo, error) {
	if rate < 0 || rate > 1 {
		return TaxInfo{}, fmt.Errorf("tax rate must be between 0 and 1 (0%% to 100%%)")
	}

	taxType = strings.TrimSpace(taxType)
	taxNumber = strings.TrimSpace(taxNumber)

	// Validate tax type if provided
	if taxType != "" {
		validTaxTypes := map[string]bool{
			"VAT": true,
			"GST": true,
			"HST": true,
			"PST": true,
			"QST": true,
			"TAX": true,
		}

		if !validTaxTypes[strings.ToUpper(taxType)] {
			return TaxInfo{}, fmt.Errorf("invalid tax type: %s", taxType)
		}
		taxType = strings.ToUpper(taxType)
	}

	// Validate tax number format if provided
	if taxNumber != "" && len(taxNumber) > 50 {
		return TaxInfo{}, fmt.Errorf("tax number cannot exceed 50 characters")
	}

	return TaxInfo{
		rate:      rate,
		taxType:   taxType,
		taxNumber: taxNumber,
		taxAmount: taxAmount,
	}, nil
}

// NewZeroTax creates a TaxInfo with zero tax
func NewZeroTax(currency Currency) TaxInfo {
	return TaxInfo{
		rate:      0,
		taxType:   "",
		taxNumber: "",
		taxAmount: Zero(currency),
	}
}

// Rate returns the tax rate as a decimal (e.g., 0.2 for 20%)
func (ti TaxInfo) Rate() float64 {
	return ti.rate
}

// RatePercentage returns the tax rate as a percentage (e.g., 20.0 for 20%)
func (ti TaxInfo) RatePercentage() float64 {
	return ti.rate * 100
}

// TaxType returns the tax type (VAT, GST, etc.)
func (ti TaxInfo) TaxType() string {
	return ti.taxType
}

// TaxNumber returns the tax number
func (ti TaxInfo) TaxNumber() string {
	return ti.taxNumber
}

// TaxAmount returns the calculated tax amount
func (ti TaxInfo) TaxAmount() Money {
	return ti.taxAmount
}

// IsZero checks if tax is zero
func (ti TaxInfo) IsZero() bool {
	return ti.rate == 0 && ti.taxAmount.IsZero()
}

// HasTaxNumber checks if tax number is provided
func (ti TaxInfo) HasTaxNumber() bool {
	return ti.taxNumber != ""
}

// CalculateTaxAmount calculates tax amount for a given subtotal
func (ti TaxInfo) CalculateTaxAmount(subtotal Money) (Money, error) {
	return subtotal.Multiply(ti.rate)
}

// Equals checks if two tax infos are equal
func (ti TaxInfo) Equals(other TaxInfo) bool {
	return ti.rate == other.rate &&
		ti.taxType == other.taxType &&
		ti.taxNumber == other.taxNumber &&
		ti.taxAmount.Equals(other.taxAmount)
}

// String returns string representation of the tax info
func (ti TaxInfo) String() string {
	if ti.IsZero() {
		return "No Tax"
	}

	result := fmt.Sprintf("%.2f%% %s", ti.RatePercentage(), ti.taxType)
	if ti.taxNumber != "" {
		result += fmt.Sprintf(" (%s)", ti.taxNumber)
	}
	result += fmt.Sprintf(" = %s", ti.taxAmount.String())
	
	return result
}

