package valueobject

import (
	"fmt"
	"strings"
)

// CompanyInfo represents company information for invoicing
type CompanyInfo struct {
	name    string
	address string
	taxID   string
}

// NewCompanyInfo creates a new CompanyInfo
func NewCompanyInfo(name, address, taxID string) (CompanyInfo, error) {
	name = strings.TrimSpace(name)
	address = strings.TrimSpace(address)
	taxID = strings.TrimSpace(taxID)

	if len(name) > 255 {
		return CompanyInfo{}, fmt.Errorf("company name cannot exceed 255 characters")
	}

	if len(address) > 1000 {
		return CompanyInfo{}, fmt.Errorf("company address cannot exceed 1000 characters")
	}

	if len(taxID) > 50 {
		return CompanyInfo{}, fmt.Errorf("company tax ID cannot exceed 50 characters")
	}

	return CompanyInfo{
		name:    name,
		address: address,
		taxID:   taxID,
	}, nil
}

// EmptyCompanyInfo creates an empty CompanyInfo
func EmptyCompanyInfo() CompanyInfo {
	return CompanyInfo{}
}

// Name returns the company name
func (c CompanyInfo) Name() string {
	return c.name
}

// Address returns the company address
func (c CompanyInfo) Address() string {
	return c.address
}

// TaxID returns the tax ID
func (c CompanyInfo) TaxID() string {
	return c.taxID
}

// IsEmpty checks if company info is empty
func (c CompanyInfo) IsEmpty() bool {
	return c.name == "" && c.address == "" && c.taxID == ""
}

// HasName checks if company has a name
func (c CompanyInfo) HasName() bool {
	return c.name != ""
}

// HasAddress checks if company has an address
func (c CompanyInfo) HasAddress() bool {
	return c.address != ""
}

// HasTaxID checks if company has a tax ID
func (c CompanyInfo) HasTaxID() bool {
	return c.taxID != ""
}

// String returns string representation
func (c CompanyInfo) String() string {
	if c.IsEmpty() {
		return ""
	}

	var parts []string
	if c.name != "" {
		parts = append(parts, c.name)
	}
	if c.address != "" {
		parts = append(parts, c.address)
	}
	if c.taxID != "" {
		parts = append(parts, "Tax ID: "+c.taxID)
	}

	return strings.Join(parts, "\n")
}

// Equals checks if two CompanyInfo are equal
func (c CompanyInfo) Equals(other CompanyInfo) bool {
	return c.name == other.name &&
		c.address == other.address &&
		c.taxID == other.taxID
}