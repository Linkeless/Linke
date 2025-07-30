package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"linke/internal/logger"

	"github.com/hibiken/asynq"
	"github.com/go-redis/redis/v8"
)

type TaskQueue struct {
	client *asynq.Client
}

type Task struct {
	ID      string                 `json:"id"`
	Type    string                 `json:"type"`
	Payload map[string]interface{} `json:"payload"`
	Retry   int                    `json:"retry"`
	MaxRetry int                   `json:"max_retry"`
	CreatedAt time.Time            `json:"created_at"`
}

type TaskHandler func(ctx context.Context, task *asynq.Task) error

type TaskProcessor struct {
	server   *asynq.Server
	handlers map[string]TaskHandler
}

func NewTaskQueue(redisClient *redis.Client) *TaskQueue {
	asynqClient := asynq.NewClient(asynq.RedisClientOpt{
		Addr:     redisClient.Options().Addr,
		Password: redisClient.Options().Password,
		DB:       redisClient.Options().DB,
	})
	return &TaskQueue{
		client: asynqClient,
	}
}

func NewTaskProcessor(redisClient *redis.Client) *TaskProcessor {
	// Map project log level to asynq log level
	logLevel := asynq.InfoLevel
	projectLogLevel := logger.GetEnvLogLevel()
	switch projectLogLevel {
	case "debug":
		logLevel = asynq.DebugLevel
	case "info":
		logLevel = asynq.InfoLevel
	case "warn":
		logLevel = asynq.WarnLevel
	case "error":
		logLevel = asynq.ErrorLevel
	case "fatal":
		logLevel = asynq.FatalLevel
	}

	asynqServer := asynq.NewServer(
		asynq.RedisClientOpt{
			Addr:     redisClient.Options().Addr,
			Password: redisClient.Options().Password,
			DB:       redisClient.Options().DB,
		},
		asynq.Config{
			Concurrency: 10,
			Logger:      logger.NewAsynqLogger(),
			LogLevel:    logLevel,
			RetryDelayFunc: func(n int, err error, task *asynq.Task) time.Duration {
				return time.Duration(n) * time.Second
			},
		},
	)
	return &TaskProcessor{
		server:   asynqServer,
		handlers: make(map[string]TaskHandler),
	}
}

func (tq *TaskQueue) Enqueue(ctx context.Context, queueName string, task *Task) error {
	task.CreatedAt = time.Now()
	
	data, err := json.Marshal(task.Payload)
	if err != nil {
		return fmt.Errorf("failed to marshal task payload: %w", err)
	}

	asynqTask := asynq.NewTask(task.Type, data, asynq.MaxRetry(task.MaxRetry), asynq.Queue(queueName))
	_, err = tq.client.Enqueue(asynqTask)
	return err
}

func (tq *TaskQueue) GetClient() *asynq.Client {
	return tq.client
}

func (tq *TaskQueue) Close() error {
	return tq.client.Close()
}

func (tq *TaskQueue) GetQueueInfo(ctx context.Context, queueName string) (map[string]interface{}, error) {
	// For asynq, we would need to use the inspector to get queue stats
	// This is a simplified implementation
	return map[string]interface{}{
		"queue_name": queueName,
		"status": "active",
	}, nil
}

func (tp *TaskProcessor) RegisterHandler(taskType string, handler TaskHandler) {
	tp.handlers[taskType] = handler
}

func (tp *TaskProcessor) Start(ctx context.Context) error {
	logger.Info("Starting asynq task processor")
	
	mux := asynq.NewServeMux()
	
	// Register all handlers
	for taskType, handler := range tp.handlers {
		mux.HandleFunc(taskType, func(ctx context.Context, task *asynq.Task) error {
			logger.Info("Processing task",
				logger.String("task_type", task.Type()),
			)
			
			err := handler(ctx, task)
			if err != nil {
				logger.Error("Task failed",
					logger.String("task_type", task.Type()),
					logger.Error2("error", err),
				)
				return err
			}
			
			logger.Info("Task completed successfully",
				logger.String("task_type", task.Type()),
			)
			return nil
		})
	}
	
	return tp.server.Run(mux)
}

func (tp *TaskProcessor) Stop() {
	logger.Info("Stopping asynq task processor")
	tp.server.Stop()
	logger.Info("Asynq task processor stopped")
}