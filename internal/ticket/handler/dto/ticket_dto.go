package dto

import (
	"time"

	"linke/internal/ticket/domain/model"
)

// CreateTicketRequest represents the request to create a ticket
type CreateTicketRequest struct {
	Title       string                 `json:"title" binding:"required,min=5,max=255" example:"Unable to access my account"`
	Description string                 `json:"description" binding:"required,min=10,max=5000" example:"I cannot log in to my account even with correct credentials"`
	Category    string                 `json:"category" binding:"required,oneof=general technical billing account feature bug subscription payment" example:"technical"`
	Priority    string                 `json:"priority" binding:"omitempty,oneof=low normal high urgent critical" example:"normal"`
	Tags        []string               `json:"tags,omitempty" example:"login,authentication"`
	Metadata    map[string]interface{} `json:"metadata,omitempty" example:"{\"browser\": \"Chrome\", \"os\": \"Windows\"}"`
}

// UpdateTicketRequest represents the request to update a ticket
type UpdateTicketRequest struct {
	Title       *string  `json:"title,omitempty" binding:"omitempty,min=5,max=255" example:"Updated ticket title"`
	Description *string  `json:"description,omitempty" binding:"omitempty,min=10,max=5000" example:"Updated description"`
	Category    *string  `json:"category,omitempty" binding:"omitempty,oneof=general technical billing account feature bug subscription payment" example:"billing"`
	Priority    *string  `json:"priority,omitempty" binding:"omitempty,oneof=low normal high urgent critical" example:"high"`
	Tags        []string `json:"tags,omitempty" example:"urgent,billing"`
}

// AssignTicketRequest represents the request to assign a ticket
type AssignTicketRequest struct {
	AssignedToID uint `json:"assigned_to_id" binding:"required" example:"2"`
}

// ChangeTicketStatusRequest represents the request to change ticket status
type ChangeTicketStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=open in_progress pending resolved closed" example:"in_progress"`
}

// ResolveTicketRequest represents the request to resolve a ticket
type ResolveTicketRequest struct {
	Resolution string `json:"resolution" binding:"required,min=10,max=5000" example:"Issue resolved by updating user permissions"`
}

// AddTicketMessageRequest represents the request to add a message to a ticket
type AddTicketMessageRequest struct {
	Content     string        `json:"content" binding:"required,min=1,max=5000" example:"Thank you for your response. I tried the suggested solution but the issue persists."`
	MessageType string        `json:"message_type" binding:"omitempty,oneof=user admin system" example:"user"`
	Attachments []Attachment  `json:"attachments,omitempty"`
	IsInternal  bool          `json:"is_internal,omitempty" example:"false"`
	Metadata    map[string]interface{} `json:"metadata,omitempty" example:"{\"client_ip\":\"192.168.1.1\"}"`
}

// Attachment represents a file attachment
type Attachment struct {
	Name string `json:"name" example:"screenshot.png"`
	URL  string `json:"url" example:"https://example.com/file.png"`
	Size int64  `json:"size,omitempty" example:"1024"`
	Type string `json:"type,omitempty" example:"image/png"`
}

// TicketResponse represents the API response for a ticket
type TicketResponse struct {
	ID           uint                `json:"id" example:"1"`
	TicketNumber string              `json:"ticket_number" example:"TK-20240101-ABC123"`
	Title        string              `json:"title" example:"Unable to access my subscription"`
	Description  string              `json:"description" example:"I am unable to access my premium subscription features"`
	Category     string              `json:"category" example:"subscription"`
	Priority     string              `json:"priority" example:"normal"`
	Status       string              `json:"status" example:"open"`
	UserID       uint                `json:"user_id" example:"1"`
	AssignedToID *uint               `json:"assigned_to_id,omitempty" example:"2"`
	AssignedAt   *time.Time          `json:"assigned_at,omitempty" example:"2024-01-01T10:00:00Z"`
	ResolvedByID *uint               `json:"resolved_by_id,omitempty" example:"2"`
	ResolvedAt   *time.Time          `json:"resolved_at,omitempty" example:"2024-01-02T15:30:00Z"`
	Resolution   string              `json:"resolution,omitempty" example:"Issue resolved by updating subscription settings"`
	FirstResponseAt *time.Time       `json:"first_response_at,omitempty" example:"2024-01-01T10:30:00Z"`
	LastResponseAt  *time.Time       `json:"last_response_at,omitempty" example:"2024-01-02T14:00:00Z"`
	ClosedAt        *time.Time       `json:"closed_at,omitempty" example:"2024-01-02T16:00:00Z"`
	Tags            []string         `json:"tags,omitempty" example:"urgent,subscription"`
	Metadata        map[string]interface{} `json:"metadata,omitempty" example:"{\"priority_escalated\": true}"`
	Messages        []TicketMessageResponse `json:"messages,omitempty"`
	CreatedAt       time.Time        `json:"created_at" example:"2024-01-01T09:00:00Z"`
	UpdatedAt       time.Time        `json:"updated_at" example:"2024-01-02T16:00:00Z"`
}

// TicketMessageResponse represents the API response for a ticket message
type TicketMessageResponse struct {
	ID          uint                `json:"id" example:"1"`
	TicketID    uint                `json:"ticket_id" example:"1"`
	UserID      uint                `json:"user_id" example:"2"`
	Content     string              `json:"content" example:"Thank you for contacting support. We will review your issue."`
	MessageType string              `json:"message_type" example:"admin"`
	Attachments []Attachment        `json:"attachments,omitempty"`
	IsInternal  bool                `json:"is_internal" example:"false"`
	Metadata    map[string]interface{} `json:"metadata,omitempty" example:"{\"priority\": \"normal\"}"`
	CreatedAt   time.Time           `json:"created_at" example:"2024-01-01T10:30:00Z"`
	UpdatedAt   time.Time           `json:"updated_at" example:"2024-01-01T10:30:00Z"`
}

// TicketListResponse represents the paginated response for ticket lists
type TicketListResponse struct {
	Tickets []TicketResponse `json:"tickets"`
	Total   int64            `json:"total"`
	Limit   int              `json:"limit"`
	Offset  int              `json:"offset"`
}

// TicketMessageListResponse represents the paginated response for ticket message lists
type TicketMessageListResponse struct {
	Messages []TicketMessageResponse `json:"messages"`
	Total    int64                   `json:"total"`
	Limit    int                     `json:"limit"`
	Offset   int                     `json:"offset"`
}

// TicketStatisticsResponse represents ticket statistics
type TicketStatisticsResponse struct {
	Total         int64            `json:"total"`
	ByStatus      map[string]int64 `json:"by_status"`
	ByPriority    map[string]int64 `json:"by_priority"`
	ByCategory    map[string]int64 `json:"by_category"`
	Unassigned    int64            `json:"unassigned"`
	OverdueCount  int64            `json:"overdue_count"`
	ResolvedToday int64            `json:"resolved_today"`
	CreatedToday  int64            `json:"created_today"`
}

// FromDomainTicket converts domain model to DTO
func FromDomainTicket(ticket *model.Ticket) TicketResponse {
	response := TicketResponse{
		ID:           ticket.ID().Value(),
		TicketNumber: ticket.TicketNumber().Value(),
		Title:        ticket.Title(),
		Description:  ticket.Description(),
		Category:     ticket.Category().Value(),
		Priority:     ticket.Priority().Value(),
		Status:       ticket.Status().Value(),
		UserID:       ticket.UserID().ToUint(),
		Resolution:   ticket.Resolution(),
		Tags:         ticket.Tags(),
		Metadata:     ticket.Metadata(),
		CreatedAt:    ticket.CreatedAt(),
		UpdatedAt:    ticket.UpdatedAt(),
	}
	
	// Handle optional fields
	if assignedTo := ticket.AssignedToID(); assignedTo != nil {
		assignedToValue := assignedTo.ToUint()
		response.AssignedToID = &assignedToValue
	}
	
	if resolvedBy := ticket.ResolvedByID(); resolvedBy != nil {
		resolvedByValue := resolvedBy.ToUint()
		response.ResolvedByID = &resolvedByValue
	}
	
	response.AssignedAt = ticket.AssignedAt()
	response.ResolvedAt = ticket.ResolvedAt()
	response.FirstResponseAt = ticket.FirstResponseAt()
	response.LastResponseAt = ticket.LastResponseAt()
	response.ClosedAt = ticket.ClosedAt()
	
	// Convert messages
	messages := ticket.Messages()
	response.Messages = make([]TicketMessageResponse, len(messages))
	for i, msg := range messages {
		response.Messages[i] = FromDomainTicketMessage(&msg)
	}
	
	return response
}

// FromDomainTicketMessage converts domain model to DTO
func FromDomainTicketMessage(message *model.TicketMessage) TicketMessageResponse {
	// Convert attachments
	attachments := make([]Attachment, len(message.Attachments()))
	for i, att := range message.Attachments() {
		attachments[i] = Attachment{
			Name: att.Name,
			URL:  att.URL,
			Size: att.Size,
			Type: att.Type,
		}
	}
	
	return TicketMessageResponse{
		ID:          message.ID(),
		TicketID:    message.TicketID().Value(),
		UserID:      message.UserID().ToUint(),
		Content:     message.Content(),
		MessageType: string(message.MessageType()),
		Attachments: attachments,
		IsInternal:  message.IsInternal(),
		Metadata:    message.Metadata(),
		CreatedAt:   message.CreatedAt(),
		UpdatedAt:   message.UpdatedAt(),
	}
}