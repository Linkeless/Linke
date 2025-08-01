package valueobject

import (
	"fmt"
	"strings"
)

// TicketCategory represents the category of a ticket
type TicketCategory struct {
	value string
}

// Valid ticket categories
const (
	TicketCategoryGeneral      = "general"
	TicketCategoryTechnical    = "technical"
	TicketCategoryBilling      = "billing"
	TicketCategoryAccount      = "account"
	TicketCategoryFeature      = "feature"
	TicketCategoryBug          = "bug"
	TicketCategorySubscription = "subscription"
	TicketCategoryPayment      = "payment"
)

var validCategories = map[string]bool{
	TicketCategoryGeneral:      true,
	TicketCategoryTechnical:    true,
	TicketCategoryBilling:      true,
	TicketCategoryAccount:      true,
	TicketCategoryFeature:      true,
	TicketCategoryBug:          true,
	TicketCategorySubscription: true,
	TicketCategoryPayment:      true,
}

// NewTicketCategory creates a new TicketCategory with validation
func NewTicketCategory(value string) (TicketCategory, error) {
	value = strings.ToLower(strings.TrimSpace(value))
	
	if value == "" {
		return TicketCategory{}, fmt.Errorf("ticket category cannot be empty")
	}
	
	if !validCategories[value] {
		return TicketCategory{}, fmt.Errorf("invalid ticket category: %s", value)
	}
	
	return TicketCategory{value: value}, nil
}

// MustTicketCategory creates a TicketCategory and panics if invalid
func MustTicketCategory(value string) TicketCategory {
	tc, err := NewTicketCategory(value)
	if err != nil {
		panic(err)
	}
	return tc
}

// DefaultTicketCategory returns the default category for new tickets
func DefaultTicketCategory() TicketCategory {
	return TicketCategory{value: TicketCategoryGeneral}
}

// Value returns the underlying value
func (t TicketCategory) Value() string {
	return t.value
}

// String returns string representation
func (t TicketCategory) String() string {
	return t.value
}

// IsGeneral checks if category is general
func (t TicketCategory) IsGeneral() bool {
	return t.value == TicketCategoryGeneral
}

// IsTechnical checks if category is technical
func (t TicketCategory) IsTechnical() bool {
	return t.value == TicketCategoryTechnical
}

// IsBilling checks if category is billing
func (t TicketCategory) IsBilling() bool {
	return t.value == TicketCategoryBilling
}

// IsAccount checks if category is account
func (t TicketCategory) IsAccount() bool {
	return t.value == TicketCategoryAccount
}

// IsFeature checks if category is feature
func (t TicketCategory) IsFeature() bool {
	return t.value == TicketCategoryFeature
}

// IsBug checks if category is bug
func (t TicketCategory) IsBug() bool {
	return t.value == TicketCategoryBug
}

// IsSubscription checks if category is subscription
func (t TicketCategory) IsSubscription() bool {
	return t.value == TicketCategorySubscription
}

// IsPayment checks if category is payment
func (t TicketCategory) IsPayment() bool {
	return t.value == TicketCategoryPayment
}

// RequiresTechnicalExpertise checks if category requires technical expertise
func (t TicketCategory) RequiresTechnicalExpertise() bool {
	return t.IsTechnical() || t.IsBug()
}

// RequiresFinancialAccess checks if category requires financial access
func (t TicketCategory) RequiresFinancialAccess() bool {
	return t.IsBilling() || t.IsPayment() || t.IsSubscription()
}

// Equals checks equality with another TicketCategory
func (t TicketCategory) Equals(other TicketCategory) bool {
	return t.value == other.value
}

// MarshalJSON implements json.Marshaler
func (t TicketCategory) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, t.value)), nil
}

// UnmarshalJSON implements json.Unmarshaler
func (t *TicketCategory) UnmarshalJSON(data []byte) error {
	str := strings.Trim(string(data), `"`)
	if str == "null" {
		*t = DefaultTicketCategory()
		return nil
	}
	
	tc, err := NewTicketCategory(str)
	if err != nil {
		return err
	}
	
	*t = tc
	return nil
}