package valueobject

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Tests for InvoiceNumber

func TestNewInvoiceNumber(t *testing.T) {
	tests := []struct {
		name        string
		value       string
		expectError bool
		errorMsg    string
		expected    string
	}{
		{
			name:        "valid invoice number",
			value:       "INV2024010001",
			expectError: false,
			expected:    "INV2024010001",
		},
		{
			name:        "valid invoice number with hyphens",
			value:       "INV-2024-01-0001",
			expectError: false,
			expected:    "INV-2024-01-0001",
		},
		{
			name:        "valid invoice number with underscores",
			value:       "INV_2024_01_0001",
			expectError: false,
			expected:    "INV_2024_01_0001",
		},
		{
			name:        "lowercase invoice number should be uppercased",
			value:       "inv2024010001",
			expectError: false,
			expected:    "INV2024010001",
		},
		{
			name:        "invoice number with whitespace should be trimmed and uppercased",
			value:       "  inv-2024-01  ",
			expectError: false,
			expected:    "INV-2024-01",
		},
		{
			name:        "numeric only invoice number",
			value:       "2024010001",
			expectError: false,
			expected:    "2024010001",
		},
		{
			name:        "empty invoice number should fail",
			value:       "",
			expectError: true,
			errorMsg:    "invoice number cannot be empty",
		},
		{
			name:        "whitespace only invoice number should fail",
			value:       "   ",
			expectError: true,
			errorMsg:    "invoice number cannot be empty",
		},
		{
			name:        "invoice number too long should fail",
			value:       strings.Repeat("A", 33),
			expectError: true,
			errorMsg:    "invoice number cannot exceed 32 characters",
		},
		{
			name:        "invoice number with invalid characters should fail",
			value:       "INV@2024#01",
			expectError: true,
			errorMsg:    "invoice number can only contain letters, numbers, hyphens, and underscores",
		},
		{
			name:        "invoice number with spaces should fail",
			value:       "INV 2024 01",
			expectError: true,
			errorMsg:    "invoice number can only contain letters, numbers, hyphens, and underscores",
		},
		{
			name:        "invoice number with dots should fail",
			value:       "INV.2024.01",
			expectError: true,
			errorMsg:    "invoice number can only contain letters, numbers, hyphens, and underscores",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			invoiceNum, err := NewInvoiceNumber(tt.value)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, invoiceNum.Value())
				assert.Equal(t, tt.expected, invoiceNum.String())
			}
		})
	}
}

func TestGenerateInvoiceNumber(t *testing.T) {
	// Test that GenerateInvoiceNumber produces a valid invoice number
	invoiceNum := GenerateInvoiceNumber()
	
	assert.False(t, invoiceNum.IsEmpty())
	assert.NotEmpty(t, invoiceNum.Value())
	assert.Contains(t, invoiceNum.Value(), "INV")
	
	// Should be able to create the same number manually
	generatedNum, err := NewInvoiceNumber(invoiceNum.Value())
	require.NoError(t, err)
	assert.True(t, invoiceNum.Equals(generatedNum))
}

func TestInvoiceNumber_IsEmpty(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		isEmpty  bool
	}{
		{
			name:    "non-empty invoice number",
			value:   "INV2024010001",
			isEmpty: false,
		},
		{
			name:    "single character invoice number",
			value:   "A",
			isEmpty: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			invoiceNum, err := NewInvoiceNumber(tt.value)
			require.NoError(t, err)
			assert.Equal(t, tt.isEmpty, invoiceNum.IsEmpty())
		})
	}
}

func TestInvoiceNumber_Equals(t *testing.T) {
	num1, _ := NewInvoiceNumber("INV2024010001")
	num2, _ := NewInvoiceNumber("inv2024010001") // Should be normalized to uppercase
	num3, _ := NewInvoiceNumber("INV2024010002")

	// Test equality (case insensitive due to normalization)
	assert.True(t, num1.Equals(num2))
	assert.True(t, num2.Equals(num1))

	// Test inequality
	assert.False(t, num1.Equals(num3))
	assert.False(t, num3.Equals(num1))
}

func TestInvoiceNumber_String(t *testing.T) {
	invoiceNum, err := NewInvoiceNumber("INV-2024-01-0001")
	require.NoError(t, err)

	assert.Equal(t, "INV-2024-01-0001", invoiceNum.String())
	assert.Equal(t, invoiceNum.String(), invoiceNum.Value())
}

// Tests for InvoiceID

func TestNewInvoiceID(t *testing.T) {
	tests := []struct {
		name     string
		value    uint
		expected uint
	}{
		{
			name:     "valid positive ID",
			value:    123,
			expected: 123,
		},
		{
			name:     "zero ID",
			value:    0,
			expected: 0,
		},
		{
			name:     "large ID",
			value:    4294967295, // Max uint32
			expected: 4294967295,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id := NewInvoiceID(tt.value)
			assert.Equal(t, tt.expected, id.Value())
		})
	}
}

func TestParseInvoiceID(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectError bool
		errorMsg    string
		expected    uint
	}{
		{
			name:        "valid positive ID string",
			input:       "123",
			expectError: false,
			expected:    123,
		},
		{
			name:        "valid zero ID string",
			input:       "0",
			expectError: false,
			expected:    0,
		},
		{
			name:        "valid large ID string",
			input:       "4294967295",
			expectError: false,
			expected:    4294967295,
		},
		{
			name:        "empty string should fail",
			input:       "",
			expectError: true,
			errorMsg:    "invoice ID cannot be empty",
		},
		{
			name:        "non-numeric string should fail",
			input:       "abc",
			expectError: true,
			errorMsg:    "invalid invoice ID format",
		},
		{
			name:        "negative number string should fail",
			input:       "-123",
			expectError: true,
			errorMsg:    "invalid invoice ID format",
		},
		{
			name:        "decimal number string should fail",
			input:       "123.45",
			expectError: true,
			errorMsg:    "invalid invoice ID format",
		},
		{
			name:        "string with spaces should fail",
			input:       "123 456",
			expectError: true,
			errorMsg:    "invalid invoice ID format",
		},
		{
			name:        "string with leading zeros",
			input:       "000123",
			expectError: false,
			expected:    123,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, err := ParseInvoiceID(tt.input)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, id.Value())
			}
		})
	}
}

func TestGenerateInvoiceID(t *testing.T) {
	// Test that GenerateInvoiceID produces a valid ID
	id := GenerateInvoiceID()
	
	// In the current implementation, it returns 0
	// In production, this would generate a proper unique ID
	assert.Equal(t, uint(0), id.Value())
	assert.True(t, id.IsZero())
}

func TestInvoiceID_IsZero(t *testing.T) {
	tests := []struct {
		name    string
		value   uint
		isZero  bool
	}{
		{
			name:   "zero ID",
			value:  0,
			isZero: true,
		},
		{
			name:   "non-zero ID",
			value:  123,
			isZero: false,
		},
		{
			name:   "large ID",
			value:  4294967295,
			isZero: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id := NewInvoiceID(tt.value)
			assert.Equal(t, tt.isZero, id.IsZero())
		})
	}
}

func TestInvoiceID_Equals(t *testing.T) {
	id1 := NewInvoiceID(123)
	id2 := NewInvoiceID(123)
	id3 := NewInvoiceID(456)

	// Test equality
	assert.True(t, id1.Equals(id2))
	assert.True(t, id2.Equals(id1))

	// Test inequality
	assert.False(t, id1.Equals(id3))
	assert.False(t, id3.Equals(id1))
}

func TestInvoiceID_String(t *testing.T) {
	tests := []struct {
		name     string
		value    uint
		expected string
	}{
		{
			name:     "zero ID string",
			value:    0,
			expected: "0",
		},
		{
			name:     "positive ID string",
			value:    123,
			expected: "123",
		},
		{
			name:     "large ID string",
			value:    4294967295,
			expected: "4294967295",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id := NewInvoiceID(tt.value)
			assert.Equal(t, tt.expected, id.String())
		})
	}
}

func TestInvoiceID_ParseAndString_Roundtrip(t *testing.T) {
	// Test that parsing and converting back to string works correctly
	originalValues := []uint{0, 1, 123, 999999, 4294967295}

	for _, value := range originalValues {
		t.Run(fmt.Sprintf("roundtrip_%d", value), func(t *testing.T) {
			// Create ID from uint
			id1 := NewInvoiceID(value)
			
			// Convert to string
			idString := id1.String()
			
			// Parse back from string
			id2, err := ParseInvoiceID(idString)
			require.NoError(t, err)
			
			// Should be equal
			assert.True(t, id1.Equals(id2))
			assert.Equal(t, value, id2.Value())
		})
	}
}

func TestInvoiceIdentifiers_Immutability(t *testing.T) {
	// Test InvoiceNumber immutability
	originalNumber := "INV2024010001"
	invoiceNum, err := NewInvoiceNumber(originalNumber)
	require.NoError(t, err)
	
	// Modifying the original string should not affect the invoice number
	originalNumber = "MODIFIED"
	assert.Equal(t, "INV2024010001", invoiceNum.Value())
	
	// Test InvoiceID immutability
	originalID := uint(123)
	invoiceID := NewInvoiceID(originalID)
	
	// Modifying the original value should not affect the invoice ID
	originalID = 999
	assert.Equal(t, uint(123), invoiceID.Value())
}