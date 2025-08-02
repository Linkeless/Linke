package interfaces

import (
	"context"
	"linke/internal/domains/ticket/entities"
)

// TicketMessageService defines the interface for ticket message operations
type TicketMessageService interface {
	// Message CRUD operations
	CreateTicketMessage(ctx context.Context, ticketID uint, userID uint, req *CreateTicketMessageRequest) (*entities.TicketMessage, error)
	GetTicketMessage(ctx context.Context, messageID uint) (*entities.TicketMessage, error)
	UpdateTicketMessage(ctx context.Context, messageID uint, req *UpdateTicketMessageRequest) (*entities.TicketMessage, error)
	DeleteTicketMessage(ctx context.Context, messageID uint) error

	// Message listing
	GetTicketMessages(ctx context.Context, req *GetTicketMessagesRequest) ([]*entities.TicketMessage, int64, error)
	GetLatestTicketMessages(ctx context.Context, ticketID uint, limit int) ([]*entities.TicketMessage, error)

	// Message management
	MarkMessageAsRead(ctx context.Context, messageID uint, userID uint) error
	MarkTicketMessagesAsRead(ctx context.Context, ticketID uint, userID uint) error

	// Internal messages
	CreateInternalMessage(ctx context.Context, ticketID uint, userID uint, content string) (*entities.TicketMessage, error)
	GetInternalMessages(ctx context.Context, ticketID uint) ([]*entities.TicketMessage, error)

	// Message statistics
	GetMessageStatistics(ctx context.Context, ticketID uint) (map[string]interface{}, error)
}

// CreateTicketMessageRequest represents the request to create a ticket message
type CreateTicketMessageRequest struct {
	Content     string `json:"content" binding:"required,min=1,max=5000" example:"Thank you for your response. I tried the suggested solution but the issue persists."`
	MessageType string `json:"message_type" binding:"omitempty,oneof=user admin system" example:"user"`
	Attachments string `json:"attachments,omitempty" example:"[{\"name\":\"screenshot.png\",\"url\":\"https://example.com/file.png\"}]"`
	IsInternal  bool   `json:"is_internal,omitempty" example:"false"`
	Metadata    string `json:"metadata,omitempty" example:"{\"client_ip\":\"192.168.1.1\"}"`
}

// UpdateTicketMessageRequest represents the request to update a ticket message
type UpdateTicketMessageRequest struct {
	Content     *string `json:"content,omitempty" binding:"omitempty,min=1,max=5000" example:"Updated message content"`
	Attachments *string `json:"attachments,omitempty" example:"[{\"name\":\"updated.png\",\"url\":\"https://example.com/updated.png\"}]"`
	IsInternal  *bool   `json:"is_internal,omitempty" example:"true"`
	Metadata    *string `json:"metadata,omitempty" example:"{\"updated_by\":\"admin\"}"`
}

// GetTicketMessagesRequest represents the request to get ticket messages
type GetTicketMessagesRequest struct {
	TicketID        uint   `form:"ticket_id" binding:"required" example:"1"`
	MessageType     string `form:"message_type" binding:"omitempty,oneof=user admin system" example:"user"`
	IncludeInternal bool   `form:"include_internal" example:"false"`
	Limit           int    `form:"limit" binding:"omitempty,min=1,max=100" example:"10"`
	Offset          int    `form:"offset" binding:"omitempty,min=0" example:"0"`
}
