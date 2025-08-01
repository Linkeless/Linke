package valueobject

import (
	"fmt"
	"regexp"
	"time"
)

// PaymentNumber represents a unique payment number
type PaymentNumber struct {
	value string
}

var paymentNumberPattern = regexp.MustCompile(`^[A-Z]{2,4}[0-9]{12,20}$`)

// NewPaymentNumber creates a new PaymentNumber with validation
func NewPaymentNumber(value string) (PaymentNumber, error) {
	if value == "" {
		return PaymentNumber{}, fmt.Errorf("payment number cannot be empty")
	}
	
	if len(value) < 15 || len(value) > 32 {
		return PaymentNumber{}, fmt.Errorf("payment number length must be between 15 and 32 characters")
	}
	
	if !paymentNumberPattern.MatchString(value) {
		return PaymentNumber{}, fmt.Errorf("payment number format is invalid")
	}
	
	return PaymentNumber{value: value}, nil
}

// GeneratePaymentNumber generates a new payment number with prefix
func GeneratePaymentNumber(prefix string) PaymentNumber {
	if prefix == "" {
		prefix = "PMT"
	}
	
	now := time.Now()
	timestamp := now.Format("20060102150405")
	nano := now.Nanosecond() % 1000
	
	value := fmt.Sprintf("%s%s%03d", prefix, timestamp, nano)
	
	// This should never fail since we control the format
	paymentNumber, _ := NewPaymentNumber(value)
	return paymentNumber
}

// GenerateRefundNumber generates a refund number based on original payment number
func GenerateRefundNumber(originalPaymentNumber PaymentNumber) PaymentNumber {
	now := time.Now()
	timestamp := now.Format("20060102150405")
	nano := now.Nanosecond() % 1000
	
	value := fmt.Sprintf("RFD%s%03d", timestamp, nano)
	
	// This should never fail since we control the format
	refundNumber, _ := NewPaymentNumber(value)
	return refundNumber
}

// Value returns the underlying string value
func (pn PaymentNumber) Value() string {
	return pn.value
}

// String returns string representation
func (pn PaymentNumber) String() string {
	return pn.value
}

// Equals checks if two PaymentNumbers are equal
func (pn PaymentNumber) Equals(other PaymentNumber) bool {
	return pn.value == other.value
}

// IsEmpty checks if the payment number is empty
func (pn PaymentNumber) IsEmpty() bool {
	return pn.value == ""
}

// IsRefund checks if this is a refund number
func (pn PaymentNumber) IsRefund() bool {
	return len(pn.value) >= 3 && pn.value[:3] == "RFD"
}