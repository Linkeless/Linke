package implementations

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"linke/internal/domains/ticket/entities"
	"linke/internal/shared/logger"

	"gorm.io/gorm"
)

type TicketService struct {
	db *gorm.DB
}

func NewTicketService(db *gorm.DB) *TicketService {
	return &TicketService{
		db: db,
	}
}

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

// AssignTicketRequest represents the request to assign a ticket
type AssignTicketRequest struct {
	AssignedToID uint `json:"assigned_to_id" binding:"required" example:"2"`
}

// ResolveTicketRequest represents the request to resolve a ticket
type ResolveTicketRequest struct {
	Resolution string `json:"resolution" binding:"required,min=10,max=5000" example:"Issue resolved by updating user permissions"`
}

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

// CreateTicket creates a new ticket
func (s *TicketService) CreateTicket(ctx context.Context, userID uint, req *CreateTicketRequest) (*entities.Ticket, error) {
	// Generate unique ticket number
	ticketNo := s.generateTicketNumber()

	// Ensure ticket number is unique
	for {
		var existing entities.Ticket
		if err := s.db.Where("ticket_no = ?", ticketNo).First(&existing).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				break // Ticket number is unique
			}
			return nil, fmt.Errorf("failed to check ticket number uniqueness: %w", err)
		}
		ticketNo = s.generateTicketNumber() // Generate new number
	}

	// Set default priority if not specified
	priority := req.Priority
	if priority == "" {
		priority = entities.TicketPriorityNormal
	}

	// Create the ticket
	ticket := &entities.Ticket{
		TicketNo:    ticketNo,
		Title:       req.Title,
		Description: req.Description,
		Category:    req.Category,
		Priority:    priority,
		Status:      entities.TicketStatusOpen,
		UserID:      userID,
		Tags:        &req.Tags,
		Metadata:    &req.Metadata,
	}

	if err := s.db.WithContext(ctx).Create(ticket).Error; err != nil {
		logger.Error("Failed to create ticket", logger.Error2("error", err))
		return nil, fmt.Errorf("failed to create ticket: %w", err)
	}

	logger.Info("Ticket created successfully",
		logger.String("ticket_no", ticket.TicketNo),
		logger.Uint("user_id", userID),
		logger.String("category", ticket.Category))

	return ticket, nil
}

// GetTicket gets a ticket by ID with relations
func (s *TicketService) GetTicket(ctx context.Context, ticketID uint) (*entities.Ticket, error) {
	var ticket entities.Ticket

	if err := s.db.WithContext(ctx).
		Preload("User").
		Preload("AssignedTo").
		Preload("ResolvedBy").
		Preload("Messages", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at ASC")
		}).
		Preload("Messages.User").
		First(&ticket, ticketID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("ticket not found")
		}
		logger.Error("Failed to get ticket", logger.Error2("error", err), logger.Uint("ticket_id", ticketID))
		return nil, fmt.Errorf("failed to get ticket: %w", err)
	}

	return &ticket, nil
}

// GetTicketByNumber gets a ticket by ticket number
func (s *TicketService) GetTicketByNumber(ctx context.Context, ticketNo string) (*entities.Ticket, error) {
	var ticket entities.Ticket

	if err := s.db.WithContext(ctx).
		Preload("User").
		Preload("AssignedTo").
		Preload("ResolvedBy").
		Preload("Messages", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at ASC")
		}).
		Preload("Messages.User").
		Where("ticket_no = ?", ticketNo).
		First(&ticket).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("ticket not found")
		}
		logger.Error("Failed to get ticket by number", logger.Error2("error", err), logger.String("ticket_no", ticketNo))
		return nil, fmt.Errorf("failed to get ticket: %w", err)
	}

	return &ticket, nil
}

// GetTickets gets tickets with filtering and pagination
func (s *TicketService) GetTickets(ctx context.Context, req *GetTicketsRequest) ([]*entities.Ticket, int64, error) {
	query := s.db.WithContext(ctx).Model(&entities.Ticket{}).
		Preload("User").
		Preload("AssignedTo").
		Preload("ResolvedBy")

	// Apply filters
	if req.UserID != 0 {
		query = query.Where("user_id = ?", req.UserID)
	}

	if req.AssignedToID != nil {
		query = query.Where("assigned_to_id = ?", *req.AssignedToID)
	}

	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}

	if req.Priority != "" {
		query = query.Where("priority = ?", req.Priority)
	}

	if req.Category != "" {
		query = query.Where("category = ?", req.Category)
	}

	if req.Search != "" {
		searchTerm := "%" + req.Search + "%"
		query = query.Where("title LIKE ? OR description LIKE ? OR ticket_no LIKE ?", searchTerm, searchTerm, searchTerm)
	}

	// Get total count
	var totalCount int64
	if err := query.Count(&totalCount).Error; err != nil {
		logger.Error("Failed to count tickets", logger.Error2("error", err))
		return nil, 0, fmt.Errorf("failed to count tickets: %w", err)
	}

	// Apply pagination and ordering
	query = query.Order("created_at DESC")

	if req.Limit > 0 {
		query = query.Limit(req.Limit)
	}

	if req.Offset > 0 {
		query = query.Offset(req.Offset)
	}

	var tickets []*entities.Ticket
	if err := query.Find(&tickets).Error; err != nil {
		logger.Error("Failed to get tickets", logger.Error2("error", err))
		return nil, 0, fmt.Errorf("failed to get tickets: %w", err)
	}

	return tickets, totalCount, nil
}

// UpdateTicket updates a ticket
func (s *TicketService) UpdateTicket(ctx context.Context, ticketID uint, req *UpdateTicketRequest) (*entities.Ticket, error) {
	// Get existing ticket
	ticket, err := s.GetTicket(ctx, ticketID)
	if err != nil {
		return nil, err
	}

	// Prepare updates
	updates := make(map[string]interface{})

	if req.Title != nil {
		updates["title"] = *req.Title
	}

	if req.Description != nil {
		updates["description"] = *req.Description
	}

	if req.Category != nil {
		updates["category"] = *req.Category
	}

	if req.Priority != nil {
		updates["priority"] = *req.Priority
	}

	if req.Status != nil {
		updates["status"] = *req.Status

		// Update timing fields based on status
		now := time.Now()
		switch *req.Status {
		case entities.TicketStatusResolved:
			updates["resolved_at"] = &now
		case entities.TicketStatusClosed:
			updates["closed_at"] = &now
		}
	}

	if req.Tags != nil {
		updates["tags"] = *req.Tags
	}

	if req.Metadata != nil {
		updates["metadata"] = *req.Metadata
	}

	// Update the ticket
	if err := s.db.WithContext(ctx).Model(ticket).Updates(updates).Error; err != nil {
		logger.Error("Failed to update ticket", logger.Error2("error", err), logger.Uint("ticket_id", ticketID))
		return nil, fmt.Errorf("failed to update ticket: %w", err)
	}

	// Reload the ticket with relations
	updatedTicket, err := s.GetTicket(ctx, ticketID)
	if err != nil {
		return nil, err
	}

	logger.Info("Ticket updated successfully", logger.Uint("ticket_id", ticketID))

	return updatedTicket, nil
}

// AssignTicket assigns a ticket to an admin
func (s *TicketService) AssignTicket(ctx context.Context, ticketID uint, req *AssignTicketRequest) (*entities.Ticket, error) {
	// Get existing ticket
	ticket, err := s.GetTicket(ctx, ticketID)
	if err != nil {
		return nil, err
	}

	// TODO: Verify that the assigned user is an admin through user service interface
	// For now, we'll just validate that the ID is provided
	if req.AssignedToID == 0 {
		return nil, fmt.Errorf("valid assigned user ID is required")
	}

	// Update assignment
	now := time.Now()
	updates := map[string]interface{}{
		"assigned_to_id": req.AssignedToID,
		"assigned_at":    &now,
	}

	// Update status to in_progress if it's currently open
	if ticket.Status == entities.TicketStatusOpen {
		updates["status"] = entities.TicketStatusInProgress
	}

	if err := s.db.WithContext(ctx).Model(ticket).Updates(updates).Error; err != nil {
		logger.Error("Failed to assign ticket", logger.Error2("error", err), logger.Uint("ticket_id", ticketID))
		return nil, fmt.Errorf("failed to assign ticket: %w", err)
	}

	// Reload the ticket with relations
	updatedTicket, err := s.GetTicket(ctx, ticketID)
	if err != nil {
		return nil, err
	}

	logger.Info("Ticket assigned successfully",
		logger.Uint("ticket_id", ticketID),
		logger.Uint("assigned_to_id", req.AssignedToID))

	return updatedTicket, nil
}

// ResolveTicket resolves a ticket
func (s *TicketService) ResolveTicket(ctx context.Context, ticketID uint, resolvedByID uint, req *ResolveTicketRequest) (*entities.Ticket, error) {
	// Get existing ticket
	ticket, err := s.GetTicket(ctx, ticketID)
	if err != nil {
		return nil, err
	}

	// Update resolution
	now := time.Now()
	updates := map[string]interface{}{
		"status":         entities.TicketStatusResolved,
		"resolved_by_id": resolvedByID,
		"resolved_at":    &now,
		"resolution":     req.Resolution,
	}

	if err := s.db.WithContext(ctx).Model(ticket).Updates(updates).Error; err != nil {
		logger.Error("Failed to resolve ticket", logger.Error2("error", err), logger.Uint("ticket_id", ticketID))
		return nil, fmt.Errorf("failed to resolve ticket: %w", err)
	}

	// Reload the ticket with relations
	updatedTicket, err := s.GetTicket(ctx, ticketID)
	if err != nil {
		return nil, err
	}

	logger.Info("Ticket resolved successfully",
		logger.Uint("ticket_id", ticketID),
		logger.Uint("resolved_by_id", resolvedByID))

	return updatedTicket, nil
}

// CloseTicket closes a ticket
func (s *TicketService) CloseTicket(ctx context.Context, ticketID uint) (*entities.Ticket, error) {
	// Get existing ticket
	ticket, err := s.GetTicket(ctx, ticketID)
	if err != nil {
		return nil, err
	}

	// Update status to closed
	now := time.Now()
	updates := map[string]interface{}{
		"status":    entities.TicketStatusClosed,
		"closed_at": &now,
	}

	if err := s.db.WithContext(ctx).Model(ticket).Updates(updates).Error; err != nil {
		logger.Error("Failed to close ticket", logger.Error2("error", err), logger.Uint("ticket_id", ticketID))
		return nil, fmt.Errorf("failed to close ticket: %w", err)
	}

	// Reload the ticket with relations
	updatedTicket, err := s.GetTicket(ctx, ticketID)
	if err != nil {
		return nil, err
	}

	logger.Info("Ticket closed successfully", logger.Uint("ticket_id", ticketID))

	return updatedTicket, nil
}

// DeleteTicket soft deletes a ticket
func (s *TicketService) DeleteTicket(ctx context.Context, ticketID uint) error {
	// Check if ticket exists
	var ticket entities.Ticket
	if err := s.db.WithContext(ctx).First(&ticket, ticketID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("ticket not found")
		}
		return fmt.Errorf("failed to check ticket existence: %w", err)
	}

	// Soft delete the ticket
	if err := s.db.WithContext(ctx).Delete(&ticket).Error; err != nil {
		logger.Error("Failed to delete ticket", logger.Error2("error", err), logger.Uint("ticket_id", ticketID))
		return fmt.Errorf("failed to delete ticket: %w", err)
	}

	logger.Info("Ticket deleted successfully", logger.Uint("ticket_id", ticketID))

	return nil
}

// GetTicketStats gets ticket statistics
func (s *TicketService) GetTicketStats(ctx context.Context) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Count tickets by status
	var statusStats []struct {
		Status string
		Count  int64
	}

	if err := s.db.WithContext(ctx).Model(&entities.Ticket{}).
		Select("status, count(*) as count").
		Group("status").
		Find(&statusStats).Error; err != nil {
		return nil, fmt.Errorf("failed to get status statistics: %w", err)
	}

	statusMap := make(map[string]int64)
	for _, stat := range statusStats {
		statusMap[stat.Status] = stat.Count
	}
	stats["by_status"] = statusMap

	// Count tickets by priority
	var priorityStats []struct {
		Priority string
		Count    int64
	}

	if err := s.db.WithContext(ctx).Model(&entities.Ticket{}).
		Select("priority, count(*) as count").
		Group("priority").
		Find(&priorityStats).Error; err != nil {
		return nil, fmt.Errorf("failed to get priority statistics: %w", err)
	}

	priorityMap := make(map[string]int64)
	for _, stat := range priorityStats {
		priorityMap[stat.Priority] = stat.Count
	}
	stats["by_priority"] = priorityMap

	// Count tickets by category
	var categoryStats []struct {
		Category string
		Count    int64
	}

	if err := s.db.WithContext(ctx).Model(&entities.Ticket{}).
		Select("category, count(*) as count").
		Group("category").
		Find(&categoryStats).Error; err != nil {
		return nil, fmt.Errorf("failed to get category statistics: %w", err)
	}

	categoryMap := make(map[string]int64)
	for _, stat := range categoryStats {
		categoryMap[stat.Category] = stat.Count
	}
	stats["by_category"] = categoryMap

	// Get total count
	var totalCount int64
	if err := s.db.WithContext(ctx).Model(&entities.Ticket{}).Count(&totalCount).Error; err != nil {
		return nil, fmt.Errorf("failed to get total count: %w", err)
	}
	stats["total"] = totalCount

	// Get unassigned count
	var unassignedCount int64
	if err := s.db.WithContext(ctx).Model(&entities.Ticket{}).
		Where("assigned_to_id IS NULL").
		Count(&unassignedCount).Error; err != nil {
		return nil, fmt.Errorf("failed to get unassigned count: %w", err)
	}
	stats["unassigned"] = unassignedCount

	return stats, nil
}

// generateTicketNumber generates a unique ticket number
func (s *TicketService) generateTicketNumber() string {
	// Generate format: TK-YYYYMMDD-XXXXXX (e.g., TK-20240718-ABC123)
	now := time.Now()
	dateStr := now.Format("20060102")

	// Generate 6-character random string
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 6)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}

	return fmt.Sprintf("TK-%s-%s", dateStr, string(b))
}
