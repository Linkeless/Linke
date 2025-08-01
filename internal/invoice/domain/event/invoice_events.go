package event

import (
	"crypto/rand"
	"fmt"
	"time"

	"linke/internal/invoice/domain/valueobject"
)

// EventVersion represents the version of events for backward compatibility
const (
	EventVersionV1 = "v1"
)

// generateEventID creates a unique event ID using UUID
func generateEventID() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		// Fallback to timestamp-based ID if UUID generation fails
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

// InvoiceCreated represents an event when an invoice is created
type InvoiceCreated struct {
	ID            string                    `json:"id"`
	Version       string                    `json:"version"`
	InvoiceID     valueobject.InvoiceID     `json:"invoice_id"`
	InvoiceNumber valueobject.InvoiceNumber `json:"invoice_number"`
	UserID        uint                      `json:"user_id"`
	OrderID       uint                      `json:"order_id"`
	InvoiceType   valueobject.InvoiceType   `json:"invoice_type"`
	TotalAmount   valueobject.Money         `json:"total_amount"`
	BillingEmail  string                    `json:"billing_email"`
	DueDate       *time.Time                `json:"due_date,omitempty"`
	Timestamp     time.Time                 `json:"timestamp"`
}

// NewInvoiceCreated creates a new InvoiceCreated event
func NewInvoiceCreated(
	invoiceID valueobject.InvoiceID,
	invoiceNumber valueobject.InvoiceNumber,
	userID uint,
	orderID uint,
	invoiceType valueobject.InvoiceType,
	totalAmount valueobject.Money,
	billingEmail string,
	dueDate *time.Time,
) *InvoiceCreated {
	return &InvoiceCreated{
		ID:            generateEventID(),
		Version:       EventVersionV1,
		InvoiceID:     invoiceID,
		InvoiceNumber: invoiceNumber,
		UserID:        userID,
		OrderID:       orderID,
		InvoiceType:   invoiceType,
		TotalAmount:   totalAmount,
		BillingEmail:  billingEmail,
		DueDate:       dueDate,
		Timestamp:     time.Now().UTC(),
	}
}

// EventID returns a unique event ID
func (e InvoiceCreated) EventID() string {
	return e.ID
}

// EventType returns the event type
func (e InvoiceCreated) EventType() string {
	return "invoice.created"
}

// AggregateID returns the aggregate ID
func (e InvoiceCreated) AggregateID() string {
	return e.InvoiceID.String()
}

// OccurredAt returns when the event occurred
func (e InvoiceCreated) OccurredAt() time.Time {
	return e.Timestamp
}

// EventData returns the event data
func (e InvoiceCreated) EventData() interface{} {
	return e
}

// InvoiceSent represents an event when an invoice is sent to customer
type InvoiceSent struct {
	ID            string                    `json:"id"`
	Version       string                    `json:"version"`
	InvoiceID     valueobject.InvoiceID     `json:"invoice_id"`
	InvoiceNumber valueobject.InvoiceNumber `json:"invoice_number"`
	UserID        uint                      `json:"user_id"`
	BillingEmail  string                    `json:"billing_email"`
	CCEmails      []string                  `json:"cc_emails,omitempty"`
	EmailSubject  string                    `json:"email_subject,omitempty"`
	SentAt        time.Time                 `json:"sent_at"`
	SendCount     int                       `json:"send_count"`
	Timestamp     time.Time                 `json:"timestamp"`
}

// NewInvoiceSent creates a new InvoiceSent event
func NewInvoiceSent(
	invoiceID valueobject.InvoiceID,
	invoiceNumber valueobject.InvoiceNumber,
	userID uint,
	billingEmail string,
	ccEmails []string,
	emailSubject string,
	sendCount int,
) *InvoiceSent {
	return &InvoiceSent{
		ID:            generateEventID(),
		Version:       EventVersionV1,
		InvoiceID:     invoiceID,
		InvoiceNumber: invoiceNumber,
		UserID:        userID,
		BillingEmail:  billingEmail,
		CCEmails:      ccEmails,
		EmailSubject:  emailSubject,
		SentAt:        time.Now().UTC(),
		SendCount:     sendCount,
		Timestamp:     time.Now().UTC(),
	}
}

// EventID returns a unique event ID
func (e InvoiceSent) EventID() string {
	return e.ID
}

// EventType returns the event type
func (e InvoiceSent) EventType() string {
	return "invoice.sent"
}

// AggregateID returns the aggregate ID
func (e InvoiceSent) AggregateID() string {
	return e.InvoiceID.String()
}

// OccurredAt returns when the event occurred
func (e InvoiceSent) OccurredAt() time.Time {
	return e.Timestamp
}

// EventData returns the event data
func (e InvoiceSent) EventData() interface{} {
	return e
}

// InvoicePaid represents an event when an invoice is paid
type InvoicePaid struct {
	ID              string                    `json:"id"`
	Version         string                    `json:"version"`
	InvoiceID       valueobject.InvoiceID     `json:"invoice_id"`
	InvoiceNumber   valueobject.InvoiceNumber `json:"invoice_number"`
	UserID          uint                      `json:"user_id"`
	OrderID         uint                      `json:"order_id"`
	PaidAmount      valueobject.Money         `json:"paid_amount"`
	TotalPaidAmount valueobject.Money         `json:"total_paid_amount"`
	TotalAmount     valueobject.Money         `json:"total_amount"`
	RemainingAmount valueobject.Money         `json:"remaining_amount"`
	IsFullyPaid     bool                      `json:"is_fully_paid"`
	PaymentRef      string                    `json:"payment_reference,omitempty"`
	PaymentMethod   string                    `json:"payment_method,omitempty"`
	PaidAt          time.Time                 `json:"paid_at"`
	Timestamp       time.Time                 `json:"timestamp"`
}

// NewInvoicePaid creates a new InvoicePaid event
func NewInvoicePaid(
	invoiceID valueobject.InvoiceID,
	invoiceNumber valueobject.InvoiceNumber,
	userID uint,
	orderID uint,
	paidAmount valueobject.Money,
	totalPaidAmount valueobject.Money,
	totalAmount valueobject.Money,
	remainingAmount valueobject.Money,
	isFullyPaid bool,
	paymentRef string,
	paymentMethod string,
) *InvoicePaid {
	return &InvoicePaid{
		ID:              generateEventID(),
		Version:         EventVersionV1,
		InvoiceID:       invoiceID,
		InvoiceNumber:   invoiceNumber,
		UserID:          userID,
		OrderID:         orderID,
		PaidAmount:      paidAmount,
		TotalPaidAmount: totalPaidAmount,
		TotalAmount:     totalAmount,
		RemainingAmount: remainingAmount,
		IsFullyPaid:     isFullyPaid,
		PaymentRef:      paymentRef,
		PaymentMethod:   paymentMethod,
		PaidAt:          time.Now().UTC(),
		Timestamp:       time.Now().UTC(),
	}
}

// EventID returns a unique event ID
func (e InvoicePaid) EventID() string {
	return e.ID
}

// EventType returns the event type
func (e InvoicePaid) EventType() string {
	return "invoice.paid"
}

// AggregateID returns the aggregate ID
func (e InvoicePaid) AggregateID() string {
	return e.InvoiceID.String()
}

// OccurredAt returns when the event occurred
func (e InvoicePaid) OccurredAt() time.Time {
	return e.Timestamp
}

// EventData returns the event data
func (e InvoicePaid) EventData() interface{} {
	return e
}

// InvoiceVoided represents an event when an invoice is voided
type InvoiceVoided struct {
	ID            string                    `json:"id"`
	Version       string                    `json:"version"`
	InvoiceID     valueobject.InvoiceID     `json:"invoice_id"`
	InvoiceNumber valueobject.InvoiceNumber `json:"invoice_number"`
	UserID        uint                      `json:"user_id"`
	OrderID       uint                      `json:"order_id"`
	PreviousStatus valueobject.InvoiceStatus `json:"previous_status"`
	Reason        string                    `json:"reason"`
	VoidedBy      uint                      `json:"voided_by,omitempty"`
	VoidedAt      time.Time                 `json:"voided_at"`
	Timestamp     time.Time                 `json:"timestamp"`
}

// NewInvoiceVoided creates a new InvoiceVoided event
func NewInvoiceVoided(
	invoiceID valueobject.InvoiceID,
	invoiceNumber valueobject.InvoiceNumber,
	userID uint,
	orderID uint,
	previousStatus valueobject.InvoiceStatus,
	reason string,
	voidedBy uint,
) *InvoiceVoided {
	return &InvoiceVoided{
		ID:             generateEventID(),
		Version:        EventVersionV1,
		InvoiceID:      invoiceID,
		InvoiceNumber:  invoiceNumber,
		UserID:         userID,
		OrderID:        orderID,
		PreviousStatus: previousStatus,
		Reason:         reason,
		VoidedBy:       voidedBy,
		VoidedAt:       time.Now().UTC(),
		Timestamp:      time.Now().UTC(),
	}
}

// EventID returns a unique event ID
func (e InvoiceVoided) EventID() string {
	return e.ID
}

// EventType returns the event type
func (e InvoiceVoided) EventType() string {
	return "invoice.voided"
}

// AggregateID returns the aggregate ID
func (e InvoiceVoided) AggregateID() string {
	return e.InvoiceID.String()
}

// OccurredAt returns when the event occurred
func (e InvoiceVoided) OccurredAt() time.Time {
	return e.Timestamp
}

// EventData returns the event data
func (e InvoiceVoided) EventData() interface{} {
	return e
}

// InvoiceOverdue represents an event when an invoice becomes overdue
type InvoiceOverdue struct {
	ID              string                    `json:"id"`
	Version         string                    `json:"version"`
	InvoiceID       valueobject.InvoiceID     `json:"invoice_id"`
	InvoiceNumber   valueobject.InvoiceNumber `json:"invoice_number"`
	UserID          uint                      `json:"user_id"`
	OrderID         uint                      `json:"order_id"`
	BillingEmail    string                    `json:"billing_email"`
	DueDate         time.Time                 `json:"due_date"`
	DaysOverdue     int                       `json:"days_overdue"`
	RemainingAmount valueobject.Money         `json:"remaining_amount"`
	TotalAmount     valueobject.Money         `json:"total_amount"`
	IsFirstOverdue  bool                      `json:"is_first_overdue"`
	Timestamp       time.Time                 `json:"timestamp"`
}

// NewInvoiceOverdue creates a new InvoiceOverdue event
func NewInvoiceOverdue(
	invoiceID valueobject.InvoiceID,
	invoiceNumber valueobject.InvoiceNumber,
	userID uint,
	orderID uint,
	billingEmail string,
	dueDate time.Time,
	daysOverdue int,
	remainingAmount valueobject.Money,
	totalAmount valueobject.Money,
	isFirstOverdue bool,
) *InvoiceOverdue {
	return &InvoiceOverdue{
		ID:              generateEventID(),
		Version:         EventVersionV1,
		InvoiceID:       invoiceID,
		InvoiceNumber:   invoiceNumber,
		UserID:          userID,
		OrderID:         orderID,
		BillingEmail:    billingEmail,
		DueDate:         dueDate,
		DaysOverdue:     daysOverdue,
		RemainingAmount: remainingAmount,
		TotalAmount:     totalAmount,
		IsFirstOverdue:  isFirstOverdue,
		Timestamp:       time.Now().UTC(),
	}
}

// EventID returns a unique event ID
func (e InvoiceOverdue) EventID() string {
	return e.ID
}

// EventType returns the event type
func (e InvoiceOverdue) EventType() string {
	return "invoice.overdue"
}

// AggregateID returns the aggregate ID
func (e InvoiceOverdue) AggregateID() string {
	return e.InvoiceID.String()
}

// OccurredAt returns when the event occurred
func (e InvoiceOverdue) OccurredAt() time.Time {
	return e.Timestamp
}

// EventData returns the event data
func (e InvoiceOverdue) EventData() interface{} {
	return e
}

// InvoiceUpdated represents an event when an invoice is updated
type InvoiceUpdated struct {
	ID               string                    `json:"id"`
	Version          string                    `json:"version"`
	InvoiceID        valueobject.InvoiceID     `json:"invoice_id"`
	InvoiceNumber    valueobject.InvoiceNumber `json:"invoice_number"`
	UserID           uint                      `json:"user_id"`
	UpdatedFields    []string                  `json:"updated_fields"`
	PreviousValues   map[string]interface{}    `json:"previous_values,omitempty"`
	NewValues        map[string]interface{}    `json:"new_values,omitempty"`
	UpdatedBy        uint                      `json:"updated_by,omitempty"`
	UpdateReason     string                    `json:"update_reason,omitempty"`
	Timestamp        time.Time                 `json:"timestamp"`
}

// NewInvoiceUpdated creates a new InvoiceUpdated event
func NewInvoiceUpdated(
	invoiceID valueobject.InvoiceID,
	invoiceNumber valueobject.InvoiceNumber,
	userID uint,
	updatedFields []string,
	previousValues map[string]interface{},
	newValues map[string]interface{},
	updatedBy uint,
	updateReason string,
) *InvoiceUpdated {
	return &InvoiceUpdated{
		ID:             generateEventID(),
		Version:        EventVersionV1,
		InvoiceID:      invoiceID,
		InvoiceNumber:  invoiceNumber,
		UserID:         userID,
		UpdatedFields:  updatedFields,
		PreviousValues: previousValues,
		NewValues:      newValues,
		UpdatedBy:      updatedBy,
		UpdateReason:   updateReason,
		Timestamp:      time.Now().UTC(),
	}
}

// EventID returns a unique event ID
func (e InvoiceUpdated) EventID() string {
	return e.ID
}

// EventType returns the event type
func (e InvoiceUpdated) EventType() string {
	return "invoice.updated"
}

// AggregateID returns the aggregate ID
func (e InvoiceUpdated) AggregateID() string {
	return e.InvoiceID.String()
}

// OccurredAt returns when the event occurred
func (e InvoiceUpdated) OccurredAt() time.Time {
	return e.Timestamp
}

// EventData returns the event data
func (e InvoiceUpdated) EventData() interface{} {
	return e
}

// InvoiceDeleted represents an event when an invoice is deleted
type InvoiceDeleted struct {
	ID            string                    `json:"id"`
	Version       string                    `json:"version"`
	InvoiceID     valueobject.InvoiceID     `json:"invoice_id"`
	InvoiceNumber valueobject.InvoiceNumber `json:"invoice_number"`
	UserID        uint                      `json:"user_id"`
	OrderID       uint                      `json:"order_id"`
	DeletedBy     uint                      `json:"deleted_by,omitempty"`
	DeleteReason  string                    `json:"delete_reason,omitempty"`
	PreviousStatus valueobject.InvoiceStatus `json:"previous_status"`
	DeletedAt     time.Time                 `json:"deleted_at"`
	Timestamp     time.Time                 `json:"timestamp"`
}

// NewInvoiceDeleted creates a new InvoiceDeleted event
func NewInvoiceDeleted(
	invoiceID valueobject.InvoiceID,
	invoiceNumber valueobject.InvoiceNumber,
	userID uint,
	orderID uint,
	deletedBy uint,
	deleteReason string,
	previousStatus valueobject.InvoiceStatus,
) *InvoiceDeleted {
	return &InvoiceDeleted{
		ID:             generateEventID(),
		Version:        EventVersionV1,
		InvoiceID:      invoiceID,
		InvoiceNumber:  invoiceNumber,
		UserID:         userID,
		OrderID:        orderID,
		DeletedBy:      deletedBy,
		DeleteReason:   deleteReason,
		PreviousStatus: previousStatus,
		DeletedAt:      time.Now().UTC(),
		Timestamp:      time.Now().UTC(),
	}
}

// EventID returns a unique event ID
func (e InvoiceDeleted) EventID() string {
	return e.ID
}

// EventType returns the event type
func (e InvoiceDeleted) EventType() string {
	return "invoice.deleted"
}

// AggregateID returns the aggregate ID
func (e InvoiceDeleted) AggregateID() string {
	return e.InvoiceID.String()
}

// OccurredAt returns when the event occurred
func (e InvoiceDeleted) OccurredAt() time.Time {
	return e.Timestamp
}

// EventData returns the event data
func (e InvoiceDeleted) EventData() interface{} {
	return e
}

// InvoicePartiallyPaid represents an event when an invoice receives a partial payment
type InvoicePartiallyPaid struct {
	ID              string                    `json:"id"`
	Version         string                    `json:"version"`
	InvoiceID       valueobject.InvoiceID     `json:"invoice_id"`
	InvoiceNumber   valueobject.InvoiceNumber `json:"invoice_number"`
	UserID          uint                      `json:"user_id"`
	OrderID         uint                      `json:"order_id"`
	PartialAmount   valueobject.Money         `json:"partial_amount"`
	TotalPaidAmount valueobject.Money         `json:"total_paid_amount"`
	RemainingAmount valueobject.Money         `json:"remaining_amount"`
	TotalAmount     valueobject.Money         `json:"total_amount"`
	PaymentRef      string                    `json:"payment_reference,omitempty"`
	PaymentMethod   string                    `json:"payment_method,omitempty"`
	PaidAt          time.Time                 `json:"paid_at"`
	Timestamp       time.Time                 `json:"timestamp"`
}

// NewInvoicePartiallyPaid creates a new InvoicePartiallyPaid event
func NewInvoicePartiallyPaid(
	invoiceID valueobject.InvoiceID,
	invoiceNumber valueobject.InvoiceNumber,
	userID uint,
	orderID uint,
	partialAmount valueobject.Money,
	totalPaidAmount valueobject.Money,
	remainingAmount valueobject.Money,
	totalAmount valueobject.Money,
	paymentRef string,
	paymentMethod string,
) *InvoicePartiallyPaid {
	return &InvoicePartiallyPaid{
		ID:              generateEventID(),
		Version:         EventVersionV1,
		InvoiceID:       invoiceID,
		InvoiceNumber:   invoiceNumber,
		UserID:          userID,
		OrderID:         orderID,
		PartialAmount:   partialAmount,
		TotalPaidAmount: totalPaidAmount,
		RemainingAmount: remainingAmount,
		TotalAmount:     totalAmount,
		PaymentRef:      paymentRef,
		PaymentMethod:   paymentMethod,
		PaidAt:          time.Now().UTC(),
		Timestamp:       time.Now().UTC(),
	}
}

// EventID returns a unique event ID
func (e InvoicePartiallyPaid) EventID() string {
	return e.ID
}

// EventType returns the event type
func (e InvoicePartiallyPaid) EventType() string {
	return "invoice.partially_paid"
}

// AggregateID returns the aggregate ID
func (e InvoicePartiallyPaid) AggregateID() string {
	return e.InvoiceID.String()
}

// OccurredAt returns when the event occurred
func (e InvoicePartiallyPaid) OccurredAt() time.Time {
	return e.Timestamp
}

// EventData returns the event data
func (e InvoicePartiallyPaid) EventData() interface{} {
	return e
}