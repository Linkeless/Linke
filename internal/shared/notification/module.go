package notification

import (
	"context"

	"go.uber.org/fx"

	"linke/internal/shared/cache"
	"linke/internal/shared/config"
	loggerPkg "linke/internal/shared/logger"
	"linke/internal/shared/queue"
)

// NotificationConfig contains notification system configuration
type NotificationConfig struct {
	TelegramBotToken string
	EnableMockMode   bool // For development/testing
}

// Module provides notification system dependencies for Fx
var Module = fx.Module("notification",
	// Provide notification configuration
	fx.Provide(
		func(cfg *config.Config) *NotificationConfig {
			return &NotificationConfig{
				TelegramBotToken: cfg.OAuth2.TelegramBotToken,
				EnableMockMode:   cfg.OAuth2.TelegramBotToken == "", // Use mock if no token provided
			}
		},
	),

	// Provide email providers
	fx.Provide(
		fx.Annotate(
			func(logger loggerPkg.Logger, cfg *NotificationConfig) EmailProvider {
				if cfg.EnableMockMode {
					return NewMockEmailProvider(logger)
				}
				// In production, this could be a real email provider like SendGrid, SES, etc.
				return NewLogOnlyEmailProvider(logger)
			},
			fx.As(new(EmailProvider)),
		),
	),

	// Provide SMS providers
	fx.Provide(
		fx.Annotate(
			func(logger loggerPkg.Logger, cfg *NotificationConfig) SMSProvider {
				if cfg.EnableMockMode {
					return NewMockSMSProvider(logger)
				}
				// In production, this could be a real SMS provider like Twilio, AWS SNS, etc.
				return NewMockSMSProvider(logger) // Using mock for now
			},
			fx.As(new(SMSProvider)),
		),
	),

	// Provide push notification providers
	fx.Provide(
		fx.Annotate(
			func(logger loggerPkg.Logger, cfg *NotificationConfig) PushProvider {
				if cfg.EnableMockMode {
					return NewMockPushProvider(logger)
				}
				// In production, this could be a real push provider like FCM, APNs, etc.
				return NewMockPushProvider(logger) // Using mock for now
			},
			fx.As(new(PushProvider)),
		),
	),

	// Provide Telegram provider
	fx.Provide(
		fx.Annotate(
			func(logger loggerPkg.Logger, cfg *NotificationConfig) TelegramProvider {
				if cfg.EnableMockMode || cfg.TelegramBotToken == "" {
					logger.Info("Using mock Telegram provider (no bot token configured)")
					return NewMockTelegramProvider(logger)
				}
				
				logger.Info("Using real Telegram Bot provider")
				return NewTelegramBotProvider(cfg.TelegramBotToken, logger)
			},
			fx.As(new(TelegramProvider)),
		),
	),

	// Provide base notification service (not exposed as NotificationService interface)
	fx.Provide(
		func(
			logger loggerPkg.Logger,
			emailProvider EmailProvider,
			smsProvider SMSProvider,
			pushProvider PushProvider,
			telegramProvider TelegramProvider,
		) *BaseNotificationService {
			service := NewBaseNotificationService(logger)
			service.SetEmailProvider(emailProvider)
			service.SetSMSProvider(smsProvider)
			service.SetPushProvider(pushProvider)
			service.SetTelegramProvider(telegramProvider)
			return service
		},
	),

	// Provide batch notification service
	fx.Provide(
		func(
			baseService *BaseNotificationService,
			taskQueue *queue.TaskQueue,
			logger loggerPkg.Logger,
		) *BatchNotificationService {
			return NewBatchNotificationService(baseService, taskQueue, logger)
		},
	),

	// Provide failure store - use memory store for now, can be switched to Redis
	fx.Provide(
		func(logger loggerPkg.Logger) FailureStore {
			return NewMemoryFailureStore(logger)
		},
	),

	// Provide retryable notification service
	fx.Provide(
		func(
			baseService *BaseNotificationService,
			taskQueue *queue.TaskQueue,
			failureStore FailureStore,
			logger loggerPkg.Logger,
		) *RetryableNotificationService {
			return NewRetryableNotificationService(baseService, taskQueue, failureStore, logger)
		},
	),

	// Provide retry task handler
	fx.Provide(
		func(
			retryService *RetryableNotificationService,
			failureStore FailureStore,
			logger loggerPkg.Logger,
		) *RetryTaskHandler {
			return NewRetryTaskHandler(retryService, failureStore, logger)
		},
	),

	// Provide notification status tracker
	fx.Provide(
		func(
			cacheStore cache.CacheStore,
			logger loggerPkg.Logger,
		) *NotificationStatusTracker {
			config := DefaultStatusTrackerConfig()
			return NewNotificationStatusTracker(cacheStore, logger, config)
		},
	),

	// Provide tracking notification service
	fx.Provide(
		fx.Annotate(
			func(
				baseService *BaseNotificationService,
				statusTracker *NotificationStatusTracker,
				logger loggerPkg.Logger,
			) *TrackingNotificationService {
				return NewTrackingNotificationService(baseService, statusTracker, logger)
			},
			fx.As(new(NotificationService)), // This is the main NotificationService implementation
		),
	),

	// Provide tracking retryable notification service
	fx.Provide(
		func(
			baseService *BaseNotificationService,
			retryableService *RetryableNotificationService,
			statusTracker *NotificationStatusTracker,
			logger loggerPkg.Logger,
		) *TrackingRetryableNotificationService {
			return NewTrackingRetryableNotificationService(
				baseService,
				retryableService,
				statusTracker,
				logger,
			)
		},
	),

	// Provide notification tracking handlers
	fx.Provide(
		func(
			statusTracker *NotificationStatusTracker,
			logger loggerPkg.Logger,
		) *NotificationTrackingHandlers {
			return NewNotificationTrackingHandlers(statusTracker, logger)
		},
	),

	// Optional: Provide notification service startup hook
	fx.Invoke(func(
		lc fx.Lifecycle,
		notificationService NotificationService,
		logger loggerPkg.Logger,
	) {
		lc.Append(fx.Hook{
			OnStart: func(ctx context.Context) error {
				// Perform health check on startup
				if err := notificationService.HealthCheck(ctx); err != nil {
					logger.Warn("Notification service health check failed", 
						loggerPkg.ErrorField(err))
				} else {
					logger.Info("Notification service initialized successfully")
				}
				return nil
			},
		})
	}),
)