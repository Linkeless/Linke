package service

import (
	"context"

	"linke/internal/ticket/domain/model"
	"linke/internal/ticket/domain/repository"
	"linke/internal/ticket/service/command"
	"linke/internal/ticket/service/query"
)

// TicketApplicationService orchestrates ticket operations
type TicketApplicationService struct {
	commandHandler *command.TicketCommandHandler
	queryHandler   *query.TicketQueryHandler
}

// NewTicketApplicationService creates a new ticket application service
func NewTicketApplicationService(
	commandHandler *command.TicketCommandHandler,
	queryHandler *query.TicketQueryHandler,
) *TicketApplicationService {
	return &TicketApplicationService{
		commandHandler: commandHandler,
		queryHandler:   queryHandler,
	}
}

// Commands

// CreateTicket creates a new ticket
func (s *TicketApplicationService) CreateTicket(ctx context.Context, cmd command.CreateTicketCommand) (*model.Ticket, error) {
	return s.commandHandler.CreateTicket(ctx, cmd)
}

// UpdateTicket updates an existing ticket
func (s *TicketApplicationService) UpdateTicket(ctx context.Context, cmd command.UpdateTicketCommand) (*model.Ticket, error) {
	return s.commandHandler.UpdateTicket(ctx, cmd)
}

// AssignTicket assigns a ticket to an admin
func (s *TicketApplicationService) AssignTicket(ctx context.Context, cmd command.AssignTicketCommand) (*model.Ticket, error) {
	return s.commandHandler.AssignTicket(ctx, cmd)
}

// ChangeTicketStatus changes the status of a ticket
func (s *TicketApplicationService) ChangeTicketStatus(ctx context.Context, cmd command.ChangeTicketStatusCommand) (*model.Ticket, error) {
	return s.commandHandler.ChangeTicketStatus(ctx, cmd)
}

// ResolveTicket resolves a ticket
func (s *TicketApplicationService) ResolveTicket(ctx context.Context, cmd command.ResolveTicketCommand) (*model.Ticket, error) {
	return s.commandHandler.ResolveTicket(ctx, cmd)
}

// CloseTicket closes a ticket
func (s *TicketApplicationService) CloseTicket(ctx context.Context, cmd command.CloseTicketCommand) (*model.Ticket, error) {
	return s.commandHandler.CloseTicket(ctx, cmd)
}

// AddTicketMessage adds a message to a ticket
func (s *TicketApplicationService) AddTicketMessage(ctx context.Context, cmd command.AddTicketMessageCommand) (*model.Ticket, error) {
	return s.commandHandler.AddTicketMessage(ctx, cmd)
}

// Queries

// GetTicket retrieves a ticket by ID
func (s *TicketApplicationService) GetTicket(ctx context.Context, query query.GetTicketQuery) (*model.Ticket, error) {
	return s.queryHandler.GetTicket(ctx, query)
}

// GetTicketByNumber retrieves a ticket by number
func (s *TicketApplicationService) GetTicketByNumber(ctx context.Context, query query.GetTicketByNumberQuery) (*model.Ticket, error) {
	return s.queryHandler.GetTicketByNumber(ctx, query)
}

// ListTickets lists tickets based on criteria
func (s *TicketApplicationService) ListTickets(ctx context.Context, query query.ListTicketsQuery) ([]*model.Ticket, int64, error) {
	return s.queryHandler.ListTickets(ctx, query)
}

// GetTicketMessages retrieves messages for a ticket
func (s *TicketApplicationService) GetTicketMessages(ctx context.Context, query query.GetTicketMessagesQuery) ([]*model.TicketMessage, int64, error) {
	return s.queryHandler.GetTicketMessages(ctx, query)
}

// GetTicketStatistics retrieves ticket statistics
func (s *TicketApplicationService) GetTicketStatistics(ctx context.Context, query query.GetTicketStatisticsQuery) (*repository.TicketStatistics, error) {
	return s.queryHandler.GetTicketStatistics(ctx, query)
}