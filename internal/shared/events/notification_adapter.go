package events

import (
	"context"

	"linke/internal/shared/notification"
)

// NotificationServiceAdapter adapts the shared notification service to the events notification interface
type NotificationServiceAdapter struct {
	service notification.NotificationService
}

// NewNotificationServiceAdapter creates a new adapter
func NewNotificationServiceAdapter(service notification.NotificationService) NotificationService {
	return &NotificationServiceAdapter{
		service: service,
	}
}

// Send adapts the notification request and delegates to the real notification service
func (a *NotificationServiceAdapter) Send(ctx context.Context, req *NotificationRequest) ([]*NotificationResult, error) {
	// Convert events.NotificationRequest to notification.NotificationRequest
	channels := make([]notification.NotificationChannel, 0, len(req.Channels))
	for _, ch := range req.Channels {
		switch ch {
		case "email":
			channels = append(channels, notification.ChannelEmail)
		case "sms":
			channels = append(channels, notification.ChannelSMS)
		case "push":
			channels = append(channels, notification.ChannelPush)
		case "telegram":
			channels = append(channels, notification.ChannelTelegram)
		case "slack":
			channels = append(channels, notification.ChannelSlack)
		}
	}

	notificationReq := &notification.NotificationRequest{
		UserID:    req.UserID,
		Email:     req.Email,
		Phone:     req.Phone,
		Channels:  channels,
		Subject:   req.Subject,
		Body:      req.Body,
		Template:  req.Template,
		Variables: req.Variables,
		Priority:  a.parsePriority(req.Priority),
		EventType: req.EventType,
		EventID:   req.EventID,
	}

	// Call the real notification service
	results, err := a.service.Send(ctx, notificationReq)
	if err != nil {
		return nil, err
	}

	// Convert notification.NotificationResult to events.NotificationResult
	eventResults := make([]*NotificationResult, len(results))
	for i, result := range results {
		eventResults[i] = &NotificationResult{
			Channel:   string(result.Channel),
			Success:   result.Success,
			Error:     result.Error,
			MessageID: result.MessageID,
			SentAt:    result.SentAt,
		}
	}

	return eventResults, nil
}

// parsePriority converts a string priority to NotificationPriority enum
func (a *NotificationServiceAdapter) parsePriority(priority string) notification.NotificationPriority {
	switch priority {
	case "low":
		return notification.PriorityLow
	case "normal":
		return notification.PriorityNormal
	case "high":
		return notification.PriorityHigh
	case "urgent", "critical":
		return notification.PriorityUrgent
	default:
		return notification.PriorityNormal
	}
}
