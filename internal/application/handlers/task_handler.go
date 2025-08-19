package handlers

import (
	"fmt"
	"time"

	"linke/internal/shared/queue"
	"linke/internal/shared/response"

	"github.com/gin-gonic/gin"
)

// CreateTaskRequest represents the request body for creating a task
type CreateTaskRequest struct {
	Type    string                 `json:"type" binding:"required" example:"email"`
	Payload map[string]interface{} `json:"payload" binding:"required" swaggertype:"object"`
}

// TaskCreatedResponse represents the response after creating a task
type TaskCreatedResponse struct {
	TaskID  string `json:"task_id" example:"task_123456"`
	Message string `json:"message" example:"Task enqueued successfully"`
}

// QueueStatusResponse represents the response for queue status
type QueueStatusResponse struct {
	QueueName    string `json:"queue_name" example:"default"`
	PendingTasks int    `json:"pending_tasks" example:"5"`
	ProcessingTasks int `json:"processing_tasks" example:"2"`
	CompletedTasks int  `json:"completed_tasks" example:"100"`
	FailedTasks    int  `json:"failed_tasks" example:"3"`
}

type TaskHandler struct {
	taskQueue *queue.TaskQueue
}

func NewTaskHandler(taskQueue *queue.TaskQueue) *TaskHandler {
	return &TaskHandler{
		taskQueue: taskQueue,
	}
}

// @Summary Create a new task
// @Description Create and enqueue a new task
// @Tags System-Task
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param task body CreateTaskRequest true "Task details"
// @Success 201 {object} TaskCreatedResponse
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /tasks [post]
func (h *TaskHandler) CreateTask(c *gin.Context) {
	var req CreateTaskRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	task := &queue.Task{
		ID:       fmt.Sprintf("task-%d", time.Now().UnixNano()),
		Type:     req.Type,
		Payload:  req.Payload,
		Retry:    0,
		MaxRetry: 3,
	}

	if err := h.taskQueue.Enqueue(c.Request.Context(), "default", task); err != nil {
		response.InternalServerError(c, "Failed to enqueue task")
		return
	}

	response.CreatedWithMessage(c, "Task enqueued successfully", gin.H{
		"task_id": task.ID,
	})
}

// @Summary Get queue status
// @Description Get the current status of the task queue
// @Tags System-Task
// @Produce json
// @Security BearerAuth
// @Success 200 {object} QueueStatusResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /tasks/status [get]
func (h *TaskHandler) GetQueueStatus(c *gin.Context) {
	info, err := h.taskQueue.GetQueueInfo(c.Request.Context(), "default")
	if err != nil {
		response.InternalServerError(c, "Failed to get queue info")
		return
	}

	response.Success(c, info)
}
