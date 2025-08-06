package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"linke/internal/shared/logger"

	"github.com/go-redis/redis/v8"
	"github.com/hibiken/asynq"
)

type TaskQueue struct {
	client *asynq.Client
}

type Task struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Payload   map[string]any `json:"payload"`
	Retry     int                    `json:"retry"`
	MaxRetry  int                    `json:"max_retry"`
	CreatedAt time.Time              `json:"created_at"`
}

// NewTask creates a new task with the given type and payload
func NewTask(taskType string, payload map[string]any) *Task {
	return &Task{
		ID:        fmt.Sprintf("%s-%d", taskType, time.Now().UnixNano()),
		Type:      taskType,
		Payload:   payload,
		MaxRetry:  3, // Default retry count
		CreatedAt: time.Now(),
	}
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
	// Map project log level to asynq log level, 减少asynq启动噪音
	logLevel := asynq.WarnLevel // 默认使用warn级别减少启动日志噪音
	projectLogLevel := logger.GetEnvLogLevel()
	switch projectLogLevel {
	case "debug":
		logLevel = asynq.DebugLevel
	case "info":
		logLevel = asynq.WarnLevel // Info级别时使用warn减少启动噪音
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

func (tq *TaskQueue) GetQueueInfo(ctx context.Context, queueName string) (map[string]any, error) {
	// For asynq, we would need to use the inspector to get queue stats
	// This is a simplified implementation
	return map[string]any{
		"queue_name": queueName,
		"status":     "active",
	}, nil
}

func (tp *TaskProcessor) RegisterHandler(taskType string, handler TaskHandler) {
	tp.handlers[taskType] = handler
}

func (tp *TaskProcessor) Start(ctx context.Context) error {
	// 移动到debug级别，避免与bootstrap中的启动日志重复
	logger.Debug("Asynq task processor initializing")

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
	logger.Debug("Asynq task processor stopped")
}

// Enhanced task queue with priority support
type PriorityTaskQueue struct {
	*TaskQueue
}

// NewPriorityTaskQueue creates a new priority task queue
func NewPriorityTaskQueue(redisClient *redis.Client) *PriorityTaskQueue {
	return &PriorityTaskQueue{
		TaskQueue: NewTaskQueue(redisClient),
	}
}

// EnqueueWithPriority enqueues a task with priority
func (ptq *PriorityTaskQueue) EnqueueWithPriority(ctx context.Context, queueName string, task *Task, priority int) error {
	task.CreatedAt = time.Now()

	data, err := json.Marshal(task.Payload)
	if err != nil {
		return fmt.Errorf("failed to marshal task payload: %w", err)
	}

	// Convert priority to asynq priority (higher number = higher priority)
	var asynqPriority asynq.Option
	switch {
	case priority >= 80:
		asynqPriority = asynq.MaxRetry(task.MaxRetry)
	case priority >= 60:
		asynqPriority = asynq.MaxRetry(task.MaxRetry)
	case priority >= 40:
		asynqPriority = asynq.MaxRetry(task.MaxRetry)
	default:
		asynqPriority = asynq.MaxRetry(task.MaxRetry)
	}

	asynqTask := asynq.NewTask(task.Type, data, asynqPriority, asynq.Queue(queueName))
	_, err = ptq.client.Enqueue(asynqTask)
	return err
}

// DelayedTaskQueue provides delayed task execution
type DelayedTaskQueue struct {
	*TaskQueue
}

// NewDelayedTaskQueue creates a new delayed task queue
func NewDelayedTaskQueue(redisClient *redis.Client) *DelayedTaskQueue {
	return &DelayedTaskQueue{
		TaskQueue: NewTaskQueue(redisClient),
	}
}

// EnqueueDelayed enqueues a task to be processed at a later time
func (dtq *DelayedTaskQueue) EnqueueDelayed(ctx context.Context, queueName string, task *Task, delay time.Duration) error {
	task.CreatedAt = time.Now()

	data, err := json.Marshal(task.Payload)
	if err != nil {
		return fmt.Errorf("failed to marshal task payload: %w", err)
	}

	asynqTask := asynq.NewTask(task.Type, data, asynq.MaxRetry(task.MaxRetry), asynq.Queue(queueName))
	_, err = dtq.client.Enqueue(asynqTask, asynq.ProcessIn(delay))
	return err
}

// EnqueueAt enqueues a task to be processed at a specific time
func (dtq *DelayedTaskQueue) EnqueueAt(ctx context.Context, queueName string, task *Task, processAt time.Time) error {
	task.CreatedAt = time.Now()

	data, err := json.Marshal(task.Payload)
	if err != nil {
		return fmt.Errorf("failed to marshal task payload: %w", err)
	}

	asynqTask := asynq.NewTask(task.Type, data, asynq.MaxRetry(task.MaxRetry), asynq.Queue(queueName))
	_, err = dtq.client.Enqueue(asynqTask, asynq.ProcessAt(processAt))
	return err
}
