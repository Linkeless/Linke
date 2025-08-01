package valueobject

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewInvoiceStatus(t *testing.T) {
	tests := []struct {
		name        string
		value       string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid draft status",
			value:       "draft",
			expectError: false,
		},
		{
			name:        "valid sent status",
			value:       "sent",
			expectError: false,
		},
		{
			name:        "valid paid status",
			value:       "paid",
			expectError: false,
		},
		{
			name:        "valid overdue status",
			value:       "overdue",
			expectError: false,
		},
		{
			name:        "valid voided status",
			value:       "voided",
			expectError: false,
		},
		{
			name:        "case insensitive - uppercase",
			value:       "DRAFT",
			expectError: false,
		},
		{
			name:        "case insensitive - mixed case",
			value:       "Sent",
			expectError: false,
		},
		{
			name:        "with whitespace",
			value:       "  paid  ",
			expectError: false,
		},
		{
			name:        "invalid status",
			value:       "invalid",
			expectError: true,
			errorMsg:    "invalid invoice status: invalid",
		},
		{
			name:        "empty status",
			value:       "",
			expectError: true,
			errorMsg:    "invalid invoice status:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, err := NewInvoiceStatus(tt.value)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				require.NoError(t, err)
				// All valid statuses should be normalized to lowercase
				expectedValue := getExpectedNormalizedValue(tt.value)
				assert.Equal(t, expectedValue, status.Value())
			}
		})
	}
}

func TestInvoiceStatus_StatusCheckers(t *testing.T) {
	tests := []struct {
		name       string
		status     InvoiceStatus
		isDraft    bool
		isSent     bool
		isPaid     bool
		isOverdue  bool
		isVoided   bool
	}{
		{
			name:       "draft status",
			status:     StatusDraft,
			isDraft:    true,
			isSent:     false,
			isPaid:     false,
			isOverdue:  false,
			isVoided:   false,
		},
		{
			name:       "sent status",
			status:     StatusSent,
			isDraft:    false,
			isSent:     true,
			isPaid:     false,
			isOverdue:  false,
			isVoided:   false,
		},
		{
			name:       "paid status",
			status:     StatusPaid,
			isDraft:    false,
			isSent:     false,
			isPaid:     true,
			isOverdue:  false,
			isVoided:   false,
		},
		{
			name:       "overdue status",
			status:     StatusOverdue,
			isDraft:    false,
			isSent:     false,
			isPaid:     false,
			isOverdue:  true,
			isVoided:   false,
		},
		{
			name:       "voided status",
			status:     StatusVoided,
			isDraft:    false,
			isSent:     false,
			isPaid:     false,
			isOverdue:  false,
			isVoided:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.isDraft, tt.status.IsDraft(), "IsDraft() check failed")
			assert.Equal(t, tt.isSent, tt.status.IsSent(), "IsSent() check failed")
			assert.Equal(t, tt.isPaid, tt.status.IsPaid(), "IsPaid() check failed")
			assert.Equal(t, tt.isOverdue, tt.status.IsOverdue(), "IsOverdue() check failed")
			assert.Equal(t, tt.isVoided, tt.status.IsVoided(), "IsVoided() check failed")
		})
	}
}

func TestInvoiceStatus_CanTransitionTo(t *testing.T) {
	tests := []struct {
		name       string
		from       InvoiceStatus
		to         InvoiceStatus
		canTransition bool
	}{
		// From draft
		{
			name:       "draft to sent",
			from:       StatusDraft,
			to:         StatusSent,
			canTransition: true,
		},
		{
			name:       "draft to voided",
			from:       StatusDraft,
			to:         StatusVoided,
			canTransition: true,
		},
		{
			name:       "draft to paid (invalid)",
			from:       StatusDraft,
			to:         StatusPaid,
			canTransition: false,
		},
		{
			name:       "draft to overdue (invalid)",
			from:       StatusDraft,
			to:         StatusOverdue,
			canTransition: false,
		},
		
		// From sent
		{
			name:       "sent to paid",
			from:       StatusSent,
			to:         StatusPaid,
			canTransition: true,
		},
		{
			name:       "sent to overdue",
			from:       StatusSent,
			to:         StatusOverdue,
			canTransition: true,
		},
		{
			name:       "sent to voided",
			from:       StatusSent,
			to:         StatusVoided,
			canTransition: true,
		},
		{
			name:       "sent to draft (invalid)",
			from:       StatusSent,
			to:         StatusDraft,
			canTransition: false,
		},
		
		// From overdue
		{
			name:       "overdue to paid",
			from:       StatusOverdue,
			to:         StatusPaid,
			canTransition: true,
		},
		{
			name:       "overdue to voided",
			from:       StatusOverdue,
			to:         StatusVoided,
			canTransition: true,
		},
		{
			name:       "overdue to sent (invalid)",
			from:       StatusOverdue,
			to:         StatusSent,
			canTransition: false,
		},
		
		// From paid - no transitions allowed
		{
			name:       "paid to voided (invalid)",
			from:       StatusPaid,
			to:         StatusVoided,
			canTransition: false,
		},
		{
			name:       "paid to overdue (invalid)",
			from:       StatusPaid,
			to:         StatusOverdue,
			canTransition: false,
		},
		
		// From voided - no transitions allowed
		{
			name:       "voided to paid (invalid)",
			from:       StatusVoided,
			to:         StatusPaid,
			canTransition: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.from.CanTransitionTo(tt.to)
			assert.Equal(t, tt.canTransition, result)
		})
	}
}

func TestInvoiceStatus_Equals(t *testing.T) {
	draft1 := StatusDraft
	draft2, _ := NewInvoiceStatus("draft")
	sent := StatusSent

	// Test equality
	assert.True(t, draft1.Equals(draft2))
	assert.True(t, draft2.Equals(draft1))
	
	// Test inequality
	assert.False(t, draft1.Equals(sent))
	assert.False(t, sent.Equals(draft1))
}

func TestInvoiceStatus_String(t *testing.T) {
	tests := []struct {
		name     string
		status   InvoiceStatus
		expected string
	}{
		{
			name:     "draft status string",
			status:   StatusDraft,
			expected: "draft",
		},
		{
			name:     "sent status string",
			status:   StatusSent,
			expected: "sent",
		},
		{
			name:     "paid status string",
			status:   StatusPaid,
			expected: "paid",
		},
		{
			name:     "overdue status string",
			status:   StatusOverdue,
			expected: "overdue",
		},
		{
			name:     "voided status string",
			status:   StatusVoided,
			expected: "voided",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.status.String())
		})
	}
}

// Helper function to get expected normalized value for test assertions
func getExpectedNormalizedValue(input string) string {
	switch strings.ToLower(strings.TrimSpace(input)) {
	case "draft":
		return "draft"
	case "sent":
		return "sent"
	case "paid":
		return "paid"
	case "overdue":
		return "overdue"
	case "voided":
		return "voided"
	default:
		return ""
	}
}

// Tests for InvoiceType

func TestNewInvoiceType(t *testing.T) {
	tests := []struct {
		name        string
		value       string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid standard type",
			value:       "standard",
			expectError: false,
		},
		{
			name:        "valid proforma type",
			value:       "proforma",
			expectError: false,
		},
		{
			name:        "valid credit note type",
			value:       "credit_note",
			expectError: false,
		},
		{
			name:        "case insensitive - uppercase",
			value:       "STANDARD",
			expectError: false,
		},
		{
			name:        "case insensitive - mixed case",
			value:       "Proforma",
			expectError: false,
		},
		{
			name:        "with whitespace",
			value:       "  credit_note  ",
			expectError: false,
		},
		{
			name:        "invalid type",
			value:       "invalid",
			expectError: true,
			errorMsg:    "invalid invoice type: invalid",
		},
		{
			name:        "empty type",
			value:       "",
			expectError: true,
			errorMsg:    "invalid invoice type:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			invoiceType, err := NewInvoiceType(tt.value)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				require.NoError(t, err)
				// All valid types should be normalized to lowercase
				expectedValue := getExpectedNormalizedInvoiceType(tt.value)
				assert.Equal(t, expectedValue, invoiceType.Value())
			}
		})
	}
}

func TestInvoiceType_TypeCheckers(t *testing.T) {
	tests := []struct {
		name           string
		invoiceType    InvoiceType
		isStandard     bool
		isProforma     bool
		isCreditNote   bool
	}{
		{
			name:         "standard type",
			invoiceType:  TypeStandard,
			isStandard:   true,
			isProforma:   false,
			isCreditNote: false,
		},
		{
			name:         "proforma type",
			invoiceType:  TypeProforma,
			isStandard:   false,
			isProforma:   true,
			isCreditNote: false,
		},
		{
			name:         "credit note type",
			invoiceType:  TypeCreditNote,
			isStandard:   false,
			isProforma:   false,
			isCreditNote: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.isStandard, tt.invoiceType.IsStandard(), "IsStandard() check failed")
			assert.Equal(t, tt.isProforma, tt.invoiceType.IsProforma(), "IsProforma() check failed")
			assert.Equal(t, tt.isCreditNote, tt.invoiceType.IsCreditNote(), "IsCreditNote() check failed")
		})
	}
}

func TestInvoiceType_Equals(t *testing.T) {
	standard1 := TypeStandard
	standard2, _ := NewInvoiceType("standard")
	proforma := TypeProforma

	// Test equality
	assert.True(t, standard1.Equals(standard2))
	assert.True(t, standard2.Equals(standard1))

	// Test inequality
	assert.False(t, standard1.Equals(proforma))
	assert.False(t, proforma.Equals(standard1))
}

func TestInvoiceType_String(t *testing.T) {
	tests := []struct {
		name        string
		invoiceType InvoiceType
		expected    string
	}{
		{
			name:        "standard type string",
			invoiceType: TypeStandard,
			expected:    "standard",
		},
		{
			name:        "proforma type string",
			invoiceType: TypeProforma,
			expected:    "proforma",
		},
		{
			name:        "credit note type string",
			invoiceType: TypeCreditNote,
			expected:    "credit_note",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.invoiceType.String())
		})
	}
}

func TestInvoiceTypeConstants(t *testing.T) {
	tests := []struct {
		name        string
		invoiceType InvoiceType
		expected    string
	}{
		{
			name:        "TypeStandard constant",
			invoiceType: TypeStandard,
			expected:    "standard",
		},
		{
			name:        "TypeProforma constant",
			invoiceType: TypeProforma,
			expected:    "proforma",
		},
		{
			name:        "TypeCreditNote constant",
			invoiceType: TypeCreditNote,
			expected:    "credit_note",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.invoiceType.Value())
			assert.Equal(t, tt.expected, tt.invoiceType.String())
		})
	}
}

func TestInvoiceType_Immutability(t *testing.T) {
	// Test that InvoiceType is immutable
	originalValue := "standard"
	invoiceType, err := NewInvoiceType(originalValue)
	require.NoError(t, err)

	originalTypeValue := invoiceType.Value()

	// Modify the original string used to create type
	originalValue = "proforma"
	invoiceType2, err := NewInvoiceType(originalValue)
	require.NoError(t, err)

	// Original type should not be affected
	assert.Equal(t, originalTypeValue, invoiceType.Value())
	assert.Equal(t, "proforma", invoiceType2.Value())
	assert.False(t, invoiceType.Equals(invoiceType2))
}

// Helper function to get expected normalized invoice type value for test assertions
func getExpectedNormalizedInvoiceType(input string) string {
	switch strings.ToLower(strings.TrimSpace(input)) {
	case "standard":
		return "standard"
	case "proforma":
		return "proforma"
	case "credit_note":
		return "credit_note"
	default:
		return ""
	}
}