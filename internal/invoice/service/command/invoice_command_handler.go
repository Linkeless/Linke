package command

import (
	"context"
	"fmt"
	"time"

	"linke/internal/invoice/domain/model"
	"linke/internal/invoice/domain/repository"
	"linke/internal/invoice/domain/service"
	"linke/internal/invoice/domain/valueobject"
	"linke/internal/shared/domain"
)

// InvoiceCommandHandler handles invoice commands
type InvoiceCommandHandler struct {
	repository             repository.InvoiceRepository
	numberGenerator        *service.InvoiceNumberGenerator
	domainService          *service.InvoiceDomainService
	eventBus               domain.EventBus
}

// NewInvoiceCommandHandler creates a new invoice command handler
func NewInvoiceCommandHandler(
	repository repository.InvoiceRepository,
	numberGenerator *service.InvoiceNumberGenerator,
	domainService *service.InvoiceDomainService,
	eventBus domain.EventBus,
) *InvoiceCommandHandler {
	return &InvoiceCommandHandler{
		repository:      repository,
		numberGenerator: numberGenerator,
		domainService:   domainService,
		eventBus:        eventBus,
	}
}

// CreateInvoice handles create invoice command
func (h *InvoiceCommandHandler) CreateInvoice(ctx context.Context, cmd CreateInvoiceCommand) (*model.Invoice, error) {
	// Validate that we can create invoice for this order
	if err := h.domainService.CanCreateInvoiceForOrder(ctx, cmd.OrderID); err != nil {
		return nil, err
	}

	// Parse invoice type
	invoiceType := valueobject.TypeStandard
	if cmd.Type != "" {
		var err error
		invoiceType, err = valueobject.NewInvoiceType(cmd.Type)
		if err != nil {
			return nil, fmt.Errorf("invalid invoice type: %w", err)
		}
	}

	// Convert to shared types for aggregate creation
	userID, err := valueobject.ConvertToSharedUserID(cmd.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to convert user ID: %w", err)
	}
	
	subtotal, err := valueobject.ConvertToSharedMoney(cmd.Subtotal)
	if err != nil {
		return nil, fmt.Errorf("failed to convert subtotal: %w", err)
	}
	
	// Create invoice
	invoice, err := model.NewInvoice(
		cmd.OrderID,
		userID,
		invoiceType,
		subtotal,
		cmd.TaxInfo,
		cmd.BillingInfo,
		cmd.Description,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create invoice: %w", err)
	}

	// Update invoice with additional information if provided
	if cmd.CompanyInfo != nil || cmd.DueDate != nil || cmd.PaymentTerms > 0 || cmd.Notes != "" {
		var paymentTerms *int
		if cmd.PaymentTerms > 0 {
			paymentTerms = &cmd.PaymentTerms
		}
		
		var notes *string
		if cmd.Notes != "" {
			notes = &cmd.Notes
		}

		err = invoice.Update(
			nil, // billing address already set
			cmd.CompanyInfo,
			nil, // tax info already set
			nil, // description already set
			notes,
			cmd.DueDate,
			paymentTerms,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to update invoice with additional info: %w", err)
		}
	}

	// Save invoice
	if err := h.repository.Save(ctx, invoice); err != nil {
		return nil, fmt.Errorf("failed to save invoice: %w", err)
	}

	// Publish domain events
	for _, event := range invoice.DomainEvents() {
		if err := h.eventBus.Publish(ctx, event); err != nil {
			// Log error but don't fail the operation
			continue
		}
	}

	invoice.ClearDomainEvents()

	return invoice, nil
}

// SendInvoice handles send invoice command
func (h *InvoiceCommandHandler) SendInvoice(ctx context.Context, cmd SendInvoiceCommand) (*model.Invoice, error) {
	// Parse invoice ID
	invoiceID, err := valueobject.ParseInvoiceID(cmd.InvoiceID)
	if err != nil {
		return nil, fmt.Errorf("invalid invoice ID: %w", err)
	}

	// Find invoice
	invoice, err := h.repository.FindByID(ctx, invoiceID)
	if err != nil {
		return nil, fmt.Errorf("failed to find invoice: %w", err)
	}

	// Send invoice
	if err := invoice.Send(); err != nil {
		return nil, fmt.Errorf("failed to send invoice: %w", err)
	}

	// Save updated invoice
	if err := h.repository.Save(ctx, invoice); err != nil {
		return nil, fmt.Errorf("failed to save sent invoice: %w", err)
	}

	// Publish domain events
	for _, event := range invoice.DomainEvents() {
		if err := h.eventBus.Publish(ctx, event); err != nil {
			continue
		}
	}

	invoice.ClearDomainEvents()

	return invoice, nil
}

// HandleUpdateInvoice handles update invoice command
func (h *InvoiceCommandHandler) HandleUpdateInvoice(ctx context.Context, cmd UpdateInvoiceCommand) (*model.Invoice, error) {
	// Parse invoice ID
	invoiceID, err := valueobject.ParseInvoiceID(cmd.InvoiceID)
	if err != nil {
		return nil, fmt.Errorf("invalid invoice ID: %w", err)
	}

	// Find invoice
	invoice, err := h.repository.FindByID(ctx, invoiceID)
	if err != nil {
		return nil, fmt.Errorf("failed to find invoice: %w", err)
	}

	// Prepare update parameters
	var billingAddress *valueobject.BillingAddress
	var companyInfo *valueobject.CompanyInfo
	var taxInfo *valueobject.TaxInfo
	var dueDate *time.Time

	// Update billing address if provided
	if cmd.BillingInfo != nil {
		billingAddress = cmd.BillingInfo
	}

	// Update company info if provided
	if cmd.CompanyInfo != nil {
		companyInfo = cmd.CompanyInfo
	}

	// Update tax info if provided - convert from DTO
	if cmd.TaxInfo != nil {
		// Convert MoneyDTO to Money
		currency, err := valueobject.NewCurrency(cmd.TaxInfo.TaxAmount.Currency)
		if err != nil {
			return nil, fmt.Errorf("invalid tax currency: %w", err)
		}
		
		taxAmount, err := valueobject.NewMoney(
			cmd.TaxInfo.TaxAmount.Amount, 
			currency,
		)
		if err != nil {
			return nil, fmt.Errorf("invalid tax amount: %w", err)
		}
		
		newTaxInfo, err := valueobject.NewTaxInfo(
			cmd.TaxInfo.TaxRate,
			cmd.TaxInfo.TaxType,
			cmd.TaxInfo.TaxNumber,
			taxAmount,
		)
		if err != nil {
			return nil, fmt.Errorf("invalid tax info: %w", err)
		}
		taxInfo = &newTaxInfo
	}

	// Set due date if provided
	if cmd.DueDate != nil {
		dueDate = cmd.DueDate
	}

	// Update invoice
	err = invoice.Update(
		billingAddress,
		companyInfo,
		taxInfo,
		cmd.Description,
		cmd.Notes,
		dueDate,
		cmd.PaymentTerms,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update invoice: %w", err)
	}

	// Save updated invoice
	if err := h.repository.Save(ctx, invoice); err != nil {
		return nil, fmt.Errorf("failed to save updated invoice: %w", err)
	}

	// Publish domain events
	for _, event := range invoice.DomainEvents() {
		if err := h.eventBus.Publish(ctx, event); err != nil {
			continue
		}
	}

	invoice.ClearDomainEvents()

	return invoice, nil
}

// HandleMarkAsPaid handles mark invoice as paid command
func (h *InvoiceCommandHandler) HandleMarkAsPaid(ctx context.Context, cmd MarkInvoiceAsPaidCommand) error {
	// Parse invoice ID
	invoiceID, err := valueobject.ParseInvoiceID(cmd.InvoiceID)
	if err != nil {
		return fmt.Errorf("invalid invoice ID: %w", err)
	}

	// Use payment amount from command
	paymentAmount := cmd.Amount

	// Find invoice
	invoice, err := h.repository.FindByID(ctx, invoiceID)
	if err != nil {
		return fmt.Errorf("failed to find invoice: %w", err)
	}

	// Validate payment
	if err := h.domainService.ValidateInvoiceForPayment(invoice, paymentAmount); err != nil {
		return fmt.Errorf("payment validation failed: %w", err)
	}

	// Mark as paid
	if err := invoice.MarkAsPaid(paymentAmount, cmd.PaymentRef); err != nil {
		return fmt.Errorf("failed to mark invoice as paid: %w", err)
	}

	// Save updated invoice
	if err := h.repository.Save(ctx, invoice); err != nil {
		return fmt.Errorf("failed to save paid invoice: %w", err)
	}

	// Publish domain events
	for _, event := range invoice.DomainEvents() {
		if err := h.eventBus.Publish(ctx, event); err != nil {
			continue
		}
	}

	invoice.ClearDomainEvents()

	return nil
}

// HandleVoidInvoice handles void invoice command
func (h *InvoiceCommandHandler) HandleVoidInvoice(ctx context.Context, cmd VoidInvoiceCommand) error {
	// Parse invoice ID
	invoiceID, err := valueobject.ParseInvoiceID(cmd.InvoiceID)
	if err != nil {
		return fmt.Errorf("invalid invoice ID: %w", err)
	}

	// Find invoice
	invoice, err := h.repository.FindByID(ctx, invoiceID)
	if err != nil {
		return fmt.Errorf("failed to find invoice: %w", err)
	}

	// Void invoice
	if err := invoice.Void(cmd.Reason); err != nil {
		return fmt.Errorf("failed to void invoice: %w", err)
	}

	// Save updated invoice
	if err := h.repository.Save(ctx, invoice); err != nil {
		return fmt.Errorf("failed to save voided invoice: %w", err)
	}

	// Publish domain events
	for _, event := range invoice.DomainEvents() {
		if err := h.eventBus.Publish(ctx, event); err != nil {
			continue
		}
	}

	invoice.ClearDomainEvents()

	return nil
}

// HandleDeleteInvoice handles delete invoice command
func (h *InvoiceCommandHandler) HandleDeleteInvoice(ctx context.Context, cmd DeleteInvoiceCommand) error {
	// Parse invoice ID
	invoiceID, err := valueobject.ParseInvoiceID(cmd.InvoiceID)
	if err != nil {
		return fmt.Errorf("invalid invoice ID: %w", err)
	}

	// Find invoice
	invoice, err := h.repository.FindByID(ctx, invoiceID)
	if err != nil {
		return fmt.Errorf("failed to find invoice: %w", err)
	}

	// Check if invoice can be deleted (only draft invoices)
	if !invoice.CanBeEdited() {
		return fmt.Errorf("only draft invoices can be deleted, current status: %s", invoice.Status().String())
	}

	// Delete invoice
	if err := h.repository.Delete(ctx, invoiceID); err != nil {
		return fmt.Errorf("failed to delete invoice: %w", err)
	}

	return nil
}

// Wrapper methods for handler compatibility

// UpdateInvoice is a wrapper for HandleUpdateInvoice
func (h *InvoiceCommandHandler) UpdateInvoice(ctx context.Context, cmd UpdateInvoiceCommand) (*model.Invoice, error) {
	return h.HandleUpdateInvoice(ctx, cmd)
}

// PayInvoice is a wrapper for HandleMarkAsPaid
func (h *InvoiceCommandHandler) PayInvoice(ctx context.Context, cmd PayInvoiceCommand) (*model.Invoice, error) {
	err := h.HandleMarkAsPaid(ctx, cmd)
	if err != nil {
		return nil, err
	}
	
	// Return the updated invoice
	invoiceID, parseErr := valueobject.ParseInvoiceID(cmd.InvoiceID)
	if parseErr != nil {
		return nil, parseErr
	}
	
	return h.repository.FindByID(ctx, invoiceID)
}

// VoidInvoice is a wrapper for HandleVoidInvoice
func (h *InvoiceCommandHandler) VoidInvoice(ctx context.Context, cmd VoidInvoiceCommand) (*model.Invoice, error) {
	err := h.HandleVoidInvoice(ctx, cmd)
	if err != nil {
		return nil, err
	}
	
	// Return the updated invoice
	invoiceID, parseErr := valueobject.ParseInvoiceID(cmd.InvoiceID)
	if parseErr != nil {
		return nil, parseErr
	}
	
	return h.repository.FindByID(ctx, invoiceID)
}

// MarkAsOverdue marks an invoice as overdue
func (h *InvoiceCommandHandler) MarkAsOverdue(ctx context.Context, cmd MarkOverdueCommand) (*model.Invoice, error) {
	// Parse invoice ID
	invoiceID, err := valueobject.ParseInvoiceID(cmd.InvoiceID)
	if err != nil {
		return nil, fmt.Errorf("invalid invoice ID: %w", err)
	}

	// Find invoice
	invoice, err := h.repository.FindByID(ctx, invoiceID)
	if err != nil {
		return nil, fmt.Errorf("failed to find invoice: %w", err)
	}

	// Mark as overdue
	if err := invoice.MarkAsOverdue(); err != nil {
		return nil, fmt.Errorf("failed to mark invoice as overdue: %w", err)
	}

	// Save updated invoice
	if err := h.repository.Save(ctx, invoice); err != nil {
		return nil, fmt.Errorf("failed to save overdue invoice: %w", err)
	}

	// Publish domain events
	for _, event := range invoice.DomainEvents() {
		if err := h.eventBus.Publish(ctx, event); err != nil {
			continue
		}
	}

	invoice.ClearDomainEvents()

	return invoice, nil
}

// DeleteInvoice is a wrapper for HandleDeleteInvoice  
func (h *InvoiceCommandHandler) DeleteInvoice(ctx context.Context, cmd DeleteInvoiceCommand) error {
	return h.HandleDeleteInvoice(ctx, cmd)
}