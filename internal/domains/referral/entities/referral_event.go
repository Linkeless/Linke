package entities

import (
	"time"

	"gorm.io/gorm"

	"linke/internal/domains/referral/constants"
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
	EventCurrency string  `json:"event_currency" gorm:"size:10;default:'CNY'"`     // Currency of the event value

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

// IsConversionEvent checks if this event represents a conversion
func (re *ReferralEvent) IsConversionEvent() bool {
	conversionEvents := map[string]bool{
		constants.EventTypeRegistration:  true,
		constants.EventTypeFirstPurchase: true,
		constants.EventTypeSubscription:  true,
		constants.EventTypeActivation:    true,
		constants.EventTypeConversion:    true,
	}
	return conversionEvents[re.EventType]
}

// IsRevenueEvent checks if this event represents revenue generation
func (re *ReferralEvent) IsRevenueEvent() bool {
	revenueEvents := map[string]bool{
		constants.EventTypeFirstPurchase: true,
		constants.EventTypeSubscription:  true,
		constants.EventTypeRenewal:       true,
	}
	return revenueEvents[re.EventType]
}

// ToResponse should be implemented in service layer to avoid import cycles
// Use dto.ToReferralEventResponse(re) instead
