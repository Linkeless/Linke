package service

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"linke/internal/ticket/domain/model"
	"linke/internal/ticket/domain/repository"
	"linke/internal/ticket/domain/valueobject"
	sharedvo "linke/internal/shared/valueobject"
)

// TicketDomainService provides domain services for tickets
type TicketDomainService struct {
	ticketRepo repository.TicketRepository
}

// NewTicketDomainService creates a new ticket domain service
func NewTicketDomainService(ticketRepo repository.TicketRepository) *TicketDomainService {
	return &TicketDomainService{
		ticketRepo: ticketRepo,
	}
}

// GenerateUniqueTicketNumber generates a unique ticket number
func (s *TicketDomainService) GenerateUniqueTicketNumber(ctx context.Context) (valueobject.TicketNumber, error) {
	maxAttempts := 10
	
	for attempt := 0; attempt < maxAttempts; attempt++ {
		// Generate a ticket number using the same logic as in the aggregate
		now := time.Now()
		dateStr := now.Format("20060102")
		
		// Generate 6-character random string
		const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
		b := make([]byte, 6)
		for i := range b {
			b[i] = charset[rand.Intn(len(charset))]
		}
		
		ticketNumberStr := fmt.Sprintf("TK-%s-%s", dateStr, string(b))
		ticketNumber, err := valueobject.NewTicketNumber(ticketNumberStr)
		if err != nil {
			continue // Try again with a new number
		}
		
		// Check if the number already exists
		exists, err := s.ticketRepo.ExistsByTicketNumber(ctx, ticketNumber)
		if err != nil {
			return valueobject.TicketNumber{}, fmt.Errorf("failed to check ticket number uniqueness: %w", err)
		}
		
		if !exists {
			return ticketNumber, nil
		}
	}
	
	return valueobject.TicketNumber{}, fmt.Errorf("failed to generate unique ticket number after %d attempts", maxAttempts)
}

// CanUserAccessTicket checks if a user can access a specific ticket
func (s *TicketDomainService) CanUserAccessTicket(
	ctx context.Context,
	userID sharedvo.UserID,
	ticketID valueobject.TicketID,
	isAdmin bool,
) (bool, error) {
	ticket, err := s.ticketRepo.FindByID(ctx, ticketID)
	if err != nil {
		return false, fmt.Errorf("failed to find ticket: %w", err)
	}
	
	// Admins can access all tickets
	if isAdmin {
		return true, nil
	}
	
	// Users can only access their own tickets
	return ticket.UserID().Equals(userID), nil
}

// ShouldEscalatePriority determines if a ticket's priority should be escalated based on age and category
func (s *TicketDomainService) ShouldEscalatePriority(ticket *model.Ticket) (bool, valueobject.TicketPriority) {
	if ticket.IsResolved() || ticket.IsClosed() {
		return false, ticket.Priority()
	}
	
	now := time.Now()
	age := now.Sub(ticket.CreatedAt())
	
	// Define escalation rules based on category and age
	switch {
	case ticket.Category().IsBug() && ticket.Priority().IsNormal() && age > 24*time.Hour:
		return true, valueobject.MustTicketPriority(valueobject.TicketPriorityHigh)
		
	case ticket.Category().IsTechnical() && ticket.Priority().IsLow() && age > 72*time.Hour:
		return true, valueobject.MustTicketPriority(valueobject.TicketPriorityNormal)
		
	case ticket.Category().RequiresFinancialAccess() && ticket.Priority().IsNormal() && age > 48*time.Hour:
		return true, valueobject.MustTicketPriority(valueobject.TicketPriorityHigh)
		
	case !ticket.HasFirstResponse() && age > 24*time.Hour:
		// Escalate if no first response within 24 hours
		if ticket.Priority().IsLow() {
			return true, valueobject.MustTicketPriority(valueobject.TicketPriorityNormal)
		} else if ticket.Priority().IsNormal() {
			return true, valueobject.MustTicketPriority(valueobject.TicketPriorityHigh)
		}
	}
	
	return false, ticket.Priority()
}

// GetSLADeadline calculates the SLA deadline for a ticket based on priority
func (s *TicketDomainService) GetSLADeadline(ticket *model.Ticket) time.Time {
	createdAt := ticket.CreatedAt()
	
	switch {
	case ticket.Priority().IsCritical():
		return createdAt.Add(2 * time.Hour) // 2 hours for critical
	case ticket.Priority().IsUrgent():
		return createdAt.Add(4 * time.Hour) // 4 hours for urgent
	case ticket.Priority().IsHigh():
		return createdAt.Add(8 * time.Hour) // 8 hours for high
	case ticket.Priority().IsNormal():
		return createdAt.Add(24 * time.Hour) // 24 hours for normal
	default:
		return createdAt.Add(72 * time.Hour) // 72 hours for low
	}
}

// IsOverdue checks if a ticket is overdue based on SLA
func (s *TicketDomainService) IsOverdue(ticket *model.Ticket) bool {
	if ticket.IsResolved() || ticket.IsClosed() {
		return false
	}
	
	// If ticket has first response, it's not considered overdue
	if ticket.HasFirstResponse() {
		return false
	}
	
	deadline := s.GetSLADeadline(ticket)
	return time.Now().After(deadline)
}

// GetRecommendedAssignee suggests the best assignee for a ticket based on category and workload
func (s *TicketDomainService) GetRecommendedAssignee(
	ctx context.Context,
	ticket *model.Ticket,
	availableAdmins []sharedvo.UserID,
) (*sharedvo.UserID, error) {
	if len(availableAdmins) == 0 {
		return nil, nil
	}
	
	// For now, use a simple round-robin approach
	// In a real system, you might consider:
	// - Admin specialization (technical vs billing)
	// - Current workload
	// - Past performance with similar tickets
	
	workloads := make(map[sharedvo.UserID]int64)
	
	// Get current workload for each admin
	for _, adminID := range availableAdmins {
		tickets, _, err := s.ticketRepo.FindByAssignedTo(ctx, adminID, 1000, 0)
		if err != nil {
			continue // Skip this admin if we can't get their workload
		}
		
		// Count only active tickets
		activeCount := int64(0)
		for _, t := range tickets {
			if t.IsActive() {
				activeCount++
			}
		}
		workloads[adminID] = activeCount
	}
	
	// Find admin with lowest workload
	var recommendedAdmin *sharedvo.UserID
	minWorkload := int64(-1)
	
	for _, adminID := range availableAdmins {
		workload := workloads[adminID]
		if minWorkload == -1 || workload < minWorkload {
			minWorkload = workload
			recommendedAdmin = &adminID
		}
	}
	
	return recommendedAdmin, nil
}

// ValidateTicketTransition validates if a ticket status transition is valid
func (s *TicketDomainService) ValidateTicketTransition(
	ticket *model.Ticket,
	newStatus valueobject.TicketStatus,
	userID sharedvo.UserID,
	isAdmin bool,
) error {
	// Basic status transition validation
	if !ticket.Status().CanTransitionTo(newStatus) {
		return fmt.Errorf("cannot transition from %s to %s", ticket.Status().String(), newStatus.String())
	}
	
	// Business rule: Only ticket owner or admins can close tickets
	if newStatus.IsClosed() && !isAdmin && !ticket.UserID().Equals(userID) {
		return fmt.Errorf("only ticket owner or admins can close tickets")
	}
	
	// Business rule: Only admins can resolve tickets
	if newStatus.IsResolved() && !isAdmin {
		return fmt.Errorf("only admins can resolve tickets")
	}
	
	// Business rule: Only admins can assign tickets (move to in_progress)
	if newStatus.IsInProgress() && !isAdmin {
		return fmt.Errorf("only admins can assign tickets")
	}
	
	return nil
}