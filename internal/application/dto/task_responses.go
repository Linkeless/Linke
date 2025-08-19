package dto

import "time"

// TaskResponse represents a task response
type TaskResponse struct {
	ID          string                 `json:"id" example:"task_123456"`
	Type        string                 `json:"type" example:"email_notification"`
	Status      string                 `json:"status" example:"queued"`
	Priority    int                    `json:"priority" example:"1"`
	Payload     map[string]interface{} `json:"payload,omitempty"`
	CreatedAt   time.Time              `json:"created_at" example:"2024-01-15T10:30:00Z"`
	ProcessedAt *time.Time             `json:"processed_at,omitempty" example:"2024-01-15T10:31:00Z"`
	CompletedAt *time.Time             `json:"completed_at,omitempty" example:"2024-01-15T10:32:00Z"`
	Error       string                 `json:"error,omitempty" example:"Failed to send email"`
	Attempts    int                    `json:"attempts" example:"1"`
	MaxAttempts int                    `json:"max_attempts" example:"3"`
}

// QueueStatusResponse represents the status of a task queue
type QueueStatusResponse struct {
	QueueName    string            `json:"queue_name" example:"default"`
	TotalTasks   int64             `json:"total_tasks" example:"1250"`
	PendingTasks int64             `json:"pending_tasks" example:"45"`
	ActiveTasks  int64             `json:"active_tasks" example:"5"`
	CompletedTasks int64           `json:"completed_tasks" example:"1200"`
	FailedTasks  int64             `json:"failed_tasks" example:"15"`
	RetryTasks   int64             `json:"retry_tasks" example:"3"`
	Workers      int               `json:"workers" example:"10"`
	ActiveWorkers int              `json:"active_workers" example:"5"`
	IsHealthy    bool              `json:"is_healthy" example:"true"`
	LastProcessed *time.Time       `json:"last_processed,omitempty" example:"2024-01-15T10:35:00Z"`
	Throughput   ThroughputStats   `json:"throughput"`
	RecentErrors []RecentError     `json:"recent_errors,omitempty"`
	UpdatedAt    time.Time         `json:"updated_at" example:"2024-01-15T10:36:00Z"`
}

// ThroughputStats represents queue throughput statistics
type ThroughputStats struct {
	TasksPerMinute  float64 `json:"tasks_per_minute" example:"25.5"`
	TasksPerHour    float64 `json:"tasks_per_hour" example:"1530"`
	AverageTaskTime float64 `json:"average_task_time" example:"2.3"`
	PeakThroughput  float64 `json:"peak_throughput" example:"45.2"`
}

// RecentError represents a recent error in the queue
type RecentError struct {
	TaskID    string    `json:"task_id" example:"task_789123"`
	Error     string    `json:"error" example:"Connection timeout"`
	Timestamp time.Time `json:"timestamp" example:"2024-01-15T10:30:00Z"`
	TaskType  string    `json:"task_type" example:"email_notification"`
}