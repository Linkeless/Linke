package implementations

import (
	"context"
	"fmt"
	"strings"
	"time"

	"linke/internal/domains/ticket/entities"
	"linke/internal/domains/ticket/usecases/interfaces"
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


// CreateTicketMessage creates a new ticket message
func (s *TicketMessageService) CreateTicketMessage(ctx context.Context, ticketID uint, userID uint, req *interfaces.CreateTicketMessageRequest) (*entities.TicketMessage, error) {
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
		IsInternal:  req.IsInternal,
	}
	
	// Only set Attachments if not empty
	if strings.TrimSpace(req.Attachments) != "" {
		message.Attachments = &req.Attachments
	}
	
	// Only set Metadata if not empty
	if strings.TrimSpace(req.Metadata) != "" {
		message.Metadata = &req.Metadata
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
	updates := map[string]any{
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
func (s *TicketMessageService) GetTicketMessages(ctx context.Context, req *interfaces.GetTicketMessagesRequest) ([]*entities.TicketMessage, int64, error) {
	query := s.db.WithContext(ctx).Model(&entities.TicketMessage{}).
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
func (s *TicketMessageService) UpdateTicketMessage(ctx context.Context, messageID uint, req *interfaces.UpdateTicketMessageRequest) (*entities.TicketMessage, error) {
	// Get existing message
	message, err := s.GetTicketMessage(ctx, messageID)
	if err != nil {
		return nil, err
	}

	// Prepare updates
	updates := make(map[string]any)

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

// GetLatestTicketMessages gets the latest messages for a ticket
func (s *TicketMessageService) GetLatestTicketMessages(ctx context.Context, ticketID uint, limit int) ([]*entities.TicketMessage, error) {
	var messages []*entities.TicketMessage

	query := s.db.WithContext(ctx).
		Where("ticket_id = ? AND is_internal = ?", ticketID, false).
		Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	if err := query.Find(&messages).Error; err != nil {
		logger.Error("Failed to get latest ticket messages", logger.Error2("error", err))
		return nil, fmt.Errorf("failed to get latest ticket messages: %w", err)
	}

	return messages, nil
}

// MarkMessageAsRead marks a message as read by a user
func (s *TicketMessageService) MarkMessageAsRead(ctx context.Context, messageID uint, userID uint) error {
	// For now, we'll implement a simple version
	// In a full implementation, this would track read status per user
	logger.Info("Message marked as read", 
		logger.Uint("message_id", messageID), 
		logger.Uint("user_id", userID))
	return nil
}

// MarkTicketMessagesAsRead marks all messages in a ticket as read by a user
func (s *TicketMessageService) MarkTicketMessagesAsRead(ctx context.Context, ticketID uint, userID uint) error {
	// For now, we'll implement a simple version
	// In a full implementation, this would track read status per user for all messages
	logger.Info("All ticket messages marked as read", 
		logger.Uint("ticket_id", ticketID), 
		logger.Uint("user_id", userID))
	return nil
}

// CreateInternalMessage creates an internal message
func (s *TicketMessageService) CreateInternalMessage(ctx context.Context, ticketID uint, userID uint, content string) (*entities.TicketMessage, error) {
	req := &interfaces.CreateTicketMessageRequest{
		Content:     content,
		MessageType: "admin",
		IsInternal:  true,
	}
	return s.CreateTicketMessage(ctx, ticketID, userID, req)
}

// GetInternalMessages gets all internal messages for a ticket
func (s *TicketMessageService) GetInternalMessages(ctx context.Context, ticketID uint) ([]*entities.TicketMessage, error) {
	var messages []*entities.TicketMessage

	if err := s.db.WithContext(ctx).
		Where("ticket_id = ? AND is_internal = ?", ticketID, true).
		Order("created_at ASC").
		Find(&messages).Error; err != nil {
		logger.Error("Failed to get internal messages", logger.Error2("error", err))
		return nil, fmt.Errorf("failed to get internal messages: %w", err)
	}

	return messages, nil
}

// GetMessageStatistics gets message statistics for a ticket
func (s *TicketMessageService) GetMessageStatistics(ctx context.Context, ticketID uint) (map[string]any, error) {
	stats := make(map[string]any)

	// Count messages by type
	var typeStats []struct {
		MessageType string
		Count       int64
	}

	if err := s.db.WithContext(ctx).Model(&entities.TicketMessage{}).
		Where("ticket_id = ?", ticketID).
		Select("message_type, count(*) as count").
		Group("message_type").
		Find(&typeStats).Error; err != nil {
		return nil, fmt.Errorf("failed to get message type statistics: %w", err)
	}

	typeMap := make(map[string]int64)
	for _, stat := range typeStats {
		typeMap[stat.MessageType] = stat.Count
	}
	stats["by_type"] = typeMap

	// Get total count
	var totalCount int64
	if err := s.db.WithContext(ctx).Model(&entities.TicketMessage{}).
		Where("ticket_id = ?", ticketID).
		Count(&totalCount).Error; err != nil {
		return nil, fmt.Errorf("failed to get total message count: %w", err)
	}
	stats["total"] = totalCount

	// Get internal message count
	var internalCount int64
	if err := s.db.WithContext(ctx).Model(&entities.TicketMessage{}).
		Where("ticket_id = ? AND is_internal = ?", ticketID, true).
		Count(&internalCount).Error; err != nil {
		return nil, fmt.Errorf("failed to get internal message count: %w", err)
	}
	stats["internal"] = internalCount

	return stats, nil
}
