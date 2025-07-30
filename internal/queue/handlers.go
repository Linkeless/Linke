package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"linke/internal/logger"

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