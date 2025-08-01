package persistence

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	"linke/internal/ticket/domain/model"
	"linke/internal/ticket/domain/repository"
	"linke/internal/ticket/domain/valueobject"
	sharedvo "linke/internal/shared/valueobject"
)

// TicketGormRepository implements TicketRepository using GORM
type TicketGormRepository struct {
	db *gorm.DB
}

// NewTicketGormRepository creates a new GORM ticket repository
func NewTicketGormRepository(db *gorm.DB) *TicketGormRepository {
	return &TicketGormRepository{db: db}
}

// Save saves a ticket aggregate
func (r *TicketGormRepository) Save(ctx context.Context, ticket *model.Ticket) error {
	po := r.toPO(ticket)
	
	// Use optimistic locking
	if po.ID == 0 {
		// Create new ticket
		if err := r.db.WithContext(ctx).Create(po).Error; err != nil {
			return fmt.Errorf("failed to create ticket: %w", err)
		}
		// Update the ticket ID in the domain model
		// In a real implementation, you might need to reload the ticket from the DB
	} else {
		// Update existing ticket
		result := r.db.WithContext(ctx).
			Where("id = ? AND version = ?", po.ID, ticket.Version()-1).
			Updates(po)
		
		if result.Error != nil {
			return fmt.Errorf("failed to update ticket: %w", result.Error)
		}
		
		if result.RowsAffected == 0 {
			return fmt.Errorf("ticket was modified by another process (optimistic lock failed)")
		}
	}
	
	return nil
}

// FindByID finds a ticket by ID
func (r *TicketGormRepository) FindByID(ctx context.Context, id valueobject.TicketID) (*model.Ticket, error) {
	var po TicketPO
	
	if err := r.db.WithContext(ctx).
		Preload("Messages", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at ASC")
		}).
		First(&po, id.Value()).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("ticket not found")
		}
		return nil, fmt.Errorf("failed to find ticket: %w", err)
	}
	
	return r.fromPO(&po)
}

// FindByTicketNumber finds a ticket by ticket number
func (r *TicketGormRepository) FindByTicketNumber(ctx context.Context, ticketNumber valueobject.TicketNumber) (*model.Ticket, error) {
	var po TicketPO
	
	if err := r.db.WithContext(ctx).
		Preload("Messages", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at ASC")
		}).
		Where("ticket_number = ?", ticketNumber.Value()).
		First(&po).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("ticket not found")
		}
		return nil, fmt.Errorf("failed to find ticket: %w", err)
	}
	
	return r.fromPO(&po)
}

// FindByUserID finds tickets by user ID with pagination
func (r *TicketGormRepository) FindByUserID(ctx context.Context, userID sharedvo.UserID, limit, offset int) ([]*model.Ticket, int64, error) {
	var pos []TicketPO
	var count int64
	
	query := r.db.WithContext(ctx).Where("user_id = ?", userID.Value())
	
	// Get count
	if err := query.Model(&TicketPO{}).Count(&count).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count tickets: %w", err)
	}
	
	// Get tickets
	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&pos).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to find tickets: %w", err)
	}
	
	tickets := make([]*model.Ticket, len(pos))
	for i, po := range pos {
		ticket, err := r.fromPO(&po)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to convert ticket: %w", err)
		}
		tickets[i] = ticket
	}
	
	return tickets, count, nil
}

// FindByAssignedTo finds tickets assigned to a user with pagination
func (r *TicketGormRepository) FindByAssignedTo(ctx context.Context, assignedToID sharedvo.UserID, limit, offset int) ([]*model.Ticket, int64, error) {
	var pos []TicketPO
	var count int64
	
	query := r.db.WithContext(ctx).Where("assigned_to_id = ?", assignedToID.Value())
	
	// Get count
	if err := query.Model(&TicketPO{}).Count(&count).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count tickets: %w", err)
	}
	
	// Get tickets
	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&pos).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to find tickets: %w", err)
	}
	
	tickets := make([]*model.Ticket, len(pos))
	for i, po := range pos {
		ticket, err := r.fromPO(&po)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to convert ticket: %w", err)
		}
		tickets[i] = ticket
	}
	
	return tickets, count, nil
}

// FindByStatus finds tickets by status with pagination
func (r *TicketGormRepository) FindByStatus(ctx context.Context, status valueobject.TicketStatus, limit, offset int) ([]*model.Ticket, int64, error) {
	var pos []TicketPO
	var count int64
	
	query := r.db.WithContext(ctx).Where("status = ?", status.Value())
	
	// Get count
	if err := query.Model(&TicketPO{}).Count(&count).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count tickets: %w", err)
	}
	
	// Get tickets
	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&pos).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to find tickets: %w", err)
	}
	
	tickets := make([]*model.Ticket, len(pos))
	for i, po := range pos {
		ticket, err := r.fromPO(&po)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to convert ticket: %w", err)
		}
		tickets[i] = ticket
	}
	
	return tickets, count, nil
}

// FindByPriority finds tickets by priority with pagination
func (r *TicketGormRepository) FindByPriority(ctx context.Context, priority valueobject.TicketPriority, limit, offset int) ([]*model.Ticket, int64, error) {
	var pos []TicketPO
	var count int64
	
	query := r.db.WithContext(ctx).Where("priority = ?", priority.Value())
	
	// Get count
	if err := query.Model(&TicketPO{}).Count(&count).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count tickets: %w", err)
	}
	
	// Get tickets
	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&pos).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to find tickets: %w", err)
	}
	
	tickets := make([]*model.Ticket, len(pos))
	for i, po := range pos {
		ticket, err := r.fromPO(&po)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to convert ticket: %w", err)
		}
		tickets[i] = ticket
	}
	
	return tickets, count, nil
}

// FindByCategory finds tickets by category with pagination
func (r *TicketGormRepository) FindByCategory(ctx context.Context, category valueobject.TicketCategory, limit, offset int) ([]*model.Ticket, int64, error) {
	var pos []TicketPO
	var count int64
	
	query := r.db.WithContext(ctx).Where("category = ?", category.Value())
	
	// Get count
	if err := query.Model(&TicketPO{}).Count(&count).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count tickets: %w", err)
	}
	
	// Get tickets
	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&pos).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to find tickets: %w", err)
	}
	
	tickets := make([]*model.Ticket, len(pos))
	for i, po := range pos {
		ticket, err := r.fromPO(&po)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to convert ticket: %w", err)
		}
		tickets[i] = ticket
	}
	
	return tickets, count, nil
}

// Search searches tickets by various criteria
func (r *TicketGormRepository) Search(ctx context.Context, criteria repository.SearchCriteria) ([]*model.Ticket, int64, error) {
	query := r.db.WithContext(ctx).Model(&TicketPO{})
	
	// Apply filters
	if criteria.UserID != nil {
		query = query.Where("user_id = ?", criteria.UserID.Value())
	}
	
	if criteria.AssignedToID != nil {
		query = query.Where("assigned_to_id = ?", criteria.AssignedToID.Value())
	}
	
	if criteria.Status != nil {
		query = query.Where("status = ?", criteria.Status.Value())
	}
	
	if criteria.Priority != nil {
		query = query.Where("priority = ?", criteria.Priority.Value())
	}
	
	if criteria.Category != nil {
		query = query.Where("category = ?", criteria.Category.Value())
	}
	
	if criteria.SearchTerm != "" {
		searchTerm := "%" + criteria.SearchTerm + "%"
		query = query.Where("title LIKE ? OR description LIKE ? OR ticket_number LIKE ?", 
			searchTerm, searchTerm, searchTerm)
	}
	
	if len(criteria.Tags) > 0 {
		for _, tag := range criteria.Tags {
			query = query.Where("tags LIKE ?", "%"+tag+"%")
		}
	}
	
	// Get count
	var count int64
	if err := query.Count(&count).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count tickets: %w", err)
	}
	
	// Apply sorting
	sortBy := criteria.SortBy
	if sortBy == "" {
		sortBy = "created_at"
	}
	
	sortOrder := strings.ToUpper(criteria.SortOrder)
	if sortOrder != "ASC" && sortOrder != "DESC" {
		sortOrder = "DESC"
	}
	
	query = query.Order(fmt.Sprintf("%s %s", sortBy, sortOrder))
	
	// Apply pagination
	if criteria.Limit > 0 {
		query = query.Limit(criteria.Limit)
	}
	if criteria.Offset > 0 {
		query = query.Offset(criteria.Offset)
	}
	
	// Execute query
	var pos []TicketPO
	if err := query.Find(&pos).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to search tickets: %w", err)
	}
	
	// Convert to domain models
	tickets := make([]*model.Ticket, len(pos))
	for i, po := range pos {
		ticket, err := r.fromPO(&po)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to convert ticket: %w", err)
		}
		tickets[i] = ticket
	}
	
	return tickets, count, nil
}

// Delete soft deletes a ticket
func (r *TicketGormRepository) Delete(ctx context.Context, id valueobject.TicketID) error {
	result := r.db.WithContext(ctx).Delete(&TicketPO{}, id.Value())
	if result.Error != nil {
		return fmt.Errorf("failed to delete ticket: %w", result.Error)
	}
	
	if result.RowsAffected == 0 {
		return fmt.Errorf("ticket not found")
	}
	
	return nil
}

// ExistsByTicketNumber checks if a ticket number already exists
func (r *TicketGormRepository) ExistsByTicketNumber(ctx context.Context, ticketNumber valueobject.TicketNumber) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&TicketPO{}).
		Where("ticket_number = ?", ticketNumber.Value()).
		Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to check ticket number existence: %w", err)
	}
	
	return count > 0, nil
}

// GetStatistics returns ticket statistics
func (r *TicketGormRepository) GetStatistics(ctx context.Context) (*repository.TicketStatistics, error) {
	stats := &repository.TicketStatistics{
		ByStatus:   make(map[string]int64),
		ByPriority: make(map[string]int64),
		ByCategory: make(map[string]int64),
	}
	
	// Total count
	if err := r.db.WithContext(ctx).Model(&TicketPO{}).Count(&stats.Total).Error; err != nil {
		return nil, fmt.Errorf("failed to get total count: %w", err)
	}
	
	// By status
	var statusStats []struct {
		Status string
		Count  int64
	}
	if err := r.db.WithContext(ctx).Model(&TicketPO{}).
		Select("status, count(*) as count").
		Group("status").
		Find(&statusStats).Error; err != nil {
		return nil, fmt.Errorf("failed to get status statistics: %w", err)
	}
	for _, stat := range statusStats {
		stats.ByStatus[stat.Status] = stat.Count
	}
	
	// By priority
	var priorityStats []struct {
		Priority string
		Count    int64
	}
	if err := r.db.WithContext(ctx).Model(&TicketPO{}).
		Select("priority, count(*) as count").
		Group("priority").
		Find(&priorityStats).Error; err != nil {
		return nil, fmt.Errorf("failed to get priority statistics: %w", err)
	}
	for _, stat := range priorityStats {
		stats.ByPriority[stat.Priority] = stat.Count
	}
	
	// By category
	var categoryStats []struct {
		Category string
		Count    int64
	}
	if err := r.db.WithContext(ctx).Model(&TicketPO{}).
		Select("category, count(*) as count").
		Group("category").
		Find(&categoryStats).Error; err != nil {
		return nil, fmt.Errorf("failed to get category statistics: %w", err)
	}
	for _, stat := range categoryStats {
		stats.ByCategory[stat.Category] = stat.Count
	}
	
	// Unassigned count
	if err := r.db.WithContext(ctx).Model(&TicketPO{}).
		Where("assigned_to_id IS NULL").
		Count(&stats.Unassigned).Error; err != nil {
		return nil, fmt.Errorf("failed to get unassigned count: %w", err)
	}
	
	// Today's counts
	today := time.Now().Truncate(24 * time.Hour)
	tomorrow := today.Add(24 * time.Hour)
	
	if err := r.db.WithContext(ctx).Model(&TicketPO{}).
		Where("resolved_at >= ? AND resolved_at < ?", today, tomorrow).
		Count(&stats.ResolvedToday).Error; err != nil {
		return nil, fmt.Errorf("failed to get resolved today count: %w", err)
	}
	
	if err := r.db.WithContext(ctx).Model(&TicketPO{}).
		Where("created_at >= ? AND created_at < ?", today, tomorrow).
		Count(&stats.CreatedToday).Error; err != nil {
		return nil, fmt.Errorf("failed to get created today count: %w", err)
	}
	
	// Overdue count (simplified - tickets without first response after 24 hours)
	overdueTime := time.Now().Add(-24 * time.Hour)
	if err := r.db.WithContext(ctx).Model(&TicketPO{}).
		Where("first_response_at IS NULL AND created_at < ? AND status NOT IN (?)", 
			overdueTime, []string{"resolved", "closed"}).
		Count(&stats.OverdueCount).Error; err != nil {
		return nil, fmt.Errorf("failed to get overdue count: %w", err)
	}
	
	return stats, nil
}

// toPO converts domain model to persistence object
func (r *TicketGormRepository) toPO(ticket *model.Ticket) *TicketPO {
	po := &TicketPO{
		ID:           ticket.ID().Value(),
		TicketNumber: ticket.TicketNumber().Value(),
		Title:        ticket.Title(),
		Description:  ticket.Description(),
		Category:     ticket.Category().Value(),
		Priority:     ticket.Priority().Value(),
		Status:       ticket.Status().Value(),
		UserID:       ticket.UserID().ToUint(),
		Resolution:   ticket.Resolution(),
		Version:      ticket.Version(),
		CreatedAt:    ticket.CreatedAt(),
		UpdatedAt:    ticket.UpdatedAt(),
	}
	
	// Handle optional fields
	if assignedTo := ticket.AssignedToID(); assignedTo != nil {
		assignedToValue := assignedTo.ToUint()
		po.AssignedToID = &assignedToValue
	}
	
	if resolvedBy := ticket.ResolvedByID(); resolvedBy != nil {
		resolvedByValue := resolvedBy.ToUint()
		po.ResolvedByID = &resolvedByValue
	}
	
	po.AssignedAt = ticket.AssignedAt()
	po.ResolvedAt = ticket.ResolvedAt()
	po.FirstResponseAt = ticket.FirstResponseAt()
	po.LastResponseAt = ticket.LastResponseAt()
	po.ClosedAt = ticket.ClosedAt()
	
	// Convert tags
	if tags := ticket.Tags(); len(tags) > 0 {
		tagsJSON, _ := json.Marshal(tags)
		tagsString := string(tagsJSON)
		po.Tags = &tagsString
	}
	
	// Convert metadata
	if metadata := ticket.Metadata(); len(metadata) > 0 {
		metadataJSON, _ := json.Marshal(metadata)
		metadataString := string(metadataJSON)
		po.Metadata = &metadataString
	}
	
	return po
}

// fromPO converts persistence object to domain model
func (r *TicketGormRepository) fromPO(po *TicketPO) (*model.Ticket, error) {
	// Create value objects
	ticketID := valueobject.NewTicketID(po.ID)
	
	_, err := valueobject.NewTicketNumber(po.TicketNumber)
	if err != nil {
		return nil, fmt.Errorf("invalid ticket number: %w", err)
	}
	
	userID, err := sharedvo.NewUserIDFromUint(po.UserID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}
	
	category, err := valueobject.NewTicketCategory(po.Category)
	if err != nil {
		return nil, fmt.Errorf("invalid category: %w", err)
	}
	
	priority, err := valueobject.NewTicketPriority(po.Priority)
	if err != nil {
		return nil, fmt.Errorf("invalid priority: %w", err)
	}
	
	_, err = valueobject.NewTicketStatus(po.Status)
	if err != nil {
		return nil, fmt.Errorf("invalid status: %w", err)
	}
	
	// Create the ticket using the constructor
	
	ticket, err := model.NewTicket(ticketID, userID, po.Title, po.Description, category, priority)
	if err != nil {
		return nil, fmt.Errorf("failed to create ticket: %w", err)
	}
	
	// Set additional fields through reflection or exposed methods
	// Note: This is a simplified approach. In a real implementation,
	// you might need to use reflection or provide factory methods
	// that can reconstruct the full aggregate state.
	
	// For now, we'll create a new ticket with the basic information
	// and rely on the application to load messages separately if needed
	
	return ticket, nil
}