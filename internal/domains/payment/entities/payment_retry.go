package entities

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"

	"linke/internal/domains/payment/constants"
)

// PaymentRetry represents a payment retry record
type PaymentRetry struct {
	// Primary Key
	ID uint `json:"id" gorm:"primaryKey"`

	// Foreign Keys
	PaymentRecordID uint `json:"payment_record_id" gorm:"not null;index"`

	// Retry Information
	AttemptNumber int       `json:"attempt_number" gorm:"not null;default:0"` // Current attempt number (0-based)
	MaxAttempts   int       `json:"max_attempts" gorm:"not null;default:3"`   // Maximum retry attempts
	NextRetryAt   time.Time `json:"next_retry_at" gorm:"not null;index"`      // Next retry time
	LastAttemptAt time.Time `json:"last_attempt_at" gorm:"not null"`          // Last attempt time
	RetryStrategy string    `json:"retry_strategy" gorm:"size:50;not null"`   // Strategy type: exponential, linear, custom

	// Retry Configuration
	InitialDelay  int     `json:"initial_delay" gorm:"not null;default:3600"`                   // Initial delay in seconds (1 hour)
	MaxDelay      int     `json:"max_delay" gorm:"not null;default:86400"`                      // Maximum delay in seconds (24 hours)
	BackoffFactor float64 `json:"backoff_factor" gorm:"type:decimal(4,2);not null;default:2.0"` // Backoff multiplier

	// Status and State
	Status           string `json:"status" gorm:"size:20;not null;default:'pending';index"` // pending, in_progress, completed, failed, cancelled
	FailureType      string `json:"failure_type" gorm:"size:30;index"`                      // temporary, permanent, network, gateway, business
	LastFailureCode  string `json:"last_failure_code" gorm:"size:50"`                       // Last error/failure code
	LastErrorMessage string `json:"last_error_message" gorm:"size:500"`                     // Last error message

	// Gateway-specific Configuration
	GatewayConfig string `json:"gateway_config,omitempty" gorm:"type:text"` // JSON config for gateway-specific retry settings

	// Tracking Information
	TotalDelayTime int        `json:"total_delay_time" gorm:"default:0"`    // Total time spent in retries (seconds)
	CompletedAt    *time.Time `json:"completed_at,omitempty" gorm:"index"`  // When retry sequence completed
	CancelledAt    *time.Time `json:"cancelled_at,omitempty" gorm:"index"`  // When retry sequence was cancelled
	SuccessfulAt   *time.Time `json:"successful_at,omitempty" gorm:"index"` // When payment finally succeeded

	// Metadata
	Metadata string `json:"metadata,omitempty" gorm:"type:text"` // Additional retry metadata (JSON)
	Notes    string `json:"notes,omitempty" gorm:"size:500"`     // Admin notes

	// Timestamp Fields
	CreatedAt time.Time      `json:"created_at" gorm:"not null;index"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"not null"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index" swaggerignore:"true"`
}

// TableName returns the table name for PaymentRetry model
func (PaymentRetry) TableName() string {
	return "payment_retries"
}


// PaymentRetryHistory represents individual retry attempt history
type PaymentRetryHistory struct {
	// Primary Key
	ID uint `json:"id" gorm:"primaryKey"`

	// Foreign Keys
	PaymentRetryID  uint `json:"payment_retry_id" gorm:"not null;index"`
	PaymentRecordID uint `json:"payment_record_id" gorm:"not null;index"`

	// Attempt Information
	AttemptNumber int       `json:"attempt_number" gorm:"not null"` // Which attempt this was
	AttemptedAt   time.Time `json:"attempted_at" gorm:"not null"`   // When this attempt was made
	Duration      int       `json:"duration" gorm:"default:0"`      // Duration of attempt in milliseconds

	// Result Information
	Status          string `json:"status" gorm:"size:20;not null"`   // success, failed, timeout, error
	ResponseCode    string `json:"response_code" gorm:"size:50"`     // Gateway response code
	ResponseMessage string `json:"response_message" gorm:"size:500"` // Gateway response message
	ErrorType       string `json:"error_type" gorm:"size:30"`        // Type of error encountered
	FailureReason   string `json:"failure_reason" gorm:"size:500"`   // Detailed failure reason

	// Technical Details
	RequestData  string `json:"request_data,omitempty" gorm:"type:text"`  // Request sent to gateway (sanitized)
	ResponseData string `json:"response_data,omitempty" gorm:"type:text"` // Response from gateway (sanitized)

	// Next Retry Information
	NextRetryAt       *time.Time `json:"next_retry_at,omitempty" gorm:"index"` // When next retry is scheduled
	DelayFromPrevious int        `json:"delay_from_previous" gorm:"default:0"` // Delay from previous attempt (seconds)

	// Metadata
	Metadata string `json:"metadata,omitempty" gorm:"type:text"` // Additional attempt metadata (JSON)

	// Timestamp Fields
	CreatedAt time.Time      `json:"created_at" gorm:"not null;index"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"not null"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index" swaggerignore:"true"`
}

// TableName returns the table name for PaymentRetryHistory model
func (PaymentRetryHistory) TableName() string {
	return "payment_retry_histories"
}


// RetryStrategyConfig represents retry configuration for different gateways
type RetryStrategyConfig struct {
	Gateway          string   `json:"gateway"`
	PaymentMethod    string   `json:"payment_method,omitempty"`
	MaxAttempts      int      `json:"max_attempts"`
	InitialDelay     int      `json:"initial_delay"` // seconds
	MaxDelay         int      `json:"max_delay"`     // seconds
	BackoffFactor    float64  `json:"backoff_factor"`
	Strategy         string   `json:"strategy"`
	CustomDelays     []int    `json:"custom_delays,omitempty"` // for custom strategy
	FailureTypes     []string `json:"failure_types"`           // which failure types to retry
	TimeoutSeconds   int      `json:"timeout_seconds"`
	EnableAfterHours bool     `json:"enable_after_hours"` // allow retries outside business hours
	MaxConcurrent    int      `json:"max_concurrent"`     // max concurrent retries for this gateway
}

// DefaultRetryStrategies defines default retry strategies for different gateways
var DefaultRetryStrategies = map[string]RetryStrategyConfig{
	constants.PaymentGatewayEpay: {
		Gateway:          constants.PaymentGatewayEpay,
		MaxAttempts:      3,
		InitialDelay:     3600,  // 1 hour
		MaxDelay:         86400, // 24 hours
		BackoffFactor:    2.0,
		Strategy:         constants.RetryStrategyExponential,
		FailureTypes:     []string{constants.FailureTypeTemporary, constants.FailureTypeNetwork, constants.FailureTypeGateway},
		TimeoutSeconds:   30,
		EnableAfterHours: true,
		MaxConcurrent:    5,
	},
	constants.PaymentGatewayEPUSDT: {
		Gateway:          constants.PaymentGatewayEPUSDT,
		MaxAttempts:      3,
		InitialDelay:     3600,  // 1 hour
		MaxDelay:         86400, // 24 hours
		BackoffFactor:    2.0,
		Strategy:         constants.RetryStrategyExponential,
		FailureTypes:     []string{constants.FailureTypeTemporary, constants.FailureTypeNetwork, constants.FailureTypeGateway},
		TimeoutSeconds:   30,
		EnableAfterHours: true,
		MaxConcurrent:    5,
	},
}

// Business methods for PaymentRetry

// IsPending checks if retry is pending
func (pr *PaymentRetry) IsPending() bool {
	return pr.Status == constants.PaymentRetryStatusPending
}

// IsInProgress checks if retry is in progress
func (pr *PaymentRetry) IsInProgress() bool {
	return pr.Status == constants.PaymentRetryStatusInProgress
}

// IsCompleted checks if retry sequence is completed
func (pr *PaymentRetry) IsCompleted() bool {
	return pr.Status == constants.PaymentRetryStatusCompleted
}

// IsFailed checks if retry sequence has failed
func (pr *PaymentRetry) IsFailed() bool {
	return pr.Status == constants.PaymentRetryStatusFailed
}

// IsCancelled checks if retry sequence was cancelled
func (pr *PaymentRetry) IsCancelled() bool {
	return pr.Status == constants.PaymentRetryStatusCancelled
}

// IsActive checks if retry is still active (pending or in progress)
func (pr *PaymentRetry) IsActive() bool {
	return pr.IsPending() || pr.IsInProgress()
}

// HasReachedMaxAttempts checks if max attempts have been reached
func (pr *PaymentRetry) HasReachedMaxAttempts() bool {
	return pr.AttemptNumber >= pr.MaxAttempts
}

// ShouldRetry determines if payment should be retried based on failure type
func (pr *PaymentRetry) ShouldRetry() bool {
	if pr.HasReachedMaxAttempts() {
		return false
	}

	// Don't retry permanent failures
	if pr.FailureType == constants.FailureTypePermanent {
		return false
	}

	// Only retry for specific failure types
	retryableTypes := []string{constants.FailureTypeTemporary, constants.FailureTypeNetwork, constants.FailureTypeGateway}
	for _, failureType := range retryableTypes {
		if pr.FailureType == failureType {
			return true
		}
	}

	return false
}

// CalculateNextRetryTime calculates the next retry time based on strategy
func (pr *PaymentRetry) CalculateNextRetryTime() time.Time {
	now := time.Now()

	switch pr.RetryStrategy {
	case constants.RetryStrategyExponential:
		delay := pr.calculateExponentialDelay()
		return now.Add(time.Duration(delay) * time.Second)

	case constants.RetryStrategyLinear:
		delay := pr.calculateLinearDelay()
		return now.Add(time.Duration(delay) * time.Second)

	case constants.RetryStrategyCustom:
		delay := pr.calculateCustomDelay()
		return now.Add(time.Duration(delay) * time.Second)

	default:
		// Default to exponential
		delay := pr.calculateExponentialDelay()
		return now.Add(time.Duration(delay) * time.Second)
	}
}

// calculateExponentialDelay calculates delay using exponential backoff
func (pr *PaymentRetry) calculateExponentialDelay() int {
	if pr.AttemptNumber == 0 {
		return pr.InitialDelay
	}

	// Calculate exponential delay: initial_delay * (backoff_factor ^ attempt_number)
	delay := float64(pr.InitialDelay)
	for i := 0; i < pr.AttemptNumber; i++ {
		delay *= pr.BackoffFactor
	}

	// Ensure we don't exceed max delay
	if int(delay) > pr.MaxDelay {
		return pr.MaxDelay
	}

	return int(delay)
}

// calculateLinearDelay calculates delay using linear backoff
func (pr *PaymentRetry) calculateLinearDelay() int {
	delay := pr.InitialDelay + (pr.AttemptNumber * pr.InitialDelay)

	if delay > pr.MaxDelay {
		return pr.MaxDelay
	}

	return delay
}

// calculateCustomDelay calculates delay using custom strategy
func (pr *PaymentRetry) calculateCustomDelay() int {
	// For custom strategy, delays could be defined in gateway config
	// For now, default to exponential
	return pr.calculateExponentialDelay()
}

// UpdateForNextAttempt updates the retry record for the next attempt
func (pr *PaymentRetry) UpdateForNextAttempt(failureType, failureCode, errorMessage string) {
	pr.AttemptNumber++
	pr.LastAttemptAt = time.Now()
	pr.FailureType = failureType
	pr.LastFailureCode = failureCode
	pr.LastErrorMessage = errorMessage

	if pr.ShouldRetry() {
		pr.NextRetryAt = pr.CalculateNextRetryTime()
		pr.Status = constants.PaymentRetryStatusPending
	} else {
		pr.Status = constants.PaymentRetryStatusFailed
		now := time.Now()
		pr.CompletedAt = &now
	}
}

// MarkAsSuccessful marks the retry as successful
func (pr *PaymentRetry) MarkAsSuccessful() {
	pr.Status = constants.PaymentRetryStatusCompleted
	now := time.Now()
	pr.SuccessfulAt = &now
	pr.CompletedAt = &now
}

// MarkAsCancelled marks the retry as cancelled
func (pr *PaymentRetry) MarkAsCancelled(reason string) {
	pr.Status = constants.PaymentRetryStatusCancelled
	now := time.Now()
	pr.CancelledAt = &now
	pr.CompletedAt = &now

	if reason != "" {
		pr.Notes = reason
	}
}

// GetGatewayConfig parses and returns gateway-specific configuration
func (pr *PaymentRetry) GetGatewayConfig() (*RetryStrategyConfig, error) {
	if pr.GatewayConfig == "" {
		return nil, nil
	}

	var config RetryStrategyConfig
	err := json.Unmarshal([]byte(pr.GatewayConfig), &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

// SetGatewayConfig sets gateway-specific configuration
func (pr *PaymentRetry) SetGatewayConfig(config *RetryStrategyConfig) error {
	if config == nil {
		pr.GatewayConfig = ""
		return nil
	}

	data, err := json.Marshal(config)
	if err != nil {
		return err
	}

	pr.GatewayConfig = string(data)
	return nil
}

