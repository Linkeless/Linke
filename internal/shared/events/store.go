package events

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"linke/internal/shared/logger"

	"gorm.io/gorm"
)

// EventStore provides persistence for events
type EventStore interface {
	Store(ctx context.Context, event Event) error
	GetEvents(ctx context.Context, filters EventFilters) ([]*StoredEvent, error)
	GetEventsByAggregateID(ctx context.Context, aggregateID string, aggregateType string) ([]*StoredEvent, error)
	GetEventsByType(ctx context.Context, eventType string, limit int, offset int) ([]*StoredEvent, error)
	Replay(ctx context.Context, fromTimestamp time.Time, handler EventHandler) error
}

// StoredEvent represents an event stored in the database
type StoredEvent struct {
	ID            uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	EventID       string    `gorm:"uniqueIndex;not null" json:"event_id"`
	EventType     string    `gorm:"index;not null" json:"event_type"`
	EventSource   string    `gorm:"index;not null" json:"event_source"`
	AggregateID   string    `gorm:"index" json:"aggregate_id,omitempty"`
	AggregateType string    `gorm:"index" json:"aggregate_type,omitempty"`
	EventVersion  string    `json:"event_version"`
	EventData     string    `gorm:"type:text" json:"event_data"`
	Metadata      string    `gorm:"type:text" json:"metadata,omitempty"`
	OccurredAt    time.Time `gorm:"index;not null" json:"occurred_at"`
	StoredAt      time.Time `gorm:"autoCreateTime" json:"stored_at"`
}

// TableName specifies the table name for GORM
func (StoredEvent) TableName() string {
	return "event_store"
}

// EventFilters provides filtering options for event queries
type EventFilters struct {
	EventTypes    []string  `json:"event_types,omitempty"`
	EventSources  []string  `json:"event_sources,omitempty"`
	AggregateID   string    `json:"aggregate_id,omitempty"`
	AggregateType string    `json:"aggregate_type,omitempty"`
	FromTime      time.Time `json:"from_time,omitempty"`
	ToTime        time.Time `json:"to_time,omitempty"`
	Limit         int       `json:"limit,omitempty"`
	Offset        int       `json:"offset,omitempty"`
}

// DatabaseEventStore implements EventStore using GORM
type DatabaseEventStore struct {
	db     *gorm.DB
	logger logger.Logger
}

// NewDatabaseEventStore creates a new database event store
func NewDatabaseEventStore(db *gorm.DB) *DatabaseEventStore {
	return &DatabaseEventStore{
		db:     db,
		logger: logger.GetGlobalLogger(),
	}
}

// Store saves an event to the database
func (s *DatabaseEventStore) Store(ctx context.Context, event Event) error {
	// Serialize event data
	eventData, err := json.Marshal(event.EventData())
	if err != nil {
		return fmt.Errorf("failed to marshal event data: %w", err)
	}

	// Serialize metadata
	var metadataJSON string
	if baseEvent, ok := event.(*BaseEvent); ok && baseEvent.Metadata != nil {
		metadata, err := json.Marshal(baseEvent.Metadata)
		if err != nil {
			return fmt.Errorf("failed to marshal event metadata: %w", err)
		}
		metadataJSON = string(metadata)
	}

	// Extract aggregate information if available
	var aggregateID, aggregateType string
	if userEvent, ok := event.(*UserEvent); ok {
		aggregateID = fmt.Sprintf("%d", userEvent.UserID)
		aggregateType = "user"
	} else if paymentEvent, ok := event.(*PaymentEvent); ok {
		aggregateID = paymentEvent.PaymentID
		aggregateType = "payment"
	} else if subscriptionEvent, ok := event.(*SubscriptionEvent); ok {
		aggregateID = fmt.Sprintf("%d", subscriptionEvent.SubscriptionID)
		aggregateType = "subscription"
	} else if orderEvent, ok := event.(*OrderEvent); ok {
		aggregateID = fmt.Sprintf("%d", orderEvent.OrderID)
		aggregateType = "order"
	} else if invoiceEvent, ok := event.(*InvoiceEvent); ok {
		aggregateID = fmt.Sprintf("%d", invoiceEvent.InvoiceID)
		aggregateType = "invoice"
	} else if serverEvent, ok := event.(*ServerEvent); ok {
		aggregateID = fmt.Sprintf("%d", serverEvent.ServerID)
		aggregateType = "server"
	}

	storedEvent := &StoredEvent{
		EventID:       event.EventID(),
		EventType:     event.EventType(),
		EventSource:   event.EventSource(),
		AggregateID:   aggregateID,
		AggregateType: aggregateType,
		EventVersion:  event.EventVersion(),
		EventData:     string(eventData),
		Metadata:      metadataJSON,
		OccurredAt:    event.EventTime(),
	}

	if err := s.db.WithContext(ctx).Create(storedEvent).Error; err != nil {
		s.logger.Error("Failed to store event",
			logger.String("event_id", event.EventID()),
			logger.String("event_type", event.EventType()),
			logger.ErrorField(err),
		)
		return fmt.Errorf("failed to store event: %w", err)
	}

	s.logger.Debug("Event stored successfully",
		logger.String("event_id", event.EventID()),
		logger.String("event_type", event.EventType()),
		logger.String("aggregate_id", aggregateID),
		logger.String("aggregate_type", aggregateType),
	)

	return nil
}

// GetEvents retrieves events based on filters
func (s *DatabaseEventStore) GetEvents(ctx context.Context, filters EventFilters) ([]*StoredEvent, error) {
	query := s.db.WithContext(ctx).Model(&StoredEvent{})

	// Apply filters
	if len(filters.EventTypes) > 0 {
		query = query.Where("event_type IN ?", filters.EventTypes)
	}
	if len(filters.EventSources) > 0 {
		query = query.Where("event_source IN ?", filters.EventSources)
	}
	if filters.AggregateID != "" {
		query = query.Where("aggregate_id = ?", filters.AggregateID)
	}
	if filters.AggregateType != "" {
		query = query.Where("aggregate_type = ?", filters.AggregateType)
	}
	if !filters.FromTime.IsZero() {
		query = query.Where("occurred_at >= ?", filters.FromTime)
	}
	if !filters.ToTime.IsZero() {
		query = query.Where("occurred_at <= ?", filters.ToTime)
	}

	// Apply pagination
	if filters.Limit > 0 {
		query = query.Limit(filters.Limit)
	}
	if filters.Offset > 0 {
		query = query.Offset(filters.Offset)
	}

	// Order by occurrence time
	query = query.Order("occurred_at ASC")

	var events []*StoredEvent
	if err := query.Find(&events).Error; err != nil {
		s.logger.Error("Failed to retrieve events",
			logger.Any("filters", filters),
			logger.ErrorField(err),
		)
		return nil, fmt.Errorf("failed to retrieve events: %w", err)
	}

	return events, nil
}

// GetEventsByAggregateID retrieves all events for a specific aggregate
func (s *DatabaseEventStore) GetEventsByAggregateID(ctx context.Context, aggregateID string, aggregateType string) ([]*StoredEvent, error) {
	var events []*StoredEvent
	query := s.db.WithContext(ctx).Where("aggregate_id = ? AND aggregate_type = ?", aggregateID, aggregateType).
		Order("occurred_at ASC")

	if err := query.Find(&events).Error; err != nil {
		s.logger.Error("Failed to retrieve events by aggregate ID",
			logger.String("aggregate_id", aggregateID),
			logger.String("aggregate_type", aggregateType),
			logger.ErrorField(err),
		)
		return nil, fmt.Errorf("failed to retrieve events by aggregate ID: %w", err)
	}

	return events, nil
}

// GetEventsByType retrieves events by type with pagination
func (s *DatabaseEventStore) GetEventsByType(ctx context.Context, eventType string, limit int, offset int) ([]*StoredEvent, error) {
	var events []*StoredEvent
	query := s.db.WithContext(ctx).Where("event_type = ?", eventType).
		Order("occurred_at ASC").
		Limit(limit).
		Offset(offset)

	if err := query.Find(&events).Error; err != nil {
		s.logger.Error("Failed to retrieve events by type",
			logger.String("event_type", eventType),
			logger.ErrorField(err),
		)
		return nil, fmt.Errorf("failed to retrieve events by type: %w", err)
	}

	return events, nil
}

// Replay replays events from a specific timestamp
func (s *DatabaseEventStore) Replay(ctx context.Context, fromTimestamp time.Time, handler EventHandler) error {
	const batchSize = 100
	offset := 0

	for {
		var events []*StoredEvent
		query := s.db.WithContext(ctx).Where("occurred_at >= ?", fromTimestamp).
			Order("occurred_at ASC").
			Limit(batchSize).
			Offset(offset)

		if err := query.Find(&events).Error; err != nil {
			return fmt.Errorf("failed to retrieve events for replay: %w", err)
		}

		if len(events) == 0 {
			break
		}

		for _, storedEvent := range events {
			// Convert stored event back to event interface
			event, err := s.deserializeStoredEvent(storedEvent)
			if err != nil {
				s.logger.Error("Failed to deserialize event during replay",
					logger.String("event_id", storedEvent.EventID),
					logger.ErrorField(err),
				)
				continue
			}

			// Check if handler supports this event type
			for _, supportedType := range handler.EventTypes() {
				if supportedType == event.EventType() {
					if err := handler.Handle(ctx, event); err != nil {
						s.logger.Error("Event handler failed during replay",
							logger.String("event_id", event.EventID()),
							logger.String("event_type", event.EventType()),
							logger.ErrorField(err),
						)
					}
					break
				}
			}
		}

		offset += batchSize
	}

	s.logger.Info("Event replay completed",
		logger.String("from_timestamp", fromTimestamp.Format(time.RFC3339)),
		logger.String("handler_types", fmt.Sprintf("%v", handler.EventTypes())),
	)

	return nil
}

// deserializeStoredEvent converts a stored event back to an Event interface
func (s *DatabaseEventStore) deserializeStoredEvent(storedEvent *StoredEvent) (Event, error) {
	// Parse event data
	var eventData any
	if err := json.Unmarshal([]byte(storedEvent.EventData), &eventData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal event data: %w", err)
	}

	// Parse metadata
	var metadata map[string]any
	if storedEvent.Metadata != "" {
		if err := json.Unmarshal([]byte(storedEvent.Metadata), &metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal event metadata: %w", err)
		}
	}

	// Create base event
	baseEvent := &BaseEvent{
		ID:       storedEvent.EventID,
		Type:     storedEvent.EventType,
		Source:   storedEvent.EventSource,
		Time:     storedEvent.OccurredAt,
		Version:  storedEvent.EventVersion,
		Data:     eventData,
		Metadata: metadata,
	}

	// Return appropriate event type based on aggregate type
	switch storedEvent.AggregateType {
	case "user":
		// Parse aggregate ID as uint for user events
		var userID uint
		if _, err := fmt.Sscanf(storedEvent.AggregateID, "%d", &userID); err != nil {
			return baseEvent, nil // Return base event if parsing fails
		}
		return &UserEvent{
			BaseEvent: baseEvent,
			UserID:    userID,
		}, nil
	case "payment":
		// For payment events, we need to extract additional fields from data
		if dataMap, ok := eventData.(map[string]any); ok {
			paymentEvent := &PaymentEvent{
				BaseEvent: baseEvent,
				PaymentID: storedEvent.AggregateID,
			}
			if amount, ok := dataMap["amount"].(float64); ok {
				paymentEvent.Amount = amount
			}
			if userID, ok := dataMap["user_id"].(float64); ok {
				paymentEvent.UserID = uint(userID)
			}
			return paymentEvent, nil
		}
		return baseEvent, nil
	case "subscription":
		var subscriptionID uint
		if _, err := fmt.Sscanf(storedEvent.AggregateID, "%d", &subscriptionID); err != nil {
			return baseEvent, nil
		}
		subscriptionEvent := &SubscriptionEvent{
			BaseEvent:      baseEvent,
			SubscriptionID: subscriptionID,
		}
		// Extract user ID from data if available
		if dataMap, ok := eventData.(map[string]any); ok {
			if userID, ok := dataMap["user_id"].(float64); ok {
				subscriptionEvent.UserID = uint(userID)
			}
		}
		return subscriptionEvent, nil
	case "order":
		var orderID uint
		if _, err := fmt.Sscanf(storedEvent.AggregateID, "%d", &orderID); err != nil {
			return baseEvent, nil
		}
		orderEvent := &OrderEvent{
			BaseEvent: baseEvent,
			OrderID:   orderID,
		}
		// Extract user ID from data if available
		if dataMap, ok := eventData.(map[string]any); ok {
			if userID, ok := dataMap["user_id"].(float64); ok {
				orderEvent.UserID = uint(userID)
			}
		}
		return orderEvent, nil
	case "invoice":
		var invoiceID uint
		if _, err := fmt.Sscanf(storedEvent.AggregateID, "%d", &invoiceID); err != nil {
			return baseEvent, nil
		}
		invoiceEvent := &InvoiceEvent{
			BaseEvent: baseEvent,
			InvoiceID: invoiceID,
		}
		// Extract additional fields from data
		if dataMap, ok := eventData.(map[string]any); ok {
			if orderID, ok := dataMap["order_id"].(float64); ok {
				invoiceEvent.OrderID = uint(orderID)
			}
			if userID, ok := dataMap["user_id"].(float64); ok {
				invoiceEvent.UserID = uint(userID)
			}
			if amount, ok := dataMap["amount"].(float64); ok {
				invoiceEvent.Amount = amount
			}
		}
		return invoiceEvent, nil
	case "server":
		var serverID uint
		if _, err := fmt.Sscanf(storedEvent.AggregateID, "%d", &serverID); err != nil {
			return baseEvent, nil
		}
		return &ServerEvent{
			BaseEvent: baseEvent,
			ServerID:  serverID,
		}, nil
	default:
		return baseEvent, nil
	}
}

// EventStoreStats provides statistics about the event store
type EventStoreStats struct {
	TotalEvents      int64            `json:"total_events"`
	EventsByType     map[string]int64 `json:"events_by_type"`
	EventsBySource   map[string]int64 `json:"events_by_source"`
	RecentEventCount int64            `json:"recent_event_count"`
	OldestEvent      time.Time        `json:"oldest_event"`
	NewestEvent      time.Time        `json:"newest_event"`
}

// GetStats returns statistics about the event store
func (s *DatabaseEventStore) GetStats(ctx context.Context) (*EventStoreStats, error) {
	stats := &EventStoreStats{
		EventsByType:   make(map[string]int64),
		EventsBySource: make(map[string]int64),
	}

	// Total events
	if err := s.db.WithContext(ctx).Model(&StoredEvent{}).Count(&stats.TotalEvents).Error; err != nil {
		return nil, fmt.Errorf("failed to count total events: %w", err)
	}

	// Events by type
	var typeStats []struct {
		EventType string `json:"event_type"`
		Count     int64  `json:"count"`
	}
	if err := s.db.WithContext(ctx).Model(&StoredEvent{}).
		Select("event_type, COUNT(*) as count").
		Group("event_type").
		Find(&typeStats).Error; err != nil {
		return nil, fmt.Errorf("failed to get events by type: %w", err)
	}
	for _, stat := range typeStats {
		stats.EventsByType[stat.EventType] = stat.Count
	}

	// Events by source
	var sourceStats []struct {
		EventSource string `json:"event_source"`
		Count       int64  `json:"count"`
	}
	if err := s.db.WithContext(ctx).Model(&StoredEvent{}).
		Select("event_source, COUNT(*) as count").
		Group("event_source").
		Find(&sourceStats).Error; err != nil {
		return nil, fmt.Errorf("failed to get events by source: %w", err)
	}
	for _, stat := range sourceStats {
		stats.EventsBySource[stat.EventSource] = stat.Count
	}

	// Recent events (last 24 hours)
	yesterday := time.Now().Add(-24 * time.Hour)
	if err := s.db.WithContext(ctx).Model(&StoredEvent{}).
		Where("occurred_at >= ?", yesterday).
		Count(&stats.RecentEventCount).Error; err != nil {
		return nil, fmt.Errorf("failed to count recent events: %w", err)
	}

	// Oldest and newest events
	var oldestEvent StoredEvent
	if err := s.db.WithContext(ctx).Model(&StoredEvent{}).
		Order("occurred_at ASC").
		First(&oldestEvent).Error; err != nil && err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("failed to get oldest event: %w", err)
	} else if err == nil {
		stats.OldestEvent = oldestEvent.OccurredAt
	}

	var newestEvent StoredEvent
	if err := s.db.WithContext(ctx).Model(&StoredEvent{}).
		Order("occurred_at DESC").
		First(&newestEvent).Error; err != nil && err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("failed to get newest event: %w", err)
	} else if err == nil {
		stats.NewestEvent = newestEvent.OccurredAt
	}

	return stats, nil
}
