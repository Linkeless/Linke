package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"linke/internal/logger"

	"github.com/go-redis/redis/v8"
)

type TaskQueue struct {
	client *redis.Client
}

type Task struct {
	ID      string                 `json:"id"`
	Type    string                 `json:"type"`
	Payload map[string]interface{} `json:"payload"`
	Retry   int                    `json:"retry"`
	MaxRetry int                   `json:"max_retry"`
	CreatedAt time.Time            `json:"created_at"`
}

type TaskHandler func(ctx context.Context, task *Task) error

type TaskProcessor struct {
	queue    *TaskQueue
	handlers map[string]TaskHandler
}

func NewTaskQueue(client *redis.Client) *TaskQueue {
	return &TaskQueue{
		client: client,
	}
}

func NewTaskProcessor(queue *TaskQueue) *TaskProcessor {
	return &TaskProcessor{
		queue:    queue,
		handlers: make(map[string]TaskHandler),
	}
}

func (tq *TaskQueue) Enqueue(ctx context.Context, queueName string, task *Task) error {
	task.CreatedAt = time.Now()
	
	data, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("failed to marshal task: %w", err)
	}

	return tq.client.LPush(ctx, queueName, data).Err()
}

func (tq *TaskQueue) Dequeue(ctx context.Context, queueName string, timeout time.Duration) (*Task, error) {
	result, err := tq.client.BRPop(ctx, timeout, queueName).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to dequeue task: %w", err)
	}

	if len(result) < 2 {
		return nil, fmt.Errorf("invalid redis response")
	}

	var task Task
	if err := json.Unmarshal([]byte(result[1]), &task); err != nil {
		return nil, fmt.Errorf("failed to unmarshal task: %w", err)
	}

	return &task, nil
}

func (tq *TaskQueue) GetQueueLength(ctx context.Context, queueName string) (int64, error) {
	return tq.client.LLen(ctx, queueName).Result()
}

func (tp *TaskProcessor) RegisterHandler(taskType string, handler TaskHandler) {
	tp.handlers[taskType] = handler
}

func (tp *TaskProcessor) ProcessTasks(ctx context.Context, queueName string) {
	logger.Info("Starting task processor", logger.String("queue", queueName))
	
	for {
		select {
		case <-ctx.Done():
			logger.Info("Task processor stopped", logger.String("queue", queueName))
			return
		default:
			task, err := tp.queue.Dequeue(ctx, queueName, 5*time.Second)
			if err != nil {
				logger.Error("Error dequeuing task", 
					logger.String("queue", queueName),
					logger.Error2("error", err),
				)
				continue
			}

			if task == nil {
				continue
			}

			if err := tp.processTask(ctx, queueName, task); err != nil {
				logger.Error("Error processing task",
					logger.String("task_id", task.ID),
					logger.String("queue", queueName),
					logger.Error2("error", err),
				)
			}
		}
	}
}

func (tp *TaskProcessor) processTask(ctx context.Context, queueName string, task *Task) error {
	handler, exists := tp.handlers[task.Type]
	if !exists {
		return fmt.Errorf("no handler registered for task type: %s", task.Type)
	}

	logger.Info("Processing task",
		logger.String("task_id", task.ID),
		logger.String("task_type", task.Type),
	)

	if err := handler(ctx, task); err != nil {
		task.Retry++
		if task.Retry < task.MaxRetry {
			logger.Warn("Task failed, retrying",
			logger.String("task_id", task.ID),
			logger.Int("retry", task.Retry),
			logger.Int("max_retry", task.MaxRetry),
		)
			return tp.queue.Enqueue(ctx, queueName, task)
		}
		
		logger.Error("Task failed after max retries, moving to dead letter queue",
			logger.String("task_id", task.ID),
			logger.Int("max_retry", task.MaxRetry),
		)
		return tp.queue.Enqueue(ctx, queueName+"_dead", task)
	}

	logger.Info("Task completed successfully",
		logger.String("task_id", task.ID),
	)
	return nil
}