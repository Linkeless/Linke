package handlers

import (
	"net/http"

	"linke/internal/application/services"
	"linke/internal/shared/logger"
	"linke/internal/shared/response"

	"github.com/gin-gonic/gin"
)

// ApplicationHandler 应用级处理器
// 处理跨领域的 HTTP 请求
type ApplicationHandler struct {
	appService *services.ApplicationService
	logger     logger.Logger
}

// NewApplicationHandler 创建应用处理器
func NewApplicationHandler(
	appService *services.ApplicationService,
	logger logger.Logger,
) *ApplicationHandler {
	return &ApplicationHandler{
		appService: appService,
		logger:     logger,
	}
}

// HealthCheck 系统健康检查接口
// @Summary 系统健康检查
// @Description 检查系统各组件健康状态
// @Tags System-Health
// @Accept json
// @Produce json
// @Success 200 {object} response.StandardResponse{data=dto.HealthCheckResponse}
// @Router /app/system/health [get]
func (h *ApplicationHandler) HealthCheck(c *gin.Context) {
	ctx := c.Request.Context()

	health := h.appService.HealthCheck(ctx)

	response.SuccessJSON(c, http.StatusOK, health)
}
