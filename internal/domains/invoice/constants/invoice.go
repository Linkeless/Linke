package constants

// Invoice type constants
const (
	InvoiceTypeStandard   = "standard"
	InvoiceTypeProforma   = "proforma"
	InvoiceTypeCreditNote = "credit_note"
)

// Invoice status constants
const (
	InvoiceStatusDraft     = "draft"
	InvoiceStatusSent      = "sent"
	InvoiceStatusPaid      = "paid"
	InvoiceStatusOverdue   = "overdue"
	InvoiceStatusCancelled = "cancelled"
	InvoiceStatusVoided    = "voided"
)
