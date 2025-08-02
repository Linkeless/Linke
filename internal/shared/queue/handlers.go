package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"linke/internal/shared/logger"

	"github.com/hibiken/asynq"
)

func EmailTaskHandler(ctx context.Context, task *asynq.Task) error {
	var payload map[string]interface{}
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal task payload: %w", err)
	}

	to, ok := payload["to"].(string)
	if !ok {
		return fmt.Errorf("missing or invalid 'to' field in email task")
	}

	subject, ok := payload["subject"].(string)
	if !ok {
		return fmt.Errorf("missing or invalid 'subject' field in email task")
	}

	body, ok := payload["body"].(string)
	if !ok {
		return fmt.Errorf("missing or invalid 'body' field in email task")
	}

	logger.Info("Sending email",
		logger.String("to", to),
		logger.String("subject", subject),
		logger.String("body", body),
		logger.String("task_type", task.Type()),
	)

	time.Sleep(2 * time.Second)

	logger.Info("Email sent successfully",
		logger.String("to", to),
		logger.String("task_type", task.Type()),
	)
	return nil
}

func NotificationTaskHandler(ctx context.Context, task *asynq.Task) error {
	var payload map[string]interface{}
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal task payload: %w", err)
	}

	userID, ok := payload["user_id"].(string)
	if !ok {
		return fmt.Errorf("missing or invalid 'user_id' field in notification task")
	}

	message, ok := payload["message"].(string)
	if !ok {
		return fmt.Errorf("missing or invalid 'message' field in notification task")
	}

	logger.Info("Sending notification",
		logger.String("user_id", userID),
		logger.String("message", message),
		logger.String("task_type", task.Type()),
	)

	time.Sleep(1 * time.Second)

	logger.Info("Notification sent successfully",
		logger.String("user_id", userID),
		logger.String("task_type", task.Type()),
	)
	return nil
}

func DataProcessingTaskHandler(ctx context.Context, task *asynq.Task) error {
	var payload map[string]interface{}
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal task payload: %w", err)
	}

	dataType, ok := payload["data_type"].(string)
	if !ok {
		return fmt.Errorf("missing or invalid 'data_type' field in data processing task")
	}

	logger.Info("Processing data",
		logger.String("data_type", dataType),
		logger.String("task_type", task.Type()),
	)

	time.Sleep(5 * time.Second)

	logger.Info("Data processing completed",
		logger.String("data_type", dataType),
		logger.String("task_type", task.Type()),
	)
	return nil
}

// Enhanced handlers with domain-specific functionality

// PaymentProcessingTaskHandler handles payment processing tasks
func PaymentProcessingTaskHandler(ctx context.Context, task *asynq.Task) error {
	var payload map[string]interface{}
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payment task payload: %w", err)
	}

	paymentID, ok := payload["payment_id"].(string)
	if !ok {
		return fmt.Errorf("missing or invalid 'payment_id' field in payment task")
	}

	amount, ok := payload["amount"].(float64)
	if !ok {
		return fmt.Errorf("missing or invalid 'amount' field in payment task")
	}

	logger.Info("Processing payment",
		logger.String("payment_id", paymentID),
		logger.Any("amount", amount),
		logger.String("task_type", task.Type()),
	)

	// Simulate payment processing
	time.Sleep(3 * time.Second)

	logger.Info("Payment processed successfully",
		logger.String("payment_id", paymentID),
		logger.String("task_type", task.Type()),
	)
	return nil
}

// SubscriptionExpiryTaskHandler handles subscription expiry notifications
func SubscriptionExpiryTaskHandler(ctx context.Context, task *asynq.Task) error {
	var payload map[string]interface{}
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal subscription expiry task payload: %w", err)
	}

	userID, ok := payload["user_id"].(string)
	if !ok {
		return fmt.Errorf("missing or invalid 'user_id' field in subscription expiry task")
	}

	subscriptionID, ok := payload["subscription_id"].(string)
	if !ok {
		return fmt.Errorf("missing or invalid 'subscription_id' field in subscription expiry task")
	}

	logger.Info("Processing subscription expiry",
		logger.String("user_id", userID),
		logger.String("subscription_id", subscriptionID),
		logger.String("task_type", task.Type()),
	)

	// Simulate subscription expiry processing
	time.Sleep(2 * time.Second)

	logger.Info("Subscription expiry processed successfully",
		logger.String("user_id", userID),
		logger.String("task_type", task.Type()),
	)
	return nil
}

// InvoiceGenerationTaskHandler handles invoice generation
func InvoiceGenerationTaskHandler(ctx context.Context, task *asynq.Task) error {
	var payload map[string]interface{}
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal invoice generation task payload: %w", err)
	}

	orderID, ok := payload["order_id"].(string)
	if !ok {
		return fmt.Errorf("missing or invalid 'order_id' field in invoice generation task")
	}

	userID, ok := payload["user_id"].(string)
	if !ok {
		return fmt.Errorf("missing or invalid 'user_id' field in invoice generation task")
	}

	logger.Info("Generating invoice",
		logger.String("order_id", orderID),
		logger.String("user_id", userID),
		logger.String("task_type", task.Type()),
	)

	// Simulate invoice generation
	time.Sleep(4 * time.Second)

	logger.Info("Invoice generated successfully",
		logger.String("order_id", orderID),
		logger.String("task_type", task.Type()),
	)
	return nil
}

// ServerHealthCheckTaskHandler handles server health checks
func ServerHealthCheckTaskHandler(ctx context.Context, task *asynq.Task) error {
	var payload map[string]interface{}
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal server health check task payload: %w", err)
	}

	serverID, ok := payload["server_id"].(string)
	if !ok {
		return fmt.Errorf("missing or invalid 'server_id' field in server health check task")
	}

	logger.Info("Performing server health check",
		logger.String("server_id", serverID),
		logger.String("task_type", task.Type()),
	)

	// Simulate health check
	time.Sleep(1 * time.Second)

	logger.Info("Server health check completed",
		logger.String("server_id", serverID),
		logger.String("task_type", task.Type()),
	)
	return nil
}

// ReferralProcessingTaskHandler handles referral processing
func ReferralProcessingTaskHandler(ctx context.Context, task *asynq.Task) error {
	var payload map[string]interface{}
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal referral processing task payload: %w", err)
	}

	referrerID, ok := payload["referrer_id"].(string)
	if !ok {
		return fmt.Errorf("missing or invalid 'referrer_id' field in referral processing task")
	}

	refereeID, ok := payload["referee_id"].(string)
	if !ok {
		return fmt.Errorf("missing or invalid 'referee_id' field in referral processing task")
	}

	logger.Info("Processing referral",
		logger.String("referrer_id", referrerID),
		logger.String("referee_id", refereeID),
		logger.String("task_type", task.Type()),
	)

	// Simulate referral processing
	time.Sleep(2 * time.Second)

	logger.Info("Referral processed successfully",
		logger.String("referrer_id", referrerID),
		logger.String("task_type", task.Type()),
	)
	return nil
}

// Task type constants
const (
	TaskTypeEmail              = "email:send"
	TaskTypeNotification       = "notification:send"
	TaskTypeDataProcessing     = "data:process"
	TaskTypePaymentProcessing  = "payment:process"
	TaskTypeSubscriptionExpiry = "subscription:expiry"
	TaskTypeInvoiceGeneration  = "invoice:generate"
	TaskTypeServerHealthCheck  = "server:health_check"
	TaskTypeReferralProcessing = "referral:process"
)

// RegisterDefaultHandlers registers all default task handlers
func RegisterDefaultHandlers(processor *TaskProcessor) {
	processor.RegisterHandler(TaskTypeEmail, EmailTaskHandler)
	processor.RegisterHandler(TaskTypeNotification, NotificationTaskHandler)
	processor.RegisterHandler(TaskTypeDataProcessing, DataProcessingTaskHandler)
	processor.RegisterHandler(TaskTypePaymentProcessing, PaymentProcessingTaskHandler)
	processor.RegisterHandler(TaskTypeSubscriptionExpiry, SubscriptionExpiryTaskHandler)
	processor.RegisterHandler(TaskTypeInvoiceGeneration, InvoiceGenerationTaskHandler)
	processor.RegisterHandler(TaskTypeServerHealthCheck, ServerHealthCheckTaskHandler)
	processor.RegisterHandler(TaskTypeReferralProcessing, ReferralProcessingTaskHandler)
}
