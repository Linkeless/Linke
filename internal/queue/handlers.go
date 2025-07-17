package queue

import (
	"context"
	"fmt"
	"time"

	"linke/internal/logger"
)

func EmailTaskHandler(ctx context.Context, task *Task) error {
	to, ok := task.Payload["to"].(string)
	if !ok {
		return fmt.Errorf("missing or invalid 'to' field in email task")
	}

	subject, ok := task.Payload["subject"].(string)
	if !ok {
		return fmt.Errorf("missing or invalid 'subject' field in email task")
	}

	body, ok := task.Payload["body"].(string)
	if !ok {
		return fmt.Errorf("missing or invalid 'body' field in email task")
	}

	logger.Info("Sending email",
		logger.String("to", to),
		logger.String("subject", subject),
		logger.String("body", body),
		logger.String("task_id", task.ID),
	)
	
	time.Sleep(2 * time.Second)
	
	logger.Info("Email sent successfully",
		logger.String("to", to),
		logger.String("task_id", task.ID),
	)
	return nil
}

func NotificationTaskHandler(ctx context.Context, task *Task) error {
	userID, ok := task.Payload["user_id"].(string)
	if !ok {
		return fmt.Errorf("missing or invalid 'user_id' field in notification task")
	}

	message, ok := task.Payload["message"].(string)
	if !ok {
		return fmt.Errorf("missing or invalid 'message' field in notification task")
	}

	logger.Info("Sending notification",
		logger.String("user_id", userID),
		logger.String("message", message),
		logger.String("task_id", task.ID),
	)
	
	time.Sleep(1 * time.Second)
	
	logger.Info("Notification sent successfully",
		logger.String("user_id", userID),
		logger.String("task_id", task.ID),
	)
	return nil
}

func DataProcessingTaskHandler(ctx context.Context, task *Task) error {
	dataType, ok := task.Payload["data_type"].(string)
	if !ok {
		return fmt.Errorf("missing or invalid 'data_type' field in data processing task")
	}

	logger.Info("Processing data",
		logger.String("data_type", dataType),
		logger.String("task_id", task.ID),
	)
	
	time.Sleep(5 * time.Second)
	
	logger.Info("Data processing completed",
		logger.String("data_type", dataType),
		logger.String("task_id", task.ID),
	)
	return nil
}