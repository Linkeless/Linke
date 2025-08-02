package handlers

import (
	"net/http"

	"linke/internal/application/services"
	"linke/internal/shared/logger"
	"linke/internal/shared/response"

	"github.com/gin-gonic/gin"
)

// SimpleApplicationHandler 简化的应用级处理器
// 当领域模块暂时不可用时使用
type SimpleApplicationHandler struct {
	appService *services.SimpleApplicationService
	logger     logger.Logger
}

// NewSimpleApplicationHandler 创建简化的应用处理器
func NewSimpleApplicationHandler(
	appService *services.SimpleApplicationService,
	logger logger.Logger,
) *SimpleApplicationHandler {
	return &SimpleApplicationHandler{
		appService: appService,
		logger:     logger,
	}
}

// HealthCheck 系统健康检查接口
// @Summary 系统健康检查
// @Description 检查系统各组件健康状态
// @Tags system
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "健康检查结果"
// @Router /api/v1/app/system/health [get]
func (h *SimpleApplicationHandler) HealthCheck(c *gin.Context) {
	ctx := c.Request.Context()
	
	health := h.appService.HealthCheck(ctx)
	
	response.SuccessJSON(c, http.StatusOK, health)
}