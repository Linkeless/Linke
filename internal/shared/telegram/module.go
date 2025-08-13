package telegram

import (
	"context"

	"linke/internal/domains/subscription/usecases/interfaces"
	ticketInterfaces "linke/internal/domains/ticket/usecases/interfaces"
	userInterfaces "linke/internal/domains/user/usecases/interfaces"
	"linke/internal/shared/config"
	"linke/internal/shared/events"
	"linke/internal/shared/logger"

	"go.uber.org/fx"
)

// Module provides the Telegram bot functionality
var Module = fx.Module("telegram",
	fx.Provide(NewBotEnhancedForFx),
	fx.Provide(NewTicketEventHandler),
	fx.Invoke(RegisterBot),
	fx.Invoke(RegisterTicketEventHandler),
)

// NewBotEnhancedForFx creates a new BotEnhanced with fx dependency injection
func NewBotEnhancedForFx(params BotParams) (*BotEnhanced, error) {
	return NewBotEnhanced(
		params.Config,
		params.UserService,
		params.SubscriptionService,
		params.SubscriptionPlanService,
		params.TicketService,
		params.TicketMessageService,
	)
}

// BotParams contains the dependencies for the Telegram bot
type BotParams struct {
	fx.In
	Config                  *config.Config
	UserService             userInterfaces.UserService
	SubscriptionService     interfaces.UserSubscriptionService
	SubscriptionPlanService interfaces.SubscriptionPlanService
	TicketService           ticketInterfaces.TicketService        `optional:"true"`
	TicketMessageService    ticketInterfaces.TicketMessageService `optional:"true"`
}

// BotResult is the result of creating a bot
type BotResult struct {
	fx.Out
	Bot *BotEnhanced `optional:"true"`
}

// RegisterBot registers the bot lifecycle hooks
func RegisterBot(lc fx.Lifecycle, bot *BotEnhanced) {
	if bot == nil {
		logger.Info("Telegram bot is not configured, skipping registration")
		return
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			// Start the bot in a goroutine
			go func() {
				logger.Info("Starting Enhanced Telegram bot")
				if err := bot.Start(context.Background()); err != nil {
					logger.Error("Enhanced Telegram bot stopped with error", logger.ErrorField(err))
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			logger.Info("Stopping Enhanced Telegram bot")
			// The bot will stop when the context is cancelled
			return nil
		},
	})
}

// RegisterTicketEventHandler registers the ticket event handler with the event bus
func RegisterTicketEventHandler(
	lc fx.Lifecycle,
	eventBus events.EventBus,
	handler *TicketEventHandler,
) {
	if handler == nil || eventBus == nil {
		logger.Info("Ticket event handler or event bus not available, skipping registration")
		return
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			// Subscribe to all ticket events at once
			if err := eventBus.Subscribe(handler.EventTypes(), handler); err != nil {
				logger.Error("Failed to subscribe to ticket events",
					logger.Any("event_types", handler.EventTypes()),
					logger.ErrorField(err))
				return err
			}
			logger.Info("Subscribed to ticket events",
				logger.Any("event_types", handler.EventTypes()),
				logger.String("handler_id", handler.ID()))
			return nil
		},
		OnStop: func(ctx context.Context) error {
			// Unsubscribe from all events
			if err := eventBus.Unsubscribe(handler.EventTypes(), handler); err != nil {
				logger.Error("Failed to unsubscribe from ticket events",
					logger.Any("event_types", handler.EventTypes()),
					logger.ErrorField(err))
			}
			return nil
		},
	})
}
