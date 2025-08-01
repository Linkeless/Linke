package persistence

import (
	"context"
	"encoding/json"
	"fmt"

	"gorm.io/gorm"

	"linke/internal/ticket/domain/model"
	"linke/internal/ticket/domain/valueobject"
	sharedvo "linke/internal/shared/valueobject"
)

// TicketMessageGormRepository implements TicketMessageRepository using GORM
type TicketMessageGormRepository struct {
	db *gorm.DB
}

// NewTicketMessageGormRepository creates a new GORM ticket message repository
func NewTicketMessageGormRepository(db *gorm.DB) *TicketMessageGormRepository {
	return &TicketMessageGormRepository{db: db}
}

// Save saves a ticket message
func (r *TicketMessageGormRepository) Save(ctx context.Context, message *model.TicketMessage) error {
	po := r.toPO(message)
	
	if po.ID == 0 {
		// Create new message
		if err := r.db.WithContext(ctx).Create(po).Error; err != nil {
			return fmt.Errorf("failed to create ticket message: %w", err)
		}
	} else {
		// Update existing message
		if err := r.db.WithContext(ctx).Save(po).Error; err != nil {
			return fmt.Errorf("failed to update ticket message: %w", err)
		}
	}
	
	return nil
}

// FindByID finds a message by ID
func (r *TicketMessageGormRepository) FindByID(ctx context.Context, id uint) (*model.TicketMessage, error) {
	var po TicketMessagePO
	
	if err := r.db.WithContext(ctx).First(&po, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("ticket message not found")
		}
		return nil, fmt.Errorf("failed to find ticket message: %w", err)
	}
	
	return r.fromPO(&po)
}

// FindByTicketID finds messages by ticket ID
func (r *TicketMessageGormRepository) FindByTicketID(ctx context.Context, ticketID valueobject.TicketID, includeInternal bool, limit, offset int) ([]*model.TicketMessage, int64, error) {
	query := r.db.WithContext(ctx).Where("ticket_id = ?", ticketID.Value())
	
	if !includeInternal {
		query = query.Where("is_internal = ?", false)
	}
	
	// Get count
	var count int64
	if err := query.Model(&TicketMessagePO{}).Count(&count).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count ticket messages: %w", err)
	}
	
	// Get messages
	var pos []TicketMessagePO
	query = query.Order("created_at ASC")
	
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}
	
	if err := query.Find(&pos).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to find ticket messages: %w", err)
	}
	
	// Convert to domain models
	messages := make([]*model.TicketMessage, len(pos))
	for i, po := range pos {
		message, err := r.fromPO(&po)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to convert ticket message: %w", err)
		}
		messages[i] = message
	}
	
	return messages, count, nil
}

// Delete soft deletes a message
func (r *TicketMessageGormRepository) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&TicketMessagePO{}, id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete ticket message: %w", result.Error)
	}
	
	if result.RowsAffected == 0 {
		return fmt.Errorf("ticket message not found")
	}
	
	return nil
}

// toPO converts domain model to persistence object
func (r *TicketMessageGormRepository) toPO(message *model.TicketMessage) *TicketMessagePO {
	po := &TicketMessagePO{
		ID:          message.ID(),
		TicketID:    message.TicketID().Value(),
		UserID:      message.UserID().ToUint(),
		Content:     message.Content(),
		MessageType: string(message.MessageType()),
		IsInternal:  message.IsInternal(),
		CreatedAt:   message.CreatedAt(),
		UpdatedAt:   message.UpdatedAt(),
	}
	
	// Convert attachments
	if attachments := message.Attachments(); len(attachments) > 0 {
		attachmentsJSON, _ := json.Marshal(attachments)
		attachmentsString := string(attachmentsJSON)
		po.Attachments = &attachmentsString
	}
	
	// Convert metadata
	if metadata := message.Metadata(); len(metadata) > 0 {
		metadataJSON, _ := json.Marshal(metadata)
		metadataString := string(metadataJSON)
		po.Metadata = &metadataString
	}
	
	return po
}

// fromPO converts persistence object to domain model
func (r *TicketMessageGormRepository) fromPO(po *TicketMessagePO) (*model.TicketMessage, error) {
	ticketID := valueobject.NewTicketID(po.TicketID)
	userID, err := sharedvo.NewUserIDFromUint(po.UserID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}
	
	message, err := model.NewTicketMessage(
		po.ID,
		ticketID,
		userID,
		po.Content,
		model.MessageType(po.MessageType),
		po.IsInternal,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create ticket message: %w", err)
	}
	
	// Parse attachments
	if po.Attachments != nil && *po.Attachments != "" {
		var attachments []model.Attachment
		if err := json.Unmarshal([]byte(*po.Attachments), &attachments); err == nil {
			message.SetAttachments(attachments)
		}
	}
	
	// Parse metadata
	if po.Metadata != nil && *po.Metadata != "" {
		var metadata map[string]interface{}
		if err := json.Unmarshal([]byte(*po.Metadata), &metadata); err == nil {
			for key, value := range metadata {
				message.SetMetadata(key, value)
			}
		}
	}
	
	return message, nil
}