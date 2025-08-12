package dto

import (
	"testing"
)

func TestValidateEpayConfig(t *testing.T) {
	tests := []struct {
		name          string
		req           *CreatePaymentConfigRequest
		expectedErrors int
	}{
		{
			name: "Valid epay config",
			req: &CreatePaymentConfigRequest{
				Method:    "epay",
				URL:       "https://api.example.com",
				PID:       "123456",
				Key:       "secret_key_123",
				NotifyURL: "https://example.com/notify",
				ReturnURL: "https://example.com/return",
			},
			expectedErrors: 0,
		},
		{
			name: "Missing required fields",
			req: &CreatePaymentConfigRequest{
				Method: "epay",
				// Missing URL, PID, Key
			},
			expectedErrors: 3,
		},
		{
			name: "Invalid URL format",
			req: &CreatePaymentConfigRequest{
				Method:    "epay",
				URL:       "invalid-url",
				PID:       "123456",
				Key:       "secret_key_123",
				NotifyURL: "invalid-notify-url",
			},
			expectedErrors: 2, // Invalid URL and NotifyURL
		},
		{
			name: "Non-epay method should not be validated",
			req: &CreatePaymentConfigRequest{
				Method: "stripe",
				// Missing fields but should not be validated for non-epay
			},
			expectedErrors: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := ValidateEpayConfig(tt.req)
			if len(errors) != tt.expectedErrors {
				t.Errorf("ValidateEpayConfig() returned %d errors, expected %d. Errors: %v",
					len(errors), tt.expectedErrors, errors)
			}
		})
	}
}

func TestValidateEpayUpdateConfig(t *testing.T) {
	tests := []struct {
		name           string
		req            *UpdatePaymentConfigRequest
		existingMethod string
		expectedErrors int
	}{
		{
			name: "Valid epay update config",
			req: &UpdatePaymentConfigRequest{
				URL:       stringPtr("https://api.example.com"),
				NotifyURL: stringPtr("https://example.com/notify"),
			},
			existingMethod: "epay",
			expectedErrors: 0,
		},
		{
			name: "Invalid URL in update",
			req: &UpdatePaymentConfigRequest{
				URL:       stringPtr("invalid-url"),
				NotifyURL: stringPtr("invalid-notify"),
			},
			existingMethod: "epay",
			expectedErrors: 2,
		},
		{
			name: "Non-epay method should not be validated",
			req: &UpdatePaymentConfigRequest{
				URL: stringPtr("invalid-url"),
			},
			existingMethod: "stripe",
			expectedErrors: 0,
		},
		{
			name: "Empty URLs should be valid",
			req: &UpdatePaymentConfigRequest{
				URL:       stringPtr(""),
				NotifyURL: stringPtr(""),
			},
			existingMethod: "epay",
			expectedErrors: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := ValidateEpayUpdateConfig(tt.req, tt.existingMethod)
			if len(errors) != tt.expectedErrors {
				t.Errorf("ValidateEpayUpdateConfig() returned %d errors, expected %d. Errors: %v",
					len(errors), tt.expectedErrors, errors)
			}
		})
	}
}

func TestIsValidURL(t *testing.T) {
	tests := []struct {
		url      string
		expected bool
	}{
		{"https://example.com", true},
		{"http://example.com", true},
		{"https://api.payment.com/webhook", true},
		{"ftp://example.com", false},
		{"invalid-url", false},
		{"", false},
		{"just-text", false},
	}

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			result := isValidURL(tt.url)
			if result != tt.expected {
				t.Errorf("isValidURL(%q) = %v, expected %v", tt.url, result, tt.expected)
			}
		})
	}
}

// Helper function to create string pointers
func stringPtr(s string) *string {
	return &s
}