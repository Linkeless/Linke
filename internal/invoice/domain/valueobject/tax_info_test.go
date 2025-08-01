package valueobject

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTaxInfo(t *testing.T) {
	usd, _ := NewCurrency("USD")
	taxAmount, _ := NewMoney(10.00, usd)

	tests := []struct {
		name        string
		rate        float64
		taxType     string
		taxNumber   string
		taxAmount   Money
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid tax info with VAT",
			rate:        0.20,
			taxType:     "VAT",
			taxNumber:   "GB123456789",
			taxAmount:   taxAmount,
			expectError: false,
		},
		{
			name:        "valid tax info with GST",
			rate:        0.15,
			taxType:     "GST",
			taxNumber:   "CA123456789",
			taxAmount:   taxAmount,
			expectError: false,
		},
		{
			name:        "valid tax info with lowercase type",
			rate:        0.08,
			taxType:     "hst",
			taxNumber:   "ON123456789",
			taxAmount:   taxAmount,
			expectError: false,
		},
		{
			name:        "valid tax info with whitespace",
			rate:        0.10,
			taxType:     "  TAX  ",
			taxNumber:   "  TX123  ",
			taxAmount:   taxAmount,
			expectError: false,
		},
		{
			name:        "valid zero tax",
			rate:        0.00,
			taxType:     "",
			taxNumber:   "",
			taxAmount:   Zero(usd),
			expectError: false,
		},
		{
			name:        "negative tax rate should fail",
			rate:        -0.10,
			taxType:     "VAT",
			taxNumber:   "GB123456789",
			taxAmount:   taxAmount,
			expectError: true,
			errorMsg:    "tax rate must be between 0 and 1",
		},
		{
			name:        "tax rate greater than 1 should fail",
			rate:        1.20,
			taxType:     "VAT",
			taxNumber:   "GB123456789",
			taxAmount:   taxAmount,
			expectError: true,
			errorMsg:    "tax rate must be between 0 and 1",
		},
		{
			name:        "invalid tax type should fail",
			rate:        0.20,
			taxType:     "UNKNOWN",
			taxNumber:   "GB123456789",
			taxAmount:   taxAmount,
			expectError: true,
			errorMsg:    "invalid tax type: UNKNOWN",
		},
		{
			name:        "tax number too long should fail",
			rate:        0.20,
			taxType:     "VAT",
			taxNumber:   "123456789012345678901234567890123456789012345678901",
			taxAmount:   taxAmount,
			expectError: true,
			errorMsg:    "tax number cannot exceed 50 characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			taxInfo, err := NewTaxInfo(tt.rate, tt.taxType, tt.taxNumber, tt.taxAmount)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.rate, taxInfo.Rate())
				assert.Equal(t, tt.rate*100, taxInfo.RatePercentage())
				
				if tt.taxType != "" {
					expectedType := "HST"
					if tt.taxType == "  TAX  " {
						expectedType = "TAX"
					} else if tt.taxType == "hst" {
						expectedType = "HST"
					} else {
						expectedType = tt.taxType
					}
					assert.Equal(t, expectedType, taxInfo.TaxType())
				}
				
				assert.Equal(t, tt.taxAmount.Amount(), taxInfo.TaxAmount().Amount())
			}
		})
	}
}

func TestNewZeroTax(t *testing.T) {
	usd, _ := NewCurrency("USD")
	
	zeroTax := NewZeroTax(usd)
	
	assert.Equal(t, 0.0, zeroTax.Rate())
	assert.Equal(t, 0.0, zeroTax.RatePercentage())
	assert.Equal(t, "", zeroTax.TaxType())
	assert.Equal(t, "", zeroTax.TaxNumber())
	assert.True(t, zeroTax.TaxAmount().IsZero())
	assert.True(t, zeroTax.IsZero())
	assert.False(t, zeroTax.HasTaxNumber())
}

func TestTaxInfo_CalculateTaxAmount(t *testing.T) {
	usd, _ := NewCurrency("USD")
	taxAmount, _ := NewMoney(20.00, usd)
	
	tests := []struct {
		name         string
		rate         float64
		subtotal     Money
		expectedTax  float64
		expectError  bool
	}{
		{
			name:        "calculate 20% VAT",
			rate:        0.20,
			subtotal:    mustCreateMoney(100.00, usd),
			expectedTax: 20.00,
			expectError: false,
		},
		{
			name:        "calculate 0% tax",
			rate:        0.00,
			subtotal:    mustCreateMoney(100.00, usd),
			expectedTax: 0.00,
			expectError: false,
		},
		{
			name:        "calculate 15% GST",
			rate:        0.15,
			subtotal:    mustCreateMoney(200.00, usd),
			expectedTax: 30.00,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			taxInfo, err := NewTaxInfo(tt.rate, "VAT", "GB123456789", taxAmount)
			require.NoError(t, err)

			calculatedTax, err := taxInfo.CalculateTaxAmount(tt.subtotal)
			
			if tt.expectError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedTax, calculatedTax.Amount())
			}
		})
	}
}

func TestTaxInfo_Checkers(t *testing.T) {
	usd, _ := NewCurrency("USD")
	
	tests := []struct {
		name          string
		rate          float64
		taxType       string
		taxNumber     string
		isZero        bool
		hasTaxNumber  bool
	}{
		{
			name:         "non-zero tax with number",
			rate:         0.20,
			taxType:      "VAT",
			taxNumber:    "GB123456789",
			isZero:       false,
			hasTaxNumber: true,
		},
		{
			name:         "non-zero tax without number",
			rate:         0.15,
			taxType:      "GST",
			taxNumber:    "",
			isZero:       false,
			hasTaxNumber: false,
		},
		{
			name:         "zero tax",
			rate:         0.00,
			taxType:      "",
			taxNumber:    "",
			isZero:       true,
			hasTaxNumber: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			taxAmount := Zero(usd)
			if tt.rate > 0 {
				taxAmount, _ = NewMoney(tt.rate*100, usd) // Example tax amount
			}
			
			taxInfo, err := NewTaxInfo(tt.rate, tt.taxType, tt.taxNumber, taxAmount)
			require.NoError(t, err)

			assert.Equal(t, tt.isZero, taxInfo.IsZero())
			assert.Equal(t, tt.hasTaxNumber, taxInfo.HasTaxNumber())
		})
	}
}

func TestTaxInfo_Equals(t *testing.T) {
	usd, _ := NewCurrency("USD")
	taxAmount1, _ := NewMoney(20.00, usd)
	taxAmount2, _ := NewMoney(20.00, usd)
	taxAmount3, _ := NewMoney(15.00, usd)

	taxInfo1, _ := NewTaxInfo(0.20, "VAT", "GB123456789", taxAmount1)
	taxInfo2, _ := NewTaxInfo(0.20, "VAT", "GB123456789", taxAmount2)
	taxInfo3, _ := NewTaxInfo(0.15, "GST", "CA123456789", taxAmount3)

	// Test equality
	assert.True(t, taxInfo1.Equals(taxInfo2))
	assert.True(t, taxInfo2.Equals(taxInfo1))

	// Test inequality
	assert.False(t, taxInfo1.Equals(taxInfo3))
	assert.False(t, taxInfo3.Equals(taxInfo1))
}

func TestTaxInfo_String(t *testing.T) {
	usd, _ := NewCurrency("USD")
	
	tests := []struct {
		name        string
		rate        float64
		taxType     string
		taxNumber   string
		taxAmount   Money
		expected    string
	}{
		{
			name:      "VAT with tax number",
			rate:      0.20,
			taxType:   "VAT",
			taxNumber: "GB123456789",
			taxAmount: mustCreateMoney(20.00, usd),
			expected:  "20.00% VAT (GB123456789) = 20.00 USD",
		},
		{
			name:      "GST without tax number",
			rate:      0.15,
			taxType:   "GST",
			taxNumber: "",
			taxAmount: mustCreateMoney(15.00, usd),
			expected:  "15.00% GST = 15.00 USD",
		},
		{
			name:      "zero tax",
			rate:      0.00,
			taxType:   "",
			taxNumber: "",
			taxAmount: Zero(usd),
			expected:  "No Tax",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			taxInfo, err := NewTaxInfo(tt.rate, tt.taxType, tt.taxNumber, tt.taxAmount)
			require.NoError(t, err)
			
			assert.Equal(t, tt.expected, taxInfo.String())
		})
	}
}

func TestNewCompanyInfo(t *testing.T) {
	tests := []struct {
		name        string
		companyName string
		address     string
		taxID       string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid company info",
			companyName: "ACME Corporation",
			address:     "123 Business St, City, State",
			taxID:       "123456789",
			expectError: false,
		},
		{
			name:        "empty company info is valid",
			companyName: "",
			address:     "",
			taxID:       "",
			expectError: false,
		},
		{
			name:        "company info with whitespace",
			companyName: "  Tech Corp  ",
			address:     "  456 Tech Ave  ",
			taxID:       "  987654321  ",
			expectError: false,
		},
		{
			name:        "company name too long should fail",
			companyName: "A" + string(make([]byte, 255)), // 256 characters
			address:     "123 Main St",
			taxID:       "123456789",
			expectError: true,
			errorMsg:    "company name cannot exceed 255 characters",
		},
		{
			name:        "tax ID too long should fail",
			companyName: "ACME Corp",
			address:     "123 Main St",
			taxID:       "123456789012345678901234567890123456789012345678901", // 51 characters
			expectError: true,
			errorMsg:    "company tax ID cannot exceed 50 characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			companyInfo, err := NewCompanyInfo(tt.companyName, tt.address, tt.taxID)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				require.NoError(t, err)
				assert.Equal(t, strings.TrimSpace(tt.companyName), companyInfo.Name())
				assert.Equal(t, strings.TrimSpace(tt.address), companyInfo.Address())
				assert.Equal(t, strings.TrimSpace(tt.taxID), companyInfo.TaxID())
			}
		})
	}
}

func TestCompanyInfo_Checkers(t *testing.T) {
	tests := []struct {
		name        string
		companyName string
		address     string
		taxID       string
		isEmpty     bool
		hasTaxID    bool
	}{
		{
			name:        "complete company info",
			companyName: "ACME Corp",
			address:     "123 Main St",
			taxID:       "123456789",
			isEmpty:     false,
			hasTaxID:    true,
		},
		{
			name:        "company without tax ID",
			companyName: "Tech Corp",
			address:     "456 Tech Ave",
			taxID:       "",
			isEmpty:     false,
			hasTaxID:    false,
		},
		{
			name:        "empty company info",
			companyName: "",
			address:     "",
			taxID:       "",
			isEmpty:     true,
			hasTaxID:    false,
		},
		{
			name:        "only tax ID provided",
			companyName: "",
			address:     "",
			taxID:       "123456789",
			isEmpty:     false,
			hasTaxID:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			companyInfo, err := NewCompanyInfo(tt.companyName, tt.address, tt.taxID)
			require.NoError(t, err)

			assert.Equal(t, tt.isEmpty, companyInfo.IsEmpty())
			assert.Equal(t, tt.hasTaxID, companyInfo.HasTaxID())
		})
	}
}

func TestCompanyInfo_Equals(t *testing.T) {
	company1, _ := NewCompanyInfo("ACME Corp", "123 Main St", "123456789")
	company2, _ := NewCompanyInfo("ACME Corp", "123 Main St", "123456789")
	company3, _ := NewCompanyInfo("Tech Corp", "456 Tech Ave", "987654321")

	// Test equality
	assert.True(t, company1.Equals(company2))
	assert.True(t, company2.Equals(company1))

	// Test inequality
	assert.False(t, company1.Equals(company3))
	assert.False(t, company3.Equals(company1))
}