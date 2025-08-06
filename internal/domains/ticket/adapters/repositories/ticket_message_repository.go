package repositories

import (
	"context"
	"linke/internal/domains/ticket/entities"
	"linke/internal/domains/ticket/usecases/interfaces"

	"gorm.io/gorm"
)

// TicketMessageRepository implements the TicketMessageRepository interface
type TicketMessageRepository struct {
	db *gorm.DB
}

// NewTicketMessageRepository creates a new ticket message repository
func NewTicketMessageRepository(db *gorm.DB) interfaces.TicketMessageRepository {
	return &TicketMessageRepository{db: db}
}

// Create creates a new ticket message
func (r *TicketMessageRepository) Create(ctx context.Context, message *entities.TicketMessage) error {
	return r.db.WithContext(ctx).Create(message).Error
}

// GetByID gets a ticket message by ID
func (r *TicketMessageRepository) GetByID(ctx context.Context, messageID uint) (*entities.TicketMessage, error) {
	var message entities.TicketMessage
	if err := r.db.WithContext(ctx).First(&message, messageID).Error; err != nil {
		return nil, err
	}
	return &message, nil
}

// Update updates a ticket message
func (r *TicketMessageRepository) Update(ctx context.Context, message *entities.TicketMessage) error {
	return r.db.WithContext(ctx).Save(message).Error
}

// Delete deletes a ticket message
func (r *TicketMessageRepository) Delete(ctx context.Context, messageID uint) error {
	return r.db.WithContext(ctx).Delete(&entities.TicketMessage{}, messageID).Error
}

// GetTicketMessages gets messages for a ticket with filtering
func (r *TicketMessageRepository) GetTicketMessages(ctx context.Context, ticketID uint, includeInternal bool, limit, offset int) ([]*entities.TicketMessage, int64, error) {
	var messages []*entities.TicketMessage
	var total int64

	query := r.db.WithContext(ctx).Where("ticket_id = ?", ticketID)

	if !includeInternal {
		query = query.Where("is_internal = ?", false)
	}

	// Get total count
	if err := query.Model(&entities.TicketMessage{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get results with pagination
	if err := query.Order("created_at ASC").Limit(limit).Offset(offset).Find(&messages).Error; err != nil {
		return nil, 0, err
	}

	return messages, total, nil
}

// GetLatestTicketMessages gets the latest messages for a ticket
func (r *TicketMessageRepository) GetLatestTicketMessages(ctx context.Context, ticketID uint, limit int) ([]*entities.TicketMessage, error) {
	var messages []*entities.TicketMessage
	if err := r.db.WithContext(ctx).
		Where("ticket_id = ? AND is_internal = ?", ticketID, false).
		Order("created_at DESC").
		Limit(limit).
		Find(&messages).Error; err != nil {
		return nil, err
	}
	return messages, nil
}

// GetInternalMessages gets internal messages for a ticket
func (r *TicketMessageRepository) GetInternalMessages(ctx context.Context, ticketID uint) ([]*entities.TicketMessage, error) {
	var messages []*entities.TicketMessage
	if err := r.db.WithContext(ctx).
		Where("ticket_id = ? AND is_internal = ?", ticketID, true).
		Order("created_at ASC").
		Find(&messages).Error; err != nil {
		return nil, err
	}
	return messages, nil
}

// MarkAsRead marks a message as read by a user
func (r *TicketMessageRepository) MarkAsRead(ctx context.Context, messageID uint, userID uint) error {
	// This could be implemented with a separate read tracking table
	// For now, we'll update the message entity if it has a read tracking field
	return r.db.WithContext(ctx).
		Model(&entities.TicketMessage{}).
		Where("id = ?", messageID).
		Update("read_at", gorm.Expr("NOW()")).Error
}

// MarkTicketMessagesAsRead marks all messages in a ticket as read by a user
func (r *TicketMessageRepository) MarkTicketMessagesAsRead(ctx context.Context, ticketID uint, userID uint) error {
	return r.db.WithContext(ctx).
		Model(&entities.TicketMessage{}).
		Where("ticket_id = ?", ticketID).
		Update("read_at", gorm.Expr("NOW()")).Error
}

// GetMessageStatistics gets statistics for messages in a ticket
func (r *TicketMessageRepository) GetMessageStatistics(ctx context.Context, ticketID uint) (map[string]any, error) {
	var stats struct {
		TotalMessages    int64 `gorm:"column:total_messages"`
		UserMessages     int64 `gorm:"column:user_messages"`
		AdminMessages    int64 `gorm:"column:admin_messages"`
		InternalMessages int64 `gorm:"column:internal_messages"`
	}

	err := r.db.WithContext(ctx).
		Model(&entities.TicketMessage{}).
		Select(`
			COUNT(*) as total_messages,
			COUNT(CASE WHEN message_type = 'user' THEN 1 END) as user_messages,
			COUNT(CASE WHEN message_type = 'admin' THEN 1 END) as admin_messages,
			COUNT(CASE WHEN is_internal = true THEN 1 END) as internal_messages
		`).
		Where("ticket_id = ?", ticketID).
		Scan(&stats).Error

	if err != nil {
		return nil, err
	}

	return map[string]any{
		"total_messages":    stats.TotalMessages,
		"user_messages":     stats.UserMessages,
		"admin_messages":    stats.AdminMessages,
		"internal_messages": stats.InternalMessages,
	}, nil
}

// BatchDelete deletes multiple ticket messages by IDs
func (r *TicketMessageRepository) BatchDelete(ctx context.Context, messageIDs []uint) (int, []uint, error) {
	if len(messageIDs) == 0 {
		return 0, nil, nil
	}
	
	result := r.db.WithContext(ctx).Delete(&entities.TicketMessage{}, messageIDs)
	if result.Error != nil {
		return 0, messageIDs, result.Error
	}
	
	successCount := int(result.RowsAffected)
	failedIDs := make([]uint, 0)
	
	// If not all rows were affected, some IDs might have failed
	if successCount < len(messageIDs) {
		for _, id := range messageIDs[successCount:] {
			failedIDs = append(failedIDs, id)
		}
	}
	
	return successCount, failedIDs, nil
}