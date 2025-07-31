package invoice

import (
	"linke/internal/handler/admin/invoice/management"
	"linke/internal/handler/admin/invoice/operation"
	"linke/internal/handler/admin/invoice/query"
	"linke/internal/service"

	"github.com/gin-gonic/gin"
)

// AdminInvoiceManager manages all invoice-related admin handlers
type AdminInvoiceManager struct {
	// Sub-handlers for different invoice management aspects
	Management *management.InvoiceCRUDHandler
	List       *query.InvoiceListHandler
	Detail     *query.InvoiceDetailHandler
	Actions    *operation.InvoiceActionHandler
	Payment    *operation.InvoicePaymentHandler
	Lifecycle  *operation.InvoiceLifecycleHandler
}

// NewAdminInvoiceManager creates a new admin invoice manager with all sub-handlers
func NewAdminInvoiceManager(invoiceService *service.InvoiceService) *AdminInvoiceManager {
	return &AdminInvoiceManager{
		Management: management.NewInvoiceCRUDHandler(invoiceService),
		List:       query.NewInvoiceListHandler(invoiceService),
		Detail:     query.NewInvoiceDetailHandler(invoiceService),
		Actions:    operation.NewInvoiceActionHandler(invoiceService),
		Payment:    operation.NewInvoicePaymentHandler(invoiceService),
		Lifecycle:  operation.NewInvoiceLifecycleHandler(invoiceService),
	}
}

// Legacy compatibility layer - maintains the same interface as the original AdminInvoiceHandler
// This allows existing code to continue working without changes while using the modular structure internally

// CreateInvoice delegates to the management module
func (m *AdminInvoiceManager) CreateInvoice(c *gin.Context) {
	m.Management.CreateInvoice(c)
}

// CreateInvoiceFromOrder delegates to the management module
func (m *AdminInvoiceManager) CreateInvoiceFromOrder(c *gin.Context) {
	m.Management.CreateInvoiceFromOrder(c)
}

// UpdateInvoice delegates to the management module
func (m *AdminInvoiceManager) UpdateInvoice(c *gin.Context) {
	m.Management.UpdateInvoice(c)
}

// ListInvoices delegates to the list module
func (m *AdminInvoiceManager) ListInvoices(c *gin.Context) {
	m.List.ListInvoices(c)
}

// GetInvoice delegates to the detail module
func (m *AdminInvoiceManager) GetInvoice(c *gin.Context) {
	m.Detail.GetInvoice(c)
}

// SendInvoice delegates to the actions module
func (m *AdminInvoiceManager) SendInvoice(c *gin.Context) {
	m.Actions.SendInvoice(c)
}

// MarkInvoiceAsPaid delegates to the payment module
func (m *AdminInvoiceManager) MarkInvoiceAsPaid(c *gin.Context) {
	m.Payment.MarkInvoiceAsPaid(c)
}

// VoidInvoice delegates to the lifecycle module
func (m *AdminInvoiceManager) VoidInvoice(c *gin.Context) {
	m.Lifecycle.VoidInvoice(c)
}

// DeleteInvoice delegates to the lifecycle module
func (m *AdminInvoiceManager) DeleteInvoice(c *gin.Context) {
	m.Lifecycle.DeleteInvoice(c)
}