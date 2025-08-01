package service

import (
	"context"
	"fmt"

	"linke/internal/invoice/domain/model"
	"linke/internal/invoice/domain/repository"
	"linke/internal/invoice/domain/valueobject"
)

// InvoiceNumberGenerator is a domain service for generating unique invoice numbers
type InvoiceNumberGenerator struct {
	repository repository.InvoiceRepository
}

// NewInvoiceNumberGenerator creates a new invoice number generator
func NewInvoiceNumberGenerator(repository repository.InvoiceRepository) *InvoiceNumberGenerator {
	return &InvoiceNumberGenerator{
		repository: repository,
	}
}

// GenerateUniqueNumber generates a unique invoice number
func (g *InvoiceNumberGenerator) GenerateUniqueNumber(ctx context.Context) (valueobject.InvoiceNumber, error) {
	const maxAttempts = 10

	for attempt := 0; attempt < maxAttempts; attempt++ {
		// Generate base number
		baseNumber := valueobject.GenerateInvoiceNumber()
		
		// Check if it exists
		exists, err := g.repository.ExistsByInvoiceNumber(ctx, baseNumber)
		if err != nil {
			return valueobject.InvoiceNumber{}, fmt.Errorf("failed to check invoice number existence: %w", err)
		}

		if !exists {
			return baseNumber, nil
		}

		// If exists, try with a suffix
		suffixedNumber, err := g.generateWithSuffix(ctx, baseNumber, attempt+1)
		if err == nil {
			return suffixedNumber, nil
		}
	}

	return valueobject.InvoiceNumber{}, fmt.Errorf("failed to generate unique invoice number after %d attempts", maxAttempts)
}

// generateWithSuffix generates a number with a suffix
func (g *InvoiceNumberGenerator) generateWithSuffix(ctx context.Context, baseNumber valueobject.InvoiceNumber, suffix int) (valueobject.InvoiceNumber, error) {
	suffixedValue := fmt.Sprintf("%s-%d", baseNumber.Value(), suffix)
	suffixedNumber, err := valueobject.NewInvoiceNumber(suffixedValue)
	if err != nil {
		return valueobject.InvoiceNumber{}, err
	}

	// Check if the suffixed number exists
	exists, err := g.repository.ExistsByInvoiceNumber(ctx, suffixedNumber)
	if err != nil {
		return valueobject.InvoiceNumber{}, fmt.Errorf("failed to check suffixed invoice number existence: %w", err)
	}

	if exists {
		return valueobject.InvoiceNumber{}, fmt.Errorf("suffixed invoice number already exists")
	}

	return suffixedNumber, nil
}

// InvoiceDomainService provides domain-level business logic for invoices
type InvoiceDomainService struct {
	repository repository.InvoiceRepository
}

// NewInvoiceDomainService creates a new invoice domain service
func NewInvoiceDomainService(repository repository.InvoiceRepository) *InvoiceDomainService {
	return &InvoiceDomainService{
		repository: repository,
	}
}

// CanCreateInvoiceForOrder checks if an invoice can be created for the given order
func (s *InvoiceDomainService) CanCreateInvoiceForOrder(ctx context.Context, orderID uint) error {
	if orderID == 0 {
		return fmt.Errorf("order ID is required")
	}

	// Check if invoice already exists for this order
	exists, err := s.repository.ExistsByOrderID(ctx, orderID)
	if err != nil {
		return fmt.Errorf("failed to check existing invoice for order: %w", err)
	}

	if exists {
		return fmt.Errorf("invoice already exists for order ID %d", orderID)
	}

	return nil
}

// ValidateInvoiceForPayment validates if an invoice can receive a payment
func (s *InvoiceDomainService) ValidateInvoiceForPayment(invoice *model.Invoice, amount valueobject.Money) error {
	if !invoice.CanBePaid() {
		return fmt.Errorf("invoice %s cannot be paid in status: %s", 
			invoice.InvoiceNumber().String(), invoice.Status().String())
	}

	// Check currency compatibility
	if !amount.Currency().Equals(invoice.TotalAmount().Currency()) {
		return fmt.Errorf("payment currency %s does not match invoice currency %s",
			amount.Currency().Code(), invoice.TotalAmount().Currency().Code())
	}

	// Check if payment amount is reasonable
	remainingAmount, err := invoice.RemainingAmount()
	if err != nil {
		return fmt.Errorf("failed to calculate remaining amount: %w", err)
	}

	// Allow overpayment but warn if significantly over
	if greater, err := amount.GreaterThan(remainingAmount); err == nil && greater {
		overpayment, _ := amount.Subtract(remainingAmount)
		// Allow up to 10% overpayment without warning
		maxOverpayment, _ := remainingAmount.Multiply(0.1)
		if overpaymentExcessive, _ := overpayment.GreaterThan(maxOverpayment); overpaymentExcessive {
			return fmt.Errorf("payment amount %s significantly exceeds remaining amount %s",
				amount.String(), remainingAmount.String())
		}
	}

	return nil
}

// CalculateTaxAmount calculates tax amount for given subtotal and tax rate
func (s *InvoiceDomainService) CalculateTaxAmount(subtotal valueobject.Money, taxRate float64) (valueobject.Money, error) {
	if taxRate < 0 || taxRate > 1 {
		return valueobject.Money{}, fmt.Errorf("tax rate must be between 0 and 1 (0%% to 100%%)")
	}

	taxAmount, err := subtotal.Multiply(taxRate)
	if err != nil {
		return valueobject.Money{}, fmt.Errorf("failed to calculate tax amount: %w", err)
	}

	return taxAmount, nil
}

// CreateTaxInfo creates tax information with calculated tax amount
func (s *InvoiceDomainService) CreateTaxInfo(
	subtotal valueobject.Money,
	taxRate float64,
	taxType, taxNumber string,
) (valueobject.TaxInfo, error) {
	taxAmount, err := s.CalculateTaxAmount(subtotal, taxRate)
	if err != nil {
		return valueobject.TaxInfo{}, err
	}

	return valueobject.NewTaxInfo(taxRate, taxType, taxNumber, taxAmount)
}

// OverdueInvoiceService handles overdue invoice detection and management
type OverdueInvoiceService struct {
	repository repository.InvoiceRepository
}

// NewOverdueInvoiceService creates a new overdue invoice service
func NewOverdueInvoiceService(repository repository.InvoiceRepository) *OverdueInvoiceService {
	return &OverdueInvoiceService{
		repository: repository,
	}
}

// FindAndMarkOverdueInvoices finds and marks invoices as overdue
func (s *OverdueInvoiceService) FindAndMarkOverdueInvoices(ctx context.Context) ([]*model.Invoice, error) {
	// Find sent invoices that might be overdue
	sentInvoices, err := s.repository.FindInvoicesByStatus(ctx, valueobject.StatusSent, 1000, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to find sent invoices: %w", err)
	}

	var overdueInvoices []*model.Invoice
	
	for _, invoice := range sentInvoices {
		if invoice.IsOverdue() && !invoice.Status().IsOverdue() {
			// Mark as overdue
			if err := invoice.MarkAsOverdue(); err != nil {
				continue // Skip this invoice but continue with others
			}

			// Save the updated invoice
			if err := s.repository.Save(ctx, invoice); err != nil {
				continue // Skip this invoice but continue with others
			}

			overdueInvoices = append(overdueInvoices, invoice)
		}
	}

	return overdueInvoices, nil
}

// GetOverdueInvoicesCount returns the count of overdue invoices
func (s *OverdueInvoiceService) GetOverdueInvoicesCount(ctx context.Context) (int64, error) {
	return s.repository.CountOverdue(ctx)
}