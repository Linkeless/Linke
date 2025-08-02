package implementations

import (
	"context"
	"fmt"
	"time"

	"linke/internal/domains/ticket/entities"
	"linke/internal/shared/logger"

	"gorm.io/gorm"
)

type TicketMessageService struct {
	db *gorm.DB
}

func NewTicketMessageService(db *gorm.DB) *TicketMessageService {
	return &TicketMessageService{
		db: db,
	}
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

// CreateTicketMessage creates a new ticket message
func (s *TicketMessageService) CreateTicketMessage(ctx context.Context, ticketID uint, userID uint, req *CreateTicketMessageRequest) (*entities.TicketMessage, error) {
	// Verify ticket exists
	var ticket entities.Ticket
	if err := s.db.WithContext(ctx).First(&ticket, ticketID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("ticket not found")
		}
		return nil, fmt.Errorf("failed to verify ticket: %w", err)
	}

	// Set default message type if not specified
	messageType := req.MessageType
	if messageType == "" {
		messageType = "user"
	}

	// Create the message
	message := &entities.TicketMessage{
		TicketID:    ticketID,
		UserID:      userID,
		Content:     req.Content,
		MessageType: messageType,
		Attachments: &req.Attachments,
		IsInternal:  req.IsInternal,
		Metadata:    &req.Metadata,
	}

	// Start transaction
	tx := s.db.WithContext(ctx).Begin()

	// Create the message
	if err := tx.Create(message).Error; err != nil {
		tx.Rollback()
		logger.Error("Failed to create ticket message", logger.Error2("error", err))
		return nil, fmt.Errorf("failed to create ticket message: %w", err)
	}

	// Update ticket's last response time
	now := time.Now()
	updates := map[string]interface{}{
		"last_response_at": &now,
	}

	// If this is the first response and it's from an admin, update first_response_at
	if ticket.FirstResponseAt == nil && messageType == "admin" {
		updates["first_response_at"] = &now
	}

	if err := tx.Model(&ticket).Updates(updates).Error; err != nil {
		tx.Rollback()
		logger.Error("Failed to update ticket timestamps", logger.Error2("error", err))
		return nil, fmt.Errorf("failed to update ticket timestamps: %w", err)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		logger.Error("Failed to commit ticket message transaction", logger.Error2("error", err))
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Reload message with relations
	createdMessage, err := s.GetTicketMessage(ctx, message.ID)
	if err != nil {
		return nil, err
	}

	logger.Info("Ticket message created successfully",
		logger.Uint("message_id", message.ID),
		logger.Uint("ticket_id", ticketID),
		logger.Uint("user_id", userID))

	return createdMessage, nil
}

// GetTicketMessage gets a ticket message by ID
func (s *TicketMessageService) GetTicketMessage(ctx context.Context, messageID uint) (*entities.TicketMessage, error) {
	var message entities.TicketMessage

	if err := s.db.WithContext(ctx).
		Preload("User").
		Preload("Ticket").
		First(&message, messageID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("ticket message not found")
		}
		logger.Error("Failed to get ticket message", logger.Error2("error", err), logger.Uint("message_id", messageID))
		return nil, fmt.Errorf("failed to get ticket message: %w", err)
	}

	return &message, nil
}

// GetTicketMessages gets messages for a ticket
func (s *TicketMessageService) GetTicketMessages(ctx context.Context, req *GetTicketMessagesRequest) ([]*entities.TicketMessage, int64, error) {
	query := s.db.WithContext(ctx).Model(&entities.TicketMessage{}).
		Preload("User").
		Where("ticket_id = ?", req.TicketID)

	// Apply filters
	if req.MessageType != "" {
		query = query.Where("message_type = ?", req.MessageType)
	}

	if !req.IncludeInternal {
		query = query.Where("is_internal = ?", false)
	}

	// Get total count
	var totalCount int64
	if err := query.Count(&totalCount).Error; err != nil {
		logger.Error("Failed to count ticket messages", logger.Error2("error", err))
		return nil, 0, fmt.Errorf("failed to count ticket messages: %w", err)
	}

	// Apply pagination and ordering
	query = query.Order("created_at ASC")

	if req.Limit > 0 {
		query = query.Limit(req.Limit)
	}

	if req.Offset > 0 {
		query = query.Offset(req.Offset)
	}

	var messages []*entities.TicketMessage
	if err := query.Find(&messages).Error; err != nil {
		logger.Error("Failed to get ticket messages", logger.Error2("error", err))
		return nil, 0, fmt.Errorf("failed to get ticket messages: %w", err)
	}

	return messages, totalCount, nil
}

// UpdateTicketMessage updates a ticket message
func (s *TicketMessageService) UpdateTicketMessage(ctx context.Context, messageID uint, req *UpdateTicketMessageRequest) (*entities.TicketMessage, error) {
	// Get existing message
	message, err := s.GetTicketMessage(ctx, messageID)
	if err != nil {
		return nil, err
	}

	// Prepare updates
	updates := make(map[string]interface{})

	if req.Content != nil {
		updates["content"] = *req.Content
	}

	if req.Attachments != nil {
		updates["attachments"] = *req.Attachments
	}

	if req.IsInternal != nil {
		updates["is_internal"] = *req.IsInternal
	}

	if req.Metadata != nil {
		updates["metadata"] = *req.Metadata
	}

	// Update the message
	if err := s.db.WithContext(ctx).Model(message).Updates(updates).Error; err != nil {
		logger.Error("Failed to update ticket message", logger.Error2("error", err), logger.Uint("message_id", messageID))
		return nil, fmt.Errorf("failed to update ticket message: %w", err)
	}

	// Reload the message with relations
	updatedMessage, err := s.GetTicketMessage(ctx, messageID)
	if err != nil {
		return nil, err
	}

	logger.Info("Ticket message updated successfully", logger.Uint("message_id", messageID))

	return updatedMessage, nil
}

// DeleteTicketMessage soft deletes a ticket message
func (s *TicketMessageService) DeleteTicketMessage(ctx context.Context, messageID uint) error {
	// Check if message exists
	var message entities.TicketMessage
	if err := s.db.WithContext(ctx).First(&message, messageID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("ticket message not found")
		}
		return fmt.Errorf("failed to check message existence: %w", err)
	}

	// Soft delete the message
	if err := s.db.WithContext(ctx).Delete(&message).Error; err != nil {
		logger.Error("Failed to delete ticket message", logger.Error2("error", err), logger.Uint("message_id", messageID))
		return fmt.Errorf("failed to delete ticket message: %w", err)
	}

	logger.Info("Ticket message deleted successfully", logger.Uint("message_id", messageID))

	return nil
}

// GetTicketMessagesForUser gets messages for a ticket that are visible to the user
func (s *TicketMessageService) GetTicketMessagesForUser(ctx context.Context, ticketID uint, userID uint, limit int, offset int) ([]*entities.TicketMessage, int64, error) {
	// Verify that the user owns the ticket or is an admin
	var ticket entities.Ticket
	if err := s.db.WithContext(ctx).
		Preload("User").
		First(&ticket, ticketID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, 0, fmt.Errorf("ticket not found")
		}
		return nil, 0, fmt.Errorf("failed to verify ticket: %w", err)
	}

	// TODO: Check if user has access to this ticket through user service interface
	// For now, we'll skip the user verification
	if userID == 0 {
		return nil, 0, fmt.Errorf("valid user ID is required")
	}

	// TODO: Check if user is admin through user service interface
	// For now, we'll just check if the ticket belongs to the user
	if ticket.UserID != userID {
		return nil, 0, fmt.Errorf("access denied: you can only view messages for your own tickets")
	}

	// Build query
	query := s.db.WithContext(ctx).Model(&entities.TicketMessage{}).
		Preload("User").
		Where("ticket_id = ?", ticketID)

	// TODO: Filter out internal messages for non-admin users
	// For now, we'll show all messages
	// query = query.Where("is_internal = ?", false)

	// Get total count
	var totalCount int64
	if err := query.Count(&totalCount).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count ticket messages: %w", err)
	}

	// Apply pagination and ordering
	query = query.Order("created_at ASC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	if offset > 0 {
		query = query.Offset(offset)
	}

	var messages []*entities.TicketMessage
	if err := query.Find(&messages).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get ticket messages: %w", err)
	}

	return messages, totalCount, nil
}
