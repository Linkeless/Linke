package valueobject

import (
	"fmt"
	"regexp"
	"strings"
)

// BillingAddress represents a billing address
type BillingAddress struct {
	name    string
	email   string
	address string
	city    string
	state   string
	country string
	zip     string
}

// NewBillingAddress creates a new BillingAddress
func NewBillingAddress(name, email, address, city, state, country, zip string) (BillingAddress, error) {
	// Trim whitespace from all fields
	name = strings.TrimSpace(name)
	email = strings.TrimSpace(email)
	address = strings.TrimSpace(address)
	city = strings.TrimSpace(city)
	state = strings.TrimSpace(state)
	country = strings.TrimSpace(country)
	zip = strings.TrimSpace(zip)

	// Validate required fields
	if name == "" {
		return BillingAddress{}, fmt.Errorf("billing name is required")
	}

	if email == "" {
		return BillingAddress{}, fmt.Errorf("billing email is required")
	}

	// Validate email format
	if err := validateEmail(email); err != nil {
		return BillingAddress{}, fmt.Errorf("invalid billing email: %w", err)
	}

	// Validate field lengths
	if len(name) > 255 {
		return BillingAddress{}, fmt.Errorf("billing name cannot exceed 255 characters")
	}

	if len(email) > 255 {
		return BillingAddress{}, fmt.Errorf("billing email cannot exceed 255 characters")
	}

	if len(city) > 100 {
		return BillingAddress{}, fmt.Errorf("billing city cannot exceed 100 characters")
	}

	if len(state) > 100 {
		return BillingAddress{}, fmt.Errorf("billing state cannot exceed 100 characters")
	}

	if len(zip) > 20 {
		return BillingAddress{}, fmt.Errorf("billing zip cannot exceed 20 characters")
	}

	// Validate country code (2 characters if provided)
	if country != "" && len(country) != 2 {
		return BillingAddress{}, fmt.Errorf("country code must be exactly 2 characters")
	}

	return BillingAddress{
		name:    name,
		email:   email,
		address: address,
		city:    city,
		state:   state,
		country: strings.ToUpper(country),
		zip:     zip,
	}, nil
}

// Name returns the billing name
func (ba BillingAddress) Name() string {
	return ba.name
}

// Email returns the billing email
func (ba BillingAddress) Email() string {
	return ba.email
}

// Address returns the street address
func (ba BillingAddress) Address() string {
	return ba.address
}

// City returns the city
func (ba BillingAddress) City() string {
	return ba.city
}

// State returns the state/province
func (ba BillingAddress) State() string {
	return ba.state
}

// Country returns the country code
func (ba BillingAddress) Country() string {
	return ba.country
}

// Zip returns the postal/zip code
func (ba BillingAddress) Zip() string {
	return ba.zip
}

// FullAddress returns the complete formatted address
func (ba BillingAddress) FullAddress() string {
	var parts []string

	if ba.address != "" {
		parts = append(parts, ba.address)
	}

	var cityStateZip []string
	if ba.city != "" {
		cityStateZip = append(cityStateZip, ba.city)
	}
	if ba.state != "" {
		if len(cityStateZip) > 0 {
			cityStateZip = append(cityStateZip, ", "+ba.state)
		} else {
			cityStateZip = append(cityStateZip, ba.state)
		}
	}
	if ba.zip != "" {
		if len(cityStateZip) > 0 {
			cityStateZip = append(cityStateZip, " "+ba.zip)
		} else {
			cityStateZip = append(cityStateZip, ba.zip)
		}
	}

	if len(cityStateZip) > 0 {
		parts = append(parts, strings.Join(cityStateZip, ""))
	}

	if ba.country != "" {
		parts = append(parts, ba.country)
	}

	return strings.Join(parts, "\n")
}

// Equals checks if two billing addresses are equal
func (ba BillingAddress) Equals(other BillingAddress) bool {
	return ba.name == other.name &&
		ba.email == other.email &&
		ba.address == other.address &&
		ba.city == other.city &&
		ba.state == other.state &&
		ba.country == other.country &&
		ba.zip == other.zip
}

// validateEmail validates email format
func validateEmail(email string) error {
	if email == "" {
		return fmt.Errorf("email cannot be empty")
	}

	// Basic email regex pattern
	pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	matched, err := regexp.MatchString(pattern, email)
	if err != nil {
		return fmt.Errorf("failed to validate email: %w", err)
	}

	if !matched {
		return fmt.Errorf("invalid email format")
	}

	return nil
}