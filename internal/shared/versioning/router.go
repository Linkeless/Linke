package versioning

import (
	"fmt"
	"net/http"
	"sort"

	"github.com/gin-gonic/gin"

	"linke/internal/shared/logger"
	"linke/internal/shared/response"
)

// HandlerFunc 表示版本感知的处理器函数
type HandlerFunc func(c *gin.Context)

// VersionedHandler 表示为特定版本注册的处理器
type VersionedHandler struct {
	Version Version
	Handler HandlerFunc
	Method  string
	Path    string
}

// VersionRouter 封装 gin.RouterGroup 以提供版本感知路由
type VersionRouter struct {
	router   *gin.RouterGroup
	config   VersionConfig
	handlers map[string][]VersionedHandler // key: method:path
	logger   logger.Logger
}

// NewVersionRouter 创建新的版本感知路由器
func NewVersionRouter(router *gin.RouterGroup, config VersionConfig, log logger.Logger) *VersionRouter {
	return &VersionRouter{
		router:   router,
		config:   config,
		handlers: make(map[string][]VersionedHandler),
		logger:   log,
	}
}

// handlerKey generates a key for storing handlers
func (vr *VersionRouter) handlerKey(method, path string) string {
	return fmt.Sprintf("%s:%s", method, path)
}

// RegisterHandler 为特定版本注册处理器
func (vr *VersionRouter) RegisterHandler(method, path string, version Version, handler HandlerFunc) {
	key := vr.handlerKey(method, path)

	versionedHandler := VersionedHandler{
		Version: version,
		Handler: handler,
		Method:  method,
		Path:    path,
	}

	vr.handlers[key] = append(vr.handlers[key], versionedHandler)

	// 按版本排序处理器（最新的在前）以支持回退逻辑
	sort.Slice(vr.handlers[key], func(i, j int) bool {
		return vr.handlers[key][i].Version.Compare(vr.handlers[key][j].Version) > 0
	})

	vr.logger.Debug("Registered versioned handler",
		logger.String("method", method),
		logger.String("path", path),
		logger.String("version", version.String()),
	)
}

// BuildRoutes builds the actual Gin routes
func (vr *VersionRouter) BuildRoutes() {
	for _, handlers := range vr.handlers {
		if len(handlers) == 0 {
			continue
		}

		// Create a version dispatch handler for this route
		dispatchHandler := vr.createVersionDispatchHandler(handlers)

		// Extract method and path from key
		method := handlers[0].Method
		path := handlers[0].Path

		// Register the route with Gin
		switch method {
		case "GET":
			vr.router.GET(path, dispatchHandler)
		case "POST":
			vr.router.POST(path, dispatchHandler)
		case "PUT":
			vr.router.PUT(path, dispatchHandler)
		case "PATCH":
			vr.router.PATCH(path, dispatchHandler)
		case "DELETE":
			vr.router.DELETE(path, dispatchHandler)
		case "HEAD":
			vr.router.HEAD(path, dispatchHandler)
		case "OPTIONS":
			vr.router.OPTIONS(path, dispatchHandler)
		default:
			vr.logger.Warn("Unsupported HTTP method for versioned route",
				logger.String("method", method),
				logger.String("path", path),
			)
		}

		vr.logger.Info("Built versioned route",
			logger.String("method", method),
			logger.String("path", path),
			logger.Int("version_count", len(handlers)),
		)
	}
}

// createVersionDispatchHandler creates a handler that dispatches to the correct version
func (vr *VersionRouter) createVersionDispatchHandler(handlers []VersionedHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		versionCtx, exists := GetVersionFromContext(c)
		if !exists {
			vr.logger.Error("No version context found in request")
			response.ErrorJSON(c, http.StatusInternalServerError, response.ErrorResponse{
				Error:   "version_context_missing",
				Message: "Version context not found in request",
			})
			return
		}

		requestedVersion := versionCtx.ResolvedVersion

		// Find exact version match first
		for _, handler := range handlers {
			if handler.Version.Compare(requestedVersion) == 0 {
				vr.logger.Debug("Exact version match found",
					logger.String("version", requestedVersion.String()),
					logger.String("path", c.Request.URL.Path),
				)
				handler.Handler(c)
				return
			}
		}

		// If auto-migration is enabled, find compatible version
		if vr.config.EnableAutoMigration {
			compatibleHandler := vr.findCompatibleHandler(handlers, requestedVersion)
			if compatibleHandler != nil {
				vr.logger.Info("Using compatible version handler",
					logger.String("requested_version", requestedVersion.String()),
					logger.String("compatible_version", compatibleHandler.Version.String()),
					logger.String("path", c.Request.URL.Path),
				)

				// Add header to indicate version migration
				c.Header("X-API-Version-Migrated-From", compatibleHandler.Version.String())
				c.Header("X-API-Version-Migrated-To", requestedVersion.String())

				compatibleHandler.Handler(c)
				return
			}
		}

		// No compatible handler found
		vr.handleNoCompatibleVersion(c, requestedVersion, handlers)
	}
}

// findCompatibleHandler finds a compatible version handler
func (vr *VersionRouter) findCompatibleHandler(handlers []VersionedHandler, requestedVersion Version) *VersionedHandler {
	// Look for backward compatible versions (same major, lower or equal minor)
	for _, handler := range handlers {
		if requestedVersion.IsCompatibleWith(handler.Version) {
			return &handler
		}
	}

	// If no backward compatible version, try the latest available version
	// that's lower than requested (for graceful degradation)
	for _, handler := range handlers {
		if handler.Version.Compare(requestedVersion) < 0 {
			return &handler
		}
	}

	return nil
}

// handleNoCompatibleVersion handles cases where no compatible version is found
func (vr *VersionRouter) handleNoCompatibleVersion(c *gin.Context, requestedVersion Version, handlers []VersionedHandler) {
	vr.logger.Warn("No compatible version handler found",
		logger.String("requested_version", requestedVersion.String()),
		logger.String("path", c.Request.URL.Path),
		logger.String("method", c.Request.Method),
	)

	availableVersions := make([]string, len(handlers))
	for i, handler := range handlers {
		availableVersions[i] = handler.Version.String()
	}

	errorResponse := response.ErrorResponse{
		Error:   "version_not_implemented",
		Message: fmt.Sprintf("Version %s is not implemented for this endpoint", requestedVersion.String()),
		Details: map[string]any{
			"requested_version":  requestedVersion.String(),
			"available_versions": availableVersions,
			"endpoint":           fmt.Sprintf("%s %s", c.Request.Method, c.Request.URL.Path),
			"migration_required": true,
		},
	}

	response.ErrorJSON(c, http.StatusNotImplemented, errorResponse)
}

// Convenience methods for registering handlers
func (vr *VersionRouter) GET(path string, version Version, handler HandlerFunc) {
	vr.RegisterHandler("GET", path, version, handler)
}

func (vr *VersionRouter) POST(path string, version Version, handler HandlerFunc) {
	vr.RegisterHandler("POST", path, version, handler)
}

func (vr *VersionRouter) PUT(path string, version Version, handler HandlerFunc) {
	vr.RegisterHandler("PUT", path, version, handler)
}

func (vr *VersionRouter) PATCH(path string, version Version, handler HandlerFunc) {
	vr.RegisterHandler("PATCH", path, version, handler)
}

func (vr *VersionRouter) DELETE(path string, version Version, handler HandlerFunc) {
	vr.RegisterHandler("DELETE", path, version, handler)
}

// Group creates a new version router group with a path prefix
func (vr *VersionRouter) Group(relativePath string) *VersionRouter {
	return NewVersionRouter(vr.router.Group(relativePath), vr.config, vr.logger)
}

// Use adds middleware to the version router
func (vr *VersionRouter) Use(middleware ...gin.HandlerFunc) gin.IRoutes {
	return vr.router.Use(middleware...)
}

// GetVersionInfo returns information about registered handlers
func (vr *VersionRouter) GetVersionInfo() map[string]any {
	routeInfo := make(map[string][]map[string]any)

	for key, handlers := range vr.handlers {
		handlerInfo := make([]map[string]any, len(handlers))
		for i, handler := range handlers {
			handlerInfo[i] = map[string]any{
				"version": handler.Version.String(),
				"method":  handler.Method,
				"path":    handler.Path,
			}
		}
		routeInfo[key] = handlerInfo
	}

	return map[string]any{
		"routes":       routeInfo,
		"total_routes": len(vr.handlers),
		"config":       vr.config,
	}
}

// VersionedHandlerWrapper wraps existing handlers to make them version-aware
type VersionedHandlerWrapper struct {
	config VersionConfig
	logger logger.Logger
}

// NewVersionedHandlerWrapper creates a new handler wrapper
func NewVersionedHandlerWrapper(config VersionConfig, log logger.Logger) *VersionedHandlerWrapper {
	return &VersionedHandlerWrapper{
		config: config,
		logger: log,
	}
}

// WrapHandler wraps an existing gin.HandlerFunc to be version-aware
func (vhw *VersionedHandlerWrapper) WrapHandler(handler gin.HandlerFunc, supportedVersions ...Version) gin.HandlerFunc {
	return func(c *gin.Context) {
		versionCtx, exists := GetVersionFromContext(c)
		if !exists {
			vhw.logger.Error("No version context found in wrapped handler")
			response.ErrorJSON(c, http.StatusInternalServerError, response.ErrorResponse{
				Error:   "version_context_missing",
				Message: "Version context not found in request",
			})
			return
		}

		// Check if handler supports the requested version
		if len(supportedVersions) > 0 {
			supported := false
			for _, version := range supportedVersions {
				if version.Compare(versionCtx.ResolvedVersion) == 0 {
					supported = true
					break
				}
			}

			if !supported {
				errorResponse := response.ErrorResponse{
					Error:   "version_not_supported",
					Message: fmt.Sprintf("This endpoint does not support version %s", versionCtx.ResolvedVersion.String()),
					Details: map[string]any{
						"requested_version":  versionCtx.ResolvedVersion.String(),
						"supported_versions": supportedVersions,
					},
				}
				response.ErrorJSON(c, http.StatusNotImplemented, errorResponse)
				return
			}
		}

		// Call the original handler
		handler(c)
	}
}
