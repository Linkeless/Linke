package aggregate

import (
	"fmt"
	"time"

	"linke/internal/payment/domain/event"
	"linke/internal/payment/domain/valueobject"
	sharedvo "linke/internal/shared/valueobject"
	"linke/internal/shared/domain"
)

// Payment represents the payment aggregate root
type Payment struct {
	// Identity and basic information
	id            valueobject.PaymentID
	paymentNumber valueobject.PaymentNumber
	
	// Related entities
	invoiceID sharedvo.InvoiceID
	userID    sharedvo.UserID
	
	// Payment details
	amount         sharedvo.Money
	status         valueobject.PaymentStatus
	paymentMethod  valueobject.PaymentMethod
	paymentGateway valueobject.PaymentGateway
	
	// Gateway information
	paymentIntentID      string
	gatewayTransactionID string
	gatewayFee          sharedvo.Money
	
	// Payment URLs and details
	paymentURL  string
	qrCodeURL   string
	redirectURL string
	
	// Time information
	expiresAt   *time.Time
	processedAt *time.Time
	completedAt *time.Time
	
	// Refund information
	refundAmount    sharedvo.Money
	refundedAt      *time.Time
	refundReason    string
	refundReference valueobject.PaymentNumber
	
	// Notification and webhook data
	webhookData          string
	notificationCount    int
	lastNotificationAt   *time.Time
	
	// Business metadata
	notes    string
	metadata string
	
	// Audit fields
	createdAt time.Time
	updatedAt time.Time
	deletedAt *time.Time
	
	// Domain events
	domainEvents []domain.DomainEvent
}

// NewPayment creates a new payment aggregate
func NewPayment(
	paymentNumber valueobject.PaymentNumber,
	invoiceID sharedvo.InvoiceID,
	userID sharedvo.UserID,
	amount sharedvo.Money,
	paymentMethod valueobject.PaymentMethod,
	paymentGateway valueobject.PaymentGateway,
) (*Payment, error) {
	
	if paymentNumber.IsEmpty() {
		return nil, fmt.Errorf("payment number cannot be empty")
	}
	
	if invoiceID.IsZero() {
		return nil, fmt.Errorf("invoice ID cannot be zero")
	}
	
	if userID.IsZero() {
		return nil, fmt.Errorf("user ID cannot be zero")
	}
	
	if !amount.IsPositive() {
		return nil, fmt.Errorf("payment amount must be positive")
	}
	
	if paymentMethod.IsEmpty() {
		return nil, fmt.Errorf("payment method cannot be empty")
	}
	
	if paymentGateway.IsEmpty() {
		return nil, fmt.Errorf("payment gateway cannot be empty")
	}
	
	now := time.Now()
	payment := &Payment{
		paymentNumber:  paymentNumber,
		invoiceID:      invoiceID,
		userID:         userID,
		amount:         amount,
		status:         valueobject.NewPendingPaymentStatus(),
		paymentMethod:  paymentMethod,
		paymentGateway: paymentGateway,
		refundAmount:   sharedvo.NewZeroMoney(amount.Currency()),
		gatewayFee:     sharedvo.NewZeroMoney(amount.Currency()),
		createdAt:      now,
		updatedAt:      now,
		domainEvents:   make([]domain.DomainEvent, 0),
	}
	
	// Add domain event
	event := event.NewPaymentCreatedEvent(
		paymentNumber,
		invoiceID,
		userID,
		amount,
		paymentMethod,
		paymentGateway,
		now,
	)
	payment.AddDomainEvent(event)
	
	return payment, nil
}

// Factory method for loading from persistence
func LoadPayment(
	id valueobject.PaymentID,
	paymentNumber valueobject.PaymentNumber,
	invoiceID sharedvo.InvoiceID,
	userID sharedvo.UserID,
	amount sharedvo.Money,
	status valueobject.PaymentStatus,
	paymentMethod valueobject.PaymentMethod,
	paymentGateway valueobject.PaymentGateway,
	paymentIntentID string,
	gatewayTransactionID string,
	gatewayFee sharedvo.Money,
	paymentURL string,
	qrCodeURL string,
	redirectURL string,
	expiresAt *time.Time,
	processedAt *time.Time,
	completedAt *time.Time,
	refundAmount sharedvo.Money,
	refundedAt *time.Time,
	refundReason string,
	refundReference valueobject.PaymentNumber,
	webhookData string,
	notificationCount int,
	lastNotificationAt *time.Time,
	notes string,
	metadata string,
	createdAt time.Time,
	updatedAt time.Time,
	deletedAt *time.Time,
) *Payment {
	return &Payment{
		id:                   id,
		paymentNumber:        paymentNumber,
		invoiceID:            invoiceID,
		userID:               userID,
		amount:               amount,
		status:               status,
		paymentMethod:        paymentMethod,
		paymentGateway:       paymentGateway,
		paymentIntentID:      paymentIntentID,
		gatewayTransactionID: gatewayTransactionID,
		gatewayFee:          gatewayFee,
		paymentURL:          paymentURL,
		qrCodeURL:           qrCodeURL,
		redirectURL:         redirectURL,
		expiresAt:           expiresAt,
		processedAt:         processedAt,
		completedAt:         completedAt,
		refundAmount:        refundAmount,
		refundedAt:          refundedAt,
		refundReason:        refundReason,
		refundReference:     refundReference,
		webhookData:         webhookData,
		notificationCount:   notificationCount,
		lastNotificationAt:  lastNotificationAt,
		notes:               notes,
		metadata:            metadata,
		createdAt:           createdAt,
		updatedAt:           updatedAt,
		deletedAt:           deletedAt,
		domainEvents:        make([]domain.DomainEvent, 0),
	}
}

// Aggregate root interface implementation
func (p *Payment) ID() valueobject.PaymentID {
	return p.id
}

func (p *Payment) DomainEvents() []domain.DomainEvent {
	return p.domainEvents
}

func (p *Payment) ClearDomainEvents() {
	p.domainEvents = make([]domain.DomainEvent, 0)
}

func (p *Payment) AddDomainEvent(event domain.DomainEvent) {
	p.domainEvents = append(p.domainEvents, event)
}

func (p *Payment) IsDeleted() bool {
	return p.deletedAt != nil
}

// Getters
func (p *Payment) PaymentNumber() valueobject.PaymentNumber {
	return p.paymentNumber
}

func (p *Payment) InvoiceID() sharedvo.InvoiceID {
	return p.invoiceID
}

func (p *Payment) UserID() sharedvo.UserID {
	return p.userID
}

func (p *Payment) Amount() sharedvo.Money {
	return p.amount
}

func (p *Payment) Status() valueobject.PaymentStatus {
	return p.status
}

func (p *Payment) PaymentMethod() valueobject.PaymentMethod {
	return p.paymentMethod
}

func (p *Payment) PaymentGateway() valueobject.PaymentGateway {
	return p.paymentGateway
}

func (p *Payment) PaymentIntentID() string {
	return p.paymentIntentID
}

func (p *Payment) GatewayTransactionID() string {
	return p.gatewayTransactionID
}

func (p *Payment) GatewayFee() sharedvo.Money {
	return p.gatewayFee
}

func (p *Payment) PaymentURL() string {
	return p.paymentURL
}

func (p *Payment) QRCodeURL() string {
	return p.qrCodeURL
}

func (p *Payment) RedirectURL() string {
	return p.redirectURL
}

func (p *Payment) ExpiresAt() *time.Time {
	return p.expiresAt
}

func (p *Payment) ProcessedAt() *time.Time {
	return p.processedAt
}

func (p *Payment) CompletedAt() *time.Time {
	return p.completedAt
}

func (p *Payment) RefundAmount() sharedvo.Money {
	return p.refundAmount
}

func (p *Payment) RefundedAt() *time.Time {
	return p.refundedAt
}

func (p *Payment) RefundReason() string {
	return p.refundReason
}

func (p *Payment) RefundReference() valueobject.PaymentNumber {
	return p.refundReference
}

func (p *Payment) WebhookData() string {
	return p.webhookData
}

func (p *Payment) NotificationCount() int {
	return p.notificationCount
}

func (p *Payment) LastNotificationAt() *time.Time {
	return p.lastNotificationAt
}

func (p *Payment) Notes() string {
	return p.notes
}

func (p *Payment) Metadata() string {
	return p.metadata
}

func (p *Payment) CreatedAt() time.Time {
	return p.createdAt
}

func (p *Payment) UpdatedAt() time.Time {
	return p.updatedAt
}

func (p *Payment) DeletedAt() *time.Time {
	return p.deletedAt
}

// Business methods

// SetPaymentDetails sets the payment gateway details
func (p *Payment) SetPaymentDetails(paymentIntentID, paymentURL, qrCodeURL, redirectURL string, expiresAt *time.Time) error {
	if p.status.IsFinished() {
		return fmt.Errorf("cannot set payment details for finished payment")
	}
	
	p.paymentIntentID = paymentIntentID
	p.paymentURL = paymentURL
	p.qrCodeURL = qrCodeURL
	p.redirectURL = redirectURL
	p.expiresAt = expiresAt
	p.updatedAt = time.Now()
	
	return nil
}

// SetGatewayTransaction sets the gateway transaction details
func (p *Payment) SetGatewayTransaction(transactionID string, fee sharedvo.Money) error {
	if p.status.IsFinished() {
		return fmt.Errorf("cannot set gateway transaction for finished payment")
	}
	
	// Validate currency matches
	if !p.amount.Currency().Equals(fee.Currency()) {
		return fmt.Errorf("gateway fee currency must match payment currency")
	}
	
	p.gatewayTransactionID = transactionID
	p.gatewayFee = fee
	p.updatedAt = time.Now()
	
	return nil
}

// Process starts payment processing
func (p *Payment) Process() error {
	if !p.status.CanTransitionTo(valueobject.NewProcessingPaymentStatus()) {
		return fmt.Errorf("cannot process payment in status: %s", p.status.String())
	}
	
	now := time.Now()
	p.status = valueobject.NewProcessingPaymentStatus()
	p.processedAt = &now
	p.updatedAt = now
	
	// Add domain event
	event := event.NewPaymentProcessingEvent(
		p.paymentNumber,
		p.invoiceID,
		p.userID,
		now,
	)
	p.AddDomainEvent(event)
	
	return nil
}

// Complete marks payment as completed
func (p *Payment) Complete(webhookData string) error {
	if !p.status.CanTransitionTo(valueobject.NewCompletedPaymentStatus()) {
		return fmt.Errorf("cannot complete payment in status: %s", p.status.String())
	}
	
	now := time.Now()
	p.status = valueobject.NewCompletedPaymentStatus()
	p.completedAt = &now
	p.webhookData = webhookData
	p.updatedAt = now
	
	// Add domain event
	event := event.NewPaymentCompletedEvent(
		p.paymentNumber,
		p.invoiceID,
		p.userID,
		p.amount,
		p.paymentMethod,
		p.paymentGateway,
		now,
	)
	p.AddDomainEvent(event)
	
	return nil
}

// Fail marks payment as failed
func (p *Payment) Fail(reason, webhookData string) error {
	if !p.status.CanTransitionTo(valueobject.NewFailedPaymentStatus()) {
		return fmt.Errorf("cannot fail payment in status: %s", p.status.String())
	}
	
	now := time.Now()
	p.status = valueobject.NewFailedPaymentStatus()
	p.notes = reason
	p.webhookData = webhookData
	p.updatedAt = now
	
	// Add domain event
	event := event.NewPaymentFailedEvent(
		p.paymentNumber,
		p.invoiceID,
		p.userID,
		p.amount,
		reason,
		now,
	)
	p.AddDomainEvent(event)
	
	return nil
}

// Cancel marks payment as cancelled
func (p *Payment) Cancel(reason string) error {
	if !p.status.CanTransitionTo(valueobject.NewCancelledPaymentStatus()) {
		return fmt.Errorf("cannot cancel payment in status: %s", p.status.String())
	}
	
	now := time.Now()
	p.status = valueobject.NewCancelledPaymentStatus()
	p.notes = reason
	p.updatedAt = now
	
	// Add domain event
	event := event.NewPaymentCancelledEvent(
		p.paymentNumber,
		p.invoiceID,
		p.userID,
		p.amount,
		reason,
		now,
	)
	p.AddDomainEvent(event)
	
	return nil
}

// AddRefund adds a refund to this payment
func (p *Payment) AddRefund(refundAmount sharedvo.Money, reason string, refundNumber valueobject.PaymentNumber) error {
	if !p.status.IsCompleted() {
		return fmt.Errorf("can only refund completed payments")
	}
	
	// Validate currency matches
	if !p.amount.Currency().Equals(refundAmount.Currency()) {
		return fmt.Errorf("refund currency must match payment currency")
	}
	
	// Check refund amount doesn't exceed available amount
	totalRefund, err := p.refundAmount.Add(refundAmount)
	if err != nil {
		return fmt.Errorf("failed to calculate total refund: %w", err)
	}
	
	if greater, err := totalRefund.IsGreaterThan(p.amount); err != nil {
		return fmt.Errorf("failed to compare refund amounts: %w", err)
	} else if greater {
		return fmt.Errorf("total refund amount cannot exceed payment amount")
	}
	
	now := time.Now()
	p.refundAmount = totalRefund
	p.refundReason = reason
	p.refundReference = refundNumber
	p.updatedAt = now
	
	// If fully refunded, set refunded timestamp
	if totalRefund.Equals(p.amount) {
		p.refundedAt = &now
	}
	
	// Add domain event
	event := event.NewPaymentRefundedEvent(
		p.paymentNumber,
		p.invoiceID,
		p.userID,
		refundAmount,
		reason,
		now,
	)
	p.AddDomainEvent(event)
	
	return nil
}

// IncrementNotificationCount increments the notification counter
func (p *Payment) IncrementNotificationCount() {
	now := time.Now()
	p.notificationCount++
	p.lastNotificationAt = &now
	p.updatedAt = now
}

// UpdateWebhookData updates the webhook data
func (p *Payment) UpdateWebhookData(webhookData string) {
	p.webhookData = webhookData
	p.updatedAt = time.Now()
}

// SetNotes sets the payment notes
func (p *Payment) SetNotes(notes string) {
	p.notes = notes
	p.updatedAt = time.Now()
}

// SetMetadata sets the payment metadata
func (p *Payment) SetMetadata(metadata string) {
	p.metadata = metadata
	p.updatedAt = time.Now()
}

// Business queries

// IsExpired checks if the payment has expired
func (p *Payment) IsExpired() bool {
	return p.expiresAt != nil && p.expiresAt.Before(time.Now())
}

// CanBeRefunded checks if the payment can be refunded
func (p *Payment) CanBeRefunded() bool {
	if !p.status.IsCompleted() {
		return false
	}
	
	// Check if there's remaining amount to refund
	if greater, err := p.refundAmount.IsGreaterThanOrEqual(p.amount); err != nil || greater {
		return false
	}
	
	return true
}

// GetRefundableAmount returns the amount that can be refunded
func (p *Payment) GetRefundableAmount() (sharedvo.Money, error) {
	if !p.CanBeRefunded() {
		return sharedvo.NewZeroMoney(p.amount.Currency()), nil
	}
	
	return p.amount.Subtract(p.refundAmount)
}

// GetNetAmount returns the net amount after gateway fees
func (p *Payment) GetNetAmount() (sharedvo.Money, error) {
	return p.amount.Subtract(p.gatewayFee)
}

// IsFullyRefunded checks if the payment is fully refunded
func (p *Payment) IsFullyRefunded() bool {
	return p.refundAmount.Equals(p.amount)
}

// SoftDelete marks the payment as deleted
func (p *Payment) SoftDelete() error {
	if p.IsDeleted() {
		return fmt.Errorf("payment is already deleted")
	}
	
	now := time.Now()
	p.deletedAt = &now
	p.updatedAt = now
	
	return nil
}