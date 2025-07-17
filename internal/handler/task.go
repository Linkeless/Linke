package handler

import (
	"fmt"
	"time"

	"linke/internal/queue"
	"linke/internal/response"

	"github.com/gin-gonic/gin"
)

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
// @Tags tasks
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param task body object true "Task details"
// @Success 201 {object} response.StandardResponse
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /tasks [post]
func (h *TaskHandler) CreateTask(c *gin.Context) {
	var req struct {
		Type    string                 `json:"type" binding:"required"`
		Payload map[string]interface{} `json:"payload" binding:"required"`
	}

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
// @Tags tasks
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.StandardResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /tasks/status [get]
func (h *TaskHandler) GetQueueStatus(c *gin.Context) {
	length, err := h.taskQueue.GetQueueLength(c.Request.Context(), "default")
	if err != nil {
		response.InternalServerError(c, "Failed to get queue length")
		return
	}

	deadLength, err := h.taskQueue.GetQueueLength(c.Request.Context(), "default_dead")
	if err != nil {
		response.InternalServerError(c, "Failed to get dead queue length")
		return
	}

	response.Success(c, gin.H{
		"queue_length":      length,
		"dead_queue_length": deadLength,
	})
}