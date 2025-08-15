package handlers

import (
	"strconv"

	"github.com/gin-gonic/gin"

	userEntities "linke/internal/domains/user/entities"
	"linke/internal/shared/constants"
	"linke/internal/shared/middleware"
	"linke/internal/shared/response"
)

// ParseIDParam 解析路径参数中的ID
func ParseIDParam(c *gin.Context, paramName string) (uint, bool) {
	idStr := c.Param(paramName)
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid "+paramName)
		return 0, false
	}
	return uint(id), true
}

// BindJSONRequest 绑定JSON请求并处理错误
func BindJSONRequest(c *gin.Context, req any) bool {
	if err := c.ShouldBindJSON(req); err != nil {
		response.BadRequest(c, constants.ErrInvalidRequestData)
		return false
	}
	return true
}

// GetCurrentUser 从上下文中获取当前用户
func GetCurrentUser(c *gin.Context) (*userEntities.User, bool) {
	userValue, exists := c.Get(middleware.AuthContextKey)
	if !exists {
		response.Unauthorized(c, constants.ErrAuthenticationRequired)
		return nil, false
	}

	user, ok := userValue.(*userEntities.User)
	if !ok {
		response.Unauthorized(c, constants.ErrInvalidUserContext)
		return nil, false
	}

	return user, true
}

// PaginationParams 分页参数结构
type PaginationParams struct {
	Limit  int `form:"limit" binding:"omitempty,min=1,max=100"`
	Offset int `form:"offset" binding:"omitempty,min=0"`
}

// SetDefaults 设置分页参数默认值
func (p *PaginationParams) SetDefaults() {
	if p.Limit == 0 {
		p.Limit = constants.DefaultPageLimit
	}
	if p.Limit > constants.MaxPageLimit {
		p.Limit = constants.MaxPageLimit
	}
}

// ParsePagination 解析分页参数
func ParsePagination(c *gin.Context) PaginationParams {
	var params PaginationParams
	c.ShouldBindQuery(&params)
	params.SetDefaults()
	return params
}
