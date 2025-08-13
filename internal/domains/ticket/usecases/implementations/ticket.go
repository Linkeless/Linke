package implementations

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"linke/internal/domains/ticket/constants"
	"linke/internal/domains/ticket/dto"
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

// CreateTicket creates a new ticket
func (s *TicketService) CreateTicket(ctx context.Context, userID uint, req *dto.CreateTicketRequest) (*entities.Ticket, error) {
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
		priority = constants.TicketPriorityNormal
	}

	// Create the ticket
	ticket := &entities.Ticket{
		TicketNo:    ticketNo,
		Title:       req.Title,
		Description: req.Description,
		Category:    req.Category,
		Priority:    priority,
		Status:      constants.TicketStatusOpen,
		UserID:      userID,
		Tags:        &req.Tags,
		Metadata:    &req.Metadata,
	}

	if err := s.db.WithContext(ctx).Create(ticket).Error; err != nil {
		logger.Error("Failed to create ticket", logger.ErrorField(err))
		return nil, fmt.Errorf("failed to create ticket: %w", err)
	}

	logger.Info("Ticket created successfully",
		logger.String("ticket_no", ticket.TicketNo),
		logger.Uint("user_id", userID),
		logger.String("category", ticket.Category))

	// Attempt auto-assignment for new tickets (non-blocking)
	go func() {
		// Use background context for async operation
		bgCtx := context.Background()
		if _, err := s.AutoAssignTicket(bgCtx, ticket.ID); err != nil {
			logger.Warn("Auto-assignment failed for new ticket",
				logger.Uint("ticket_id", ticket.ID),
				logger.String("ticket_no", ticket.TicketNo),
				logger.ErrorField(err))
		}
	}()

	return ticket, nil
}

// GetTicket gets a ticket by ID with relations
func (s *TicketService) GetTicket(ctx context.Context, ticketID uint) (*entities.Ticket, error) {
	var ticket entities.Ticket

	if err := s.db.WithContext(ctx).
		Preload("Messages", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at ASC")
		}).
		First(&ticket, ticketID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("ticket not found")
		}
		logger.Error("Failed to get ticket", logger.Uint("ticketID", uint(ticketID)))
		return nil, fmt.Errorf("failed to get ticket: %w", err)
	}

	return &ticket, nil
}

// GetTicketByNumber gets a ticket by ticket number
func (s *TicketService) GetTicketByNumber(ctx context.Context, ticketNo string) (*entities.Ticket, error) {
	var ticket entities.Ticket

	if err := s.db.WithContext(ctx).
		Preload("Messages", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at ASC")
		}).
		Where("ticket_no = ?", ticketNo).
		First(&ticket).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("ticket not found")
		}
		logger.Error("Failed to get ticket by number", logger.String("ticket_no", ticketNo), logger.ErrorField(err))
		return nil, fmt.Errorf("failed to get ticket: %w", err)
	}

	return &ticket, nil
}

// GetTickets gets tickets with filtering and pagination
func (s *TicketService) GetTickets(ctx context.Context, req *dto.GetTicketsRequest) ([]*entities.Ticket, int64, error) {
	query := s.db.WithContext(ctx).Model(&entities.Ticket{})

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
		logger.Error("Failed to count tickets", logger.ErrorField(err))
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
		logger.Error("Failed to get tickets", logger.ErrorField(err))
		return nil, 0, fmt.Errorf("failed to get tickets: %w", err)
	}

	return tickets, totalCount, nil
}

// UpdateTicket updates a ticket
func (s *TicketService) UpdateTicket(ctx context.Context, ticketID uint, req *dto.UpdateTicketRequest) (*entities.Ticket, error) {
	// Get existing ticket
	ticket, err := s.GetTicket(ctx, ticketID)
	if err != nil {
		return nil, err
	}

	// Prepare updates
	updates := make(map[string]any)

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
		case constants.TicketStatusResolved:
			updates["resolved_at"] = &now
		case constants.TicketStatusClosed:
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
		logger.Error("Failed to update ticket", logger.Uint("ticketID", uint(ticketID)))
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
func (s *TicketService) AssignTicket(ctx context.Context, ticketID uint, req *dto.AssignTicketRequest) (*entities.Ticket, error) {
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
	updates := map[string]any{
		"assigned_to_id": req.AssignedToID,
		"assigned_at":    &now,
	}

	// Update status to in_progress if it's currently open
	if ticket.Status == constants.TicketStatusOpen {
		updates["status"] = constants.TicketStatusInProgress
	}

	if err := s.db.WithContext(ctx).Model(ticket).Updates(updates).Error; err != nil {
		logger.Error("Failed to assign ticket", logger.Uint("ticketID", uint(ticketID)))
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
func (s *TicketService) ResolveTicket(ctx context.Context, ticketID uint, resolvedByID uint, req *dto.ResolveTicketRequest) (*entities.Ticket, error) {
	// Get existing ticket
	ticket, err := s.GetTicket(ctx, ticketID)
	if err != nil {
		return nil, err
	}

	// Update resolution
	now := time.Now()
	updates := map[string]any{
		"status":         constants.TicketStatusResolved,
		"resolved_by_id": resolvedByID,
		"resolved_at":    &now,
		"resolution":     req.Resolution,
	}

	if err := s.db.WithContext(ctx).Model(ticket).Updates(updates).Error; err != nil {
		logger.Error("Failed to resolve ticket", logger.Uint("ticketID", uint(ticketID)))
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
func (s *TicketService) CloseTicket(ctx context.Context, ticketID uint, reason string) (*entities.Ticket, error) {
	// Get existing ticket
	ticket, err := s.GetTicket(ctx, ticketID)
	if err != nil {
		return nil, err
	}

	// Update status to closed
	now := time.Now()
	updates := map[string]any{
		"status":    constants.TicketStatusClosed,
		"closed_at": &now,
	}

	if err := s.db.WithContext(ctx).Model(ticket).Updates(updates).Error; err != nil {
		logger.Error("Failed to close ticket", logger.Uint("ticketID", uint(ticketID)))
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
		logger.Error("Failed to delete ticket", logger.Uint("ticketID", uint(ticketID)))
		return fmt.Errorf("failed to delete ticket: %w", err)
	}

	logger.Info("Ticket deleted successfully", logger.Uint("ticket_id", ticketID))

	return nil
}

// GetUserTickets gets tickets for a specific user
func (s *TicketService) GetUserTickets(ctx context.Context, userID uint, limit, offset int) ([]*entities.Ticket, int64, error) {
	req := &dto.GetTicketsRequest{
		UserID: userID,
		Limit:  limit,
		Offset: offset,
	}
	return s.GetTickets(ctx, req)
}

// GetAssignedTickets gets tickets assigned to a specific agent
func (s *TicketService) GetAssignedTickets(ctx context.Context, assignedToID uint, limit, offset int) ([]*entities.Ticket, int64, error) {
	req := &dto.GetTicketsRequest{
		AssignedToID: &assignedToID,
		Limit:        limit,
		Offset:       offset,
	}
	return s.GetTickets(ctx, req)
}

// UnassignTicket removes assignment from a ticket
func (s *TicketService) UnassignTicket(ctx context.Context, ticketID uint) (*entities.Ticket, error) {
	// Get existing ticket
	ticket, err := s.GetTicket(ctx, ticketID)
	if err != nil {
		return nil, err
	}

	// Remove assignment
	updates := map[string]any{
		"assigned_to_id": nil,
		"assigned_at":    nil,
	}

	// Update status back to open if it was in_progress
	if ticket.Status == constants.TicketStatusInProgress {
		updates["status"] = constants.TicketStatusOpen
	}

	if err := s.db.WithContext(ctx).Model(ticket).Updates(updates).Error; err != nil {
		logger.Error("Failed to unassign ticket", logger.Uint("ticketID", uint(ticketID)))
		return nil, fmt.Errorf("failed to unassign ticket: %w", err)
	}

	// Reload the ticket with relations
	updatedTicket, err := s.GetTicket(ctx, ticketID)
	if err != nil {
		return nil, err
	}

	logger.Info("Ticket unassigned successfully", logger.Uint("ticket_id", ticketID))

	return updatedTicket, nil
}

// UpdateTicketStatus updates a ticket's status
func (s *TicketService) UpdateTicketStatus(ctx context.Context, ticketID uint, status string) (*entities.Ticket, error) {
	req := &dto.UpdateTicketRequest{
		Status: &status,
	}
	return s.UpdateTicket(ctx, ticketID, req)
}

// UpdateTicketPriority updates a ticket's priority
func (s *TicketService) UpdateTicketPriority(ctx context.Context, ticketID uint, priority string) (*entities.Ticket, error) {
	req := &dto.UpdateTicketRequest{
		Priority: &priority,
	}
	return s.UpdateTicket(ctx, ticketID, req)
}

// ReopenTicket reopens a closed ticket
func (s *TicketService) ReopenTicket(ctx context.Context, ticketID uint, reason string) (*entities.Ticket, error) {
	// Get existing ticket
	ticket, err := s.GetTicket(ctx, ticketID)
	if err != nil {
		return nil, err
	}

	// Update status to open and clear resolved fields
	updates := map[string]any{
		"status":         constants.TicketStatusOpen,
		"resolved_by_id": nil,
		"resolved_at":    nil,
		"resolution":     "",
		"closed_at":      nil,
	}

	if err := s.db.WithContext(ctx).Model(ticket).Updates(updates).Error; err != nil {
		logger.Error("Failed to reopen ticket", logger.Uint("ticketID", uint(ticketID)))
		return nil, fmt.Errorf("failed to reopen ticket: %w", err)
	}

	// Reload the ticket with relations
	updatedTicket, err := s.GetTicket(ctx, ticketID)
	if err != nil {
		return nil, err
	}

	logger.Info("Ticket reopened successfully",
		logger.Uint("ticket_id", ticketID),
		logger.String("reason", reason))

	return updatedTicket, nil
}

// GetTicketStatistics gets ticket statistics for a date range
func (s *TicketService) GetTicketStatistics(ctx context.Context, fromDate, toDate string) (map[string]any, error) {
	stats := make(map[string]any)

	query := s.db.WithContext(ctx).Model(&entities.Ticket{})

	// Apply date filters if provided
	if fromDate != "" {
		query = query.Where("created_at >= ?", fromDate)
	}
	if toDate != "" {
		query = query.Where("created_at <= ?", toDate)
	}

	// Count tickets by status
	var statusStats []struct {
		Status string
		Count  int64
	}

	if err := query.Select("status, count(*) as count").Group("status").Find(&statusStats).Error; err != nil {
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

	if err := query.Select("priority, count(*) as count").Group("priority").Find(&priorityStats).Error; err != nil {
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

	if err := query.Select("category, count(*) as count").Group("category").Find(&categoryStats).Error; err != nil {
		return nil, fmt.Errorf("failed to get category statistics: %w", err)
	}

	categoryMap := make(map[string]int64)
	for _, stat := range categoryStats {
		categoryMap[stat.Category] = stat.Count
	}
	stats["by_category"] = categoryMap

	// Get total count
	var totalCount int64
	if err := query.Count(&totalCount).Error; err != nil {
		return nil, fmt.Errorf("failed to get total count: %w", err)
	}
	stats["total"] = totalCount

	// Get unassigned count
	var unassignedCount int64
	if err := query.Where("assigned_to_id IS NULL").Count(&unassignedCount).Error; err != nil {
		return nil, fmt.Errorf("failed to get unassigned count: %w", err)
	}
	stats["unassigned"] = unassignedCount

	return stats, nil
}

// GetUserTicketStatistics gets ticket statistics for a specific user
func (s *TicketService) GetUserTicketStatistics(ctx context.Context, userID uint) (map[string]any, error) {
	stats := make(map[string]any)

	// Count tickets by status for this user
	var statusStats []struct {
		Status string
		Count  int64
	}

	if err := s.db.WithContext(ctx).Model(&entities.Ticket{}).
		Where("user_id = ?", userID).
		Select("status, count(*) as count").
		Group("status").
		Find(&statusStats).Error; err != nil {
		return nil, fmt.Errorf("failed to get user status statistics: %w", err)
	}

	statusMap := make(map[string]int64)
	for _, stat := range statusStats {
		statusMap[stat.Status] = stat.Count
	}
	stats["by_status"] = statusMap

	// Get total count for user
	var totalCount int64
	if err := s.db.WithContext(ctx).Model(&entities.Ticket{}).
		Where("user_id = ?", userID).
		Count(&totalCount).Error; err != nil {
		return nil, fmt.Errorf("failed to get user total count: %w", err)
	}
	stats["total"] = totalCount

	return stats, nil
}

// GetAgentTicketStatistics gets ticket statistics for a specific agent
func (s *TicketService) GetAgentTicketStatistics(ctx context.Context, agentID uint) (map[string]any, error) {
	stats := make(map[string]any)

	// Count tickets by status assigned to this agent
	var statusStats []struct {
		Status string
		Count  int64
	}

	if err := s.db.WithContext(ctx).Model(&entities.Ticket{}).
		Where("assigned_to_id = ?", agentID).
		Select("status, count(*) as count").
		Group("status").
		Find(&statusStats).Error; err != nil {
		return nil, fmt.Errorf("failed to get agent status statistics: %w", err)
	}

	statusMap := make(map[string]int64)
	for _, stat := range statusStats {
		statusMap[stat.Status] = stat.Count
	}
	stats["by_status"] = statusMap

	// Get total count for agent
	var totalAssigned int64
	if err := s.db.WithContext(ctx).Model(&entities.Ticket{}).
		Where("assigned_to_id = ?", agentID).
		Count(&totalAssigned).Error; err != nil {
		return nil, fmt.Errorf("failed to get agent total count: %w", err)
	}
	stats["total_assigned"] = totalAssigned

	// Get resolved count for agent
	var resolvedCount int64
	if err := s.db.WithContext(ctx).Model(&entities.Ticket{}).
		Where("resolved_by_id = ?", agentID).
		Count(&resolvedCount).Error; err != nil {
		return nil, fmt.Errorf("failed to get agent resolved count: %w", err)
	}
	stats["total_resolved"] = resolvedCount

	return stats, nil
}

// BulkAssignTickets assigns multiple tickets to an agent
func (s *TicketService) BulkAssignTickets(ctx context.Context, ticketIDs []uint, assignedToID uint) error {
	if len(ticketIDs) == 0 {
		return fmt.Errorf("no ticket IDs provided")
	}

	now := time.Now()
	updates := map[string]any{
		"assigned_to_id": assignedToID,
		"assigned_at":    &now,
	}

	// Update tickets that are currently open to in_progress
	if err := s.db.WithContext(ctx).Model(&entities.Ticket{}).
		Where("id IN ? AND status = ?", ticketIDs, constants.TicketStatusOpen).
		Update("status", constants.TicketStatusInProgress).Error; err != nil {
		logger.Error("Failed to update ticket status during bulk assignment", logger.ErrorField(err))
		return fmt.Errorf("failed to update ticket status: %w", err)
	}

	// Update all tickets with assignment
	if err := s.db.WithContext(ctx).Model(&entities.Ticket{}).
		Where("id IN ?", ticketIDs).
		Updates(updates).Error; err != nil {
		logger.Error("Failed to bulk assign tickets", logger.ErrorField(err))
		return fmt.Errorf("failed to bulk assign tickets: %w", err)
	}

	logger.Info("Bulk assigned tickets successfully",
		logger.Any("ticket_ids", ticketIDs),
		logger.Uint("assigned_to_id", assignedToID))

	return nil
}

// BulkUpdateTicketStatus updates the status of multiple tickets
func (s *TicketService) BulkUpdateTicketStatus(ctx context.Context, ticketIDs []uint, status string) error {
	if len(ticketIDs) == 0 {
		return fmt.Errorf("no ticket IDs provided")
	}

	updates := map[string]any{
		"status": status,
	}

	// Set timing fields based on status
	now := time.Now()
	switch status {
	case constants.TicketStatusResolved:
		updates["resolved_at"] = &now
	case constants.TicketStatusClosed:
		updates["closed_at"] = &now
	}

	if err := s.db.WithContext(ctx).Model(&entities.Ticket{}).
		Where("id IN ?", ticketIDs).
		Updates(updates).Error; err != nil {
		logger.Error("Failed to bulk update ticket status", logger.ErrorField(err))
		return fmt.Errorf("failed to bulk update ticket status: %w", err)
	}

	logger.Info("Bulk updated ticket status successfully",
		logger.Any("ticket_ids", ticketIDs),
		logger.String("status", status))

	return nil
}

// generateTicketNumber generates a unique ticket number
func (s *TicketService) generateTicketNumber() string {
	// Generate format: TK-YYYYMMDD-XXXXXX (e.g., TK-20240718-ABC123)
	now := time.Now()
	dateStr := now.Format("20060102")

	// Generate random string using constant length
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, constants.TicketNumberLength)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}

	return fmt.Sprintf("%s-%s-%s", constants.TicketNumberPrefix, dateStr, string(b))
}

// AutoAssignTicket automatically assigns a ticket to the best available agent
func (s *TicketService) AutoAssignTicket(ctx context.Context, ticketID uint) (*entities.Ticket, error) {
	// Get the ticket
	var ticket entities.Ticket
	if err := s.db.WithContext(ctx).First(&ticket, ticketID).Error; err != nil {
		return nil, fmt.Errorf("ticket not found: %w", err)
	}

	// Check if ticket is already assigned
	if ticket.AssignedToID != nil {
		logger.Info("Ticket already assigned, skipping auto-assignment",
			logger.Uint("ticket_id", ticketID),
			logger.Uint("assigned_to", *ticket.AssignedToID))
		return &ticket, nil
	}

	// Find the best agent for this ticket
	bestAgentID, err := s.FindBestAgentForTicket(ctx, &ticket)
	if err != nil {
		logger.Error("Failed to find best agent for auto-assignment",
			logger.Uint("ticket_id", ticketID),
			logger.ErrorField(err))
		return &ticket, nil // Return original ticket without assignment
	}

	if bestAgentID == 0 {
		logger.Warn("No available agent found for auto-assignment", logger.Uint("ticket_id", ticketID))
		return &ticket, nil
	}

	// Assign the ticket to the best agent
	now := time.Now()
	updates := map[string]any{
		"assigned_to_id": bestAgentID,
		"assigned_at":    &now,
		"status":         constants.TicketStatusInProgress,
		"updated_at":     now,
	}

	if err := s.db.WithContext(ctx).Model(&ticket).Updates(updates).Error; err != nil {
		logger.Error("Failed to auto-assign ticket",
			logger.Uint("ticket_id", ticketID),
			logger.Uint("agent_id", bestAgentID),
			logger.ErrorField(err))
		return nil, fmt.Errorf("failed to auto-assign ticket: %w", err)
	}

	// Update the ticket object with new values
	ticket.AssignedToID = &bestAgentID
	ticket.AssignedAt = &now
	ticket.Status = constants.TicketStatusInProgress
	ticket.UpdatedAt = now

	logger.Info("Ticket auto-assigned successfully",
		logger.Uint("ticket_id", ticketID),
		logger.Uint("agent_id", bestAgentID),
		logger.String("category", ticket.Category),
		logger.String("priority", ticket.Priority))

	return &ticket, nil
}

// GetAgentWorkload returns the current workload (number of assigned tickets) for an agent
func (s *TicketService) GetAgentWorkload(ctx context.Context, agentID uint) (int, error) {
	var count int64
	if err := s.db.WithContext(ctx).Model(&entities.Ticket{}).
		Where("assigned_to_id = ? AND status IN ?", agentID,
			[]string{constants.TicketStatusOpen, constants.TicketStatusInProgress, constants.TicketStatusPending}).
		Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to get agent workload: %w", err)
	}
	return int(count), nil
}

// GetAvailableAgents returns a list of agents available for assignment in a specific category
func (s *TicketService) GetAvailableAgents(ctx context.Context, category string) ([]*dto.AgentInfo, error) {
	// Query for admin users (agents) - simplified approach
	// In a real implementation, you might have a separate agents table or role system
	var agents []struct {
		ID    uint   `json:"id"`
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	// Get all admin users as potential agents
	if err := s.db.WithContext(ctx).
		Table("users").
		Select("id, name, email").
		Where("role = ? AND status = ?", "admin", "active").
		Find(&agents).Error; err != nil {
		return nil, fmt.Errorf("failed to get available agents: %w", err)
	}

	var agentInfos []*dto.AgentInfo
	for _, agent := range agents {
		// Get current workload for this agent
		workload, err := s.GetAgentWorkload(ctx, agent.ID)
		if err != nil {
			logger.Warn("Failed to get workload for agent",
				logger.Uint("agent_id", agent.ID),
				logger.ErrorField(err))
			workload = 0
		}

		// Create agent info with defaults (in real implementation, these would come from agent profile)
		agentInfo := &dto.AgentInfo{
			UserID:            agent.ID,
			Name:              agent.Name,
			Email:             agent.Email,
			Specialties:       []string{"general", category}, // Simplified: assume all agents can handle any category
			CurrentLoad:       workload,
			MaxLoad:           10,   // Default max load
			IsOnline:          true, // Simplified: assume all agents are online
			LastActiveAt:      time.Now().Format(time.RFC3339),
			AvgResponseTime:   30,  // Default 30 minutes
			SatisfactionScore: 8.5, // Default score
		}

		// Only include agents who are not at max capacity
		if agentInfo.CurrentLoad < agentInfo.MaxLoad {
			agentInfos = append(agentInfos, agentInfo)
		}
	}

	return agentInfos, nil
}

// FindBestAgentForTicket finds the best agent to assign a ticket to based on workload and specialties
func (s *TicketService) FindBestAgentForTicket(ctx context.Context, ticket *entities.Ticket) (uint, error) {
	// Get available agents for this ticket category
	availableAgents, err := s.GetAvailableAgents(ctx, ticket.Category)
	if err != nil {
		return 0, fmt.Errorf("failed to get available agents: %w", err)
	}

	if len(availableAgents) == 0 {
		return 0, fmt.Errorf("no available agents found")
	}

	// Simple assignment algorithm: choose agent with lowest current workload
	// In a more sophisticated implementation, you could consider:
	// - Agent specialties matching ticket category
	// - Agent response time history
	// - Customer satisfaction scores
	// - Agent availability/online status
	// - Priority-based assignment for urgent tickets

	var bestAgent *dto.AgentInfo
	lowestWorkload := int(^uint(0) >> 1) // Max int

	for _, agent := range availableAgents {
		// Calculate assignment score (lower is better)
		score := agent.CurrentLoad

		// Bonus for category specialization
		hasSpecialty := false
		for _, specialty := range agent.Specialties {
			if specialty == ticket.Category {
				hasSpecialty = true
				break
			}
		}

		if hasSpecialty {
			score -= 1 // Reduce score (better) if agent specializes in this category
		}

		// For urgent tickets, prefer agents with better response times
		if ticket.Priority == constants.TicketPriorityUrgent || ticket.Priority == constants.TicketPriorityCritical {
			if agent.AvgResponseTime < 30 { // Less than 30 minutes
				score -= 1
			}
		}

		if score < lowestWorkload {
			lowestWorkload = score
			bestAgent = agent
		}
	}

	if bestAgent == nil {
		return 0, fmt.Errorf("could not determine best agent")
	}

	logger.Info("Selected best agent for ticket assignment",
		logger.Uint("ticket_id", ticket.ID),
		logger.Uint("agent_id", bestAgent.UserID),
		logger.String("agent_name", bestAgent.Name),
		logger.Int("agent_workload", bestAgent.CurrentLoad),
		logger.String("ticket_category", ticket.Category),
		logger.String("ticket_priority", ticket.Priority))

	return bestAgent.UserID, nil
}
