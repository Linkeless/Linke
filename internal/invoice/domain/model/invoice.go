package model

import (
	"fmt"
	"time"

	"linke/internal/invoice/domain/event"
	"linke/internal/invoice/domain/valueobject"
	"linke/internal/shared/domain"
	sharedvo "linke/internal/shared/valueobject"
)

// Invoice represents the invoice aggregate root
type Invoice struct {
	// Identity
	id            sharedvo.InvoiceID
	invoiceNumber valueobject.InvoiceNumber
	orderID       uint
	userID        sharedvo.UserID

	// Classification
	invoiceType valueobject.InvoiceType
	status      valueobject.InvoiceStatus

	// Financial Information
	subtotal    sharedvo.Money
	taxInfo     valueobject.TaxInfo
	totalAmount sharedvo.Money
	paidAmount  sharedvo.Money

	// Billing Information
	billingAddress valueobject.BillingAddress
	companyInfo    valueobject.CompanyInfo

	// Payment Terms
	issuedAt         *time.Time
	dueAt            *time.Time
	paymentTermsDays int
	paidAt           *time.Time

	// Content
	description   string
	lineItems     string // JSON string
	notes         string
	internalNotes string

	// Document Management
	pdfPath  string
	pdfSize  int
	template string
	language string

	// Sending Records
	sentAt         *time.Time
	sendCount      int
	lastReminderAt *time.Time

	// Voiding Information
	voidedAt   *time.Time
	voidReason string

	// Metadata
	metadata string // JSON string

	// Timestamps
	createdAt time.Time
	updatedAt time.Time
	deletedAt *time.Time

	// Domain Events
	events []domain.DomainEvent
}

// NewInvoice creates a new invoice
func NewInvoice(
	orderID uint,
	userID sharedvo.UserID,
	invoiceType valueobject.InvoiceType,
	subtotal sharedvo.Money,
	taxInfo valueobject.TaxInfo,
	billingAddress valueobject.BillingAddress,
	description string,
) (*Invoice, error) {
	if orderID == 0 {
		return nil, fmt.Errorf("order ID is required")
	}

	if userID.IsZero() {
		return nil, fmt.Errorf("user ID is required")
	}

	if subtotal.IsZero() {
		return nil, fmt.Errorf("subtotal must be greater than zero")
	}

	if description == "" {
		return nil, fmt.Errorf("description is required")
	}

	// Convert subtotal to domain Money for tax calculation
	domainSubtotal, err := valueobject.ConvertFromSharedMoney(subtotal)
	if err != nil {
		return nil, fmt.Errorf("failed to convert subtotal: %w", err)
	}
	
	// Calculate tax amount based on subtotal
	domainTaxAmount, err := taxInfo.CalculateTaxAmount(domainSubtotal)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate tax amount: %w", err)
	}
	
	// Convert tax amount back to shared Money
	taxAmount, err := valueobject.ConvertToSharedMoney(domainTaxAmount)
	if err != nil {
		return nil, fmt.Errorf("failed to convert tax amount: %w", err)
	}
	
	// Calculate total amount
	totalAmount, err := subtotal.Add(taxAmount)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate total amount: %w", err)
	}

	now := time.Now()
	
	// Generate placeholder invoice ID (will be assigned on persistence)
	sharedInvoiceID := sharedvo.GenerateInvoiceID()
	
	// Create zero money with same currency as subtotal
	zeroPaidAmount := sharedvo.Zero(subtotal.Currency())
	
	invoice := &Invoice{
		id:               sharedInvoiceID,
		invoiceNumber:    valueobject.GenerateInvoiceNumber(),
		orderID:          orderID,
		userID:           userID,
		invoiceType:      invoiceType,
		status:           valueobject.StatusDraft,
		subtotal:         subtotal,
		taxInfo:          taxInfo,
		totalAmount:      totalAmount,
		paidAmount:       zeroPaidAmount,
		billingAddress:   billingAddress,
		paymentTermsDays: 30, // Default payment terms
		description:      description,
		template:         "default",
		language:         "en",
		createdAt:        now,
		updatedAt:        now,
		events:           make([]domain.DomainEvent, 0),
	}

	// Raise domain event (convert back to domain types for event)
	domainInvoiceID := valueobject.ConvertFromSharedInvoiceID(invoice.id)
	domainUserID := valueobject.ConvertFromSharedUserID(invoice.userID)
	domainTotalAmount, _ := valueobject.ConvertFromSharedMoney(invoice.totalAmount)
	
	invoice.raiseEvent(event.InvoiceCreated{
		InvoiceID:     domainInvoiceID,
		InvoiceNumber: invoice.invoiceNumber,
		UserID:        domainUserID,
		OrderID:       invoice.orderID,
		TotalAmount:   domainTotalAmount,
		BillingEmail:  invoice.billingAddress.Email(),
		Timestamp:     now,
	})

	return invoice, nil
}

// Reconstruction method for loading from persistence
func ReconstructInvoice(
	id valueobject.InvoiceID,
	invoiceNumber valueobject.InvoiceNumber,
	orderID, userID uint,
	invoiceType valueobject.InvoiceType,
	status valueobject.InvoiceStatus,
	subtotal, totalAmount, paidAmount valueobject.Money,
	taxInfo valueobject.TaxInfo,
	billingAddress valueobject.BillingAddress,
	companyInfo valueobject.CompanyInfo,
	issuedAt, dueAt, paidAt, sentAt, lastReminderAt, voidedAt *time.Time,
	paymentTermsDays, sendCount int,
	description, lineItems, notes, internalNotes, pdfPath, template, language, voidReason, metadata string,
	pdfSize int,
	createdAt, updatedAt time.Time,
	deletedAt *time.Time,
) *Invoice {
	// Convert domain types to shared types for Invoice struct fields
	sharedID, _ := valueobject.ConvertToSharedInvoiceID(id)
	sharedUserID, _ := valueobject.ConvertToSharedUserID(userID)
	sharedSubtotal, _ := valueobject.ConvertToSharedMoney(subtotal)
	sharedTotalAmount, _ := valueobject.ConvertToSharedMoney(totalAmount)
	sharedPaidAmount, _ := valueobject.ConvertToSharedMoney(paidAmount)
	
	return &Invoice{
		id:               sharedID,
		invoiceNumber:    invoiceNumber,
		orderID:          orderID,
		userID:           sharedUserID,
		invoiceType:      invoiceType,
		status:           status,
		subtotal:         sharedSubtotal,
		taxInfo:          taxInfo,
		totalAmount:      sharedTotalAmount,
		paidAmount:       sharedPaidAmount,
		billingAddress:   billingAddress,
		companyInfo:      companyInfo,
		issuedAt:         issuedAt,
		dueAt:            dueAt,
		paymentTermsDays: paymentTermsDays,
		paidAt:           paidAt,
		description:      description,
		lineItems:        lineItems,
		notes:            notes,
		internalNotes:    internalNotes,
		pdfPath:          pdfPath,
		pdfSize:          pdfSize,
		template:         template,
		language:         language,
		sentAt:           sentAt,
		sendCount:        sendCount,
		lastReminderAt:   lastReminderAt,
		voidedAt:         voidedAt,
		voidReason:       voidReason,
		metadata:         metadata,
		createdAt:        createdAt,
		updatedAt:        updatedAt,
		deletedAt:        deletedAt,
		events:           make([]domain.DomainEvent, 0),
	}
}

// Identity methods
func (i *Invoice) ID() valueobject.InvoiceID {
	return valueobject.ConvertFromSharedInvoiceID(i.id)
}

func (i *Invoice) InvoiceNumber() valueobject.InvoiceNumber {
	return i.invoiceNumber
}

func (i *Invoice) OrderID() uint {
	return i.orderID
}

func (i *Invoice) UserID() uint {
	return valueobject.ConvertFromSharedUserID(i.userID)
}

// Classification methods
func (i *Invoice) InvoiceType() valueobject.InvoiceType {
	return i.invoiceType
}

func (i *Invoice) Status() valueobject.InvoiceStatus {
	return i.status
}

// Financial methods
func (i *Invoice) Subtotal() valueobject.Money {
	domainMoney, _ := valueobject.ConvertFromSharedMoney(i.subtotal)
	return domainMoney
}

func (i *Invoice) TaxInfo() valueobject.TaxInfo {
	return i.taxInfo
}

func (i *Invoice) TotalAmount() valueobject.Money {
	domainMoney, _ := valueobject.ConvertFromSharedMoney(i.totalAmount)
	return domainMoney
}

func (i *Invoice) PaidAmount() valueobject.Money {
	domainMoney, _ := valueobject.ConvertFromSharedMoney(i.paidAmount)
	return domainMoney
}

func (i *Invoice) RemainingAmount() (valueobject.Money, error) {
	remaining, err := i.totalAmount.Subtract(i.paidAmount)
	if err != nil {
		return valueobject.Money{}, err
	}
	return valueobject.ConvertFromSharedMoney(remaining)
}

// Address and company methods
func (i *Invoice) BillingAddress() valueobject.BillingAddress {
	return i.billingAddress
}

func (i *Invoice) CompanyInfo() valueobject.CompanyInfo {
	return i.companyInfo
}

// Payment terms methods
func (i *Invoice) IssuedAt() *time.Time {
	return i.issuedAt
}

func (i *Invoice) DueAt() *time.Time {
	return i.dueAt
}

func (i *Invoice) PaymentTermsDays() int {
	return i.paymentTermsDays
}

func (i *Invoice) PaidAt() *time.Time {
	return i.paidAt
}

// Content methods
func (i *Invoice) Description() string {
	return i.description
}

func (i *Invoice) Notes() string {
	return i.notes
}

func (i *Invoice) InternalNotes() string {
	return i.internalNotes
}

// Document methods
func (i *Invoice) PDFPath() string {
	return i.pdfPath
}

func (i *Invoice) PDFSize() int {
	return i.pdfSize
}

func (i *Invoice) Template() string {
	return i.template
}

func (i *Invoice) Language() string {
	return i.language
}

// Content methods - additional
func (i *Invoice) LineItems() string {
	return i.lineItems
}

func (i *Invoice) Metadata() string {
	return i.metadata
}

// Audit methods
func (i *Invoice) CreatedAt() time.Time {
	return i.createdAt
}

func (i *Invoice) UpdatedAt() time.Time {
	return i.updatedAt
}

// Sending and status methods
func (i *Invoice) SentAt() *time.Time {
	return i.sentAt
}

func (i *Invoice) SendCount() int {
	return i.sendCount
}

func (i *Invoice) LastReminderAt() *time.Time {
	return i.lastReminderAt
}

func (i *Invoice) VoidedAt() *time.Time {
	return i.voidedAt
}

func (i *Invoice) VoidReason() string {
	return i.voidReason
}

// Business logic methods

// CanBeEdited checks if the invoice can be edited
func (i *Invoice) CanBeEdited() bool {
	return i.status.IsDraft()
}

// CanBeSent checks if the invoice can be sent
func (i *Invoice) CanBeSent() bool {
	return i.status.IsDraft()
}

// CanBePaid checks if the invoice can be marked as paid
func (i *Invoice) CanBePaid() bool {
	return (i.status.IsSent() || i.status.IsOverdue()) && !i.IsFullyPaid()
}

// CanBeVoided checks if the invoice can be voided
func (i *Invoice) CanBeVoided() bool {
	return !i.status.IsPaid() && !i.status.IsVoided()
}

// IsFullyPaid checks if the invoice is fully paid
func (i *Invoice) IsFullyPaid() bool {
	paid, err := i.paidAmount.GreaterThanOrEqual(i.totalAmount)
	if err != nil {
		return false
	}
	return paid
}

// IsOverdue checks if the invoice is overdue
func (i *Invoice) IsOverdue() bool {
	if i.status.IsOverdue() {
		return true
	}
	return i.dueAt != nil && time.Now().After(*i.dueAt) && !i.IsFullyPaid()
}

// DaysOverdue calculates days overdue
func (i *Invoice) DaysOverdue() int {
	if !i.IsOverdue() || i.dueAt == nil {
		return 0
	}
	
	days := int(time.Since(*i.dueAt).Hours() / 24)
	if days < 0 {
		return 0
	}
	return days
}

// Business operations

// Send marks the invoice as sent
func (i *Invoice) Send() error {
	if !i.CanBeSent() {
		return fmt.Errorf("invoice cannot be sent in status: %s", i.status.String())
	}

	now := time.Now()
	i.status = valueobject.StatusSent
	i.sentAt = &now
	i.sendCount++
	i.updatedAt = now

	// Set issued date if not already set
	if i.issuedAt == nil {
		i.issuedAt = &now
	}

	// Set due date if not already set and payment terms are specified
	if i.dueAt == nil && i.paymentTermsDays > 0 {
		dueDate := now.AddDate(0, 0, i.paymentTermsDays)
		i.dueAt = &dueDate
	}

	// Convert to domain types for event
	domainInvoiceID := valueobject.ConvertFromSharedInvoiceID(i.id)
	
	i.raiseEvent(event.InvoiceSent{
		InvoiceID:     domainInvoiceID,
		InvoiceNumber: i.invoiceNumber,
		BillingEmail:  i.billingAddress.Email(),
		SentAt:        now,
		SendCount:     i.sendCount,
		Timestamp:     now,
	})

	return nil
}

// MarkAsPaid records a payment for the invoice
func (i *Invoice) MarkAsPaid(amount valueobject.Money, paymentRef string) error {
	if !i.CanBePaid() {
		return fmt.Errorf("invoice cannot be marked as paid in status: %s", i.status.String())
	}

	// Convert domain Money to shared Money for calculation
	sharedAmount, err := valueobject.ConvertToSharedMoney(amount)
	if err != nil {
		return fmt.Errorf("failed to convert payment amount: %w", err)
	}

	// Ensure same currency
	if !sharedAmount.Currency().Equals(i.totalAmount.Currency()) {
		return fmt.Errorf("payment currency %s does not match invoice currency %s",
			sharedAmount.Currency().Code(), i.totalAmount.Currency().Code())
	}

	now := time.Now()
	
	// Add to paid amount
	newPaidAmount, err := i.paidAmount.Add(sharedAmount)
	if err != nil {
		return fmt.Errorf("failed to add payment amount: %w", err)
	}

	i.paidAmount = newPaidAmount
	i.paidAt = &now
	i.updatedAt = now

	// Check if fully paid
	isFullyPaid := i.IsFullyPaid()
	if isFullyPaid {
		i.status = valueobject.StatusPaid
	}

	// Convert to domain types for event
	domainInvoiceID := valueobject.ConvertFromSharedInvoiceID(i.id)
	domainUserID := valueobject.ConvertFromSharedUserID(i.userID)
	domainTotalPaidAmount, _ := valueobject.ConvertFromSharedMoney(i.paidAmount)
	domainTotalAmount, _ := valueobject.ConvertFromSharedMoney(i.totalAmount)
	
	i.raiseEvent(event.InvoicePaid{
		InvoiceID:       domainInvoiceID,
		InvoiceNumber:   i.invoiceNumber,
		UserID:          domainUserID,
		OrderID:         i.orderID,
		PaidAmount:      amount,
		TotalPaidAmount: domainTotalPaidAmount,
		TotalAmount:     domainTotalAmount,
		IsFullyPaid:     isFullyPaid,
		PaymentRef:      paymentRef,
		PaidAt:          now,
		Timestamp:       now,
	})

	return nil
}

// Void voids the invoice
func (i *Invoice) Void(reason string) error {
	if !i.CanBeVoided() {
		return fmt.Errorf("invoice cannot be voided in status: %s", i.status.String())
	}

	if reason == "" {
		return fmt.Errorf("void reason is required")
	}

	now := time.Now()
	i.status = valueobject.StatusVoided
	i.voidedAt = &now
	i.voidReason = reason
	i.updatedAt = now

	// Convert to domain types for event
	domainInvoiceID := valueobject.ConvertFromSharedInvoiceID(i.id)
	domainUserID := valueobject.ConvertFromSharedUserID(i.userID)
	
	i.raiseEvent(event.InvoiceVoided{
		InvoiceID:     domainInvoiceID,
		InvoiceNumber: i.invoiceNumber,
		UserID:        domainUserID,
		OrderID:       i.orderID,
		Reason:        reason,
		VoidedAt:      now,
		Timestamp:     now,
	})

	return nil
}

// MarkAsOverdue marks the invoice as overdue
func (i *Invoice) MarkAsOverdue() error {
	if !i.IsOverdue() {
		return fmt.Errorf("invoice is not overdue")
	}

	if i.status.IsOverdue() {
		return nil // Already marked as overdue
	}

	now := time.Now()
	i.status = valueobject.StatusOverdue
	i.updatedAt = now

	remainingAmount, _ := i.RemainingAmount()

	// Convert to domain types for event
	domainInvoiceID := valueobject.ConvertFromSharedInvoiceID(i.id)
	domainUserID := valueobject.ConvertFromSharedUserID(i.userID)
	
	i.raiseEvent(event.InvoiceOverdue{
		InvoiceID:       domainInvoiceID,
		InvoiceNumber:   i.invoiceNumber,
		UserID:          domainUserID,
		OrderID:         i.orderID,
		BillingEmail:    i.billingAddress.Email(),
		DueDate:         *i.dueAt,
		DaysOverdue:     i.DaysOverdue(),
		RemainingAmount: remainingAmount,
		Timestamp:       now,
	})

	return nil
}

// Update updates editable fields of the invoice
func (i *Invoice) Update(
	billingAddress *valueobject.BillingAddress,
	companyInfo *valueobject.CompanyInfo,
	taxInfo *valueobject.TaxInfo,
	description *string,
	notes *string,
	dueDate *time.Time,
	paymentTermsDays *int,
) error {
	if !i.CanBeEdited() {
		return fmt.Errorf("invoice cannot be edited in status: %s", i.status.String())
	}

	updatedFields := make([]string, 0)

	if billingAddress != nil {
		i.billingAddress = *billingAddress
		updatedFields = append(updatedFields, "billing_address")
	}

	if companyInfo != nil {
		i.companyInfo = *companyInfo
		updatedFields = append(updatedFields, "company_info")
	}

	if taxInfo != nil {
		i.taxInfo = *taxInfo
		// Recalculate total amount
		domainTaxAmount := taxInfo.TaxAmount()
		sharedTaxAmount, err := valueobject.ConvertToSharedMoney(domainTaxAmount)
		if err != nil {
			return fmt.Errorf("failed to convert tax amount: %w", err)
		}
		totalAmount, err := i.subtotal.Add(sharedTaxAmount)
		if err != nil {
			return fmt.Errorf("failed to recalculate total amount: %w", err)
		}
		i.totalAmount = totalAmount
		updatedFields = append(updatedFields, "tax_info", "total_amount")
	}

	if description != nil {
		if *description == "" {
			return fmt.Errorf("description cannot be empty")
		}
		i.description = *description
		updatedFields = append(updatedFields, "description")
	}

	if notes != nil {
		i.notes = *notes
		updatedFields = append(updatedFields, "notes")
	}

	if dueDate != nil {
		i.dueAt = dueDate
		updatedFields = append(updatedFields, "due_date")
	}

	if paymentTermsDays != nil {
		if *paymentTermsDays < 0 {
			return fmt.Errorf("payment terms days cannot be negative")
		}
		i.paymentTermsDays = *paymentTermsDays
		updatedFields = append(updatedFields, "payment_terms")
	}

	if len(updatedFields) > 0 {
		i.updatedAt = time.Now()

		// Convert to domain types for event
		domainInvoiceID := valueobject.ConvertFromSharedInvoiceID(i.id)
		domainUserID := valueobject.ConvertFromSharedUserID(i.userID)
		
		i.raiseEvent(event.InvoiceUpdated{
			InvoiceID:     domainInvoiceID,
			InvoiceNumber: i.invoiceNumber,
			UserID:        domainUserID,
			UpdatedFields: updatedFields,
			Timestamp:     i.updatedAt,
		})
	}

	return nil
}

// Domain events handling

// DomainEvents returns the domain events
func (i *Invoice) DomainEvents() []domain.DomainEvent {
	return i.events
}

// ClearDomainEvents clears the domain events
func (i *Invoice) ClearDomainEvents() {
	i.events = make([]domain.DomainEvent, 0)
}

// raiseEvent adds a domain event
func (i *Invoice) raiseEvent(event domain.DomainEvent) {
	i.events = append(i.events, event)
}