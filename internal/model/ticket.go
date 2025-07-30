package model

import (
	"time"

	"gorm.io/gorm"
)

// Ticket status constants
const (
	TicketStatusOpen       = "open"
	TicketStatusInProgress = "in_progress"
	TicketStatusPending    = "pending"
	TicketStatusResolved   = "resolved"
	TicketStatusClosed     = "closed"
)

// Ticket priority constants
const (
	TicketPriorityLow      = "low"
	TicketPriorityNormal   = "normal"
	TicketPriorityHigh     = "high"
	TicketPriorityUrgent   = "urgent"
	TicketPriorityCritical = "critical"
)

// Ticket category constants
const (
	TicketCategoryGeneral      = "general"
	TicketCategoryTechnical    = "technical"
	TicketCategoryBilling      = "billing"
	TicketCategoryAccount      = "account"
	TicketCategoryFeature      = "feature"
	TicketCategoryBug          = "bug"
	TicketCategorySubscription = "subscription"
	TicketCategoryPayment      = "payment"
)

// Ticket represents a support ticket
type Ticket struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	TicketNo    string    `gorm:"uniqueIndex;size:32;not null" json:"ticket_no"`
	Title       string    `gorm:"size:255;not null" json:"title"`
	Description string    `gorm:"type:text;not null" json:"description"`
	Category    string    `gorm:"size:50;not null;default:'general'" json:"category"`
	Priority    string    `gorm:"size:20;not null;default:'normal'" json:"priority"`
	Status      string    `gorm:"size:20;not null;default:'open'" json:"status"`
	
	// User information
	UserID    uint   `gorm:"not null;index" json:"user_id"`
	User      *User  `gorm:"foreignKey:UserID" json:"user,omitempty"`
	
	// Assignment information
	AssignedToID *uint `gorm:"index" json:"assigned_to_id"`
	AssignedTo   *User `gorm:"foreignKey:AssignedToID" json:"assigned_to,omitempty"`
	AssignedAt   *time.Time `json:"assigned_at"`
	
	// Resolution information
	ResolvedByID *uint `gorm:"index" json:"resolved_by_id"`
	ResolvedBy   *User `gorm:"foreignKey:ResolvedByID" json:"resolved_by,omitempty"`
	ResolvedAt   *time.Time `json:"resolved_at"`
	Resolution   string     `gorm:"type:text" json:"resolution"`
	
	// Timing information
	FirstResponseAt *time.Time `json:"first_response_at"`
	LastResponseAt  *time.Time `json:"last_response_at"`
	ClosedAt        *time.Time `json:"closed_at"`
	
	// Metadata
	Tags     *string    `gorm:"type:text" json:"tags"`
	Metadata *string    `gorm:"type:json" json:"metadata"`
	
	// Relationships
	Messages []TicketMessage `gorm:"foreignKey:TicketID" json:"messages,omitempty"`
	
	// Timestamps
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// TicketMessage represents a message in a ticket conversation
type TicketMessage struct {
	ID       uint   `gorm:"primaryKey" json:"id"`
	TicketID uint   `gorm:"not null;index" json:"ticket_id"`
	Ticket   *Ticket `gorm:"foreignKey:TicketID" json:"ticket,omitempty"`
	
	// User information
	UserID uint  `gorm:"not null;index" json:"user_id"`
	User   *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
	
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

// TicketResponse represents the API response for a ticket
type TicketResponse struct {
	ID          uint                  `json:"id" example:"1"`
	TicketNo    string                `json:"ticket_no" example:"TKT-20240101-001"`
	Title       string                `json:"title" example:"Unable to access my subscription"`
	Description string                `json:"description" example:"I am unable to access my premium subscription features"`
	Category    string                `json:"category" example:"subscription"`
	Priority    string                `json:"priority" example:"normal"`
	Status      string                `json:"status" example:"open"`
	UserID      uint                  `json:"user_id" example:"1"`
	User        *UserResponse         `json:"user,omitempty"`
	AssignedToID *uint                `json:"assigned_to_id" example:"2"`
	AssignedTo   *UserResponse        `json:"assigned_to,omitempty"`
	AssignedAt   *time.Time           `json:"assigned_at" example:"2024-01-01T10:00:00Z"`
	ResolvedByID *uint                `json:"resolved_by_id" example:"2"`
	ResolvedBy   *UserResponse        `json:"resolved_by,omitempty"`
	ResolvedAt   *time.Time           `json:"resolved_at" example:"2024-01-02T15:30:00Z"`
	Resolution   string               `json:"resolution" example:"Issue resolved by updating subscription settings"`
	FirstResponseAt *time.Time        `json:"first_response_at" example:"2024-01-01T10:30:00Z"`
	LastResponseAt  *time.Time        `json:"last_response_at" example:"2024-01-02T14:00:00Z"`
	ClosedAt        *time.Time        `json:"closed_at" example:"2024-01-02T16:00:00Z"`
	Tags         *string               `json:"tags" example:"urgent,subscription"`
	Metadata     *string               `json:"metadata" example:"{\"priority_escalated\": true}"`
	Messages     []TicketMessageResponse `json:"messages,omitempty"`
	CreatedAt    time.Time            `json:"created_at" example:"2024-01-01T09:00:00Z"`
	UpdatedAt    time.Time            `json:"updated_at" example:"2024-01-02T16:00:00Z"`
}

// TicketMessageResponse represents the API response for a ticket message
type TicketMessageResponse struct {
	ID          uint          `json:"id" example:"1"`
	TicketID    uint          `json:"ticket_id" example:"1"`
	UserID      uint          `json:"user_id" example:"2"`
	User        *UserResponse `json:"user,omitempty"`
	Content     string        `json:"content" example:"Thank you for contacting support. We will review your issue."`
	MessageType string        `json:"message_type" example:"admin"`
	Attachments *string        `json:"attachments" example:"[{\"name\":\"screenshot.png\",\"url\":\"/uploads/screenshot.png\"}]"`
	IsInternal  bool          `json:"is_internal" example:"false"`
	Metadata    *string        `json:"metadata" example:"{\"priority\": \"normal\"}"`
	CreatedAt   time.Time     `json:"created_at" example:"2024-01-01T10:30:00Z"`
	UpdatedAt   time.Time     `json:"updated_at" example:"2024-01-01T10:30:00Z"`
}

// TicketUserResponse represents the ticket response for regular users (limited information)
type TicketUserResponse struct {
	ID          uint                      `json:"id" example:"1"`
	TicketNo    string                    `json:"ticket_no" example:"TKT-20240101-001"`
	Title       string                    `json:"title" example:"Unable to access my subscription"`
	Description string                    `json:"description" example:"I am unable to access my premium subscription features"`
	Category    string                    `json:"category" example:"subscription"`
	Priority    string                    `json:"priority" example:"normal"`
	Status      string                    `json:"status" example:"open"`
	Resolution  string                    `json:"resolution" example:"Issue resolved by updating subscription settings"`
	FirstResponseAt *time.Time            `json:"first_response_at" example:"2024-01-01T10:30:00Z"`
	LastResponseAt  *time.Time            `json:"last_response_at" example:"2024-01-02T14:00:00Z"`
	ClosedAt        *time.Time            `json:"closed_at" example:"2024-01-02T16:00:00Z"`
	Messages        []TicketMessageUserResponse `json:"messages,omitempty"`
	CreatedAt       time.Time             `json:"created_at" example:"2024-01-01T09:00:00Z"`
	UpdatedAt       time.Time             `json:"updated_at" example:"2024-01-02T16:00:00Z"`
}

// TicketMessageUserResponse represents the ticket message response for regular users
type TicketMessageUserResponse struct {
	ID          uint      `json:"id" example:"1"`
	TicketID    uint      `json:"ticket_id" example:"1"`
	Content     string    `json:"content" example:"Thank you for contacting support. We will review your issue."`
	MessageType string    `json:"message_type" example:"admin"`
	Attachments *string    `json:"attachments" example:"[{\"name\":\"screenshot.png\",\"url\":\"/uploads/screenshot.png\"}]"`
	CreatedAt   time.Time `json:"created_at" example:"2024-01-01T10:30:00Z"`
	UpdatedAt   time.Time `json:"updated_at" example:"2024-01-01T10:30:00Z"`
}

// ToResponse converts a Ticket to TicketResponse (admin view)
func (t *Ticket) ToResponse() *TicketResponse {
	response := &TicketResponse{
		ID:              t.ID,
		TicketNo:        t.TicketNo,
		Title:           t.Title,
		Description:     t.Description,
		Category:        t.Category,
		Priority:        t.Priority,
		Status:          t.Status,
		UserID:          t.UserID,
		AssignedToID:    t.AssignedToID,
		AssignedAt:      t.AssignedAt,
		ResolvedByID:    t.ResolvedByID,
		ResolvedAt:      t.ResolvedAt,
		Resolution:      t.Resolution,
		FirstResponseAt: t.FirstResponseAt,
		LastResponseAt:  t.LastResponseAt,
		ClosedAt:        t.ClosedAt,
		Tags:            t.Tags,
		Metadata:        t.Metadata,
		CreatedAt:       t.CreatedAt,
		UpdatedAt:       t.UpdatedAt,
	}

	if t.User != nil {
		response.User = t.User.ToResponse()
	}

	if t.AssignedTo != nil {
		response.AssignedTo = t.AssignedTo.ToResponse()
	}

	if t.ResolvedBy != nil {
		response.ResolvedBy = t.ResolvedBy.ToResponse()
	}

	if t.Messages != nil {
		response.Messages = make([]TicketMessageResponse, len(t.Messages))
		for i, msg := range t.Messages {
			response.Messages[i] = *msg.ToResponse()
		}
	}

	return response
}

// ToUserResponse converts a Ticket to TicketUserResponse (user view)
func (t *Ticket) ToUserResponse() *TicketUserResponse {
	response := &TicketUserResponse{
		ID:              t.ID,
		TicketNo:        t.TicketNo,
		Title:           t.Title,
		Description:     t.Description,
		Category:        t.Category,
		Priority:        t.Priority,
		Status:          t.Status,
		Resolution:      t.Resolution,
		FirstResponseAt: t.FirstResponseAt,
		LastResponseAt:  t.LastResponseAt,
		ClosedAt:        t.ClosedAt,
		CreatedAt:       t.CreatedAt,
		UpdatedAt:       t.UpdatedAt,
	}

	if t.Messages != nil {
		response.Messages = make([]TicketMessageUserResponse, 0)
		for _, msg := range t.Messages {
			// Filter out internal messages from user view
			if !msg.IsInternal {
				response.Messages = append(response.Messages, *msg.ToUserResponse())
			}
		}
	}

	return response
}

// ToResponse converts a TicketMessage to TicketMessageResponse (admin view)
func (tm *TicketMessage) ToResponse() *TicketMessageResponse {
	response := &TicketMessageResponse{
		ID:          tm.ID,
		TicketID:    tm.TicketID,
		UserID:      tm.UserID,
		Content:     tm.Content,
		MessageType: tm.MessageType,
		Attachments: tm.Attachments,
		IsInternal:  tm.IsInternal,
		Metadata:    tm.Metadata,
		CreatedAt:   tm.CreatedAt,
		UpdatedAt:   tm.UpdatedAt,
	}

	if tm.User != nil {
		response.User = tm.User.ToResponse()
	}

	return response
}

// ToUserResponse converts a TicketMessage to TicketMessageUserResponse (user view)
func (tm *TicketMessage) ToUserResponse() *TicketMessageUserResponse {
	return &TicketMessageUserResponse{
		ID:          tm.ID,
		TicketID:    tm.TicketID,
		Content:     tm.Content,
		MessageType: tm.MessageType,
		Attachments: tm.Attachments,
		CreatedAt:   tm.CreatedAt,
		UpdatedAt:   tm.UpdatedAt,
	}
}

// TableName returns the table name for Ticket
func (Ticket) TableName() string {
	return "tickets"
}

// TableName returns the table name for TicketMessage
func (TicketMessage) TableName() string {
	return "ticket_messages"
}