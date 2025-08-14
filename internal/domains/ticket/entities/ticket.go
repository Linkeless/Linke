package entities

import (
	"time"

	"gorm.io/gorm"

	"linke/internal/domains/ticket/constants"
)

// Ticket represents a support ticket
// Fields are ordered for optimal memory alignment
type Ticket struct {
	// 8-byte aligned fields first
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"not null;index" json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Pointer fields (8 bytes on 64-bit systems)
	AssignedToID    *uint          `gorm:"index" json:"assigned_to_id"`
	ResolvedByID    *uint          `gorm:"index" json:"resolved_by_id"`
	AssignedAt      *time.Time     `json:"assigned_at"`
	ResolvedAt      *time.Time     `json:"resolved_at"`
	FirstResponseAt *time.Time     `json:"first_response_at"`
	LastResponseAt  *time.Time     `json:"last_response_at"`
	ClosedAt        *time.Time     `json:"closed_at"`
	Tags            *string        `gorm:"type:text" json:"tags"`
	Metadata        *string        `gorm:"type:json" json:"metadata"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"deleted_at"`

	// String fields (variable length)
	TicketNo    string `gorm:"uniqueIndex;size:32;not null" json:"ticket_no"`
	Title       string `gorm:"size:255;not null" json:"title"`
	Description string `gorm:"type:text;not null" json:"description"`
	Category    string `gorm:"size:50;not null;default:'general'" json:"category"`
	Priority    string `gorm:"size:20;not null;default:'normal'" json:"priority"`
	Status      string `gorm:"size:20;not null;default:'open'" json:"status"`
	Resolution  string `gorm:"type:text" json:"resolution"`

	// Relationships (loaded separately to avoid N+1 queries)
	Messages []TicketMessage `gorm:"foreignKey:TicketID" json:"messages,omitempty"`
}

// TicketMessage represents a message in a ticket conversation
// Fields are ordered for optimal memory alignment
type TicketMessage struct {
	// 8-byte aligned fields first
	ID        uint      `gorm:"primaryKey" json:"id"`
	TicketID  uint      `gorm:"not null;index" json:"ticket_id"`
	UserID    uint      `gorm:"not null;index" json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Pointer fields (8 bytes on 64-bit systems)
	Ticket      *Ticket        `gorm:"foreignKey:TicketID" json:"ticket,omitempty"`
	Attachments *string        `gorm:"type:json" json:"attachments"`
	Metadata    *string        `gorm:"type:json" json:"metadata"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at"`

	// String fields (variable length)
	Content     string `gorm:"type:text;not null" json:"content"`
	MessageType string `gorm:"size:20;not null;default:'user'" json:"message_type"` // user, admin, system

	// Boolean fields (1 byte, placed last for alignment)
	IsInternal bool `gorm:"default:false" json:"is_internal"` // Internal notes visible only to admins
}

// TableName returns the table name for Ticket
func (Ticket) TableName() string {
	return constants.TableTickets
}

// TableName returns the table name for TicketMessage
func (TicketMessage) TableName() string {
	return constants.TableTicketMessages
}
