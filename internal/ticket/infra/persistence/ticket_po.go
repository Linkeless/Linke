package persistence

import (
	"time"

	"gorm.io/gorm"
)

// TicketPO represents the ticket persistence object for GORM
type TicketPO struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	TicketNumber string         `gorm:"uniqueIndex;size:32;not null" json:"ticket_number"`
	Title        string         `gorm:"size:255;not null" json:"title"`
	Description  string         `gorm:"type:text;not null" json:"description"`
	Category     string         `gorm:"size:50;not null;default:'general'" json:"category"`
	Priority     string         `gorm:"size:20;not null;default:'normal'" json:"priority"`
	Status       string         `gorm:"size:20;not null;default:'open'" json:"status"`
	
	// User information
	UserID uint `gorm:"not null;index" json:"user_id"`
	
	// Assignment information
	AssignedToID *uint      `gorm:"index" json:"assigned_to_id"`
	AssignedAt   *time.Time `json:"assigned_at"`
	
	// Resolution information
	ResolvedByID *uint      `gorm:"index" json:"resolved_by_id"`
	ResolvedAt   *time.Time `json:"resolved_at"`
	Resolution   string     `gorm:"type:text" json:"resolution"`
	
	// Timing information
	FirstResponseAt *time.Time `json:"first_response_at"`
	LastResponseAt  *time.Time `json:"last_response_at"`
	ClosedAt        *time.Time `json:"closed_at"`
	
	// Metadata
	Tags     *string `gorm:"type:text" json:"tags"`
	Metadata *string `gorm:"type:json" json:"metadata"`
	
	// Version for optimistic locking
	Version int `gorm:"not null;default:1" json:"version"`
	
	// Timestamps
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	
	// Relationships
	Messages []TicketMessagePO `gorm:"foreignKey:TicketID" json:"messages,omitempty"`
}

// TicketMessagePO represents the ticket message persistence object for GORM
type TicketMessagePO struct {
	ID       uint `gorm:"primaryKey" json:"id"`
	TicketID uint `gorm:"not null;index" json:"ticket_id"`
	
	// User information
	UserID uint `gorm:"not null;index" json:"user_id"`
	
	// Message content
	Content     string `gorm:"type:text;not null" json:"content"`
	MessageType string `gorm:"size:20;not null;default:'user'" json:"message_type"` // user, admin, system
	
	// Attachments
	Attachments *string `gorm:"type:json" json:"attachments"`
	
	// Metadata
	IsInternal bool    `gorm:"default:false" json:"is_internal"` // Internal notes visible only to admins
	Metadata   *string `gorm:"type:json" json:"metadata"`
	
	// Timestamps
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// TableName returns the table name for TicketPO
func (TicketPO) TableName() string {
	return "tickets"
}

// TableName returns the table name for TicketMessagePO
func (TicketMessagePO) TableName() string {
	return "ticket_messages"
}