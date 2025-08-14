package dto

import (
	"time"

	sharedDto "linke/internal/shared/dto"
)

// Base ticket request structures

// CreateTicketRequest represents the request to create a ticket
type CreateTicketRequest struct {
	Title       string `json:"title" binding:"required,min=5,max=255" example:"Unable to access my account"`
	Description string `json:"description" binding:"required,min=10,max=5000" example:"I cannot log in to my account even with correct credentials"`
	Category    string `json:"category" binding:"required,oneof=general technical billing account feature bug subscription payment" example:"technical"`
	Priority    string `json:"priority" binding:"omitempty,oneof=low normal high urgent critical" example:"normal"`
	Tags        string `json:"tags,omitempty" example:"login,authentication"`
	Metadata    string `json:"metadata,omitempty" example:"{\"browser\": \"Chrome\", \"os\": \"Windows\"}"`
}

// UpdateTicketRequest represents the request to update a ticket
type UpdateTicketRequest struct {
	Title       *string `json:"title,omitempty" binding:"omitempty,min=5,max=255" example:"Updated ticket title"`
	Description *string `json:"description,omitempty" binding:"omitempty,min=10,max=5000" example:"Updated description"`
	Category    *string `json:"category,omitempty" binding:"omitempty,oneof=general technical billing account feature bug subscription payment" example:"billing"`
	Priority    *string `json:"priority,omitempty" binding:"omitempty,oneof=low normal high urgent critical" example:"high"`
	Status      *string `json:"status,omitempty" binding:"omitempty,oneof=open in_progress pending resolved closed" example:"in_progress"`
	Tags        *string `json:"tags,omitempty" example:"urgent,billing"`
	Metadata    *string `json:"metadata,omitempty" example:"{\"updated_by\": \"admin\"}"`
}

// User-specific ticket request structures

// UserCreateTicketRequest represents the request body for user ticket creation
type UserCreateTicketRequest struct {
	Title       string `json:"title" binding:"required,min=5,max=255" example:"Unable to access my subscription"`
	Description string `json:"description" binding:"required,min=10,max=5000" example:"I am unable to access my premium subscription features"`
	Category    string `json:"category" binding:"required,oneof=general technical billing account feature bug subscription payment" example:"subscription"`
	Priority    string `json:"priority" binding:"omitempty,oneof=low normal high urgent critical" example:"normal"`
	Tags        string `json:"tags,omitempty" example:"urgent,subscription"`
	Metadata    string `json:"metadata,omitempty" example:"{\"browser\": \"Chrome\", \"os\": \"Windows\"}"`
}

// CloseTicketRequest represents the request body for closing a ticket
type CloseTicketRequest struct {
	Reason string `json:"reason,omitempty" example:"Issue resolved, closing ticket"`
}

// Admin-specific ticket request structures

// AdminCreateTicketRequest represents the request body for admin ticket creation
type AdminCreateTicketRequest struct {
	Title        string `json:"title" binding:"required,min=5,max=255" example:"Unable to access account"`
	Description  string `json:"description" binding:"required,min=10,max=5000" example:"Customer reports unable to log in to their account"`
	Category     string `json:"category" binding:"required,oneof=general technical billing account feature bug subscription payment" example:"technical"`
	Priority     string `json:"priority" binding:"omitempty,oneof=low normal high urgent critical" example:"normal"`
	UserID       uint   `json:"user_id" binding:"required" example:"123"`
	AssignedToID *uint  `json:"assigned_to_id,omitempty" example:"456"`
	Tags         string `json:"tags,omitempty" example:"urgent,login"`
	Metadata     string `json:"metadata,omitempty" example:"{\"source\": \"admin_created\"}"`
}

// AdminUpdateTicketRequest represents the request body for admin ticket updates
type AdminUpdateTicketRequest struct {
	Title       *string `json:"title,omitempty" binding:"omitempty,min=5,max=255" example:"Updated ticket title"`
	Description *string `json:"description,omitempty" binding:"omitempty,min=10,max=5000" example:"Updated description"`
	Category    *string `json:"category,omitempty" binding:"omitempty,oneof=general technical billing account feature bug subscription payment" example:"billing"`
	Priority    *string `json:"priority,omitempty" binding:"omitempty,oneof=low normal high urgent critical" example:"high"`
	Status      *string `json:"status,omitempty" binding:"omitempty,oneof=open in_progress pending resolved closed" example:"in_progress"`
	Tags        *string `json:"tags,omitempty" example:"urgent,billing"`
	Metadata    *string `json:"metadata,omitempty" example:"{\"updated_by\": \"admin\"}"`
}

// Assignment and escalation structures

// AssignTicketRequest represents the request body for ticket assignment
type AssignTicketRequest struct {
	AssignedToID uint   `json:"assigned_to_id" binding:"required" example:"456"`
	Note         string `json:"note,omitempty" example:"Assigned to senior support agent"`
}

// EscalateTicketRequest represents the request body for ticket escalation
type EscalateTicketRequest struct {
	EscalatedToID    uint    `json:"escalated_to_id" binding:"required" example:"789"`
	EscalationReason string  `json:"escalation_reason" binding:"required,min=10,max=1000" example:"Customer demands supervisor escalation"`
	Priority         *string `json:"priority,omitempty" binding:"omitempty,oneof=high urgent critical" example:"urgent"`
}

// ResolveTicketRequest represents the request to resolve a ticket
type ResolveTicketRequest struct {
	Resolution string `json:"resolution" binding:"required,min=10,max=5000" example:"Issue resolved by updating user permissions"`
}

// Message-related request structures

// CreateTicketMessageRequest represents the request to create a ticket message
type CreateTicketMessageRequest struct {
	Content     string `json:"content" binding:"required,min=1,max=5000" example:"Thank you for your response. I tried the suggested solution but the issue persists."`
	MessageType string `json:"message_type" binding:"omitempty,oneof=user admin system" example:"user"`
	Attachments string `json:"attachments,omitempty" example:"[{\"name\":\"screenshot.png\",\"url\":\"https://example.com/file.png\"}]"`
	IsInternal  bool   `json:"is_internal,omitempty" example:"false"`
	Metadata    string `json:"metadata,omitempty" example:"{\"client_ip\":\"192.168.1.1\"}"`
}

// UserTicketMessageRequest represents the request body for user message creation
type UserTicketMessageRequest struct {
	Content     string `json:"content" binding:"required,min=1,max=5000" example:"Thank you for your response. I tried the suggested solution but the issue persists."`
	Attachments string `json:"attachments,omitempty" example:"[{\"name\":\"screenshot.png\",\"url\":\"https://example.com/file.png\"}]"`
	Metadata    string `json:"metadata,omitempty" example:"{\"client_ip\":\"192.168.1.1\"}"`
}

// AdminTicketMessageRequest represents the request body for admin message creation
type AdminTicketMessageRequest struct {
	Content     string `json:"content" binding:"required,min=1,max=5000" example:"Thank you for contacting support. We are reviewing your issue."`
	MessageType string `json:"message_type" binding:"omitempty,oneof=admin system" example:"admin"`
	IsInternal  bool   `json:"is_internal,omitempty" example:"false"`
	Attachments string `json:"attachments,omitempty" example:"[{\"name\":\"response.pdf\",\"url\":\"https://example.com/file.pdf\"}]"`
	Metadata    string `json:"metadata,omitempty" example:"{\"agent_id\":\"456\"}"`
}

// UpdateTicketMessageRequest represents the request to update a ticket message
type UpdateTicketMessageRequest struct {
	Content     *string `json:"content,omitempty" binding:"omitempty,min=1,max=5000" example:"Updated message content"`
	Attachments *string `json:"attachments,omitempty" example:"[{\"name\":\"updated.png\",\"url\":\"https://example.com/updated.png\"}]"`
	IsInternal  *bool   `json:"is_internal,omitempty" example:"true"`
	Metadata    *string `json:"metadata,omitempty" example:"{\"updated_by\":\"admin\"}"`
}

// Query and filter structures

// GetTicketsRequest represents the request to get tickets with filters
type GetTicketsRequest struct {
	UserID       uint   `form:"user_id" example:"1"`
	AssignedToID *uint  `form:"assigned_to_id" example:"2"`
	Status       string `form:"status" binding:"omitempty,oneof=open in_progress pending resolved closed" example:"open"`
	Priority     string `form:"priority" binding:"omitempty,oneof=low normal high urgent critical" example:"high"`
	Category     string `form:"category" binding:"omitempty,oneof=general technical billing account feature bug subscription payment" example:"technical"`
	Search       string `form:"search" example:"login issue"`
	Limit        int    `form:"limit" binding:"omitempty,min=1,max=100" example:"10"`
	Offset       int    `form:"offset" binding:"omitempty,min=0" example:"0"`
}

// GetTicketMessagesRequest represents the request to get ticket messages
type GetTicketMessagesRequest struct {
	TicketID        uint   `form:"ticket_id" binding:"required" example:"1"`
	MessageType     string `form:"message_type" binding:"omitempty,oneof=user admin system" example:"user"`
	IncludeInternal bool   `form:"include_internal" example:"false"`
	Limit           int    `form:"limit" binding:"omitempty,min=1,max=100" example:"10"`
	Offset          int    `form:"offset" binding:"omitempty,min=0" example:"0"`
}

// Advanced search and analytics structures

// TicketSearchRequest represents the request for advanced ticket search
type TicketSearchRequest struct {
	Query         string     `form:"query" example:"login issue"`
	UserID        uint       `form:"user_id" example:"123"`
	AssignedToID  *uint      `form:"assigned_to_id" example:"456"`
	Status        string     `form:"status" binding:"omitempty,oneof=open in_progress pending resolved closed" example:"open"`
	Priority      string     `form:"priority" binding:"omitempty,oneof=low normal high urgent critical" example:"high"`
	Category      string     `form:"category" binding:"omitempty,oneof=general technical billing account feature bug subscription payment" example:"technical"`
	CreatedAfter  *time.Time `form:"created_after" time_format:"2006-01-02" example:"2024-01-01"`
	CreatedBefore *time.Time `form:"created_before" time_format:"2006-01-02" example:"2024-12-31"`
	Tags          string     `form:"tags" example:"urgent,billing"`
	Limit         int        `form:"limit" binding:"omitempty,min=1,max=100" example:"20"`
	Offset        int        `form:"offset" binding:"omitempty,min=0" example:"0"`
}

// TicketAnalyticsRequest represents the request for ticket analytics
type TicketAnalyticsRequest struct {
	StartDate string `form:"start_date" time_format:"2006-01-02" example:"2024-01-01"`
	EndDate   string `form:"end_date" time_format:"2006-01-02" example:"2024-12-31"`
	GroupBy   string `form:"group_by" binding:"omitempty,oneof=day week month agent category priority" example:"day"`
	AgentID   *uint  `form:"agent_id" example:"456"`
	Category  string `form:"category" binding:"omitempty,oneof=general technical billing account feature bug subscription payment" example:"technical"`
	Priority  string `form:"priority" binding:"omitempty,oneof=low normal high urgent critical" example:"high"`
}

// Bulk operations structures

// BulkTicketActionRequest represents the request body for bulk ticket operations
type BulkTicketActionRequest struct {
	TicketIDs    []uint  `json:"ticket_ids" binding:"required,min=1,max=100"`
	Action       string  `json:"action" binding:"required,oneof=assign close reopen update_priority update_status" example:"assign"`
	AssignedToID *uint   `json:"assigned_to_id,omitempty" example:"456"`
	Status       *string `json:"status,omitempty" binding:"omitempty,oneof=open in_progress pending resolved closed" example:"closed"`
	Priority     *string `json:"priority,omitempty" binding:"omitempty,oneof=low normal high urgent critical" example:"high"`
	Reason       string  `json:"reason,omitempty" example:"Bulk closing resolved tickets"`
}

// AgentInfo represents information about an agent for ticket assignment
type AgentInfo struct {
	UserID            uint     `json:"user_id"`
	Name              string   `json:"name"`
	Email             string   `json:"email"`
	Specialties       []string `json:"specialties"`  // Categories this agent specializes in
	CurrentLoad       int      `json:"current_load"` // Number of currently assigned tickets
	MaxLoad           int      `json:"max_load"`     // Maximum tickets this agent can handle
	IsOnline          bool     `json:"is_online"`
	LastActiveAt      string   `json:"last_active_at,omitempty"`
	AvgResponseTime   int      `json:"avg_response_time"`  // In minutes
	SatisfactionScore float64  `json:"satisfaction_score"` // 0-10 scale
}

// TicketResponse represents the API response for a ticket (admin view)
type TicketResponse struct {
	ID              uint                    `json:"id" example:"1"`
	TicketNo        string                  `json:"ticket_no" example:"TKT-20240101-001"`
	Title           string                  `json:"title" example:"Unable to access my subscription"`
	Description     string                  `json:"description" example:"I am unable to access my premium subscription features"`
	Category        string                  `json:"category" example:"subscription"`
	Priority        string                  `json:"priority" example:"normal"`
	Status          string                  `json:"status" example:"open"`
	UserID          uint                    `json:"user_id" example:"1"`
	User            *sharedDto.UserBasicDTO `json:"user,omitempty"`
	AssignedToID    *uint                   `json:"assigned_to_id" example:"2"`
	AssignedTo      *sharedDto.UserBasicDTO `json:"assigned_to,omitempty"`
	AssignedAt      *time.Time              `json:"assigned_at" example:"2024-01-01T10:00:00Z"`
	ResolvedByID    *uint                   `json:"resolved_by_id" example:"2"`
	ResolvedBy      *sharedDto.UserBasicDTO `json:"resolved_by,omitempty"`
	ResolvedAt      *time.Time              `json:"resolved_at" example:"2024-01-02T15:30:00Z"`
	Resolution      string                  `json:"resolution" example:"Issue resolved by updating subscription settings"`
	FirstResponseAt *time.Time              `json:"first_response_at" example:"2024-01-01T10:30:00Z"`
	LastResponseAt  *time.Time              `json:"last_response_at" example:"2024-01-02T14:00:00Z"`
	ClosedAt        *time.Time              `json:"closed_at" example:"2024-01-02T16:00:00Z"`
	Tags            *string                 `json:"tags" example:"urgent,subscription"`
	Metadata        *string                 `json:"metadata" example:"{\"priority_escalated\": true}"`
	Messages        []TicketMessageResponse `json:"messages,omitempty"`
	CreatedAt       time.Time               `json:"created_at" example:"2024-01-01T09:00:00Z"`
	UpdatedAt       time.Time               `json:"updated_at" example:"2024-01-02T16:00:00Z"`
}

// TicketMessageResponse represents the API response for a ticket message (admin view)
type TicketMessageResponse struct {
	ID          uint                    `json:"id" example:"1"`
	TicketID    uint                    `json:"ticket_id" example:"1"`
	UserID      uint                    `json:"user_id" example:"2"`
	User        *sharedDto.UserBasicDTO `json:"user,omitempty"`
	Content     string                  `json:"content" example:"Thank you for contacting support. We will review your issue."`
	MessageType string                  `json:"message_type" example:"admin"`
	Attachments *string                 `json:"attachments" example:"[{\"name\":\"screenshot.png\",\"url\":\"/uploads/screenshot.png\"}]"`
	IsInternal  bool                    `json:"is_internal" example:"false"`
	Metadata    *string                 `json:"metadata" example:"{\"priority\": \"normal\"}"`
	CreatedAt   time.Time               `json:"created_at" example:"2024-01-01T10:30:00Z"`
	UpdatedAt   time.Time               `json:"updated_at" example:"2024-01-01T10:30:00Z"`
}

// TicketUserResponse represents the ticket response for regular users (limited information)
type TicketUserResponse struct {
	ID              uint                        `json:"id" example:"1"`
	TicketNo        string                      `json:"ticket_no" example:"TKT-20240101-001"`
	Title           string                      `json:"title" example:"Unable to access my subscription"`
	Description     string                      `json:"description" example:"I am unable to access my premium subscription features"`
	Category        string                      `json:"category" example:"subscription"`
	Priority        string                      `json:"priority" example:"normal"`
	Status          string                      `json:"status" example:"open"`
	Resolution      string                      `json:"resolution" example:"Issue resolved by updating subscription settings"`
	FirstResponseAt *time.Time                  `json:"first_response_at" example:"2024-01-01T10:30:00Z"`
	LastResponseAt  *time.Time                  `json:"last_response_at" example:"2024-01-02T14:00:00Z"`
	ClosedAt        *time.Time                  `json:"closed_at" example:"2024-01-02T16:00:00Z"`
	Messages        []TicketMessageUserResponse `json:"messages,omitempty"`
	CreatedAt       time.Time                   `json:"created_at" example:"2024-01-01T09:00:00Z"`
	UpdatedAt       time.Time                   `json:"updated_at" example:"2024-01-02T16:00:00Z"`
}

// TicketMessageUserResponse represents the ticket message response for regular users
type TicketMessageUserResponse struct {
	ID          uint      `json:"id" example:"1"`
	TicketID    uint      `json:"ticket_id" example:"1"`
	Content     string    `json:"content" example:"Thank you for contacting support. We will review your issue."`
	MessageType string    `json:"message_type" example:"admin"`
	Attachments *string   `json:"attachments" example:"[{\"name\":\"screenshot.png\",\"url\":\"/uploads/screenshot.png\"}]"`
	CreatedAt   time.Time `json:"created_at" example:"2024-01-01T10:30:00Z"`
	UpdatedAt   time.Time `json:"updated_at" example:"2024-01-01T10:30:00Z"`
}
