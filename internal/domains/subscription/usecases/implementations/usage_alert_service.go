package implementations

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"linke/internal/domains/subscription/constants"
	"linke/internal/domains/subscription/entities"
	"linke/internal/domains/subscription/usecases/interfaces"
	subscriptionInterfaces "linke/internal/domains/subscription/usecases/interfaces"
	"linke/internal/shared/logger"
	"linke/internal/shared/notification"

	"go.uber.org/zap"
)

// usageAlertService implements the UsageAlertService interface
type usageAlertService struct {
	alertRepo       interfaces.AlertRepository
	usageRepo       interfaces.UsageRepository
	subscriptionSvc subscriptionInterfaces.UserSubscriptionService
	notificationSvc notification.NotificationService
	logger          *zap.Logger
}

// NewUsageAlertService creates a new usage alert service instance
func NewUsageAlertService(
	alertRepo interfaces.AlertRepository,
	usageRepo interfaces.UsageRepository,
	subscriptionSvc subscriptionInterfaces.UserSubscriptionService,
	notificationSvc notification.NotificationService,
) interfaces.UsageAlertService {
	return &usageAlertService{
		alertRepo:       alertRepo,
		usageRepo:       usageRepo,
		subscriptionSvc: subscriptionSvc,
		notificationSvc: notificationSvc,
		logger:          logger.GetLogger(),
	}
}

// Alert Configuration Management

func (s *usageAlertService) CreateAlertConfiguration(ctx context.Context, req *interfaces.CreateAlertConfigRequest) (*entities.AlertConfiguration, error) {
	// Validate request
	if req.UserSubscriptionID == 0 {
		return nil, fmt.Errorf("user_subscription_id is required")
	}
	if req.UsageType == "" {
		return nil, fmt.Errorf("usage_type is required")
	}
	if req.Threshold <= 0 {
		return nil, fmt.Errorf("threshold must be positive")
	}
	if req.ThresholdType != constants.ThresholdTypePercentage && req.ThresholdType != constants.ThresholdTypeAbsolute {
		return nil, fmt.Errorf("invalid threshold_type")
	}

	// Verify subscription exists
	_, err := s.subscriptionSvc.GetUserSubscription(ctx, req.UserSubscriptionID)
	if err != nil {
		return nil, fmt.Errorf("subscription not found: %w", err)
	}

	// Set defaults
	if req.Priority == "" {
		req.Priority = constants.PriorityMedium
	}
	if req.CooldownMinutes == 0 {
		req.CooldownMinutes = constants.DefaultCooldownMinutes
	}

	// Create alert configuration
	config := &entities.AlertConfiguration{
		UserSubscriptionID: req.UserSubscriptionID,
		UsageType:          req.UsageType,
		ThresholdType:      req.ThresholdType,
		Threshold:          req.Threshold,
		IsEnabled:          req.IsEnabled,
		CooldownMinutes:    req.CooldownMinutes,
		Name:               req.Name,
		Description:        req.Description,
		Priority:           req.Priority,
	}

	// Set notification channels
	if err := config.SetNotificationChannels(req.NotificationChannels); err != nil {
		return nil, fmt.Errorf("failed to set notification channels: %w", err)
	}

	// Save configuration
	if err := s.alertRepo.CreateAlertConfiguration(ctx, config); err != nil {
		return nil, fmt.Errorf("failed to create alert configuration: %w", err)
	}

	return config, nil
}

func (s *usageAlertService) UpdateAlertConfiguration(ctx context.Context, req *interfaces.UpdateAlertConfigRequest) (*entities.AlertConfiguration, error) {
	// Get existing configuration
	config, err := s.alertRepo.GetAlertConfiguration(ctx, req.ID)
	if err != nil {
		return nil, fmt.Errorf("alert configuration not found: %w", err)
	}

	// Update fields
	if req.ThresholdType != nil {
		if *req.ThresholdType != constants.ThresholdTypePercentage && *req.ThresholdType != constants.ThresholdTypeAbsolute {
			return nil, fmt.Errorf("invalid threshold_type")
		}
		config.ThresholdType = *req.ThresholdType
	}
	if req.Threshold != nil {
		if *req.Threshold <= 0 {
			return nil, fmt.Errorf("threshold must be positive")
		}
		config.Threshold = *req.Threshold
	}
	if req.Name != nil {
		config.Name = *req.Name
	}
	if req.Description != nil {
		config.Description = *req.Description
	}
	if req.Priority != nil {
		config.Priority = *req.Priority
	}
	if req.IsEnabled != nil {
		config.IsEnabled = *req.IsEnabled
	}
	if req.CooldownMinutes != nil {
		config.CooldownMinutes = *req.CooldownMinutes
	}
	if req.NotificationChannels != nil {
		if err := config.SetNotificationChannels(req.NotificationChannels); err != nil {
			return nil, fmt.Errorf("failed to set notification channels: %w", err)
		}
	}

	// Save updated configuration
	if err := s.alertRepo.UpdateAlertConfiguration(ctx, config); err != nil {
		return nil, fmt.Errorf("failed to update alert configuration: %w", err)
	}

	return config, nil
}

func (s *usageAlertService) DeleteAlertConfiguration(ctx context.Context, configID uint) error {
	// Check if configuration exists
	_, err := s.alertRepo.GetAlertConfiguration(ctx, configID)
	if err != nil {
		return fmt.Errorf("alert configuration not found: %w", err)
	}

	// Delete configuration (soft delete)
	return s.alertRepo.DeleteAlertConfiguration(ctx, configID)
}

func (s *usageAlertService) GetAlertConfiguration(ctx context.Context, configID uint) (*entities.AlertConfiguration, error) {
	return s.alertRepo.GetAlertConfiguration(ctx, configID)
}

func (s *usageAlertService) GetAlertConfigurations(ctx context.Context, req *interfaces.GetAlertConfigsRequest) (*interfaces.GetAlertConfigsResponse, error) {
	// Set defaults
	if req.Limit == 0 {
		req.Limit = constants.DefaultPageSize
	}
	if req.Limit > constants.MaxPageSize {
		req.Limit = constants.MaxPageSize
	}

	// Build filter
	filter := interfaces.AlertConfigurationFilter{
		UserSubscriptionID: req.UserSubscriptionID,
		UsageType:          req.UsageType,
		IsEnabled:          req.IsEnabled,
		Priority:           req.Priority,
		Limit:              req.Limit,
		Offset:             req.Offset,
		OrderBy:            req.OrderBy,
		OrderDirection:     req.OrderDirection,
		IncludeDeleted:     false,
	}

	// Get configurations
	configs, err := s.alertRepo.GetAlertConfigurations(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get alert configurations: %w", err)
	}

	// Convert to response format
	configResponses := make([]*entities.AlertConfigurationResponse, len(configs))
	for i, config := range configs {
		configResponses[i] = config.ToResponse()
	}

	// Get total count for pagination
	totalFilter := filter
	totalFilter.Limit = 0
	totalFilter.Offset = 0
	allConfigs, err := s.alertRepo.GetAlertConfigurations(ctx, totalFilter)
	if err != nil {
		return nil, fmt.Errorf("failed to get total count: %w", err)
	}
	totalCount := int64(len(allConfigs))

	// Build pagination info
	totalPages := int((totalCount + int64(req.Limit) - 1) / int64(req.Limit))
	currentPage := (req.Offset / req.Limit) + 1

	pagination := &interfaces.PaginationInfo{
		CurrentPage: currentPage,
		PageSize:    req.Limit,
		TotalPages:  totalPages,
		TotalCount:  totalCount,
		HasNext:     currentPage < totalPages,
		HasPrevious: currentPage > 1,
	}

	return &interfaces.GetAlertConfigsResponse{
		AlertConfigurations: configResponses,
		Pagination:          pagination,
		TotalCount:          totalCount,
	}, nil
}

// Alert Processing

func (s *usageAlertService) CheckUsageThresholds(ctx context.Context, subscriptionID uint, usageType string) ([]*entities.UsageAlert, error) {
	// Get active alert configurations for this subscription and usage type
	configs, err := s.alertRepo.GetActiveAlertConfigurations(ctx, subscriptionID, usageType)
	if err != nil {
		return nil, fmt.Errorf("failed to get alert configurations: %w", err)
	}

	if len(configs) == 0 {
		return nil, nil // No alerts configured
	}

	// Get subscription to determine limits
	subscription, err := s.subscriptionSvc.GetUserSubscription(ctx, subscriptionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get subscription: %w", err)
	}

	// Get current usage
	currentUsage, err := s.usageRepo.GetCurrentUsage(ctx, subscriptionID, usageType)
	if err != nil {
		return nil, fmt.Errorf("failed to get current usage: %w", err)
	}

	// Determine usage limit based on type
	var usageLimit int64
	switch usageType {
	case constants.UsageTypeTraffic:
		usageLimit = subscription.TrafficLimit
	default:
		usageLimit = 0 // Unlimited for other types
	}

	var firedAlerts []*entities.UsageAlert

	// Check each configuration
	for _, config := range configs {
		if !config.ShouldTrigger(currentUsage, usageLimit) {
			continue
		}

		// Check if there's already an active alert for this configuration
		existingAlerts, err := s.alertRepo.GetAlertsForConfiguration(ctx, config.ID, 1)
		if err != nil {
			continue // Skip if error getting existing alerts
		}

		// If there's an active alert and it's in cooldown, skip
		if len(existingAlerts) > 0 {
			lastAlert := existingAlerts[0]
			if lastAlert.IsActive() && lastAlert.CanSendNotification(config.CooldownMinutes) == false {
				continue
			}
		}

		// Fire new alert
		fireReq := &interfaces.FireAlertRequest{
			UserSubscriptionID:   subscriptionID,
			AlertConfigurationID: config.ID,
			CurrentUsage:         currentUsage,
			UsageLimit:           usageLimit,
			Message:              s.generateAlertMessage(config, currentUsage, usageLimit),
			Severity:             config.GetSeverityLevel(currentUsage, usageLimit),
			ForceNotification:    false,
		}

		alert, err := s.FireAlert(ctx, fireReq)
		if err != nil {
			// Log error but continue with other alerts
			continue
		}

		firedAlerts = append(firedAlerts, alert)
	}

	return firedAlerts, nil
}

func (s *usageAlertService) ProcessUsageUpdate(ctx context.Context, subscriptionID uint, usageType string, currentUsage int64) error {
	// Check thresholds and fire alerts if needed
	alerts, err := s.CheckUsageThresholds(ctx, subscriptionID, usageType)
	if err != nil {
		return fmt.Errorf("failed to check usage thresholds: %w", err)
	}

	// Send notifications for fired alerts
	for _, alert := range alerts {
		config, err := s.alertRepo.GetAlertConfiguration(ctx, alert.AlertConfigurationID)
		if err != nil {
			continue // Skip if can't get configuration
		}

		channels := config.GetNotificationChannels()
		if len(channels) > 0 {
			// Send notifications asynchronously
			go func(alert *entities.UsageAlert, channels []entities.NotificationChannel) {
				notifyCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
				defer cancel()

				if err := s.SendNotification(notifyCtx, alert, channels); err != nil {
					s.logger.Error("Failed to send notification for alert",
						zap.Uint("alert_id", alert.ID),
						zap.Error(err))
				}
			}(alert, channels)
		}
	}

	return nil
}

func (s *usageAlertService) FireAlert(ctx context.Context, req *interfaces.FireAlertRequest) (*entities.UsageAlert, error) {
	// Validate request
	if req.UserSubscriptionID == 0 {
		return nil, fmt.Errorf("user_subscription_id is required")
	}
	if req.AlertConfigurationID == 0 {
		return nil, fmt.Errorf("alert_configuration_id is required")
	}

	// Get alert configuration
	config, err := s.alertRepo.GetAlertConfiguration(ctx, req.AlertConfigurationID)
	if err != nil {
		return nil, fmt.Errorf("alert configuration not found: %w", err)
	}

	// Calculate usage percentage
	var usagePercent float64
	if req.UsageLimit > 0 {
		usagePercent = float64(req.CurrentUsage) / float64(req.UsageLimit) * 100
		if usagePercent > 100 {
			usagePercent = 100
		}
	}

	// Set defaults
	severity := req.Severity
	if severity == "" {
		severity = config.GetSeverityLevel(req.CurrentUsage, req.UsageLimit)
	}

	message := req.Message
	if message == "" {
		message = s.generateAlertMessage(config, req.CurrentUsage, req.UsageLimit)
	}

	// Create alert
	alert := &entities.UsageAlert{
		UserSubscriptionID:   req.UserSubscriptionID,
		AlertConfigurationID: req.AlertConfigurationID,
		UsageType:            config.UsageType,
		CurrentUsage:         req.CurrentUsage,
		UsageLimit:           req.UsageLimit,
		ThresholdValue:       config.Threshold,
		UsagePercent:         usagePercent,
		Status:               constants.AlertStatusFired,
		Severity:             severity,
		FiredAt:              time.Now(),
		Message:              message,
		NotificationChannels: config.NotificationChannels,
	}

	// Save alert
	if err := s.alertRepo.CreateUsageAlert(ctx, alert); err != nil {
		return nil, fmt.Errorf("failed to create usage alert: %w", err)
	}

	return alert, nil
}

func (s *usageAlertService) ResolveAlert(ctx context.Context, alertID uint, reason string) error {
	// Get alert
	alert, err := s.alertRepo.GetUsageAlert(ctx, alertID)
	if err != nil {
		return fmt.Errorf("alert not found: %w", err)
	}

	if alert.IsResolved() {
		return fmt.Errorf("alert is already resolved")
	}

	// Update alert metadata with resolution reason
	if reason != "" {
		metadata := fmt.Sprintf("Resolved: %s", reason)
		alert.Metadata = metadata
	}

	// Resolve alert
	alert.Resolve()

	// Save updated alert
	return s.alertRepo.UpdateUsageAlert(ctx, alert)
}

// Alert Management

func (s *usageAlertService) GetUsageAlerts(ctx context.Context, req *interfaces.GetUsageAlertsRequest) (*interfaces.GetUsageAlertsResponse, error) {
	// Set defaults
	if req.Limit == 0 {
		req.Limit = constants.DefaultPageSize
	}
	if req.Limit > constants.MaxPageSize {
		req.Limit = constants.MaxPageSize
	}

	// Build filter
	filter := interfaces.UsageAlertFilter{
		UserSubscriptionID:   req.UserSubscriptionID,
		AlertConfigurationID: req.AlertConfigurationID,
		UsageType:            req.UsageType,
		Status:               req.Status,
		Severity:             req.Severity,
		IsActive:             req.IsActive,
		IsResolved:           req.IsResolved,
		StartTime:            req.StartTime,
		EndTime:              req.EndTime,
		Limit:                req.Limit,
		Offset:               req.Offset,
		OrderBy:              req.OrderBy,
		OrderDirection:       req.OrderDirection,
		IncludeDeleted:       false,
	}

	// Get alerts
	alerts, err := s.alertRepo.GetUsageAlerts(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get usage alerts: %w", err)
	}

	// Convert to response format
	alertResponses := make([]*entities.UsageAlertResponse, len(alerts))
	for i, alert := range alerts {
		alertResponses[i] = alert.ToResponse()
	}

	// Get summary if subscription ID is provided
	var summary *interfaces.AlertsSummary
	if req.UserSubscriptionID != 0 {
		// Create basic summary from the alerts we have
		summary = s.createBasicAlertsSummary(alerts)
	}

	// Get total count for pagination
	totalFilter := filter
	totalFilter.Limit = 0
	totalFilter.Offset = 0
	allAlerts, err := s.alertRepo.GetUsageAlerts(ctx, totalFilter)
	if err != nil {
		return nil, fmt.Errorf("failed to get total count: %w", err)
	}
	totalCount := int64(len(allAlerts))

	// Build pagination info
	totalPages := int((totalCount + int64(req.Limit) - 1) / int64(req.Limit))
	currentPage := (req.Offset / req.Limit) + 1

	pagination := &interfaces.PaginationInfo{
		CurrentPage: currentPage,
		PageSize:    req.Limit,
		TotalPages:  totalPages,
		TotalCount:  totalCount,
		HasNext:     currentPage < totalPages,
		HasPrevious: currentPage > 1,
	}

	return &interfaces.GetUsageAlertsResponse{
		UsageAlerts: alertResponses,
		Pagination:  pagination,
		Summary:     summary,
		TotalCount:  totalCount,
	}, nil
}

func (s *usageAlertService) AcknowledgeAlert(ctx context.Context, alertID uint, acknowledgedBy uint) error {
	// Get alert
	alert, err := s.alertRepo.GetUsageAlert(ctx, alertID)
	if err != nil {
		return fmt.Errorf("alert not found: %w", err)
	}

	if !alert.IsActive() {
		return fmt.Errorf("alert is not active")
	}

	// Acknowledge alert
	return s.alertRepo.AcknowledgeAlert(ctx, alertID)
}

func (s *usageAlertService) SuppressAlert(ctx context.Context, alertID uint, duration time.Duration, reason string) error {
	// Get alert
	alert, err := s.alertRepo.GetUsageAlert(ctx, alertID)
	if err != nil {
		return fmt.Errorf("alert not found: %w", err)
	}

	if !alert.IsActive() {
		return fmt.Errorf("alert is not active")
	}

	// Update metadata with suppression info
	metadata := fmt.Sprintf("Suppressed for %v: %s", duration, reason)
	alert.Metadata = metadata

	// Suppress alert
	return s.alertRepo.SuppressAlert(ctx, alertID, duration)
}

func (s *usageAlertService) BulkResolveAlerts(ctx context.Context, req *interfaces.BulkResolveAlertsRequest) (*interfaces.BulkResolveAlertsResponse, error) {
	var resolvedCount int64
	var failedIDs []uint
	var errors []string

	for _, alertID := range req.AlertIDs {
		if err := s.ResolveAlert(ctx, alertID, req.Reason); err != nil {
			failedIDs = append(failedIDs, alertID)
			errors = append(errors, fmt.Sprintf("Alert %d: %v", alertID, err))
		} else {
			resolvedCount++
		}
	}

	return &interfaces.BulkResolveAlertsResponse{
		ResolvedCount: resolvedCount,
		FailedIDs:     failedIDs,
		Errors:        errors,
	}, nil
}

// Notification Management

func (s *usageAlertService) SendNotification(ctx context.Context, alert *entities.UsageAlert, channels []entities.NotificationChannel) error {
	results := make([]entities.NotificationResult, 0, len(channels))

	for _, channel := range channels {
		if !channel.Enabled {
			continue
		}

		result := entities.NotificationResult{
			Channel: channel.Type,
			Target:  channel.Target,
			SentAt:  time.Now(),
		}

		// Send notification based on channel type
		switch channel.Type {
		case constants.NotificationChannelEmail:
			err := s.sendEmailNotification(ctx, alert, channel)
			result.Success = err == nil
			if err != nil {
				result.Error = err.Error()
				result.Message = "Failed to send email"
			} else {
				result.Message = "Email sent successfully"
			}

		case constants.NotificationChannelWebhook:
			err := s.sendWebhookNotification(ctx, alert, channel)
			result.Success = err == nil
			if err != nil {
				result.Error = err.Error()
				result.Message = "Failed to send webhook"
			} else {
				result.Message = "Webhook sent successfully"
			}

		case constants.NotificationChannelInApp:
			err := s.sendInAppNotification(ctx, alert, channel)
			result.Success = err == nil
			if err != nil {
				result.Error = err.Error()
				result.Message = "Failed to create in-app notification"
			} else {
				result.Message = "In-app notification created"
			}

		case constants.NotificationChannelTelegram:
			err := s.sendTelegramNotification(ctx, alert, channel)
			result.Success = err == nil
			if err != nil {
				result.Error = err.Error()
				result.Message = "Failed to send telegram notification"
			} else {
				result.Message = "Telegram notification sent successfully"
			}

		default:
			result.Success = false
			result.Error = "Unsupported channel type"
			result.Message = "Unsupported notification channel"
		}

		results = append(results, result)
	}

	// Update alert with notification results
	if err := alert.SetNotificationResults(results); err != nil {
		return fmt.Errorf("failed to set notification results: %w", err)
	}

	// Update notification tracking
	alert.NotificationsSent = len(results)
	now := time.Now()
	alert.LastNotificationSent = &now

	// Save updated alert
	return s.alertRepo.UpdateUsageAlert(ctx, alert)
}

func (s *usageAlertService) TestNotificationChannel(ctx context.Context, req *interfaces.TestNotificationRequest) (*interfaces.TestNotificationResponse, error) {
	startTime := time.Now()

	// Create a test alert
	testAlert := &entities.UsageAlert{
		UserSubscriptionID: req.UserSubscriptionID,
		UsageType:          constants.UsageTypeTraffic,
		CurrentUsage:       8589934592,  // 8GB
		UsageLimit:         10737418240, // 10GB
		UsagePercent:       80.0,
		Severity:           constants.AlertSeverityWarning,
		Message:            req.TestMessage,
		FiredAt:            time.Now(),
	}

	if testAlert.Message == "" {
		testAlert.Message = "This is a test notification from your usage monitoring system."
	}

	// Test the notification channel
	err := s.SendNotification(ctx, testAlert, []entities.NotificationChannel{req.Channel})

	responseTime := time.Since(startTime)

	response := &interfaces.TestNotificationResponse{
		Success:      err == nil,
		ResponseTime: responseTime.String(),
	}

	if err != nil {
		response.Message = "Test notification failed"
		response.Error = err.Error()
	} else {
		response.Message = "Test notification sent successfully"
		response.Details = fmt.Sprintf("Notification sent to %s via %s", req.Channel.Target, req.Channel.Type)
	}

	return response, nil
}

func (s *usageAlertService) UpdateNotificationPreferences(ctx context.Context, req *interfaces.UpdateNotificationPrefsRequest) error {
	// TODO: Implement notification preferences update
	// This would update user-level notification settings
	// For now, return success
	return nil
}

// Alert Analytics

func (s *usageAlertService) GetAlertStatistics(ctx context.Context, req *interfaces.AlertStatsRequest) (*interfaces.AlertStatisticsResponse, error) {
	// Parse period (for future use with repository methods)
	_ = req.Period // Acknowledge parameter

	// For now, create empty responses since we don't have the extended repository methods
	// In a full implementation, these would use the repository methods
	summary := &interfaces.AlertsSummary{
		TotalAlerts:     0,
		ActiveAlerts:    0,
		ResolvedAlerts:  0,
		CriticalAlerts:  0,
		HighAlerts:      0,
		MediumAlerts:    0,
		LowAlerts:       0,
		AlertsToday:     0,
		AlertsThisWeek:  0,
		AlertsThisMonth: 0,
	}

	// Get trends - simplified implementation
	var trends []*interfaces.AlertTrendEntry

	// Get top configurations - simplified implementation
	var topConfigurations []*interfaces.TopAlertConfiguration

	// Build distribution (simplified)
	distribution := &interfaces.AlertDistribution{
		BySeverity:  make(map[string]int64),
		ByUsageType: make(map[string]int64),
		ByHour:      make(map[int]int64),
		ByDayOfWeek: make(map[int]int64),
		ByStatus:    make(map[string]int64),
	}

	return &interfaces.AlertStatisticsResponse{
		Summary:           summary,
		Trends:            trends,
		Distribution:      distribution,
		TopConfigurations: topConfigurations,
		Period:            req.Period,
		GeneratedAt:       time.Now(),
	}, nil
}

func (s *usageAlertService) GetAlertHistory(ctx context.Context, req *interfaces.AlertHistoryRequest) (*interfaces.AlertHistoryResponse, error) {
	// Set defaults
	if req.Limit == 0 {
		req.Limit = constants.DefaultPageSize
	}
	if req.Limit > constants.MaxPageSize {
		req.Limit = constants.MaxPageSize
	}

	// Build filter for alerts
	filter := interfaces.UsageAlertFilter{
		UserSubscriptionID:   req.UserSubscriptionID,
		AlertConfigurationID: req.AlertConfigurationID,
		UsageType:            req.UsageType,
		StartTime:            req.StartTime,
		EndTime:              req.EndTime,
		Limit:                req.Limit,
		Offset:               req.Offset,
		OrderBy:              "fired_at",
		OrderDirection:       "DESC",
		IncludeDeleted:       false,
	}

	if !req.IncludeResolved {
		isResolved := false
		filter.IsResolved = &isResolved
	}

	// Get alerts
	alerts, err := s.alertRepo.GetUsageAlerts(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get alert history: %w", err)
	}

	// Build history entries
	historyEntries := make([]*interfaces.AlertHistoryEntry, len(alerts))

	for i, alert := range alerts {
		entry := &interfaces.AlertHistoryEntry{
			Alert: alert.ToResponse(),
		}

		// Get configuration
		if config, err := s.alertRepo.GetAlertConfiguration(ctx, alert.AlertConfigurationID); err == nil {
			entry.Configuration = config.ToResponse()
		}

		// Add notification history if requested
		if req.IncludeNotifications {
			results := alert.GetNotificationResults()
			entry.NotificationHistory = make([]*entities.NotificationResult, len(results))
			for j, result := range results {
				entry.NotificationHistory[j] = &result
			}
		}

		// Calculate resolution time
		if alert.IsResolved() && alert.ResolvedAt != nil {
			resolutionTime := alert.ResolvedAt.Sub(alert.FiredAt).String()
			entry.ResolutionTime = &resolutionTime
		}

		historyEntries[i] = entry
	}

	// Calculate summary
	summary := s.calculateAlertHistorySummary(alerts)

	// Get total count for pagination
	totalFilter := filter
	totalFilter.Limit = 0
	totalFilter.Offset = 0
	allAlerts, err := s.alertRepo.GetUsageAlerts(ctx, totalFilter)
	if err != nil {
		return nil, fmt.Errorf("failed to get total count: %w", err)
	}
	totalCount := int64(len(allAlerts))

	// Build pagination info
	totalPages := int((totalCount + int64(req.Limit) - 1) / int64(req.Limit))
	currentPage := (req.Offset / req.Limit) + 1

	pagination := &interfaces.PaginationInfo{
		CurrentPage: currentPage,
		PageSize:    req.Limit,
		TotalPages:  totalPages,
		TotalCount:  totalCount,
		HasNext:     currentPage < totalPages,
		HasPrevious: currentPage > 1,
	}

	return &interfaces.AlertHistoryResponse{
		AlertHistory: historyEntries,
		Summary:      summary,
		Pagination:   pagination,
		TotalCount:   totalCount,
	}, nil
}

// Default Configurations

func (s *usageAlertService) CreateDefaultAlertConfigurations(ctx context.Context, subscriptionID uint) error {
	// Default alert configurations based on common thresholds
	defaultConfigs := []interfaces.CreateAlertConfigRequest{
		{
			UserSubscriptionID: subscriptionID,
			UsageType:          constants.UsageTypeTraffic,
			ThresholdType:      constants.ThresholdTypePercentage,
			Threshold:          50.0,
			Name:               "Traffic 50% Warning",
			Description:        "Alert when traffic usage reaches 50% of limit",
			Priority:           constants.PriorityLow,
			IsEnabled:          true,
			NotificationChannels: []entities.NotificationChannel{
				{
					Type:    constants.NotificationChannelInApp,
					Target:  "user",
					Enabled: true,
				},
			},
			CooldownMinutes: 1440, // 24 hours
		},
		{
			UserSubscriptionID: subscriptionID,
			UsageType:          constants.UsageTypeTraffic,
			ThresholdType:      constants.ThresholdTypePercentage,
			Threshold:          80.0,
			Name:               "Traffic 80% Alert",
			Description:        "Alert when traffic usage reaches 80% of limit",
			Priority:           constants.PriorityMedium,
			IsEnabled:          true,
			NotificationChannels: []entities.NotificationChannel{
				{
					Type:    constants.NotificationChannelEmail,
					Target:  "user",
					Enabled: true,
				},
				{
					Type:    constants.NotificationChannelInApp,
					Target:  "user",
					Enabled: true,
				},
			},
			CooldownMinutes: 720, // 12 hours
		},
		{
			UserSubscriptionID: subscriptionID,
			UsageType:          constants.UsageTypeTraffic,
			ThresholdType:      constants.ThresholdTypePercentage,
			Threshold:          90.0,
			Name:               "Traffic 90% Critical",
			Description:        "Critical alert when traffic usage reaches 90% of limit",
			Priority:           constants.PriorityHigh,
			IsEnabled:          true,
			NotificationChannels: []entities.NotificationChannel{
				{
					Type:    constants.NotificationChannelEmail,
					Target:  "user",
					Enabled: true,
				},
				{
					Type:    constants.NotificationChannelInApp,
					Target:  "user",
					Enabled: true,
				},
			},
			CooldownMinutes: 360, // 6 hours
		},
		{
			UserSubscriptionID: subscriptionID,
			UsageType:          constants.UsageTypeTraffic,
			ThresholdType:      constants.ThresholdTypePercentage,
			Threshold:          100.0,
			Name:               "Traffic Limit Exceeded",
			Description:        "Critical alert when traffic limit is exceeded",
			Priority:           constants.PriorityCritical,
			IsEnabled:          true,
			NotificationChannels: []entities.NotificationChannel{
				{
					Type:    constants.NotificationChannelEmail,
					Target:  "user",
					Enabled: true,
				},
				{
					Type:    constants.NotificationChannelInApp,
					Target:  "user",
					Enabled: true,
				},
			},
			CooldownMinutes: 60, // 1 hour
		},
	}

	// Create each default configuration
	for _, configReq := range defaultConfigs {
		_, err := s.CreateAlertConfiguration(ctx, &configReq)
		if err != nil {
			s.logger.Error("Failed to create default alert configuration",
				zap.String("config_name", configReq.Name),
				zap.Error(err))
		}
	}

	return nil
}

func (s *usageAlertService) CopyAlertConfigurationsFromPlan(ctx context.Context, subscriptionID uint, planID uint) error {
	// TODO: This would copy alert configurations from a subscription plan template
	// For now, just create default configurations
	return s.CreateDefaultAlertConfigurations(ctx, subscriptionID)
}

// Helper methods

func (s *usageAlertService) generateAlertMessage(config *entities.AlertConfiguration, currentUsage, usageLimit int64) string {
	if usageLimit == 0 {
		return fmt.Sprintf("Usage alert: %s usage has reached %d", config.UsageType, currentUsage)
	}

	usagePercent := float64(currentUsage) / float64(usageLimit) * 100
	return fmt.Sprintf("Usage alert: %s usage has reached %.1f%% (%d/%d) of your limit",
		config.UsageType, usagePercent, currentUsage, usageLimit)
}

func (s *usageAlertService) calculateAlertHistorySummary(alerts []*entities.UsageAlert) *interfaces.AlertHistorySummary {
	if len(alerts) == 0 {
		return &interfaces.AlertHistorySummary{}
	}

	var totalResolutionTime time.Duration
	var resolvedCount int64
	var unresolvedCount int64
	var notificationsSent int64
	var notificationFailures int64
	var fastestResolution time.Duration = time.Hour * 24 * 365 // Start with a large value
	var slowestResolution time.Duration

	for _, alert := range alerts {
		if alert.IsResolved() && alert.ResolvedAt != nil {
			resolvedCount++
			resolutionTime := alert.ResolvedAt.Sub(alert.FiredAt)
			totalResolutionTime += resolutionTime

			if resolutionTime < fastestResolution {
				fastestResolution = resolutionTime
			}
			if resolutionTime > slowestResolution {
				slowestResolution = resolutionTime
			}
		} else {
			unresolvedCount++
		}

		notificationsSent += int64(alert.NotificationsSent)

		// Count notification failures from results
		results := alert.GetNotificationResults()
		for _, result := range results {
			if !result.Success {
				notificationFailures++
			}
		}
	}

	var averageResolutionTime time.Duration
	if resolvedCount > 0 {
		averageResolutionTime = totalResolutionTime / time.Duration(resolvedCount)
	}

	if resolvedCount == 0 {
		fastestResolution = 0
	}

	return &interfaces.AlertHistorySummary{
		TotalAlerts:           int64(len(alerts)),
		AverageResolutionTime: averageResolutionTime.String(),
		FastestResolution:     fastestResolution.String(),
		SlowestResolution:     slowestResolution.String(),
		UnresolvedCount:       unresolvedCount,
		NotificationsSent:     notificationsSent,
		NotificationFailures:  notificationFailures,
	}
}

// Notification implementations (stubs for now)

func (s *usageAlertService) sendEmailNotification(ctx context.Context, alert *entities.UsageAlert, channel entities.NotificationChannel) error {
	// TODO: Implement email notification
	// This would integrate with an email service like SendGrid, AWS SES, etc.
	return nil
}

func (s *usageAlertService) sendWebhookNotification(ctx context.Context, alert *entities.UsageAlert, channel entities.NotificationChannel) error {
	// TODO: Implement webhook notification
	// This would send HTTP POST request to the webhook URL
	return nil
}

func (s *usageAlertService) sendInAppNotification(ctx context.Context, alert *entities.UsageAlert, channel entities.NotificationChannel) error {
	// TODO: Implement in-app notification
	// This would create a notification record in the database
	return nil
}

func (s *usageAlertService) sendTelegramNotification(ctx context.Context, alert *entities.UsageAlert, channel entities.NotificationChannel) error {
	// Parse chat ID from channel target
	chatIDStr := channel.Target
	chatID, err := strconv.ParseInt(chatIDStr, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid telegram chat ID: %s", chatIDStr)
	}

	// Create notification request using appropriate template
	template := "telegram_subscription_expired"
	if alert.Severity == constants.AlertSeverityCritical {
		template = "telegram_invoice_overdue" // Use urgent template for critical alerts
	}

	variables := map[string]string{
		"user_name":     fmt.Sprintf("User %d", alert.UserSubscriptionID), // TODO: Get actual user name
		"usage_type":    alert.UsageType,
		"current_usage": s.formatBytes(alert.CurrentUsage),
		"usage_limit":   s.formatBytes(alert.UsageLimit),
		"usage_percent": fmt.Sprintf("%.1f%%", alert.UsagePercent),
		"severity":      alert.Severity,
		"message":       alert.Message,
		"alert_id":      fmt.Sprintf("ALT-%d", alert.ID),
	}

	req := &notification.NotificationRequest{
		UserID:         alert.UserSubscriptionID,
		TelegramChatID: chatID,
		Channels:       []notification.NotificationChannel{notification.ChannelTelegram},
		Subject:        "Usage Alert",
		Template:       template,
		Variables:      variables,
	}

	results, err := s.notificationSvc.Send(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to send telegram notification: %w", err)
	}

	// Check if notification was successful
	for _, result := range results {
		if result.Channel == notification.ChannelTelegram && !result.Success {
			return fmt.Errorf("telegram notification failed: %s", result.Error)
		}
	}

	return nil
}

// Helper method to format bytes in human readable format
func (s *usageAlertService) formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// Helper method to create basic alerts summary from a list of alerts
func (s *usageAlertService) createBasicAlertsSummary(alerts []*entities.UsageAlert) *interfaces.AlertsSummary {
	if len(alerts) == 0 {
		return &interfaces.AlertsSummary{
			TotalAlerts:     0,
			ActiveAlerts:    0,
			ResolvedAlerts:  0,
			CriticalAlerts:  0,
			HighAlerts:      0,
			MediumAlerts:    0,
			LowAlerts:       0,
			AlertsToday:     0,
			AlertsThisWeek:  0,
			AlertsThisMonth: 0,
		}
	}

	summary := &interfaces.AlertsSummary{
		TotalAlerts: int64(len(alerts)),
	}

	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	weekStart := todayStart.AddDate(0, 0, -int(now.Weekday()))
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	for _, alert := range alerts {
		// Count by status
		if alert.IsActive() {
			summary.ActiveAlerts++
		}
		if alert.IsResolved() {
			summary.ResolvedAlerts++
		}

		// Count by severity
		switch alert.Severity {
		case constants.AlertSeverityCritical:
			summary.CriticalAlerts++
		case constants.AlertSeverityError:
			summary.HighAlerts++
		case constants.AlertSeverityWarning:
			summary.MediumAlerts++
		case constants.AlertSeverityInfo:
			summary.LowAlerts++
		}

		// Count by time period
		if alert.FiredAt.After(todayStart) {
			summary.AlertsToday++
		}
		if alert.FiredAt.After(weekStart) {
			summary.AlertsThisWeek++
		}
		if alert.FiredAt.After(monthStart) {
			summary.AlertsThisMonth++
		}
	}

	return summary
}
