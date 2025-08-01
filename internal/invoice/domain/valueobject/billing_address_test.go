package valueobject

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewBillingAddress(t *testing.T) {
	tests := []struct {
		name        string
		billingName string
		email       string
		address     string
		city        string
		state       string
		country     string
		zip         string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid complete billing address",
			billingName: "John Doe",
			email:       "john@example.com",
			address:     "123 Main Street",
			city:        "New York",
			state:       "NY",
			country:     "US",
			zip:         "10001",
			expectError: false,
		},
		{
			name:        "valid minimal billing address",
			billingName: "Jane Smith",
			email:       "jane@example.com",
			address:     "",
			city:        "",
			state:       "",
			country:     "",
			zip:         "",
			expectError: false,
		},
		{
			name:        "billing address with whitespace",
			billingName: "  Bob Johnson  ",
			email:       "  bob@example.com  ",
			address:     "  456 Oak Ave  ",
			city:        "  Boston  ",
			state:       "  MA  ",
			country:     "  us  ",
			zip:         "  02101  ",
			expectError: false,
		},
		{
			name:        "empty billing name should fail",
			billingName: "",
			email:       "test@example.com",
			address:     "123 Main St",
			city:        "City",
			state:       "State",
			country:     "US",
			zip:         "12345",
			expectError: true,
			errorMsg:    "billing name is required",
		},
		{
			name:        "whitespace only billing name should fail",
			billingName: "   ",
			email:       "test@example.com",
			address:     "123 Main St",
			city:        "City",
			state:       "State",
			country:     "US",
			zip:         "12345",
			expectError: true,
			errorMsg:    "billing name is required",
		},
		{
			name:        "empty email should fail",
			billingName: "John Doe",
			email:       "",
			address:     "123 Main St",
			city:        "City",
			state:       "State",
			country:     "US",
			zip:         "12345",
			expectError: true,
			errorMsg:    "billing email is required",
		},
		{
			name:        "invalid email format should fail",
			billingName: "John Doe",
			email:       "invalid-email",
			address:     "123 Main St",
			city:        "City",
			state:       "State",
			country:     "US",
			zip:         "12345",
			expectError: true,
			errorMsg:    "invalid billing email",
		},
		{
			name:        "billing name too long should fail",
			billingName: strings.Repeat("A", 256),
			email:       "test@example.com",
			address:     "123 Main St",
			city:        "City",
			state:       "State",
			country:     "US",
			zip:         "12345",
			expectError: true,
			errorMsg:    "billing name cannot exceed 255 characters",
		},
		{
			name:        "email too long should fail",
			billingName: "John Doe",
			email:       strings.Repeat("a", 250) + "@example.com",
			address:     "123 Main St",
			city:        "City",
			state:       "State",
			country:     "US",
			zip:         "12345",
			expectError: true,
			errorMsg:    "billing email cannot exceed 255 characters",
		},
		{
			name:        "city too long should fail",
			billingName: "John Doe",
			email:       "test@example.com",
			address:     "123 Main St",
			city:        strings.Repeat("C", 101),
			state:       "State",
			country:     "US",
			zip:         "12345",
			expectError: true,
			errorMsg:    "billing city cannot exceed 100 characters",
		},
		{
			name:        "state too long should fail",
			billingName: "John Doe",
			email:       "test@example.com",
			address:     "123 Main St",
			city:        "City",
			state:       strings.Repeat("S", 101),
			country:     "US",
			zip:         "12345",
			expectError: true,
			errorMsg:    "billing state cannot exceed 100 characters",
		},
		{
			name:        "zip too long should fail",
			billingName: "John Doe",
			email:       "test@example.com",
			address:     "123 Main St",
			city:        "City",
			state:       "State",
			country:     "US",
			zip:         strings.Repeat("1", 21),
			expectError: true,
			errorMsg:    "billing zip cannot exceed 20 characters",
		},
		{
			name:        "country code too short should fail",
			billingName: "John Doe",
			email:       "test@example.com",
			address:     "123 Main St",
			city:        "City",
			state:       "State",
			country:     "U",
			zip:         "12345",
			expectError: true,
			errorMsg:    "country code must be exactly 2 characters",
		},
		{
			name:        "country code too long should fail",
			billingName: "John Doe",
			email:       "test@example.com",
			address:     "123 Main St",
			city:        "City",
			state:       "State",
			country:     "USA",
			zip:         "12345",
			expectError: true,
			errorMsg:    "country code must be exactly 2 characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			billingAddr, err := NewBillingAddress(
				tt.billingName,
				tt.email,
				tt.address,
				tt.city,
				tt.state,
				tt.country,
				tt.zip,
			)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				require.NoError(t, err)
				assert.Equal(t, strings.TrimSpace(tt.billingName), billingAddr.Name())
				assert.Equal(t, strings.TrimSpace(tt.email), billingAddr.Email())
				assert.Equal(t, strings.TrimSpace(tt.address), billingAddr.Address())
				assert.Equal(t, strings.TrimSpace(tt.city), billingAddr.City())
				assert.Equal(t, strings.TrimSpace(tt.state), billingAddr.State())
				assert.Equal(t, strings.ToUpper(strings.TrimSpace(tt.country)), billingAddr.Country())
				assert.Equal(t, strings.TrimSpace(tt.zip), billingAddr.Zip())
			}
		})
	}
}

func TestBillingAddress_Getters(t *testing.T) {
	billingAddr, err := NewBillingAddress(
		"John Doe",
		"john@example.com",
		"123 Main Street",
		"New York",
		"NY",
		"US",
		"10001",
	)
	require.NoError(t, err)

	assert.Equal(t, "John Doe", billingAddr.Name())
	assert.Equal(t, "john@example.com", billingAddr.Email())
	assert.Equal(t, "123 Main Street", billingAddr.Address())
	assert.Equal(t, "New York", billingAddr.City())
	assert.Equal(t, "NY", billingAddr.State())
	assert.Equal(t, "US", billingAddr.Country())
	assert.Equal(t, "10001", billingAddr.Zip())
}

func TestBillingAddress_FullAddress(t *testing.T) {
	tests := []struct {
		name        string
		billingName string
		email       string
		address     string
		city        string
		state       string
		country     string
		zip         string
		expected    string
	}{
		{
			name:        "complete address",
			billingName: "John Doe",
			email:       "john@example.com",
			address:     "123 Main Street",
			city:        "New York",
			state:       "NY",
			country:     "US",
			zip:         "10001",
			expected:    "123 Main Street\nNew York, NY 10001\nUS",
		},
		{
			name:        "address without street",
			billingName: "Jane Smith",
			email:       "jane@example.com",
			address:     "",
			city:        "Boston",
			state:       "MA",
			country:     "US",
			zip:         "02101",
			expected:    "Boston, MA 02101\nUS",
		},
		{
			name:        "address with only city",
			billingName: "Bob Johnson",
			email:       "bob@example.com",
			address:     "",
			city:        "Chicago",
			state:       "",
			country:     "",
			zip:         "",
			expected:    "Chicago",
		},
		{
			name:        "address with only zip",
			billingName: "Alice Brown",
			email:       "alice@example.com",
			address:     "",
			city:        "",
			state:       "",
			country:     "",
			zip:         "90210",
			expected:    "90210",
		},
		{
			name:        "minimal address",
			billingName: "Charlie Wilson",
			email:       "charlie@example.com",
			address:     "",
			city:        "",
			state:       "",
			country:     "",
			zip:         "",
			expected:    "",
		},
		{
			name:        "only country",
			billingName: "David Lee",
			email:       "david@example.com",
			address:     "",
			city:        "",
			state:       "",
			country:     "CA",
			zip:         "",
			expected:    "CA",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			billingAddr, err := NewBillingAddress(
				tt.billingName,
				tt.email,
				tt.address,
				tt.city,
				tt.state,
				tt.country,
				tt.zip,
			)
			require.NoError(t, err)

			assert.Equal(t, tt.expected, billingAddr.FullAddress())
		})
	}
}

func TestBillingAddress_Equals(t *testing.T) {
	addr1, _ := NewBillingAddress(
		"John Doe",
		"john@example.com",
		"123 Main St",
		"New York",
		"NY",
		"US",
		"10001",
	)

	addr2, _ := NewBillingAddress(
		"John Doe",
		"john@example.com",
		"123 Main St",
		"New York",
		"NY",
		"US",
		"10001",
	)

	addr3, _ := NewBillingAddress(
		"Jane Smith",
		"jane@example.com",
		"456 Oak Ave",
		"Boston",
		"MA",
		"US",
		"02101",
	)

	// Test equality
	assert.True(t, addr1.Equals(addr2))
	assert.True(t, addr2.Equals(addr1))

	// Test inequality
	assert.False(t, addr1.Equals(addr3))
	assert.False(t, addr3.Equals(addr1))
}

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		name        string
		email       string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid email",
			email:       "test@example.com",
			expectError: false,
		},
		{
			name:        "valid email with subdomain",
			email:       "user@mail.example.com",
			expectError: false,
		},
		{
			name:        "valid email with numbers",
			email:       "user123@example123.com",
			expectError: false,
		},
		{
			name:        "valid email with special characters",
			email:       "user.name+tag@example.com",
			expectError: false,
		},
		{
			name:        "empty email should fail",
			email:       "",
			expectError: true,
			errorMsg:    "email cannot be empty",
		},
		{
			name:        "email without @ should fail",
			email:       "testexample.com",
			expectError: true,
			errorMsg:    "invalid email format",
		},
		{
			name:        "email without domain should fail",
			email:       "test@",
			expectError: true,
			errorMsg:    "invalid email format",
		},
		{
			name:        "email without local part should fail",
			email:       "@example.com",
			expectError: true,
			errorMsg:    "invalid email format",
		},
		{
			name:        "email with invalid characters should fail",
			email:       "test@exam ple.com",
			expectError: true,
			errorMsg:    "invalid email format",
		},
		{
			name:        "email without TLD should fail",
			email:       "test@example",
			expectError: true,
			errorMsg:    "invalid email format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateEmail(tt.email)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestBillingAddress_CountryNormalization(t *testing.T) {
	tests := []struct {
		name     string
		country  string
		expected string
	}{
		{
			name:     "lowercase country code",
			country:  "us",
			expected: "US",
		},
		{
			name:     "mixed case country code",
			country:  "Ca",
			expected: "CA",
		},
		{
			name:     "uppercase country code",
			country:  "GB",
			expected: "GB",
		},
		{
			name:     "empty country code",
			country:  "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			addr, err := NewBillingAddress(
				"Test User",
				"test@example.com",
				"123 Test St",
				"Test City",
				"Test State",
				tt.country,
				"12345",
			)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, addr.Country())
		})
	}
}

func TestBillingAddress_Immutability(t *testing.T) {
	// Test that BillingAddress is immutable
	originalName := "John Doe"
	originalEmail := "john@example.com"

	addr, err := NewBillingAddress(
		originalName,
		originalEmail,
		"123 Main St",
		"New York",
		"NY",
		"US",
		"10001",
	)
	require.NoError(t, err)

	// Verify original values
	assert.Equal(t, originalName, addr.Name())
	assert.Equal(t, originalEmail, addr.Email())

	// Modifying the original strings should not affect the address
	originalName = "Modified Name"
	originalEmail = "modified@example.com"

	// Address should still have original values
	assert.Equal(t, "John Doe", addr.Name())
	assert.Equal(t, "john@example.com", addr.Email())
}