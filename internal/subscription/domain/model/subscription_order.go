package model

import (
	"errors"
	"time"

	"linke/internal/subscription/domain/valueobject"
)

type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "pending"
	OrderStatusFulfilled OrderStatus = "fulfilled"
	OrderStatusCancelled OrderStatus = "cancelled"
	OrderStatusRefunded  OrderStatus = "refunded"
)

type SubscriptionOrder struct {
	id               uint
	userID           valueobject.UserID
	planID           valueobject.PlanID
	status           OrderStatus
	billingCycle     valueobject.BillingCycle
	billingInterval  int
	servicePeriod    int
	totalAmount      valueobject.Price
	discountAmount   valueobject.Price
	taxAmount        valueobject.Price
	finalAmount      valueobject.Price
	currency         valueobject.Currency
	couponCode       string
	inviteCode       string
	serviceStartDate *time.Time
	serviceEndDate   *time.Time
	notes            string
	metadata         map[string]interface{}
	paymentMethod    string
	paymentStatus    string
	createdAt        time.Time
	updatedAt        time.Time
	deletedAt        *time.Time
}

func NewSubscriptionOrder(
	userID valueobject.UserID,
	planID valueobject.PlanID,
	billingCycle valueobject.BillingCycle,
	billingInterval int,
	servicePeriod int,
	totalAmount valueobject.Price,
	currency valueobject.Currency,
) (*SubscriptionOrder, error) {
	if billingInterval <= 0 {
		return nil, errors.New("billing interval must be positive")
	}

	if servicePeriod <= 0 {
		return nil, errors.New("service period must be positive")
	}

	zeroPrice, _ := valueobject.NewPrice(0, currency)
	now := time.Now()

	return &SubscriptionOrder{
		userID:          userID,
		planID:          planID,
		status:          OrderStatusPending,
		billingCycle:    billingCycle,
		billingInterval: billingInterval,
		servicePeriod:   servicePeriod,
		totalAmount:     totalAmount,
		discountAmount:  *zeroPrice,
		taxAmount:       *zeroPrice,
		finalAmount:     totalAmount,
		currency:        currency,
		metadata:        make(map[string]interface{}),
		createdAt:       now,
		updatedAt:       now,
	}, nil
}

func (o *SubscriptionOrder) ID() uint {
	return o.id
}

func (o *SubscriptionOrder) SetID(id uint) {
	o.id = id
}

func (o *SubscriptionOrder) UserID() valueobject.UserID {
	return o.userID
}

func (o *SubscriptionOrder) PlanID() valueobject.PlanID {
	return o.planID
}

func (o *SubscriptionOrder) Status() OrderStatus {
	return o.status
}

func (o *SubscriptionOrder) BillingCycle() valueobject.BillingCycle {
	return o.billingCycle
}

func (o *SubscriptionOrder) BillingInterval() int {
	return o.billingInterval
}

func (o *SubscriptionOrder) ServicePeriod() int {
	return o.servicePeriod
}

func (o *SubscriptionOrder) TotalAmount() valueobject.Price {
	return o.totalAmount
}

func (o *SubscriptionOrder) DiscountAmount() valueobject.Price {
	return o.discountAmount
}

func (o *SubscriptionOrder) TaxAmount() valueobject.Price {
	return o.taxAmount
}

func (o *SubscriptionOrder) FinalAmount() valueobject.Price {
	return o.finalAmount
}

func (o *SubscriptionOrder) Currency() valueobject.Currency {
	return o.currency
}

func (o *SubscriptionOrder) CouponCode() string {
	return o.couponCode
}

func (o *SubscriptionOrder) InviteCode() string {
	return o.inviteCode
}

func (o *SubscriptionOrder) ServiceStartDate() *time.Time {
	return o.serviceStartDate
}

func (o *SubscriptionOrder) ServiceEndDate() *time.Time {
	return o.serviceEndDate
}

func (o *SubscriptionOrder) Notes() string {
	return o.notes
}

func (o *SubscriptionOrder) Metadata() map[string]interface{} {
	return o.metadata
}

func (o *SubscriptionOrder) PaymentMethod() string {
	return o.paymentMethod
}

func (o *SubscriptionOrder) PaymentStatus() string {
	return o.paymentStatus
}

func (o *SubscriptionOrder) CreatedAt() time.Time {
	return o.createdAt
}

func (o *SubscriptionOrder) UpdatedAt() time.Time {
	return o.updatedAt
}

func (o *SubscriptionOrder) DeletedAt() *time.Time {
	return o.deletedAt
}

func (o *SubscriptionOrder) IsPending() bool {
	return o.status == OrderStatusPending
}

func (o *SubscriptionOrder) IsFulfilled() bool {
	return o.status == OrderStatusFulfilled
}

func (o *SubscriptionOrder) IsCancelled() bool {
	return o.status == OrderStatusCancelled
}

func (o *SubscriptionOrder) IsRefunded() bool {
	return o.status == OrderStatusRefunded
}

func (o *SubscriptionOrder) IsDeleted() bool {
	return o.deletedAt != nil
}

func (o *SubscriptionOrder) ApplyCoupon(couponCode string, discountAmount valueobject.Price) error {
	if o.status != OrderStatusPending {
		return errors.New("can only apply coupon to pending orders")
	}

	if discountAmount.Currency() != o.currency {
		return errors.New("discount currency must match order currency")
	}

	o.couponCode = couponCode
	o.discountAmount = discountAmount
	o.recalculateFinalAmount()
	o.updatedAt = time.Now()

	return nil
}

func (o *SubscriptionOrder) ApplyInviteCode(inviteCode string) error {
	if o.status != OrderStatusPending {
		return errors.New("can only apply invite code to pending orders")
	}

	o.inviteCode = inviteCode
	o.updatedAt = time.Now()

	return nil
}

func (o *SubscriptionOrder) SetTaxAmount(taxAmount valueobject.Price) error {
	if taxAmount.Currency() != o.currency {
		return errors.New("tax currency must match order currency")
	}

	o.taxAmount = taxAmount
	o.recalculateFinalAmount()
	o.updatedAt = time.Now()

	return nil
}

func (o *SubscriptionOrder) SetServiceDates(startDate, endDate time.Time) error {
	if startDate.After(endDate) {
		return errors.New("service start date cannot be after end date")
	}

	o.serviceStartDate = &startDate
	o.serviceEndDate = &endDate
	o.updatedAt = time.Now()

	return nil
}

func (o *SubscriptionOrder) SetPaymentInfo(method, status string) {
	o.paymentMethod = method
	o.paymentStatus = status
	o.updatedAt = time.Now()
}

func (o *SubscriptionOrder) SetNotes(notes string) {
	o.notes = notes
	o.updatedAt = time.Now()
}

func (o *SubscriptionOrder) SetMetadata(key string, value interface{}) {
	if o.metadata == nil {
		o.metadata = make(map[string]interface{})
	}
	o.metadata[key] = value
	o.updatedAt = time.Now()
}

func (o *SubscriptionOrder) GetMetadata(key string) (interface{}, bool) {
	value, exists := o.metadata[key]
	return value, exists
}

func (o *SubscriptionOrder) Fulfill() error {
	if o.status != OrderStatusPending {
		return errors.New("can only fulfill pending orders")
	}

	o.status = OrderStatusFulfilled
	o.updatedAt = time.Now()

	return nil
}

func (o *SubscriptionOrder) Cancel(reason string) error {
	if o.status == OrderStatusFulfilled {
		return errors.New("cannot cancel fulfilled orders")
	}

	if o.status == OrderStatusCancelled {
		return errors.New("order is already cancelled")
	}

	o.status = OrderStatusCancelled
	o.SetNotes(reason)
	o.updatedAt = time.Now()

	return nil
}

func (o *SubscriptionOrder) Refund(reason string) error {
	if o.status != OrderStatusFulfilled {
		return errors.New("can only refund fulfilled orders")
	}

	o.status = OrderStatusRefunded
	o.SetNotes(reason)
	o.updatedAt = time.Now()

	return nil
}

func (o *SubscriptionOrder) Delete() {
	now := time.Now()
	o.deletedAt = &now
	o.updatedAt = now
}

func (o *SubscriptionOrder) recalculateFinalAmount() {
	finalAmount := o.totalAmount.Amount() - o.discountAmount.Amount() + o.taxAmount.Amount()
	if finalAmount < 0 {
		finalAmount = 0
	}

	finalPrice, _ := valueobject.NewPrice(finalAmount, o.currency)
	o.finalAmount = *finalPrice
}