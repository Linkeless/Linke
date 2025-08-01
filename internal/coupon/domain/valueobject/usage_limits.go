package valueobject

import "fmt"

// UsageLimits represents the usage limitations of a coupon
type UsageLimits struct {
	maxUses        int
	usedCount      int
	maxUsesPerUser int
}

// NewUsageLimits creates a new usage limits with validation
func NewUsageLimits(maxUses, usedCount, maxUsesPerUser int) (UsageLimits, error) {
	if maxUses < 0 {
		return UsageLimits{}, fmt.Errorf("maxUses cannot be negative: %d", maxUses)
	}
	
	if usedCount < 0 {
		return UsageLimits{}, fmt.Errorf("usedCount cannot be negative: %d", usedCount)
	}
	
	if maxUsesPerUser < 1 {
		return UsageLimits{}, fmt.Errorf("maxUsesPerUser must be at least 1: %d", maxUsesPerUser)
	}
	
	// UsedCount should not exceed MaxUses (if MaxUses is limited)
	if maxUses > 0 && usedCount > maxUses {
		return UsageLimits{}, fmt.Errorf("usedCount (%d) cannot exceed maxUses (%d)", usedCount, maxUses)
	}
	
	return UsageLimits{
		maxUses:        maxUses,
		usedCount:      usedCount,
		maxUsesPerUser: maxUsesPerUser,
	}, nil
}

// MustNewUsageLimits creates a new usage limits and panics on error
func MustNewUsageLimits(maxUses, usedCount, maxUsesPerUser int) UsageLimits {
	ul, err := NewUsageLimits(maxUses, usedCount, maxUsesPerUser)
	if err != nil {
		panic(err)
	}
	return ul
}

// NewUnlimitedUsageLimits creates usage limits with no global limit
func NewUnlimitedUsageLimits(maxUsesPerUser int) (UsageLimits, error) {
	return NewUsageLimits(0, 0, maxUsesPerUser)
}

// MaxUses returns the maximum total uses allowed (0 = unlimited)
func (ul UsageLimits) MaxUses() int {
	return ul.maxUses
}

// UsedCount returns the current number of uses
func (ul UsageLimits) UsedCount() int {
	return ul.usedCount
}

// MaxUsesPerUser returns the maximum uses per user
func (ul UsageLimits) MaxUsesPerUser() int {
	return ul.maxUsesPerUser
}

// RemainingUses returns the number of remaining uses (nil if unlimited)
func (ul UsageLimits) RemainingUses() *int {
	if ul.maxUses == 0 {
		return nil // Unlimited
	}
	
	remaining := ul.maxUses - ul.usedCount
	if remaining < 0 {
		remaining = 0
	}
	return &remaining
}

// IsUnlimited checks if the coupon has unlimited uses
func (ul UsageLimits) IsUnlimited() bool {
	return ul.maxUses == 0
}

// IsExhausted checks if the coupon has reached its usage limit
func (ul UsageLimits) IsExhausted() bool {
	if ul.maxUses == 0 {
		return false // Unlimited
	}
	return ul.usedCount >= ul.maxUses
}

// CanBeUsed checks if the coupon can still be used
func (ul UsageLimits) CanBeUsed() bool {
	return !ul.IsExhausted()
}

// IncrementUsedCount returns new usage limits with incremented used count
func (ul UsageLimits) IncrementUsedCount() (UsageLimits, error) {
	if ul.IsExhausted() {
		return UsageLimits{}, fmt.Errorf("coupon usage limit has been reached")
	}
	
	return UsageLimits{
		maxUses:        ul.maxUses,
		usedCount:      ul.usedCount + 1,
		maxUsesPerUser: ul.maxUsesPerUser,
	}, nil
}

// String returns string representation
func (ul UsageLimits) String() string {
	if ul.IsUnlimited() {
		return fmt.Sprintf("used:%d/unlimited, max_per_user:%d", ul.usedCount, ul.maxUsesPerUser)
	}
	return fmt.Sprintf("used:%d/%d, max_per_user:%d", ul.usedCount, ul.maxUses, ul.maxUsesPerUser)
}

// Equals checks if two usage limits are equal
func (ul UsageLimits) Equals(other UsageLimits) bool {
	return ul.maxUses == other.maxUses &&
		ul.usedCount == other.usedCount &&
		ul.maxUsesPerUser == other.maxUsesPerUser
}

// MarshalJSON implements json.Marshaler
func (ul UsageLimits) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`{"max_uses":%d,"used_count":%d,"max_uses_per_user":%d}`, 
		ul.maxUses, ul.usedCount, ul.maxUsesPerUser)), nil
}