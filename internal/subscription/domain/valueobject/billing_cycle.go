package valueobject

import "errors"

type BillingCycle string

const (
	Monthly  BillingCycle = "monthly"
	Yearly   BillingCycle = "yearly"
	Lifetime BillingCycle = "lifetime"
)

func NewBillingCycle(value string) (*BillingCycle, error) {
	cycle := BillingCycle(value)
	if !cycle.IsValid() {
		return nil, errors.New("invalid billing cycle")
	}
	return &cycle, nil
}

func (b BillingCycle) IsValid() bool {
	switch b {
	case Monthly, Yearly, Lifetime:
		return true
	default:
		return false
	}
}

func (b BillingCycle) String() string {
	return string(b)
}

func (b BillingCycle) IsRecurring() bool {
	return b != Lifetime
}

func (b BillingCycle) IsLifetime() bool {
	return b == Lifetime
}