package event

import (
	"context"
	"encoding/json"
	"fmt"

	"go.uber.org/zap"
	"linke/internal/logger"
	"linke/internal/subscription/domain/event"
)

type DomainEventPublisher struct {
	logger *zap.Logger
}

func NewDomainEventPublisher() *DomainEventPublisher {
	return &DomainEventPublisher{
		logger: logger.GetLogger(),
	}
}

func (p *DomainEventPublisher) PublishDomainEvents(ctx context.Context, events []interface{}) error {
	for _, eventInterface := range events {
		domainEvent, ok := eventInterface.(event.DomainEvent)
		if !ok {
			p.logger.Warn("Event is not a domain event, skipping", zap.Any("event", eventInterface))
			continue
		}

		if err := p.publishEvent(ctx, domainEvent); err != nil {
			p.logger.Error("Failed to publish domain event",
				zap.Error(err),
				zap.String("event_type", domainEvent.EventType()))
			return fmt.Errorf("failed to publish event %s: %w", domainEvent.EventType(), err)
		}

		p.logger.Info("Domain event published successfully",
			zap.String("event_type", domainEvent.EventType()),
			zap.Time("occurred_on", domainEvent.OccurredOn()))
	}

	return nil
}

func (p *DomainEventPublisher) publishEvent(ctx context.Context, domainEvent event.DomainEvent) error {
	eventData, err := json.Marshal(domainEvent)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	p.logger.Debug("Publishing domain event",
		logger.String("event_type", domainEvent.EventType()),
		logger.String("event_data", string(eventData)))

	return nil
}