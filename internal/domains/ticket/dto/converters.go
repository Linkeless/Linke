package dto

import (
	"sync"
	"sync/atomic"

	"linke/internal/domains/ticket/entities"
)

// Pool statistics for monitoring
var (
	ticketResponsePoolHits   int64
	ticketResponsePoolMisses int64
	
	ticketUserResponsePoolHits   int64
	ticketUserResponsePoolMisses int64
	
	ticketMessageResponsePoolHits   int64
	ticketMessageResponsePoolMisses int64
	
	ticketMessageUserResponsePoolHits   int64
	ticketMessageUserResponsePoolMisses int64
)

// Object pools for performance optimization
var (
	ticketResponsePool = sync.Pool{
		New: func() any {
			atomic.AddInt64(&ticketResponsePoolMisses, 1)
			return &TicketResponse{}
		},
	}

	ticketUserResponsePool = sync.Pool{
		New: func() any {
			atomic.AddInt64(&ticketUserResponsePoolMisses, 1)
			return &TicketUserResponse{}
		},
	}

	ticketMessageResponsePool = sync.Pool{
		New: func() any {
			atomic.AddInt64(&ticketMessageResponsePoolMisses, 1)
			return &TicketMessageResponse{}
		},
	}

	ticketMessageUserResponsePool = sync.Pool{
		New: func() any {
			atomic.AddInt64(&ticketMessageUserResponsePoolMisses, 1)
			return &TicketMessageUserResponse{}
		},
	}
)

// PoolStats provides statistics about object pool usage
type PoolStats struct {
	TicketResponseHits   int64 `json:"ticket_response_hits"`
	TicketResponseMisses int64 `json:"ticket_response_misses"`
	
	TicketUserResponseHits   int64 `json:"ticket_user_response_hits"`
	TicketUserResponseMisses int64 `json:"ticket_user_response_misses"`
	
	TicketMessageResponseHits   int64 `json:"ticket_message_response_hits"`
	TicketMessageResponseMisses int64 `json:"ticket_message_response_misses"`
	
	TicketMessageUserResponseHits   int64 `json:"ticket_message_user_response_hits"`
	TicketMessageUserResponseMisses int64 `json:"ticket_message_user_response_misses"`
}

// GetPoolStats returns current pool statistics
func GetPoolStats() PoolStats {
	return PoolStats{
		TicketResponseHits:   atomic.LoadInt64(&ticketResponsePoolHits),
		TicketResponseMisses: atomic.LoadInt64(&ticketResponsePoolMisses),
		
		TicketUserResponseHits:   atomic.LoadInt64(&ticketUserResponsePoolHits),
		TicketUserResponseMisses: atomic.LoadInt64(&ticketUserResponsePoolMisses),
		
		TicketMessageResponseHits:   atomic.LoadInt64(&ticketMessageResponsePoolHits),
		TicketMessageResponseMisses: atomic.LoadInt64(&ticketMessageResponsePoolMisses),
		
		TicketMessageUserResponseHits:   atomic.LoadInt64(&ticketMessageUserResponsePoolHits),
		TicketMessageUserResponseMisses: atomic.LoadInt64(&ticketMessageUserResponsePoolMisses),
	}
}

// resetTicketResponse safely resets a TicketResponse object
func resetTicketResponse(resp *TicketResponse) {
	resp.ID = 0
	resp.TicketNo = ""
	resp.Title = ""
	resp.Description = ""
	resp.Category = ""
	resp.Priority = ""
	resp.Status = ""
	resp.UserID = 0
	resp.User = nil
	resp.AssignedToID = nil
	resp.AssignedTo = nil
	resp.AssignedAt = nil
	resp.ResolvedByID = nil
	resp.ResolvedBy = nil
	resp.ResolvedAt = nil
	resp.Resolution = ""
	resp.FirstResponseAt = nil
	resp.LastResponseAt = nil
	resp.ClosedAt = nil
	resp.Tags = nil
	resp.Metadata = nil
	// Preserve slice capacity but reset length
	if resp.Messages != nil {
		resp.Messages = resp.Messages[:0]
	}
}

// resetTicketUserResponse safely resets a TicketUserResponse object  
func resetTicketUserResponse(resp *TicketUserResponse) {
	resp.ID = 0
	resp.TicketNo = ""
	resp.Title = ""
	resp.Description = ""
	resp.Category = ""
	resp.Priority = ""
	resp.Status = ""
	resp.Resolution = ""
	resp.FirstResponseAt = nil
	resp.LastResponseAt = nil
	resp.ClosedAt = nil
	// Preserve slice capacity but reset length
	if resp.Messages != nil {
		resp.Messages = resp.Messages[:0]
	}
}

// resetTicketMessageResponse safely resets a TicketMessageResponse object
func resetTicketMessageResponse(resp *TicketMessageResponse) {
	resp.ID = 0
	resp.TicketID = 0
	resp.UserID = 0
	resp.User = nil
	resp.Content = ""
	resp.MessageType = ""
	resp.Attachments = nil
	resp.IsInternal = false
	resp.Metadata = nil
}

// resetTicketMessageUserResponse safely resets a TicketMessageUserResponse object
func resetTicketMessageUserResponse(resp *TicketMessageUserResponse) {
	resp.ID = 0
	resp.TicketID = 0
	resp.Content = ""
	resp.MessageType = ""
	resp.Attachments = nil
}

// GetTicketResponse retrieves a TicketResponse from the pool
func GetTicketResponse() *TicketResponse {
	atomic.AddInt64(&ticketResponsePoolHits, 1)
	return ticketResponsePool.Get().(*TicketResponse)
}

// PutTicketResponse returns a TicketResponse to the pool after safe reset
func PutTicketResponse(resp *TicketResponse) {
	if resp != nil {
		resetTicketResponse(resp)
		ticketResponsePool.Put(resp)
	}
}

// GetTicketUserResponse retrieves a TicketUserResponse from the pool
func GetTicketUserResponse() *TicketUserResponse {
	atomic.AddInt64(&ticketUserResponsePoolHits, 1)
	return ticketUserResponsePool.Get().(*TicketUserResponse)
}

// PutTicketUserResponse returns a TicketUserResponse to the pool after safe reset
func PutTicketUserResponse(resp *TicketUserResponse) {
	if resp != nil {
		resetTicketUserResponse(resp)
		ticketUserResponsePool.Put(resp)
	}
}

// GetTicketMessageResponse retrieves a TicketMessageResponse from the pool
func GetTicketMessageResponse() *TicketMessageResponse {
	atomic.AddInt64(&ticketMessageResponsePoolHits, 1)
	return ticketMessageResponsePool.Get().(*TicketMessageResponse)
}

// PutTicketMessageResponse returns a TicketMessageResponse to the pool after safe reset
func PutTicketMessageResponse(resp *TicketMessageResponse) {
	if resp != nil {
		resetTicketMessageResponse(resp)
		ticketMessageResponsePool.Put(resp)
	}
}

// GetTicketMessageUserResponse retrieves a TicketMessageUserResponse from the pool
func GetTicketMessageUserResponse() *TicketMessageUserResponse {
	atomic.AddInt64(&ticketMessageUserResponsePoolHits, 1)
	return ticketMessageUserResponsePool.Get().(*TicketMessageUserResponse)
}

// PutTicketMessageUserResponse returns a TicketMessageUserResponse to the pool after safe reset
func PutTicketMessageUserResponse(resp *TicketMessageUserResponse) {
	if resp != nil {
		resetTicketMessageUserResponse(resp)
		ticketMessageUserResponsePool.Put(resp)
	}
}

// ToTicketResponse converts a Ticket entity to TicketResponse (admin view)
func ToTicketResponse(ticket *entities.Ticket) *TicketResponse {
	if ticket == nil {
		return nil
	}

	response := GetTicketResponse()
	response.ID = ticket.ID
	response.TicketNo = ticket.TicketNo
	response.Title = ticket.Title
	response.Description = ticket.Description
	response.Category = ticket.Category
	response.Priority = ticket.Priority
	response.Status = ticket.Status
	response.UserID = ticket.UserID
	response.AssignedToID = ticket.AssignedToID
	response.AssignedAt = ticket.AssignedAt
	response.ResolvedByID = ticket.ResolvedByID
	response.ResolvedAt = ticket.ResolvedAt
	response.Resolution = ticket.Resolution
	response.FirstResponseAt = ticket.FirstResponseAt
	response.LastResponseAt = ticket.LastResponseAt
	response.ClosedAt = ticket.ClosedAt
	response.Tags = ticket.Tags
	response.Metadata = ticket.Metadata
	response.CreatedAt = ticket.CreatedAt
	response.UpdatedAt = ticket.UpdatedAt

	// Convert messages if present
	if ticket.Messages != nil {
		response.Messages = make([]TicketMessageResponse, len(ticket.Messages))
		for i, msg := range ticket.Messages {
			msgResp := ToTicketMessageResponse(&msg)
			response.Messages[i] = *msgResp
			PutTicketMessageResponse(msgResp)
		}
	}

	return response
}

// ToTicketUserResponse converts a Ticket entity to TicketUserResponse (user view)
func ToTicketUserResponse(ticket *entities.Ticket) *TicketUserResponse {
	if ticket == nil {
		return nil
	}

	response := GetTicketUserResponse()
	response.ID = ticket.ID
	response.TicketNo = ticket.TicketNo
	response.Title = ticket.Title
	response.Description = ticket.Description
	response.Category = ticket.Category
	response.Priority = ticket.Priority
	response.Status = ticket.Status
	response.Resolution = ticket.Resolution
	response.FirstResponseAt = ticket.FirstResponseAt
	response.LastResponseAt = ticket.LastResponseAt
	response.ClosedAt = ticket.ClosedAt
	response.CreatedAt = ticket.CreatedAt
	response.UpdatedAt = ticket.UpdatedAt

	// Convert non-internal messages only
	if ticket.Messages != nil {
		var userMessages []TicketMessageUserResponse
		for _, msg := range ticket.Messages {
			if !msg.IsInternal {
				msgResp := ToTicketMessageUserResponse(&msg)
				userMessages = append(userMessages, *msgResp)
				PutTicketMessageUserResponse(msgResp)
			}
		}
		response.Messages = userMessages
	}

	return response
}

// ToTicketMessageResponse converts a TicketMessage entity to TicketMessageResponse (admin view)
func ToTicketMessageResponse(message *entities.TicketMessage) *TicketMessageResponse {
	if message == nil {
		return nil
	}

	response := GetTicketMessageResponse()
	response.ID = message.ID
	response.TicketID = message.TicketID
	response.UserID = message.UserID
	response.Content = message.Content
	response.MessageType = message.MessageType
	response.Attachments = message.Attachments
	response.IsInternal = message.IsInternal
	response.Metadata = message.Metadata
	response.CreatedAt = message.CreatedAt
	response.UpdatedAt = message.UpdatedAt

	return response
}

// ToTicketMessageUserResponse converts a TicketMessage entity to TicketMessageUserResponse (user view)
func ToTicketMessageUserResponse(message *entities.TicketMessage) *TicketMessageUserResponse {
	if message == nil {
		return nil
	}

	response := GetTicketMessageUserResponse()
	response.ID = message.ID
	response.TicketID = message.TicketID
	response.Content = message.Content
	response.MessageType = message.MessageType
	response.Attachments = message.Attachments
	response.CreatedAt = message.CreatedAt
	response.UpdatedAt = message.UpdatedAt

	return response
}

// ToTicketSliceResponse converts a slice of Ticket entities to TicketResponse slice
func ToTicketSliceResponse(tickets []*entities.Ticket) []*TicketResponse {
	if tickets == nil {
		return nil
	}

	responses := make([]*TicketResponse, len(tickets))
	for i, ticket := range tickets {
		responses[i] = ToTicketResponse(ticket)
	}

	return responses
}

// ToTicketUserSliceResponse converts a slice of Ticket entities to TicketUserResponse slice
func ToTicketUserSliceResponse(tickets []*entities.Ticket) []*TicketUserResponse {
	if tickets == nil {
		return nil
	}

	responses := make([]*TicketUserResponse, len(tickets))
	for i, ticket := range tickets {
		responses[i] = ToTicketUserResponse(ticket)
	}

	return responses
}
