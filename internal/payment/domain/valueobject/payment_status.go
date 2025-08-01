package valueobject

import "fmt"

// PaymentStatus represents the status of a payment
type PaymentStatus struct {
	value string
}

// Payment status constants
const (
	PaymentStatusPending    = "pending"
	PaymentStatusProcessing = "processing"
	PaymentStatusCompleted  = "completed"
	PaymentStatusFailed     = "failed"
	PaymentStatusCancelled  = "cancelled"
)

var validPaymentStatuses = map[string]bool{
	PaymentStatusPending:    true,
	PaymentStatusProcessing: true,
	PaymentStatusCompleted:  true,
	PaymentStatusFailed:     true,
	PaymentStatusCancelled:  true,
}

// Status transition rules
var validStatusTransitions = map[string][]string{
	PaymentStatusPending: {
		PaymentStatusProcessing,
		PaymentStatusCompleted,
		PaymentStatusFailed,
		PaymentStatusCancelled,
	},
	PaymentStatusProcessing: {
		PaymentStatusCompleted,
		PaymentStatusFailed,
		PaymentStatusCancelled,
	},
	PaymentStatusCompleted: {
		// Completed payments cannot change status
		// Refunds are handled as separate payment records
	},
	PaymentStatusFailed: {
		// Failed payments cannot change status
	},
	PaymentStatusCancelled: {
		// Cancelled payments cannot change status
	},
}

// NewPaymentStatus creates a new PaymentStatus with validation
func NewPaymentStatus(value string) (PaymentStatus, error) {
	if value == "" {
		return PaymentStatus{}, fmt.Errorf("payment status cannot be empty")
	}
	
	if !validPaymentStatuses[value] {
		return PaymentStatus{}, fmt.Errorf("invalid payment status: %s", value)
	}
	
	return PaymentStatus{value: value}, nil
}

// NewPendingPaymentStatus creates a new pending payment status
func NewPendingPaymentStatus() PaymentStatus {
	// This should never fail since we use a valid constant
	status, _ := NewPaymentStatus(PaymentStatusPending)
	return status
}

// NewProcessingPaymentStatus creates a new processing payment status
func NewProcessingPaymentStatus() PaymentStatus {
	status, _ := NewPaymentStatus(PaymentStatusProcessing)
	return status
}

// NewCompletedPaymentStatus creates a new completed payment status
func NewCompletedPaymentStatus() PaymentStatus {
	status, _ := NewPaymentStatus(PaymentStatusCompleted)
	return status
}

// NewFailedPaymentStatus creates a new failed payment status
func NewFailedPaymentStatus() PaymentStatus {
	status, _ := NewPaymentStatus(PaymentStatusFailed)
	return status
}

// NewCancelledPaymentStatus creates a new cancelled payment status
func NewCancelledPaymentStatus() PaymentStatus {
	status, _ := NewPaymentStatus(PaymentStatusCancelled)
	return status
}

// Value returns the underlying string value
func (ps PaymentStatus) Value() string {
	return ps.value
}

// String returns string representation
func (ps PaymentStatus) String() string {
	return ps.value
}

// Equals checks if two PaymentStatuses are equal
func (ps PaymentStatus) Equals(other PaymentStatus) bool {
	return ps.value == other.value
}

// IsPending checks if the status is pending
func (ps PaymentStatus) IsPending() bool {
	return ps.value == PaymentStatusPending
}

// IsProcessing checks if the status is processing
func (ps PaymentStatus) IsProcessing() bool {
	return ps.value == PaymentStatusProcessing
}

// IsCompleted checks if the status is completed
func (ps PaymentStatus) IsCompleted() bool {
	return ps.value == PaymentStatusCompleted
}

// IsFailed checks if the status is failed
func (ps PaymentStatus) IsFailed() bool {
	return ps.value == PaymentStatusFailed
}

// IsCancelled checks if the status is cancelled
func (ps PaymentStatus) IsCancelled() bool {
	return ps.value == PaymentStatusCancelled
}

// IsFinished checks if the payment is in a final state
func (ps PaymentStatus) IsFinished() bool {
	return ps.IsCompleted() || ps.IsFailed() || ps.IsCancelled()
}

// CanTransitionTo checks if the status can transition to the target status
func (ps PaymentStatus) CanTransitionTo(target PaymentStatus) bool {
	allowedTransitions, exists := validStatusTransitions[ps.value]
	if !exists {
		return false
	}
	
	for _, allowedStatus := range allowedTransitions {
		if allowedStatus == target.value {
			return true
		}
	}
	
	return false
}

// GetAllowedTransitions returns all allowed status transitions from current status
func (ps PaymentStatus) GetAllowedTransitions() []PaymentStatus {
	allowedTransitions, exists := validStatusTransitions[ps.value]
	if !exists {
		return []PaymentStatus{}
	}
	
	var transitions []PaymentStatus
	for _, status := range allowedTransitions {
		// This should never fail since we control the values
		if ps, err := NewPaymentStatus(status); err == nil {
			transitions = append(transitions, ps)
		}
	}
	
	return transitions
}