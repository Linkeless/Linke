package valueobject

import (
	"fmt"
	"time"
)

// ValidityPeriod represents a time period during which a coupon is valid
type ValidityPeriod struct {
	validFrom  *time.Time
	validUntil *time.Time
}

// NewValidityPeriod creates a new validity period with validation
func NewValidityPeriod(validFrom, validUntil *time.Time) (ValidityPeriod, error) {
	// Both nil is valid (no time restrictions)
	if validFrom == nil && validUntil == nil {
		return ValidityPeriod{}, nil
	}
	
	// If both are provided, validFrom must be before validUntil
	if validFrom != nil && validUntil != nil {
		if validFrom.After(*validUntil) || validFrom.Equal(*validUntil) {
			return ValidityPeriod{}, fmt.Errorf("validFrom must be before validUntil: %v >= %v", validFrom, validUntil)
		}
	}
	
	return ValidityPeriod{
		validFrom:  validFrom,
		validUntil: validUntil,
	}, nil
}

// MustNewValidityPeriod creates a new validity period and panics on error
func MustNewValidityPeriod(validFrom, validUntil *time.Time) ValidityPeriod {
	vp, err := NewValidityPeriod(validFrom, validUntil)
	if err != nil {
		panic(err)
	}
	return vp
}

// NewUnlimitedValidityPeriod creates a validity period with no time restrictions
func NewUnlimitedValidityPeriod() ValidityPeriod {
	return ValidityPeriod{}
}

// NewValidityPeriodFrom creates a validity period starting from a specific time
func NewValidityPeriodFrom(validFrom time.Time) ValidityPeriod {
	return ValidityPeriod{
		validFrom: &validFrom,
	}
}

// NewValidityPeriodUntil creates a validity period ending at a specific time
func NewValidityPeriodUntil(validUntil time.Time) ValidityPeriod {
	return ValidityPeriod{
		validUntil: &validUntil,
	}
}

// ValidFrom returns the start time of validity period
func (vp ValidityPeriod) ValidFrom() *time.Time {
	return vp.validFrom
}

// ValidUntil returns the end time of validity period
func (vp ValidityPeriod) ValidUntil() *time.Time {
	return vp.validUntil
}

// IsValidAt checks if the period is valid at a specific time
func (vp ValidityPeriod) IsValidAt(checkTime time.Time) bool {
	// Check validFrom
	if vp.validFrom != nil && checkTime.Before(*vp.validFrom) {
		return false
	}
	
	// Check validUntil
	if vp.validUntil != nil && checkTime.After(*vp.validUntil) {
		return false
	}
	
	return true
}

// IsValidNow checks if the period is valid at the current time
func (vp ValidityPeriod) IsValidNow() bool {
	return vp.IsValidAt(time.Now())
}

// IsExpired checks if the validity period has expired
func (vp ValidityPeriod) IsExpired() bool {
	return vp.validUntil != nil && time.Now().After(*vp.validUntil)
}

// IsNotYetValid checks if the validity period has not yet started
func (vp ValidityPeriod) IsNotYetValid() bool {
	return vp.validFrom != nil && time.Now().Before(*vp.validFrom)
}

// HasStartTime checks if the period has a start time
func (vp ValidityPeriod) HasStartTime() bool {
	return vp.validFrom != nil
}

// HasEndTime checks if the period has an end time
func (vp ValidityPeriod) HasEndTime() bool {
	return vp.validUntil != nil
}

// IsUnlimited checks if the period has no time restrictions
func (vp ValidityPeriod) IsUnlimited() bool {
	return vp.validFrom == nil && vp.validUntil == nil
}

// Duration calculates the duration of the validity period
// Returns nil if either end is unlimited
func (vp ValidityPeriod) Duration() *time.Duration {
	if vp.validFrom == nil || vp.validUntil == nil {
		return nil
	}
	
	duration := vp.validUntil.Sub(*vp.validFrom)
	return &duration
}

// String returns string representation
func (vp ValidityPeriod) String() string {
	if vp.IsUnlimited() {
		return "unlimited"
	}
	
	var from, until string
	if vp.validFrom != nil {
		from = vp.validFrom.Format(time.RFC3339)
	} else {
		from = "unlimited"
	}
	
	if vp.validUntil != nil {
		until = vp.validUntil.Format(time.RFC3339)
	} else {
		until = "unlimited"
	}
	
	return fmt.Sprintf("from:%s until:%s", from, until)
}

// Equals checks if two validity periods are equal
func (vp ValidityPeriod) Equals(other ValidityPeriod) bool {
	// Compare validFrom
	if (vp.validFrom == nil) != (other.validFrom == nil) {
		return false
	}
	if vp.validFrom != nil && !vp.validFrom.Equal(*other.validFrom) {
		return false
	}
	
	// Compare validUntil
	if (vp.validUntil == nil) != (other.validUntil == nil) {
		return false
	}
	if vp.validUntil != nil && !vp.validUntil.Equal(*other.validUntil) {
		return false
	}
	
	return true
}

// MarshalJSON implements json.Marshaler
func (vp ValidityPeriod) MarshalJSON() ([]byte, error) {
	type jsonValidityPeriod struct {
		ValidFrom  *time.Time `json:"valid_from"`
		ValidUntil *time.Time `json:"valid_until"`
	}
	
	return []byte(fmt.Sprintf(`{"valid_from":%s,"valid_until":%s}`, 
		timeToJSON(vp.validFrom), timeToJSON(vp.validUntil))), nil
}

// Helper function to convert time pointer to JSON
func timeToJSON(t *time.Time) string {
	if t == nil {
		return "null"
	}
	return fmt.Sprintf("\"%s\"", t.Format(time.RFC3339))
}