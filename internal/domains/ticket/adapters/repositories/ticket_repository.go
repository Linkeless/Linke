package repositories

import (
	"context"
	"fmt"
	"time"

	"linke/internal/domains/ticket/entities"
	"linke/internal/domains/ticket/usecases/interfaces"
	"linke/internal/shared/framework"
	"linke/internal/shared/repository"

	"gorm.io/gorm"
)

// ticketRepository implements the TicketRepository interface
type ticketRepository struct {
	*repository.UserScopedTimeBasedRepositoryImpl[entities.Ticket, uint]
}

// NewTicketRepository creates a new TicketRepository implementation
func NewTicketRepository(db *gorm.DB, logger framework.Logger) interfaces.TicketRepository {
	return &ticketRepository{
		UserScopedTimeBasedRepositoryImpl: repository.NewUserScopedTimeBasedRepository[entities.Ticket, uint](db, logger),
	}
}

// ListOpen retrieves open tickets with pagination
func (r *ticketRepository) ListOpen(ctx context.Context, limit, offset int) ([]*entities.Ticket, int64, error) {
	return r.ListByStatus(ctx, entities.TicketStatusOpen, limit, offset)
}

// ListClosed retrieves closed tickets with pagination
func (r *ticketRepository) ListClosed(ctx context.Context, limit, offset int) ([]*entities.Ticket, int64, error) {
	return r.ListByStatus(ctx, entities.TicketStatusClosed, limit, offset)
}

// ListByPriority retrieves tickets by priority with pagination
func (r *ticketRepository) ListByPriority(ctx context.Context, priority string, limit, offset int) ([]*entities.Ticket, int64, error) {
	var tickets []*entities.Ticket
	var total int64

	// Count total tickets by priority
	if err := r.GetDB().WithContext(ctx).Model(&entities.Ticket{}).Where("priority = ?", priority).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count tickets by priority: %w", err)
	}

	// Get paginated tickets by priority
	query := r.GetDB().WithContext(ctx).Where("priority = ?", priority).Order("created_at DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&tickets).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list tickets by priority: %w", err)
	}

	return tickets, total, nil
}

// ListHighPriority retrieves high priority tickets with pagination
func (r *ticketRepository) ListHighPriority(ctx context.Context, limit, offset int) ([]*entities.Ticket, int64, error) {
	var tickets []*entities.Ticket
	var total int64

	// Count high and urgent priority tickets
	query := r.GetDB().WithContext(ctx).Model(&entities.Ticket{}).
		Where("priority IN (?)", []string{entities.TicketPriorityHigh, entities.TicketPriorityUrgent, entities.TicketPriorityCritical})

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count high priority tickets: %w", err)
	}

	// Get paginated high priority tickets
	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&tickets).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list high priority tickets: %w", err)
	}

	return tickets, total, nil
}

// ListUnassigned retrieves unassigned tickets with pagination
func (r *ticketRepository) ListUnassigned(ctx context.Context, limit, offset int) ([]*entities.Ticket, int64, error) {
	var tickets []*entities.Ticket
	var total int64

	// Count unassigned tickets
	query := r.GetDB().WithContext(ctx).Model(&entities.Ticket{}).Where("assigned_to_id IS NULL")

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count unassigned tickets: %w", err)
	}

	// Get paginated unassigned tickets
	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&tickets).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list unassigned tickets: %w", err)
	}

	return tickets, total, nil
}

// ListByAssignee retrieves tickets assigned to a specific user with pagination
func (r *ticketRepository) ListByAssignee(ctx context.Context, assigneeID uint, limit, offset int) ([]*entities.Ticket, int64, error) {
	var tickets []*entities.Ticket
	var total int64

	// Count tickets by assignee
	if err := r.GetDB().WithContext(ctx).Model(&entities.Ticket{}).Where("assigned_to_id = ?", assigneeID).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count tickets by assignee: %w", err)
	}

	// Get paginated tickets by assignee
	query := r.GetDB().WithContext(ctx).Where("assigned_to_id = ?", assigneeID).Order("created_at DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&tickets).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list tickets by assignee: %w", err)
	}

	return tickets, total, nil
}

// UpdateAssignee updates the assignee of a ticket
func (r *ticketRepository) UpdateAssignee(ctx context.Context, id uint, assigneeID *uint) error {
	updates := map[string]interface{}{
		"assigned_to_id": assigneeID,
	}

	// Set assigned_at timestamp if assigning to someone
	if assigneeID != nil {
		now := time.Now()
		updates["assigned_at"] = &now
	} else {
		updates["assigned_at"] = nil
	}

	result := r.GetDB().WithContext(ctx).Model(&entities.Ticket{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		return fmt.Errorf("failed to update ticket assignee: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("ticket with id %d not found", id)
	}
	return nil
}

// ListRecentTickets retrieves tickets created after a specific time
func (r *ticketRepository) ListRecentTickets(ctx context.Context, since time.Time, limit, offset int) ([]*entities.Ticket, int64, error) {
	return r.ListCreatedAfter(ctx, since, limit, offset)
}

// GetStatusStats returns statistics about tickets grouped by status
func (r *ticketRepository) GetStatusStats(ctx context.Context) (map[string]int64, error) {
	var results []struct {
		Status string
		Count  int64
	}

	if err := r.GetDB().WithContext(ctx).Model(&entities.Ticket{}).
		Select("status, COUNT(*) as count").
		Group("status").
		Scan(&results).Error; err != nil {
		return nil, fmt.Errorf("failed to get status stats: %w", err)
	}

	stats := make(map[string]int64)
	for _, result := range results {
		stats[result.Status] = result.Count
	}

	return stats, nil
}

// GetPriorityStats returns statistics about tickets grouped by priority
func (r *ticketRepository) GetPriorityStats(ctx context.Context) (map[string]int64, error) {
	var results []struct {
		Priority string
		Count    int64
	}

	if err := r.GetDB().WithContext(ctx).Model(&entities.Ticket{}).
		Select("priority, COUNT(*) as count").
		Group("priority").
		Scan(&results).Error; err != nil {
		return nil, fmt.Errorf("failed to get priority stats: %w", err)
	}

	stats := make(map[string]int64)
	for _, result := range results {
		stats[result.Priority] = result.Count
	}

	return stats, nil
}