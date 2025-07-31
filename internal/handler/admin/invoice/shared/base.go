package shared

import (
	"linke/internal/service"
)

// BaseHandler provides common dependencies for invoice handlers
type BaseHandler struct {
	InvoiceService *service.InvoiceService
	Validator      *InvoiceValidator
}

// NewBaseHandler creates a new base handler with common dependencies
func NewBaseHandler(invoiceService *service.InvoiceService) *BaseHandler {
	return &BaseHandler{
		InvoiceService: invoiceService,
		Validator:      NewInvoiceValidator(),
	}
}