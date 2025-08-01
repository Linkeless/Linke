package aggregate

import (
	"encoding/json"
	"fmt"
	"time"

	"linke/internal/coupon/domain/entity"
	"linke/internal/coupon/domain/event"
	"linke/internal/coupon/domain/valueobject"
	sharedvo "linke/internal/shared/valueobject"
	"linke/internal/shared/domain"
)

// Coupon represents the coupon aggregate root
type Coupon struct {
	// Identity
	id   valueobject.CouponID
	code valueobject.CouponCode

	// Basic Information
	name        string
	description string

	// Discount Configuration
	discountValue  valueobject.DiscountValue
	minOrderAmount sharedvo.Money

	// Validity
	validityPeriod valueobject.ValidityPeriod
	status         valueobject.CouponStatus

	// Usage Configuration
	usageLimits valueobject.UsageLimits

	// Applicable Plans
	applicablePlans []uint64

	// Visibility
	isPublic bool

	// Audit
	createdBy sharedvo.UserID
	createdAt time.Time
	updatedAt time.Time
	deletedAt *time.Time

	// Domain Events
	domainEvents []domain.DomainEvent

	// Entities (managed by this aggregate)
	usages map[entity.CouponUsageID]*entity.CouponUsage
}

// NewCoupon creates a new coupon aggregate
func NewCoupon(
	id valueobject.CouponID,
	code valueobject.CouponCode,
	name string,
	description string,
	discountValue valueobject.DiscountValue,
	minOrderAmount sharedvo.Money,
	validityPeriod valueobject.ValidityPeriod,
	usageLimits valueobject.UsageLimits,
	applicablePlans []uint64,
	isPublic bool,
	createdBy sharedvo.UserID,
) (*Coupon, error) {
	// Validation
	if id.IsEmpty() {
		return nil, fmt.Errorf("coupon ID cannot be empty")
	}

	if code.IsEmpty() {
		return nil, fmt.Errorf("coupon code cannot be empty")
	}

	if name == "" {
		return nil, fmt.Errorf("coupon name cannot be empty")
	}

	if createdBy.IsZero() {
		return nil, fmt.Errorf("created by cannot be empty")
	}

	// Create coupon
	now := time.Now()
	coupon := &Coupon{
		id:              id,
		code:            code,
		name:            name,
		description:     description,
		discountValue:   discountValue,
		minOrderAmount:  minOrderAmount,
		validityPeriod:  validityPeriod,
		status:          valueobject.CouponStatusActive,
		usageLimits:     usageLimits,
		applicablePlans: applicablePlans,
		isPublic:        isPublic,
		createdBy:       createdBy,
		createdAt:       now,
		updatedAt:       now,
		domainEvents:    make([]domain.DomainEvent, 0),
		usages:          make(map[entity.CouponUsageID]*entity.CouponUsage),
	}

	// Add domain event (convert shared types to domain types for event)
	domainMinOrderAmount, err := valueobject.ConvertFromSharedMoney(minOrderAmount)
	if err != nil {
		return nil, fmt.Errorf("failed to convert min order amount for event: %w", err)
	}
	
	coupon.addDomainEvent(event.NewCouponCreatedEvent(
		fmt.Sprintf("coupon-created-%s-%d", id.String(), time.Now().UnixNano()),
		id,
		code,
		discountValue.Type(),
		discountValue,
		domainMinOrderAmount,
		validityPeriod,
		usageLimits,
		createdBy,
		name,
		description,
		applicablePlans,
		isPublic,
	))

	return coupon, nil
}

// ID returns the coupon ID
func (c *Coupon) ID() valueobject.CouponID {
	return c.id
}

// Code returns the coupon code
func (c *Coupon) Code() valueobject.CouponCode {
	return c.code
}

// Name returns the coupon name
func (c *Coupon) Name() string {
	return c.name
}

// Description returns the coupon description
func (c *Coupon) Description() string {
	return c.description
}

// DiscountValue returns the discount value
func (c *Coupon) DiscountValue() valueobject.DiscountValue {
	return c.discountValue
}

// MinOrderAmount returns the minimum order amount
func (c *Coupon) MinOrderAmount() sharedvo.Money {
	return c.minOrderAmount
}

// ValidityPeriod returns the validity period
func (c *Coupon) ValidityPeriod() valueobject.ValidityPeriod {
	return c.validityPeriod
}

// Status returns the coupon status
func (c *Coupon) Status() valueobject.CouponStatus {
	return c.status
}

// UsageLimits returns the usage limits
func (c *Coupon) UsageLimits() valueobject.UsageLimits {
	return c.usageLimits
}

// ApplicablePlans returns the applicable plans
func (c *Coupon) ApplicablePlans() []uint64 {
	return c.applicablePlans
}

// IsPublic returns whether the coupon is public
func (c *Coupon) IsPublic() bool {
	return c.isPublic
}

// CreatedBy returns who created the coupon
func (c *Coupon) CreatedBy() sharedvo.UserID {
	return c.createdBy
}

// CreatedAt returns the creation time
func (c *Coupon) CreatedAt() time.Time {
	return c.createdAt
}

// UpdatedAt returns the last update time
func (c *Coupon) UpdatedAt() time.Time {
	return c.updatedAt
}

// DeletedAt returns the deletion time
func (c *Coupon) DeletedAt() *time.Time {
	return c.deletedAt
}

// IsDeleted checks if the coupon is soft deleted
func (c *Coupon) IsDeleted() bool {
	return c.deletedAt != nil
}

// DomainEvents returns the domain events
func (c *Coupon) DomainEvents() []domain.DomainEvent {
	return c.domainEvents
}

// ClearDomainEvents clears the domain events
func (c *Coupon) ClearDomainEvents() {
	c.domainEvents = make([]domain.DomainEvent, 0)
}

// IsValidForUse checks if the coupon is currently valid for use
func (c *Coupon) IsValidForUse() (bool, string) {
	// Check if deleted
	if c.IsDeleted() {
		return false, "Coupon has been deleted"
	}

	// Check status
	if !c.status.IsActive() {
		return false, fmt.Sprintf("Coupon is %s", c.status.String())
	}

	// Check validity period
	if !c.validityPeriod.IsValidNow() {
		if c.validityPeriod.IsExpired() {
			return false, "Coupon has expired"
		}
		if c.validityPeriod.IsNotYetValid() {
			return false, "Coupon is not yet valid"
		}
	}

	// Check usage limits
	if c.usageLimits.IsExhausted() {
		return false, "Coupon usage limit has been reached"
	}

	return true, ""
}

// CanBeUsedBy checks if the coupon can be used by a specific user for a specific order
func (c *Coupon) CanBeUsedBy(userID sharedvo.UserID, orderAmount sharedvo.Money, planID uint64) (bool, string) {
	// First check basic validity
	if valid, reason := c.IsValidForUse(); !valid {
		return false, reason
	}

	// Check minimum order amount
	if isGreaterOrEqual, err := orderAmount.GreaterThanOrEqual(c.minOrderAmount); err != nil {
		return false, fmt.Sprintf("Error comparing order amounts: %v", err)
	} else if !isGreaterOrEqual {
		return false, fmt.Sprintf("Minimum order amount is %s", c.minOrderAmount.String())
	}

	// Check currency match
	if orderAmount.Currency() != c.minOrderAmount.Currency() {
		return false, fmt.Sprintf("Order currency %s does not match coupon currency %s", 
			orderAmount.Currency(), c.minOrderAmount.Currency())
	}

	// Check applicable plans
	if len(c.applicablePlans) > 0 {
		planApplicable := false
		for _, applicablePlan := range c.applicablePlans {
			if applicablePlan == planID {
				planApplicable = true
				break
			}
		}
		if !planApplicable {
			return false, "Coupon is not applicable to this plan"
		}
	}

	// Check per-user usage limit
	userUsageCount := c.getUserUsageCount(userID)
	if userUsageCount >= c.usageLimits.MaxUsesPerUser() {
		return false, "You have already used this coupon the maximum number of times"
	}

	return true, ""
}

// CalculateDiscount calculates the discount amount for a given order
func (c *Coupon) CalculateDiscount(orderAmount sharedvo.Money) (sharedvo.Money, error) {
	// Convert to domain-specific Money for discount calculation
	domainOrderAmount, err := valueobject.ConvertFromSharedMoney(orderAmount)
	if err != nil {
		return sharedvo.Money{}, fmt.Errorf("failed to convert order amount: %w", err)
	}
	
	domainDiscount, err := c.discountValue.CalculateDiscount(domainOrderAmount)
	if err != nil {
		return sharedvo.Money{}, err
	}
	
	// Convert back to shared Money
	return valueobject.ConvertToSharedMoney(domainDiscount)
}

// Use records the usage of the coupon by a user for an order
func (c *Coupon) Use(usageID entity.CouponUsageID, userID sharedvo.UserID, orderID uint64, orderAmount sharedvo.Money) error {
	// Validate that coupon can be used
	if canUse, reason := c.CanBeUsedBy(userID, orderAmount, 0); !canUse {
		return fmt.Errorf("coupon cannot be used: %s", reason)
	}

	// Calculate discount
	discountAmount, err := c.CalculateDiscount(orderAmount)
	if err != nil {
		return fmt.Errorf("failed to calculate discount: %w", err)
	}

	// Convert to domain-specific types for entity creation
	domainDiscountAmount, err := valueobject.ConvertFromSharedMoney(discountAmount)
	if err != nil {
		return fmt.Errorf("failed to convert discount amount: %w", err)
	}
	domainOrderAmount, err := valueobject.ConvertFromSharedMoney(orderAmount)
	if err != nil {
		return fmt.Errorf("failed to convert order amount: %w", err)
	}
	
	// Create usage entity
	usage, err := entity.NewCouponUsage(usageID, c.id, userID, orderID, domainDiscountAmount, domainOrderAmount)
	if err != nil {
		return fmt.Errorf("failed to create coupon usage: %w", err)
	}

	// Add usage to aggregate
	c.usages[usageID] = usage

	// Update usage limits
	newUsageLimits, err := c.usageLimits.IncrementUsedCount()
	if err != nil {
		return fmt.Errorf("failed to increment usage count: %w", err)
	}
	c.usageLimits = newUsageLimits
	c.updatedAt = time.Now()

	// Add domain event (convert back to domain types for event)
	domainDiscountAmountForEvent, _ := valueobject.ConvertFromSharedMoney(discountAmount)
	domainOrderAmountForEvent, _ := valueobject.ConvertFromSharedMoney(orderAmount)
	c.addDomainEvent(event.NewCouponUsedEvent(
		fmt.Sprintf("coupon-used-%s-%d", c.id.String(), time.Now().UnixNano()),
		c.id,
		userID,
		orderID,
		domainDiscountAmountForEvent,
		domainOrderAmountForEvent,
		c.usageLimits.RemainingUses(),
	))

	// Check if usage limit reached
	if c.usageLimits.IsExhausted() {
		c.addDomainEvent(event.NewCouponUsageLimitReachedEvent(
			fmt.Sprintf("coupon-limit-reached-%s-%d", c.id.String(), time.Now().UnixNano()),
			c.id,
			c.code,
			c.usageLimits,
		))
	}

	return nil
}

// ChangeStatus changes the coupon status
func (c *Coupon) ChangeStatus(newStatus valueobject.CouponStatus, changedBy sharedvo.UserID, reason string) error {
	// Check if transition is allowed
	if !c.status.CanTransitionTo(newStatus) {
		return fmt.Errorf("cannot transition from %s to %s", c.status.String(), newStatus.String())
	}

	oldStatus := c.status
	c.status = newStatus
	c.updatedAt = time.Now()

	// Add domain event (convert to domain type for event)
	c.addDomainEvent(event.NewCouponStatusChangedEvent(
		fmt.Sprintf("coupon-status-changed-%s-%d", c.id.String(), time.Now().UnixNano()),
		c.id,
		oldStatus,
		newStatus,
		changedBy,
		reason,
	))

	return nil
}

// Expire marks the coupon as expired
func (c *Coupon) Expire() error {
	if c.status.IsExpired() {
		return fmt.Errorf("coupon is already expired")
	}

	c.status = valueobject.CouponStatusExpired
	c.updatedAt = time.Now()

	// Add domain event
	c.addDomainEvent(event.NewCouponExpiredEvent(
		fmt.Sprintf("coupon-expired-%s-%d", c.id.String(), time.Now().UnixNano()),
		c.id,
		c.code,
		time.Now(),
	))

	return nil
}

// SoftDelete marks the coupon as deleted
func (c *Coupon) SoftDelete() {
	if c.deletedAt == nil {
		now := time.Now()
		c.deletedAt = &now
		c.updatedAt = now
	}
}

// UpdateBasicInfo updates basic information
func (c *Coupon) UpdateBasicInfo(name, description string) error {
	if name == "" {
		return fmt.Errorf("coupon name cannot be empty")
	}

	c.name = name
	c.description = description
	c.updatedAt = time.Now()

	return nil
}

// UpdateDiscountValue updates the discount value
func (c *Coupon) UpdateDiscountValue(discountValue valueobject.DiscountValue) {
	c.discountValue = discountValue
	c.updatedAt = time.Now()
}

// UpdateMinOrderAmount updates the minimum order amount 
func (c *Coupon) UpdateMinOrderAmount(minOrderAmount sharedvo.Money) {
	c.minOrderAmount = minOrderAmount
	c.updatedAt = time.Now()
}

// UpdateValidityPeriod updates the validity period
func (c *Coupon) UpdateValidityPeriod(validityPeriod valueobject.ValidityPeriod) {
	c.validityPeriod = validityPeriod
	c.updatedAt = time.Now()
}

// UpdateApplicablePlans updates the applicable plans
func (c *Coupon) UpdateApplicablePlans(applicablePlans []uint64) {
	c.applicablePlans = applicablePlans
	c.updatedAt = time.Now()
}

// UpdateVisibility updates the public visibility
func (c *Coupon) UpdateVisibility(isPublic bool) {
	c.isPublic = isPublic
	c.updatedAt = time.Now()
}

// GetUsages returns all usage records
func (c *Coupon) GetUsages() []*entity.CouponUsage {
	usages := make([]*entity.CouponUsage, 0, len(c.usages))
	for _, usage := range c.usages {
		usages = append(usages, usage)
	}
	return usages
}

// GetUsageByID returns a usage record by ID
func (c *Coupon) GetUsageByID(usageID entity.CouponUsageID) *entity.CouponUsage {
	return c.usages[usageID]
}

// Helper methods

// addDomainEvent adds a domain event
func (c *Coupon) addDomainEvent(event domain.DomainEvent) {
	c.domainEvents = append(c.domainEvents, event)
}

// getUserUsageCount returns the number of times a user has used this coupon
func (c *Coupon) getUserUsageCount(userID sharedvo.UserID) int {
	count := 0
	for _, usage := range c.usages {
		if usage.UserID().Equals(userID) && !usage.IsDeleted() {
			count++
		}
	}
	return count
}

// String returns string representation
func (c *Coupon) String() string {
	return fmt.Sprintf("Coupon{ID:%s, Code:%s, Name:%s, Status:%s}",
		c.id.String(), c.code.String(), c.name, c.status.String())
}

// MarshalJSON implements json.Marshaler for API responses
func (c *Coupon) MarshalJSON() ([]byte, error) {
	type couponJSON struct {
		ID              string                  `json:"id"`
		Code            string                  `json:"code"`
		Name            string                  `json:"name"`
		Description     string                  `json:"description"`
		DiscountValue   valueobject.DiscountValue `json:"discount_value"`
		MinOrderAmount  sharedvo.Money          `json:"min_order_amount"`
		ValidityPeriod  valueobject.ValidityPeriod `json:"validity_period"`
		Status          string                  `json:"status"`
		UsageLimits     valueobject.UsageLimits `json:"usage_limits"`
		ApplicablePlans []uint64               `json:"applicable_plans"`
		IsPublic        bool                   `json:"is_public"`
		CreatedBy       string                 `json:"created_by"`
		CreatedAt       time.Time              `json:"created_at"`
		UpdatedAt       time.Time              `json:"updated_at"`
	}

	return json.Marshal(couponJSON{
		ID:              c.id.String(),
		Code:            c.code.String(),
		Name:            c.name,
		Description:     c.description,
		DiscountValue:   c.discountValue,
		MinOrderAmount:  c.minOrderAmount,
		ValidityPeriod:  c.validityPeriod,
		Status:          c.status.String(),
		UsageLimits:     c.usageLimits,
		ApplicablePlans: c.applicablePlans,
		IsPublic:        c.isPublic,
		CreatedBy:       c.createdBy.String(),
		CreatedAt:       c.createdAt,
		UpdatedAt:       c.updatedAt,
	})
}