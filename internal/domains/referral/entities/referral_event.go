package entities

import (
	"time"

	"gorm.io/gorm"

	"linke/internal/shared/dto"
)

// ReferralEvent represents events that occur during the referral lifecycle
type ReferralEvent struct {
	// Primary Key
	ID uint `json:"id" gorm:"primaryKey"`

	// Foreign Keys
	ReferralID uint `json:"referral_id" gorm:"not null;index"`
	UserID     uint `json:"user_id" gorm:"not null;index"` // User who triggered the event

	// Event Data
	EventType        string `json:"event_type" gorm:"size:50;not null;index"` // Type of event
	EventDescription string `json:"event_description" gorm:"size:255"`        // Human-readable description
	EventData        string `json:"event_data,omitempty" gorm:"type:text"`    // JSON data for the event

	// Attribution
	IPAddress   string `json:"ip_address" gorm:"size:45"`    // IP address
	UserAgent   string `json:"user_agent" gorm:"size:500"`   // User agent
	ReferrerURL string `json:"referrer_url" gorm:"size:500"` // Referrer URL
	PageURL     string `json:"page_url" gorm:"size:500"`     // Page URL where event occurred

	// UTM Parameters
	UTMSource   string `json:"utm_source" gorm:"size:100"`   // UTM source
	UTMCampaign string `json:"utm_campaign" gorm:"size:100"` // UTM campaign
	UTMMedium   string `json:"utm_medium" gorm:"size:100"`   // UTM medium
	UTMTerm     string `json:"utm_term" gorm:"size:100"`     // UTM term
	UTMContent  string `json:"utm_content" gorm:"size:100"`  // UTM content

	// Event Value
	EventValue    float64 `json:"event_value" gorm:"type:decimal(10,2);default:0"` // Monetary value associated with event
	EventCurrency string  `json:"event_currency" gorm:"size:10;default:'USD'"`     // Currency of the event value

	// Metadata
	Metadata    string     `json:"metadata,omitempty" gorm:"type:text"` // Additional event metadata (JSON)
	ProcessedAt *time.Time `json:"processed_at,omitempty" gorm:"index"` // When event was processed

	// Note: Relationships removed to avoid cross-domain dependencies
	// Related data should be fetched and assembled at the application layer

	// Timestamp Fields
	CreatedAt time.Time      `json:"created_at" gorm:"not null;index"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"not null"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

// TableName returns the table name for ReferralEvent model
func (ReferralEvent) TableName() string {
	return "referral_events"
}

// Event Type constants
const (
	EventTypeClick         = "click"
	EventTypeView          = "view"
	EventTypeRegistration  = "registration"
	EventTypeActivation    = "activation"
	EventTypeFirstPurchase = "first_purchase"
	EventTypeSubscription  = "subscription"
	EventTypeRenewal       = "renewal"
	EventTypeCancellation  = "cancellation"
	EventTypeRefund        = "refund"
	EventTypeReward        = "reward"
	EventTypeConversion    = "conversion"
	EventTypeExpired       = "expired"
	EventTypeBlocked       = "blocked"
)

// IsConversionEvent checks if this event represents a conversion
func (re *ReferralEvent) IsConversionEvent() bool {
	conversionEvents := map[string]bool{
		EventTypeRegistration:  true,
		EventTypeFirstPurchase: true,
		EventTypeSubscription:  true,
		EventTypeActivation:    true,
		EventTypeConversion:    true,
	}
	return conversionEvents[re.EventType]
}

// IsRevenueEvent checks if this event represents revenue generation
func (re *ReferralEvent) IsRevenueEvent() bool {
	revenueEvents := map[string]bool{
		EventTypeFirstPurchase: true,
		EventTypeSubscription:  true,
		EventTypeRenewal:       true,
	}
	return revenueEvents[re.EventType]
}

// ReferralEventResponse represents the referral event data structure for API responses
type ReferralEventResponse struct {
	ID               uint       `json:"id" example:"1"`
	ReferralID       uint       `json:"referral_id" example:"1"`
	UserID           uint       `json:"user_id" example:"2"`
	EventType        string     `json:"event_type" example:"registration"`
	EventDescription string     `json:"event_description" example:"User completed registration"`
	EventData        string     `json:"event_data,omitempty" example:"{\"signup_method\":\"email\"}"`
	IPAddress        string     `json:"ip_address" example:"192.168.1.100"`
	UserAgent        string     `json:"user_agent" example:"Mozilla/5.0..."`
	ReferrerURL      string     `json:"referrer_url" example:"https://example.com/ref"`
	PageURL          string     `json:"page_url" example:"https://example.com/signup"`
	UTMSource        string     `json:"utm_source" example:"facebook"`
	UTMCampaign      string     `json:"utm_campaign" example:"summer_referral"`
	UTMMedium        string     `json:"utm_medium" example:"social"`
	UTMTerm          string     `json:"utm_term" example:"referral"`
	UTMContent       string     `json:"utm_content" example:"banner"`
	EventValue       float64    `json:"event_value" example:"29.99"`
	EventCurrency    string     `json:"event_currency" example:"USD"`
	ProcessedAt      *time.Time `json:"processed_at,omitempty" example:"2024-01-01T00:00:00Z"`
	CreatedAt        time.Time  `json:"created_at" example:"2024-01-01T00:00:00Z"`
	UpdatedAt        time.Time  `json:"updated_at" example:"2024-01-01T00:00:00Z"`

	// Optional related data
	Referral *ReferralResponse `json:"referral,omitempty"`
	User     *dto.UserBasicDTO `json:"user,omitempty"`
}

// ToResponse converts ReferralEvent to ReferralEventResponse
func (re *ReferralEvent) ToResponse() *ReferralEventResponse {
	resp := &ReferralEventResponse{
		ID:               re.ID,
		ReferralID:       re.ReferralID,
		UserID:           re.UserID,
		EventType:        re.EventType,
		EventDescription: re.EventDescription,
		EventData:        re.EventData,
		IPAddress:        re.IPAddress,
		UserAgent:        re.UserAgent,
		ReferrerURL:      re.ReferrerURL,
		PageURL:          re.PageURL,
		UTMSource:        re.UTMSource,
		UTMCampaign:      re.UTMCampaign,
		UTMMedium:        re.UTMMedium,
		UTMTerm:          re.UTMTerm,
		UTMContent:       re.UTMContent,
		EventValue:       re.EventValue,
		EventCurrency:    re.EventCurrency,
		ProcessedAt:      re.ProcessedAt,
		CreatedAt:        re.CreatedAt,
		UpdatedAt:        re.UpdatedAt,
	}

	// Note: Related data should be populated at the application layer
	// to avoid cross-domain dependencies

	return resp
}
