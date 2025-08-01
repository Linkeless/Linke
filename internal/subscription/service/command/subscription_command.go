package command

import (
	"time"

	"linke/internal/subscription/domain/valueobject"
)

type CreateSubscriptionCommand struct {
	UserID             valueobject.UserID
	PlanID             valueobject.PlanID
	OrderID            uint
	StartDate          time.Time
	EndDate            time.Time
	BillingCycle       valueobject.BillingCycle
	BillingInterval    int
	Price              valueobject.Price
	AutoRenew          bool
	TrialEndDate       *time.Time
	ServerGroupIDs     []uint
	Notes              string
}

type RenewSubscriptionCommand struct {
	SubscriptionID     valueobject.SubscriptionID
	NewPeriodStart     time.Time
	NewPeriodEnd       time.Time
	NextBillingDate    *time.Time
}

type CancelSubscriptionCommand struct {
	SubscriptionID     valueobject.SubscriptionID
	Reason             string
	Immediately        bool
}

type PauseSubscriptionCommand struct {
	SubscriptionID     valueobject.SubscriptionID
	Reason             string
}

type ResumeSubscriptionCommand struct {
	SubscriptionID     valueobject.SubscriptionID
}

type UpdateSubscriptionUsageCommand struct {
	SubscriptionID     valueobject.SubscriptionID
}

type UpdateSubscriptionServerGroupsCommand struct {
	SubscriptionID     valueobject.SubscriptionID
	ServerGroupIDs     []uint
}

type UpdateSubscriptionNotesCommand struct {
	SubscriptionID     valueobject.SubscriptionID
	Notes              string
}

type FailSubscriptionRenewalCommand struct {
	SubscriptionID     valueobject.SubscriptionID
	Reason             string
}

type ExpireSubscriptionCommand struct {
	SubscriptionID     valueobject.SubscriptionID
}

type DeleteSubscriptionCommand struct {
	SubscriptionID     valueobject.SubscriptionID
}