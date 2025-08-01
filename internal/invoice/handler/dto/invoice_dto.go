package dto

import (
	"time"

	"linke/internal/invoice/domain/model"
	"linke/internal/invoice/domain/valueobject"
)

// CreateInvoiceRequest represents the request to create a new invoice
type CreateInvoiceRequest struct {
	OrderID     uint                  `json:"order_id" validate:"required,min=1"`
	UserID      uint                  `json:"user_id" validate:"required,min=1"`
	Type        string                `json:"type" validate:"required,oneof=standard refund proforma"`
	Subtotal    MoneyDTO              `json:"subtotal" validate:"required"`
	TaxInfo     TaxInfoDTO            `json:"tax_info"`
	BillingInfo BillingAddressDTO     `json:"billing_info" validate:"required"`
	CompanyInfo CompanyInfoDTO        `json:"company_info"`
	Description string                `json:"description" validate:"required,min=1,max=1000"`
	Notes       string                `json:"notes" validate:"max=2000"`
	DueDate     *time.Time            `json:"due_date"`
	PaymentTerms int                  `json:"payment_terms" validate:"min=0,max=365"`
	Template    string                `json:"template" validate:"max=50"`
	Language    string                `json:"language" validate:"max=10"`
}

// UpdateInvoiceRequest represents the request to update an invoice
type UpdateInvoiceRequest struct {
	BillingInfo     *BillingAddressDTO `json:"billing_info"`
	CompanyInfo     *CompanyInfoDTO    `json:"company_info"`
	TaxInfo         *TaxInfoDTO        `json:"tax_info"`
	Description     *string            `json:"description" validate:"omitempty,min=1,max=1000"`
	Notes           *string            `json:"notes" validate:"omitempty,max=2000"`
	DueDate         *time.Time         `json:"due_date"`
	PaymentTerms    *int               `json:"payment_terms" validate:"omitempty,min=0,max=365"`
	InternalNotes   *string            `json:"internal_notes" validate:"omitempty,max=2000"`
}

// SendInvoiceRequest represents the request to send an invoice
type SendInvoiceRequest struct {
	Email     string `json:"email" validate:"required,email"`
	Subject   string `json:"subject" validate:"max=200"`
	Message   string `json:"message" validate:"max=2000"`
	CCEmails  []string `json:"cc_emails" validate:"dive,email"`
}

// PayInvoiceRequest represents the request to mark an invoice as paid
type PayInvoiceRequest struct {
	Amount     MoneyDTO `json:"amount" validate:"required"`
	PaymentRef string   `json:"payment_ref" validate:"required,max=255"`
	Notes      string   `json:"notes" validate:"max=500"`
}

// VoidInvoiceRequest represents the request to void an invoice
type VoidInvoiceRequest struct {
	Reason string `json:"reason" validate:"required,min=1,max=500"`
}

// InvoiceListQuery represents query parameters for listing invoices
type InvoiceListQuery struct {
	Page       int      `form:"page,default=1" validate:"min=1"`
	Size       int      `form:"size,default=20" validate:"min=1,max=100"`
	UserID     *uint    `form:"user_id"`
	Status     []string `form:"status" validate:"dive,oneof=draft sent paid overdue voided"`
	Type       *string  `form:"type" validate:"omitempty,oneof=standard refund proforma"`
	DateFrom   *time.Time `form:"date_from"`
	DateTo     *time.Time `form:"date_to"`
	Search     string   `form:"search" validate:"max=100"`
	SortBy     string   `form:"sort_by,default=created_at" validate:"oneof=created_at updated_at due_date total_amount"`
	SortOrder  string   `form:"sort_order,default=desc" validate:"oneof=asc desc"`
}

// InvoiceResponse represents the invoice response
type InvoiceResponse struct {
	ID              string            `json:"id"`
	InvoiceNumber   string            `json:"invoice_number"`
	OrderID         uint              `json:"order_id"`
	UserID          uint              `json:"user_id"`
	Type            string            `json:"type"`
	Status          string            `json:"status"`
	Subtotal        MoneyDTO          `json:"subtotal"`
	TaxInfo         TaxInfoDTO        `json:"tax_info"`
	TotalAmount     MoneyDTO          `json:"total_amount"`
	PaidAmount      MoneyDTO          `json:"paid_amount"`
	RemainingAmount MoneyDTO          `json:"remaining_amount"`
	BillingInfo     BillingAddressDTO `json:"billing_info"`
	CompanyInfo     CompanyInfoDTO    `json:"company_info"`
	Description     string            `json:"description"`
	Notes           string            `json:"notes"`
	InternalNotes   string            `json:"internal_notes,omitempty"`
	IssuedAt        *time.Time        `json:"issued_at"`
	DueAt           *time.Time        `json:"due_at"`
	PaymentTerms    int               `json:"payment_terms"`
	PaidAt          *time.Time        `json:"paid_at"`
	SentAt          *time.Time        `json:"sent_at"`
	SendCount       int               `json:"send_count"`
	LastReminderAt  *time.Time        `json:"last_reminder_at"`
	VoidedAt        *time.Time        `json:"voided_at"`
	VoidReason      string            `json:"void_reason,omitempty"`
	PDFPath         string            `json:"pdf_path,omitempty"`
	Template        string            `json:"template"`
	Language        string            `json:"language"`
	IsOverdue       bool              `json:"is_overdue"`
	DaysOverdue     int               `json:"days_overdue"`
	CreatedAt       time.Time         `json:"created_at"`
	UpdatedAt       time.Time         `json:"updated_at"`
}

// InvoiceListResponse represents the paginated invoice list response
type InvoiceListResponse struct {
	Items      []InvoiceResponse `json:"items"`
	Total      int64             `json:"total"`
	Page       int               `json:"page"`
	Size       int               `json:"size"`
	TotalPages int               `json:"total_pages"`
	HasNext    bool              `json:"has_next"`
	HasPrev    bool              `json:"has_prev"`
}

// InvoiceStatsResponse represents invoice statistics
type InvoiceStatsResponse struct {
	TotalInvoices    int64    `json:"total_invoices"`
	DraftCount       int64    `json:"draft_count"`
	SentCount        int64    `json:"sent_count"`
	PaidCount        int64    `json:"paid_count"`
	OverdueCount     int64    `json:"overdue_count"`
	VoidedCount      int64    `json:"voided_count"`
	TotalAmount      MoneyDTO `json:"total_amount"`
	PaidAmount       MoneyDTO `json:"paid_amount"`
	OutstandingAmount MoneyDTO `json:"outstanding_amount"`
	OverdueAmount    MoneyDTO `json:"overdue_amount"`
}

// MoneyDTO represents a monetary amount
type MoneyDTO struct {
	Amount       float64 `json:"amount"`
	Currency     string  `json:"currency"`
	DisplayValue string  `json:"display_value"`
}

// TaxInfoDTO represents tax information
type TaxInfoDTO struct {
	TaxRate      float64  `json:"tax_rate"`
	TaxAmount    MoneyDTO `json:"tax_amount"`
	TaxNumber    string   `json:"tax_number"`
	TaxType      string   `json:"tax_type"`
	TaxExempt    bool     `json:"tax_exempt"`
	ExemptReason string   `json:"exempt_reason"`
}

// BillingAddressDTO represents billing address information
type BillingAddressDTO struct {
	Name        string `json:"name" validate:"required,max=100"`
	Email       string `json:"email" validate:"required,email"`
	Phone       string `json:"phone" validate:"max=20"`
	Company     string `json:"company" validate:"max=100"`
	AddressLine1 string `json:"address_line1" validate:"required,max=200"`
	AddressLine2 string `json:"address_line2" validate:"max=200"`
	City        string `json:"city" validate:"required,max=100"`
	State       string `json:"state" validate:"max=100"`
	PostalCode  string `json:"postal_code" validate:"required,max=20"`
	Country     string `json:"country" validate:"required,len=2"`
}

// CompanyInfoDTO represents company information
type CompanyInfoDTO struct {
	Name          string `json:"name" validate:"max=200"`
	TaxNumber     string `json:"tax_number" validate:"max=50"`
	RegNumber     string `json:"reg_number" validate:"max=50"`
	AddressLine1  string `json:"address_line1" validate:"max=200"`
	AddressLine2  string `json:"address_line2" validate:"max=200"`
	City          string `json:"city" validate:"max=100"`
	State         string `json:"state" validate:"max=100"`
	PostalCode    string `json:"postal_code" validate:"max=20"`
	Country       string `json:"country" validate:"len=2"`
	Website       string `json:"website" validate:"omitempty,url"`
	Email         string `json:"email" validate:"omitempty,email"`
	Phone         string `json:"phone" validate:"max=20"`
}

// FromInvoice converts a domain invoice to response DTO
func FromInvoice(invoice *model.Invoice) InvoiceResponse {
	remainingAmount, _ := invoice.RemainingAmount()
	
	return InvoiceResponse{
		ID:              invoice.ID().String(),
		InvoiceNumber:   invoice.InvoiceNumber().String(),
		OrderID:         invoice.OrderID(),
		UserID:          invoice.UserID(),
		Type:            invoice.InvoiceType().String(),
		Status:          invoice.Status().String(),
		Subtotal:        FromMoney(invoice.Subtotal()),
		TaxInfo:         FromTaxInfo(invoice.TaxInfo()),
		TotalAmount:     FromMoney(invoice.TotalAmount()),
		PaidAmount:      FromMoney(invoice.PaidAmount()),
		RemainingAmount: FromMoney(remainingAmount),
		BillingInfo:     FromBillingAddress(invoice.BillingAddress()),
		CompanyInfo:     FromCompanyInfo(invoice.CompanyInfo()),
		Description:     invoice.Description(),
		Notes:           invoice.Notes(),
		InternalNotes:   invoice.InternalNotes(),
		IssuedAt:        invoice.IssuedAt(),
		DueAt:           invoice.DueAt(),
		PaymentTerms:    invoice.PaymentTermsDays(),
		PaidAt:          invoice.PaidAt(),
		SentAt:          invoice.SentAt(),
		SendCount:       invoice.SendCount(),
		LastReminderAt:  invoice.LastReminderAt(),
		VoidedAt:        invoice.VoidedAt(),
		VoidReason:      invoice.VoidReason(),
		PDFPath:         invoice.PDFPath(),
		Template:        invoice.Template(),
		Language:        invoice.Language(),
		IsOverdue:       invoice.IsOverdue(),
		DaysOverdue:     invoice.DaysOverdue(),
		CreatedAt:       invoice.CreatedAt(),
		UpdatedAt:       invoice.UpdatedAt(),
	}
}

// FromInvoices converts a list of domain invoices to response DTOs
func FromInvoices(invoices []*model.Invoice) []InvoiceResponse {
	responses := make([]InvoiceResponse, len(invoices))
	for i, invoice := range invoices {
		responses[i] = FromInvoice(invoice)
	}
	return responses
}

// FromMoney converts a domain Money to DTO
func FromMoney(money valueobject.Money) MoneyDTO {
	return MoneyDTO{
		Amount:       money.Amount(),
		Currency:     money.Currency().Code(),
		DisplayValue: money.DisplayValue(),
	}
}

// ToMoney converts a DTO Money to domain Money
func (m MoneyDTO) ToMoney() (valueobject.Money, error) {
	currency, err := valueobject.NewCurrency(m.Currency)
	if err != nil {
		return valueobject.Money{}, err
	}
	return valueobject.NewMoney(m.Amount, currency)
}

// FromTaxInfo converts domain TaxInfo to DTO
func FromTaxInfo(taxInfo valueobject.TaxInfo) TaxInfoDTO {
	return TaxInfoDTO{
		TaxRate:      taxInfo.Rate(),
		TaxAmount:    FromMoney(taxInfo.TaxAmount()),
		TaxNumber:    taxInfo.TaxNumber(),
		TaxType:      taxInfo.TaxType(),
		TaxExempt:    false, // Will need to add this method to TaxInfo
		ExemptReason: "",    // Will need to add this method to TaxInfo
	}
}

// ToTaxInfo converts DTO TaxInfo to domain TaxInfo
func (t TaxInfoDTO) ToTaxInfo(subtotal valueobject.Money) (valueobject.TaxInfo, error) {
	taxAmount, err := t.TaxAmount.ToMoney()
	if err != nil {
		return valueobject.TaxInfo{}, err
	}
	return valueobject.NewTaxInfo(t.TaxRate, t.TaxType, t.TaxNumber, taxAmount)
}

// FromBillingAddress converts domain BillingAddress to DTO
func FromBillingAddress(addr valueobject.BillingAddress) BillingAddressDTO {
	return BillingAddressDTO{
		Name:         addr.Name(),
		Email:        addr.Email(),
		Phone:        "", // BillingAddress doesn't have phone field
		Company:      "", // BillingAddress doesn't have company field
		AddressLine1: addr.Address(),
		AddressLine2: "", // BillingAddress doesn't have separate address lines
		City:         addr.City(),
		State:        addr.State(),
		PostalCode:   addr.Zip(),
		Country:      addr.Country(),
	}
}

// ToBillingAddress converts DTO BillingAddress to domain BillingAddress
func (b BillingAddressDTO) ToBillingAddress() (valueobject.BillingAddress, error) {
	// Combine address lines
	address := b.AddressLine1
	if b.AddressLine2 != "" {
		address += ", " + b.AddressLine2
	}
	
	return valueobject.NewBillingAddress(
		b.Name,
		b.Email,
		address,
		b.City,
		b.State,
		b.Country,
		b.PostalCode,
	)
}

// FromCompanyInfo converts domain CompanyInfo to DTO
func FromCompanyInfo(info valueobject.CompanyInfo) CompanyInfoDTO {
	return CompanyInfoDTO{
		Name:         info.Name(),
		TaxNumber:    info.TaxID(),
		RegNumber:    "", // CompanyInfo doesn't have reg number
		AddressLine1: info.Address(),
		AddressLine2: "", // CompanyInfo doesn't have separate address lines
		City:         "", // CompanyInfo doesn't have separate city
		State:        "", // CompanyInfo doesn't have separate state
		PostalCode:   "", // CompanyInfo doesn't have separate postal code
		Country:      "", // CompanyInfo doesn't have separate country
		Website:      "", // CompanyInfo doesn't have website
		Email:        "", // CompanyInfo doesn't have email
		Phone:        "", // CompanyInfo doesn't have phone
	}
}

// ToCompanyInfo converts DTO CompanyInfo to domain CompanyInfo
func (c CompanyInfoDTO) ToCompanyInfo() (valueobject.CompanyInfo, error) {
	// Combine address information
	address := c.AddressLine1
	if c.AddressLine2 != "" {
		address += ", " + c.AddressLine2
	}
	if c.City != "" {
		address += ", " + c.City
	}
	if c.State != "" {
		address += ", " + c.State
	}
	if c.PostalCode != "" {
		address += " " + c.PostalCode
	}
	if c.Country != "" {
		address += ", " + c.Country
	}
	
	return valueobject.NewCompanyInfo(c.Name, address, c.TaxNumber)
}