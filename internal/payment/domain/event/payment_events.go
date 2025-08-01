package event

import (
	"time"

	"linke/internal/payment/domain/valueobject"
	sharedvo "linke/internal/shared/valueobject"
	"linke/internal/shared/domain"
)

// PaymentCreatedEvent represents the event when a payment is created
type PaymentCreatedEvent struct {
	eventID        string
	occurredAt     time.Time
	paymentNumber  valueobject.PaymentNumber
	invoiceID      sharedvo.InvoiceID
	userID         sharedvo.UserID
	amount         sharedvo.Money
	paymentMethod  valueobject.PaymentMethod
	paymentGateway valueobject.PaymentGateway
}

// NewPaymentCreatedEvent creates a new PaymentCreatedEvent
func NewPaymentCreatedEvent(
	paymentNumber valueobject.PaymentNumber,
	invoiceID sharedvo.InvoiceID,
	userID sharedvo.UserID,
	amount sharedvo.Money,
	paymentMethod valueobject.PaymentMethod,
	paymentGateway valueobject.PaymentGateway,
	occurredAt time.Time,
) *PaymentCreatedEvent {
	return &PaymentCreatedEvent{
		eventID:        domain.NewEventID(),
		occurredAt:     occurredAt,
		paymentNumber:  paymentNumber,
		invoiceID:      invoiceID,
		userID:         userID,
		amount:         amount,
		paymentMethod:  paymentMethod,
		paymentGateway: paymentGateway,
	}
}

func (e *PaymentCreatedEvent) EventID() string {
	return e.eventID
}

func (e *PaymentCreatedEvent) OccurredAt() time.Time {
	return e.occurredAt
}

func (e *PaymentCreatedEvent) EventType() string {
	return "PaymentCreated"
}

func (e *PaymentCreatedEvent) PaymentNumber() valueobject.PaymentNumber {
	return e.paymentNumber
}

func (e *PaymentCreatedEvent) InvoiceID() sharedvo.InvoiceID {
	return e.invoiceID
}

func (e *PaymentCreatedEvent) UserID() sharedvo.UserID {
	return e.userID
}

func (e *PaymentCreatedEvent) Amount() sharedvo.Money {
	return e.amount
}

func (e *PaymentCreatedEvent) PaymentMethod() valueobject.PaymentMethod {
	return e.paymentMethod
}

func (e *PaymentCreatedEvent) PaymentGateway() valueobject.PaymentGateway {
	return e.paymentGateway
}

func (e *PaymentCreatedEvent) AggregateID() string {
	return e.paymentNumber.String()
}

func (e *PaymentCreatedEvent) EventData() interface{} {
	return e
}

// PaymentProcessingEvent represents the event when a payment starts processing
type PaymentProcessingEvent struct {
	eventID       string
	occurredAt    time.Time
	paymentNumber valueobject.PaymentNumber
	invoiceID     sharedvo.InvoiceID
	userID        sharedvo.UserID
}

// NewPaymentProcessingEvent creates a new PaymentProcessingEvent
func NewPaymentProcessingEvent(
	paymentNumber valueobject.PaymentNumber,
	invoiceID sharedvo.InvoiceID,
	userID sharedvo.UserID,
	occurredAt time.Time,
) *PaymentProcessingEvent {
	return &PaymentProcessingEvent{
		eventID:       domain.NewEventID(),
		occurredAt:    occurredAt,
		paymentNumber: paymentNumber,
		invoiceID:     invoiceID,
		userID:        userID,
	}
}

func (e *PaymentProcessingEvent) EventID() string {
	return e.eventID
}

func (e *PaymentProcessingEvent) OccurredAt() time.Time {
	return e.occurredAt
}

func (e *PaymentProcessingEvent) EventType() string {
	return "PaymentProcessing"
}

func (e *PaymentProcessingEvent) PaymentNumber() valueobject.PaymentNumber {
	return e.paymentNumber
}

func (e *PaymentProcessingEvent) InvoiceID() sharedvo.InvoiceID {
	return e.invoiceID
}

func (e *PaymentProcessingEvent) UserID() sharedvo.UserID {
	return e.userID
}

func (e *PaymentProcessingEvent) AggregateID() string {
	return e.paymentNumber.String()
}

func (e *PaymentProcessingEvent) EventData() interface{} {
	return e
}

// PaymentCompletedEvent represents the event when a payment is completed
type PaymentCompletedEvent struct {
	eventID        string
	occurredAt     time.Time
	paymentNumber  valueobject.PaymentNumber
	invoiceID      sharedvo.InvoiceID
	userID         sharedvo.UserID
	amount         sharedvo.Money
	paymentMethod  valueobject.PaymentMethod
	paymentGateway valueobject.PaymentGateway
}

// NewPaymentCompletedEvent creates a new PaymentCompletedEvent
func NewPaymentCompletedEvent(
	paymentNumber valueobject.PaymentNumber,
	invoiceID sharedvo.InvoiceID,
	userID sharedvo.UserID,
	amount sharedvo.Money,
	paymentMethod valueobject.PaymentMethod,
	paymentGateway valueobject.PaymentGateway,
	occurredAt time.Time,
) *PaymentCompletedEvent {
	return &PaymentCompletedEvent{
		eventID:        domain.NewEventID(),
		occurredAt:     occurredAt,
		paymentNumber:  paymentNumber,
		invoiceID:      invoiceID,
		userID:         userID,
		amount:         amount,
		paymentMethod:  paymentMethod,
		paymentGateway: paymentGateway,
	}
}

func (e *PaymentCompletedEvent) EventID() string {
	return e.eventID
}

func (e *PaymentCompletedEvent) OccurredAt() time.Time {
	return e.occurredAt
}

func (e *PaymentCompletedEvent) EventType() string {
	return "PaymentCompleted"
}

func (e *PaymentCompletedEvent) PaymentNumber() valueobject.PaymentNumber {
	return e.paymentNumber
}

func (e *PaymentCompletedEvent) InvoiceID() sharedvo.InvoiceID {
	return e.invoiceID
}

func (e *PaymentCompletedEvent) UserID() sharedvo.UserID {
	return e.userID
}

func (e *PaymentCompletedEvent) Amount() sharedvo.Money {
	return e.amount
}

func (e *PaymentCompletedEvent) PaymentMethod() valueobject.PaymentMethod {
	return e.paymentMethod
}

func (e *PaymentCompletedEvent) PaymentGateway() valueobject.PaymentGateway {
	return e.paymentGateway
}

func (e *PaymentCompletedEvent) AggregateID() string {
	return e.paymentNumber.String()
}

func (e *PaymentCompletedEvent) EventData() interface{} {
	return e
}

// PaymentFailedEvent represents the event when a payment fails
type PaymentFailedEvent struct {
	eventID       string
	occurredAt    time.Time
	paymentNumber valueobject.PaymentNumber
	invoiceID     sharedvo.InvoiceID
	userID        sharedvo.UserID
	amount        sharedvo.Money
	reason        string
}

// NewPaymentFailedEvent creates a new PaymentFailedEvent
func NewPaymentFailedEvent(
	paymentNumber valueobject.PaymentNumber,
	invoiceID sharedvo.InvoiceID,
	userID sharedvo.UserID,
	amount sharedvo.Money,
	reason string,
	occurredAt time.Time,
) *PaymentFailedEvent {
	return &PaymentFailedEvent{
		eventID:       domain.NewEventID(),
		occurredAt:    occurredAt,
		paymentNumber: paymentNumber,
		invoiceID:     invoiceID,
		userID:        userID,
		amount:        amount,
		reason:        reason,
	}
}

func (e *PaymentFailedEvent) EventID() string {
	return e.eventID
}

func (e *PaymentFailedEvent) OccurredAt() time.Time {
	return e.occurredAt
}

func (e *PaymentFailedEvent) EventType() string {
	return "PaymentFailed"
}

func (e *PaymentFailedEvent) PaymentNumber() valueobject.PaymentNumber {
	return e.paymentNumber
}

func (e *PaymentFailedEvent) InvoiceID() sharedvo.InvoiceID {
	return e.invoiceID
}

func (e *PaymentFailedEvent) UserID() sharedvo.UserID {
	return e.userID
}

func (e *PaymentFailedEvent) Amount() sharedvo.Money {
	return e.amount
}

func (e *PaymentFailedEvent) Reason() string {
	return e.reason
}

func (e *PaymentFailedEvent) AggregateID() string {
	return e.paymentNumber.String()
}

func (e *PaymentFailedEvent) EventData() interface{} {
	return e
}

// PaymentCancelledEvent represents the event when a payment is cancelled
type PaymentCancelledEvent struct {
	eventID       string
	occurredAt    time.Time
	paymentNumber valueobject.PaymentNumber
	invoiceID     sharedvo.InvoiceID
	userID        sharedvo.UserID
	amount        sharedvo.Money
	reason        string
}

// NewPaymentCancelledEvent creates a new PaymentCancelledEvent
func NewPaymentCancelledEvent(
	paymentNumber valueobject.PaymentNumber,
	invoiceID sharedvo.InvoiceID,
	userID sharedvo.UserID,
	amount sharedvo.Money,
	reason string,
	occurredAt time.Time,
) *PaymentCancelledEvent {
	return &PaymentCancelledEvent{
		eventID:       domain.NewEventID(),
		occurredAt:    occurredAt,
		paymentNumber: paymentNumber,
		invoiceID:     invoiceID,
		userID:        userID,
		amount:        amount,
		reason:        reason,
	}
}

func (e *PaymentCancelledEvent) EventID() string {
	return e.eventID
}

func (e *PaymentCancelledEvent) OccurredAt() time.Time {
	return e.occurredAt
}

func (e *PaymentCancelledEvent) EventType() string {
	return "PaymentCancelled"
}

func (e *PaymentCancelledEvent) PaymentNumber() valueobject.PaymentNumber {
	return e.paymentNumber
}

func (e *PaymentCancelledEvent) InvoiceID() sharedvo.InvoiceID {
	return e.invoiceID
}

func (e *PaymentCancelledEvent) UserID() sharedvo.UserID {
	return e.userID
}

func (e *PaymentCancelledEvent) Amount() sharedvo.Money {
	return e.amount
}

func (e *PaymentCancelledEvent) Reason() string {
	return e.reason
}

func (e *PaymentCancelledEvent) AggregateID() string {
	return e.paymentNumber.String()
}

func (e *PaymentCancelledEvent) EventData() interface{} {
	return e
}

// PaymentRefundedEvent represents the event when a payment is refunded
type PaymentRefundedEvent struct {
	eventID       string
	occurredAt    time.Time
	paymentNumber valueobject.PaymentNumber
	invoiceID     sharedvo.InvoiceID
	userID        sharedvo.UserID
	refundAmount  sharedvo.Money
	reason        string
}

// NewPaymentRefundedEvent creates a new PaymentRefundedEvent
func NewPaymentRefundedEvent(
	paymentNumber valueobject.PaymentNumber,
	invoiceID sharedvo.InvoiceID,
	userID sharedvo.UserID,
	refundAmount sharedvo.Money,
	reason string,
	occurredAt time.Time,
) *PaymentRefundedEvent {
	return &PaymentRefundedEvent{
		eventID:       domain.NewEventID(),
		occurredAt:    occurredAt,
		paymentNumber: paymentNumber,
		invoiceID:     invoiceID,
		userID:        userID,
		refundAmount:  refundAmount,
		reason:        reason,
	}
}

func (e *PaymentRefundedEvent) EventID() string {
	return e.eventID
}

func (e *PaymentRefundedEvent) OccurredAt() time.Time {
	return e.occurredAt
}

func (e *PaymentRefundedEvent) EventType() string {
	return "PaymentRefunded"
}

func (e *PaymentRefundedEvent) PaymentNumber() valueobject.PaymentNumber {
	return e.paymentNumber
}

func (e *PaymentRefundedEvent) InvoiceID() sharedvo.InvoiceID {
	return e.invoiceID
}

func (e *PaymentRefundedEvent) UserID() sharedvo.UserID {
	return e.userID
}

func (e *PaymentRefundedEvent) RefundAmount() sharedvo.Money {
	return e.refundAmount
}

func (e *PaymentRefundedEvent) Reason() string {
	return e.reason
}

func (e *PaymentRefundedEvent) AggregateID() string {
	return e.paymentNumber.String()
}

func (e *PaymentRefundedEvent) EventData() interface{} {
	return e
}