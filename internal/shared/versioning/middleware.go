package versioning

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"linke/internal/shared/logger"
	"linke/internal/shared/response"
)

// VersionMiddleware 处理 API 版本协商
type VersionMiddleware struct {
	config    VersionConfig
	extractor VersionExtractor
	logger    logger.Logger
}

// NewVersionMiddleware 创建新的版本中间件
func NewVersionMiddleware(config VersionConfig, log logger.Logger) *VersionMiddleware {
	extractor := CreateExtractor(config)

	return &VersionMiddleware{
		config:    config,
		extractor: extractor,
		logger:    log,
	}
}

// Middleware returns the Gin middleware function
func (vm *VersionMiddleware) Middleware() gin.HandlerFunc {
	return vm.VersionNegotiation
}

// VersionNegotiation performs version negotiation for incoming requests
func (vm *VersionMiddleware) VersionNegotiation(c *gin.Context) {
	startTime := time.Now()

	// Extract version from request
	requestedVersion, err := vm.extractor.ExtractVersion(c)
	var resolvedVersion Version
	var versionInfo *VersionInfo

	if err != nil {
		// Use default version if extraction fails
		resolvedVersion = vm.config.DefaultVersion
		versionInfo = vm.config.GetVersionInfo(resolvedVersion)

		vm.logger.Debug("Using default version due to extraction failure",
			zap.String("path", c.Request.URL.Path),
			zap.String("method", c.Request.Method),
			zap.String("error", err.Error()),
			zap.String("default_version", resolvedVersion.String()),
		)
	} else {
		// Validate requested version
		if !vm.config.IsVersionSupported(requestedVersion) {
			vm.handleUnsupportedVersion(c, requestedVersion)
			return
		}

		// Check if version is sunset
		versionInfo = vm.config.GetVersionInfo(requestedVersion)
		if versionInfo != nil && versionInfo.IsSunset() {
			vm.handleSunsetVersion(c, requestedVersion, *versionInfo)
			return
		}

		resolvedVersion = requestedVersion

		vm.logger.Debug("Version negotiation successful",
			zap.String("path", c.Request.URL.Path),
			zap.String("method", c.Request.Method),
			zap.String("requested_version", requestedVersion.String()),
			zap.String("resolved_version", resolvedVersion.String()),
			zap.String("strategy", string(vm.config.Strategy)),
		)
	}

	// Create version context
	versionCtx := &VersionContext{
		RequestedVersion: requestedVersion,
		ResolvedVersion:  resolvedVersion,
		Strategy:         vm.config.Strategy,
	}

	if versionInfo != nil {
		versionCtx.VersionInfo = *versionInfo
	}

	// Set version in context
	SetVersionInContext(c, versionCtx)

	// Add version headers to response
	vm.addVersionHeaders(c, versionCtx)

	// Add deprecation headers if needed
	if vm.config.EnableDeprecationHeaders && versionInfo != nil && versionInfo.IsDeprecated() {
		vm.addDeprecationHeaders(c, *versionInfo)
	}

	// Log version negotiation metrics
	duration := time.Since(startTime)
	vm.logger.Debug("Version negotiation completed",
		zap.String("resolved_version", resolvedVersion.String()),
		zap.Duration("duration", duration),
	)

	c.Next()
}

// handleUnsupportedVersion handles requests for unsupported versions
func (vm *VersionMiddleware) handleUnsupportedVersion(c *gin.Context, requestedVersion Version) {
	vm.logger.Warn("Unsupported API version requested",
		zap.String("path", c.Request.URL.Path),
		zap.String("method", c.Request.Method),
		zap.String("requested_version", requestedVersion.String()),
		zap.String("min_version", vm.config.MinVersion.String()),
		zap.String("max_version", vm.config.MaxVersion.String()),
		zap.String("client_ip", c.ClientIP()),
		zap.String("user_agent", c.GetHeader("User-Agent")),
	)

	supportedVersions := make([]string, len(vm.config.SupportedVersions))
	for i, vi := range vm.config.SupportedVersions {
		supportedVersions[i] = vi.Version.String()
	}

	errorResponse := response.ErrorResponse{
		Error:   "unsupported_api_version",
		Message: fmt.Sprintf("API version %s is not supported", requestedVersion.String()),
		Details: map[string]any{
			"requested_version":  requestedVersion.String(),
			"supported_versions": supportedVersions,
			"min_version":        vm.config.MinVersion.String(),
			"max_version":        vm.config.MaxVersion.String(),
			"latest_version":     vm.config.GetLatestVersion().String(),
		},
	}

	// Add recommendation header
	c.Header("X-API-Recommendation", fmt.Sprintf("Use version %s", vm.config.GetLatestVersion().String()))

	response.ErrorJSON(c, http.StatusBadRequest, errorResponse)
	c.Abort()
}

// handleSunsetVersion handles requests for sunset versions
func (vm *VersionMiddleware) handleSunsetVersion(c *gin.Context, requestedVersion Version, versionInfo VersionInfo) {
	vm.logger.Warn("Sunset API version requested",
		zap.String("path", c.Request.URL.Path),
		zap.String("method", c.Request.Method),
		zap.String("requested_version", requestedVersion.String()),
		zap.Time("sunset_date", *versionInfo.SunsetDate),
		zap.String("client_ip", c.ClientIP()),
		zap.String("user_agent", c.GetHeader("User-Agent")),
	)

	errorResponse := response.ErrorResponse{
		Error:   "api_version_sunset",
		Message: fmt.Sprintf("API version %s has been sunset and is no longer available", requestedVersion.String()),
		Details: map[string]any{
			"requested_version": requestedVersion.String(),
			"sunset_date":       versionInfo.SunsetDate.Format(time.RFC3339),
			"latest_version":    vm.config.GetLatestVersion().String(),
			"migration_guide":   fmt.Sprintf("Please migrate to version %s", vm.config.GetLatestVersion().String()),
		},
	}

	// Add migration recommendation header
	c.Header("X-API-Migration", fmt.Sprintf("Migrate to version %s", vm.config.GetLatestVersion().String()))

	response.ErrorJSON(c, http.StatusGone, errorResponse)
	c.Abort()
}

// addVersionHeaders adds version-related headers to the response
func (vm *VersionMiddleware) addVersionHeaders(c *gin.Context, versionCtx *VersionContext) {
	c.Header("X-API-Version", versionCtx.ResolvedVersion.String())
	c.Header("X-API-Version-Strategy", string(versionCtx.Strategy))

	if versionCtx.RequestedVersion.String() != "" {
		c.Header("X-API-Version-Requested", versionCtx.RequestedVersion.String())
	}

	// Add latest version info
	latestVersion := vm.config.GetLatestVersion()
	c.Header("X-API-Version-Latest", latestVersion.String())

	// Add supported versions
	supportedVersions := make([]string, len(vm.config.SupportedVersions))
	for i, vi := range vm.config.SupportedVersions {
		supportedVersions[i] = vi.Version.String()
	}
	c.Header("X-API-Versions-Supported", fmt.Sprintf("%v", supportedVersions))
}

// addDeprecationHeaders adds deprecation headers for deprecated versions
func (vm *VersionMiddleware) addDeprecationHeaders(c *gin.Context, versionInfo VersionInfo) {
	// Add deprecation warning
	c.Header("Warning", fmt.Sprintf(`299 - "API version %s is deprecated"`, versionInfo.Version.String()))

	// Add sunset header if sunset date is set
	if versionInfo.SunsetDate != nil {
		c.Header("Sunset", versionInfo.SunsetDate.Format(time.RFC1123))

		// Add days until sunset
		daysUntilSunset := versionInfo.DaysUntilSunset()
		if daysUntilSunset >= 0 {
			c.Header("X-API-Sunset-Days", strconv.Itoa(daysUntilSunset))
		}
	}

	// Add link to migration guide
	latestVersion := vm.config.GetLatestVersion()
	c.Header("Link", fmt.Sprintf(`<%s>; rel="successor-version"`, latestVersion.String()))

	vm.logger.Info("Deprecation headers added",
		zap.String("deprecated_version", versionInfo.Version.String()),
		zap.String("status", versionInfo.Status),
		zap.String("path", c.Request.URL.Path),
	)
}

// VersionInfo returns version information endpoint handler
func (vm *VersionMiddleware) VersionInfo() gin.HandlerFunc {
	return func(c *gin.Context) {
		versionInfo := map[string]any{
			"current_version":    vm.config.DefaultVersion.String(),
			"latest_version":     vm.config.GetLatestVersion().String(),
			"min_version":        vm.config.MinVersion.String(),
			"max_version":        vm.config.MaxVersion.String(),
			"supported_versions": vm.config.SupportedVersions,
			"strategy":           vm.config.Strategy,
			"deprecation_policy": map[string]any{
				"enable_deprecation_headers": vm.config.EnableDeprecationHeaders,
				"enable_auto_migration":      vm.config.EnableAutoMigration,
			},
		}

		c.JSON(http.StatusOK, response.SuccessResponse{
			Message: "API version information",
			Data:    versionInfo,
		})
	}
}

// HealthCheck returns a health check handler that includes version info
func (vm *VersionMiddleware) HealthCheck() gin.HandlerFunc {
	return func(c *gin.Context) {
		versionCtx, _ := GetVersionFromContext(c)

		health := map[string]any{
			"status":  "healthy",
			"service": "linke-api",
		}

		if versionCtx != nil {
			health["version"] = versionCtx.ResolvedVersion.String()
			health["version_info"] = versionCtx.VersionInfo
		} else {
			health["version"] = vm.config.DefaultVersion.String()
		}

		c.JSON(http.StatusOK, health)
	}
}
