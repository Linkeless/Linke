package event

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"linke/internal/shared/domain"
)

// InMemoryEventPublisher is a simple in-memory event publisher for development/testing
type InMemoryEventPublisher struct {
	events []domain.DomainEvent
}

// NewInMemoryEventPublisher creates a new InMemoryEventPublisher
func NewInMemoryEventPublisher() domain.EventPublisher {
	return &InMemoryEventPublisher{
		events: make([]domain.DomainEvent, 0),
	}
}

// Publish publishes a domain event
func (p *InMemoryEventPublisher) Publish(ctx context.Context, event domain.DomainEvent) error {
	// In a real implementation, this would publish to a message broker like Redis, RabbitMQ, etc.
	// For now, we'll just store in memory and log
	p.events = append(p.events, event)
	
	// Log the event (in production, use proper logging)
	eventData, _ := json.Marshal(map[string]interface{}{
		"event_id":     event.EventID(),
		"aggregate_id": event.AggregateID(),
		"event_type":   event.EventType(),
		"occurred_at":  event.OccurredAt(),
		"event_data":   event.EventData(),
	})
	
	fmt.Printf("Event published: %s\n", string(eventData))
	
	return nil
}

// PublishBatch publishes multiple domain events
func (p *InMemoryEventPublisher) PublishBatch(ctx context.Context, events []domain.DomainEvent) error {
	for _, event := range events {
		if err := p.Publish(ctx, event); err != nil {
			return fmt.Errorf("failed to publish event %s: %w", event.EventID(), err)
		}
	}
	return nil
}

// GetEvents returns all published events (for testing)
func (p *InMemoryEventPublisher) GetEvents() []domain.DomainEvent {
	return p.events
}

// Clear clears all events (for testing)
func (p *InMemoryEventPublisher) Clear() {
	p.events = make([]domain.DomainEvent, 0)
}

// AsyncEventPublisher is an asynchronous event publisher that uses channels
type AsyncEventPublisher struct {
	eventChan chan domain.DomainEvent
	batchChan chan []domain.DomainEvent
	done      chan bool
}

// NewAsyncEventPublisher creates a new AsyncEventPublisher
func NewAsyncEventPublisher() *AsyncEventPublisher {
	publisher := &AsyncEventPublisher{
		eventChan: make(chan domain.DomainEvent, 1000),
		batchChan: make(chan []domain.DomainEvent, 100),
		done:      make(chan bool),
	}
	
	// Start background worker
	go publisher.worker()
	
	return publisher
}

// Publish publishes a domain event asynchronously
func (p *AsyncEventPublisher) Publish(ctx context.Context, event domain.DomainEvent) error {
	select {
	case p.eventChan <- event:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	default:
		return fmt.Errorf("event channel is full")
	}
}

// PublishBatch publishes multiple domain events asynchronously
func (p *AsyncEventPublisher) PublishBatch(ctx context.Context, events []domain.DomainEvent) error {
	select {
	case p.batchChan <- events:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	default:
		return fmt.Errorf("batch channel is full")
	}
}

// Close closes the event publisher
func (p *AsyncEventPublisher) Close() {
	close(p.eventChan)
	close(p.batchChan)
	<-p.done
}

// worker processes events in the background
func (p *AsyncEventPublisher) worker() {
	defer func() {
		p.done <- true
	}()
	
	for {
		select {
		case event, ok := <-p.eventChan:
			if !ok {
				return
			}
			p.processEvent(event)
			
		case events, ok := <-p.batchChan:
			if !ok {
				return
			}
			p.processBatch(events)
		}
	}
}

// processEvent processes a single event
func (p *AsyncEventPublisher) processEvent(event domain.DomainEvent) {
	// In a real implementation, this would publish to a message broker
	eventData, _ := json.Marshal(map[string]interface{}{
		"event_id":     event.EventID(),
		"aggregate_id": event.AggregateID(),
		"event_type":   event.EventType(),
		"occurred_at":  event.OccurredAt(),
		"event_data":   event.EventData(),
		"processed_at": time.Now(),
	})
	
	fmt.Printf("Event processed: %s\n", string(eventData))
}

// processBatch processes multiple events
func (p *AsyncEventPublisher) processBatch(events []domain.DomainEvent) {
	for _, event := range events {
		p.processEvent(event)
	}
}

// RedisEventPublisher publishes events to Redis (placeholder implementation)
type RedisEventPublisher struct {
	// redisClient redis.Client - placeholder
	channel string
}

// NewRedisEventPublisher creates a new RedisEventPublisher
func NewRedisEventPublisher(channel string) *RedisEventPublisher {
	return &RedisEventPublisher{
		channel: channel,
	}
}

// Publish publishes a domain event to Redis
func (p *RedisEventPublisher) Publish(ctx context.Context, event domain.DomainEvent) error {
	// Placeholder implementation
	// In a real implementation, you would use Redis PUBLISH command
	eventData, err := json.Marshal(map[string]interface{}{
		"event_id":     event.EventID(),
		"aggregate_id": event.AggregateID(),
		"event_type":   event.EventType(),
		"occurred_at":  event.OccurredAt(),
		"event_data":   event.EventData(),
	})
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}
	
	// redisClient.Publish(ctx, p.channel, eventData)
	fmt.Printf("Redis event published to %s: %s\n", p.channel, string(eventData))
	
	return nil
}

// PublishBatch publishes multiple domain events to Redis
func (p *RedisEventPublisher) PublishBatch(ctx context.Context, events []domain.DomainEvent) error {
	for _, event := range events {
		if err := p.Publish(ctx, event); err != nil {
			return fmt.Errorf("failed to publish event %s: %w", event.EventID(), err)
		}
	}
	return nil
}