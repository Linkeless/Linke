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
// @Tags Task-System
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param task body CreateTaskRequest true "Task details"
// @Success 201 {object} response.StandardResponse
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
// @Tags Task-System
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.StandardResponse
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
