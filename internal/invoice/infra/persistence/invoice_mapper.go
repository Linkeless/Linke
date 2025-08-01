package persistence

import (
	"fmt"

	"linke/internal/invoice/domain/model"
	"linke/internal/invoice/domain/valueobject"
)

// InvoiceMapper handles mapping between domain model and persistence object
type InvoiceMapper struct{}

// NewInvoiceMapper creates a new invoice mapper
func NewInvoiceMapper() *InvoiceMapper {
	return &InvoiceMapper{}
}

// ToModel converts persistence object to domain model
func (m *InvoiceMapper) ToModel(po *InvoicePO) (*model.Invoice, error) {
	// Parse value objects
	domainID := valueobject.NewInvoiceID(po.ID)
	
	// Convert to shared types for aggregate creation
	id, err := valueobject.ConvertToSharedInvoiceID(domainID)
	if err != nil {
		return nil, fmt.Errorf("failed to convert invoice ID: %w", err)
	}
	
	invoiceNumber, err := valueobject.NewInvoiceNumber(po.InvoiceNumber)
	if err != nil {
		return nil, fmt.Errorf("invalid invoice number in database: %w", err)
	}

	invoiceType, err := valueobject.NewInvoiceType(po.InvoiceType)
	if err != nil {
		return nil, fmt.Errorf("invalid invoice type in database: %w", err)
	}

	status, err := valueobject.NewInvoiceStatus(po.Status)
	if err != nil {
		return nil, fmt.Errorf("invalid invoice status in database: %w", err)
	}

	// Create domain types first for value object creation
	domainCurrency, err := valueobject.NewCurrency(po.Currency)
	if err != nil {
		return nil, fmt.Errorf("invalid currency in database: %w", err)
	}

	domainSubtotal, err := valueobject.NewMoney(po.Subtotal, domainCurrency)
	if err != nil {
		return nil, fmt.Errorf("invalid subtotal in database: %w", err)
	}

	domainTotalAmount, err := valueobject.NewMoney(po.TotalAmount, domainCurrency)
	if err != nil {
		return nil, fmt.Errorf("invalid total amount in database: %w", err)
	}
	
	// Convert to shared types for aggregate
	subtotal, err := valueobject.ConvertToSharedMoney(domainSubtotal)
	if err != nil {
		return nil, fmt.Errorf("failed to convert subtotal: %w", err)
	}
	
	totalAmount, err := valueobject.ConvertToSharedMoney(domainTotalAmount)
	if err != nil {
		return nil, fmt.Errorf("failed to convert total amount: %w", err)
	}

	domainPaidAmount, err := valueobject.NewMoney(po.PaidAmount, domainCurrency)
	if err != nil {
		return nil, fmt.Errorf("invalid paid amount in database: %w", err)
	}

	taxAmount, err := valueobject.NewMoney(po.TaxAmount, domainCurrency)
	if err != nil {
		return nil, fmt.Errorf("invalid tax amount in database: %w", err)
	}

	taxInfo, err := valueobject.NewTaxInfo(po.TaxRate, po.TaxType, po.TaxNumber, taxAmount)
	if err != nil {
		return nil, fmt.Errorf("invalid tax info in database: %w", err)
	}

	billingAddress, err := valueobject.NewBillingAddress(
		po.BillingName,
		po.BillingEmail,
		po.BillingAddress,
		po.BillingCity,
		po.BillingState,
		po.BillingCountry,
		po.BillingZip,
	)
	if err != nil {
		return nil, fmt.Errorf("invalid billing address in database: %w", err)
	}

	companyInfo, err := valueobject.NewCompanyInfo(
		po.CompanyName,
		po.CompanyAddress,
		po.CompanyTaxID,
	)
	if err != nil {
		return nil, fmt.Errorf("invalid company info in database: %w", err)
	}

	// Reconstruct domain model - ReconstructInvoice expects domain types
	domainInvoiceID := valueobject.ConvertFromSharedInvoiceID(id)
	domainSubtotalReconvert, _ := valueobject.ConvertFromSharedMoney(subtotal)
	domainTotalAmountReconvert, _ := valueobject.ConvertFromSharedMoney(totalAmount)
	// Use domainPaidAmount directly since we already have it
	
	invoice := model.ReconstructInvoice(
		domainInvoiceID,
		invoiceNumber,
		po.OrderID,
		po.UserID,
		invoiceType,
		status,
		domainSubtotalReconvert,
		domainTotalAmountReconvert,
		domainPaidAmount,
		taxInfo,
		billingAddress,
		companyInfo,
		po.IssuedAt,
		po.DueAt,
		po.PaidAt,
		po.SentAt,
		po.LastReminderAt,
		po.VoidedAt,
		po.PaymentTermsDays,
		po.SendCount,
		po.Description,
		po.LineItems,
		po.Notes,
		po.InternalNotes,
		po.PDFPath,
		po.Template,
		po.Language,
		po.VoidReason,
		po.Metadata,
		po.PDFSize,
		po.CreatedAt,
		po.UpdatedAt,
		nil, // deletedAt will be handled by GORM
	)

	return invoice, nil
}

// ToPersistence converts domain model to persistence object
func (m *InvoiceMapper) ToPersistence(invoice *model.Invoice) *InvoicePO {
	po := &InvoicePO{
		ID:               invoice.ID().Value(),
		OrderID:          invoice.OrderID(),
		UserID:           invoice.UserID(),
		InvoiceNumber:    invoice.InvoiceNumber().Value(),
		InvoiceType:      invoice.InvoiceType().Value(),
		Status:           invoice.Status().Value(),
		Subtotal:         invoice.Subtotal().Amount(),
		TaxRate:          invoice.TaxInfo().Rate(),
		TaxAmount:        invoice.TaxInfo().TaxAmount().Amount(),
		TotalAmount:      invoice.TotalAmount().Amount(),
		Currency:         invoice.TotalAmount().Currency().Code(),
		IssuedAt:         invoice.IssuedAt(),
		DueAt:            invoice.DueAt(),
		PaymentTermsDays: invoice.PaymentTermsDays(),
		PaidAmount:       invoice.PaidAmount().Amount(),
		PaidAt:           invoice.PaidAt(),
		BillingName:      invoice.BillingAddress().Name(),
		BillingEmail:     invoice.BillingAddress().Email(),
		BillingAddress:   invoice.BillingAddress().Address(),
		BillingCity:      invoice.BillingAddress().City(),
		BillingState:     invoice.BillingAddress().State(),
		BillingCountry:   invoice.BillingAddress().Country(),
		BillingZip:       invoice.BillingAddress().Zip(),
		TaxNumber:        invoice.TaxInfo().TaxNumber(),
		TaxType:          invoice.TaxInfo().TaxType(),
		CompanyName:      invoice.CompanyInfo().Name(),
		CompanyAddress:   invoice.CompanyInfo().Address(),
		CompanyTaxID:     invoice.CompanyInfo().TaxID(),
		Description:      invoice.Description(),
		Notes:            invoice.Notes(),
		InternalNotes:    invoice.InternalNotes(),
		PDFPath:          invoice.PDFPath(),
		Template:         invoice.Template(),
		Language:         invoice.Language(),
		
		// 新增缺失的字段映射
		SentAt:         invoice.SentAt(),
		SendCount:      invoice.SendCount(),
		LastReminderAt: invoice.LastReminderAt(),
		VoidedAt:       invoice.VoidedAt(),
		VoidReason:     invoice.VoidReason(),
		PDFSize:        invoice.PDFSize(),
		LineItems:      invoice.LineItems(),
		Metadata:       invoice.Metadata(),
		
		CreatedAt:        invoice.CreatedAt(),
		UpdatedAt:        invoice.UpdatedAt(),
	}

	return po
}

// UpdatePersistenceFromModel updates persistence object from domain model
func (m *InvoiceMapper) UpdatePersistenceFromModel(po *InvoicePO, invoice *model.Invoice) {
	po.Status = invoice.Status().Value()
	po.TaxRate = invoice.TaxInfo().Rate()
	po.TaxAmount = invoice.TaxInfo().TaxAmount().Amount()
	po.TotalAmount = invoice.TotalAmount().Amount()
	po.IssuedAt = invoice.IssuedAt()
	po.DueAt = invoice.DueAt()
	po.PaymentTermsDays = invoice.PaymentTermsDays()
	po.PaidAmount = invoice.PaidAmount().Amount()
	po.PaidAt = invoice.PaidAt()
	po.BillingName = invoice.BillingAddress().Name()
	po.BillingEmail = invoice.BillingAddress().Email()
	po.BillingAddress = invoice.BillingAddress().Address()
	po.BillingCity = invoice.BillingAddress().City()
	po.BillingState = invoice.BillingAddress().State()
	po.BillingCountry = invoice.BillingAddress().Country()
	po.BillingZip = invoice.BillingAddress().Zip()
	po.TaxNumber = invoice.TaxInfo().TaxNumber()
	po.TaxType = invoice.TaxInfo().TaxType()
	po.CompanyName = invoice.CompanyInfo().Name()
	po.CompanyAddress = invoice.CompanyInfo().Address()
	po.CompanyTaxID = invoice.CompanyInfo().TaxID()
	po.Description = invoice.Description()
	po.Notes = invoice.Notes()
	po.InternalNotes = invoice.InternalNotes()
	po.PDFPath = invoice.PDFPath()
	po.Template = invoice.Template()
	po.Language = invoice.Language()
	
	// 新增缺失字段的更新
	po.SentAt = invoice.SentAt()
	po.SendCount = invoice.SendCount()
	po.LastReminderAt = invoice.LastReminderAt()
	po.VoidedAt = invoice.VoidedAt()
	po.VoidReason = invoice.VoidReason()
	po.PDFSize = invoice.PDFSize()
	po.LineItems = invoice.LineItems()
	po.Metadata = invoice.Metadata()
	
	po.UpdatedAt = invoice.UpdatedAt()
}