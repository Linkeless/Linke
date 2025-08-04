package versioning

import (
	"time"

	"go.uber.org/fx"

	"linke/internal/shared/config"
)

// Module 提供版本管理系统依赖
var Module = fx.Module("versioning",
	// 提供版本配置
	fx.Provide(NewVersionConfigFromConfig),

	// 提供版本中间件
	fx.Provide(NewVersionMiddleware),

	// Provide migration registry
	fx.Provide(NewMigrationRegistry),

	// Provide response migrator
	fx.Provide(NewResponseMigrator),

	// Provide auto migration builder
	fx.Provide(NewAutoMigrationBuilder),
)

// NewVersionConfigFromConfig 从主配置创建版本配置
func NewVersionConfigFromConfig(cfg *config.Config) VersionConfig {
	// 解析版本
	defaultVersion, _ := ParseVersion(cfg.Versioning.DefaultVersion)
	minVersion, _ := ParseVersion(cfg.Versioning.MinVersion)
	maxVersion, _ := ParseVersion(cfg.Versioning.MaxVersion)

	// Parse strategy
	var strategy VersioningStrategy
	switch cfg.Versioning.Strategy {
	case "header":
		strategy = HeaderStrategy
	case "query":
		strategy = QueryStrategy
	case "content_type":
		strategy = ContentTypeStrategy
	default:
		strategy = URLPathStrategy
	}

	// Parse sunset date
	var sunsetDate *time.Time
	if cfg.Versioning.SunsetV1Date != "" {
		if parsed, err := time.Parse(time.RFC3339, cfg.Versioning.SunsetV1Date); err == nil {
			sunsetDate = &parsed
		}
	}

	// 创建版本信息
	now := time.Now()
	supportedVersions := []VersionInfo{
		{
			Version:     NewVersion(1, 0, 0),
			Status:      "deprecated",
			SunsetDate:  sunsetDate,
			Description: "Initial API version (deprecated)",
			Released:    now.AddDate(0, -12, 0), // Released 12 months ago
		},
		{
			Version:     NewVersion(2, 0, 0),
			Status:      "active",
			Description: "Enhanced API with improved features",
			Released:    now.AddDate(0, -6, 0), // Released 6 months ago
		},
	}

	return VersionConfig{
		Strategy:                 strategy,
		DefaultVersion:           defaultVersion,
		MinVersion:               minVersion,
		MaxVersion:               maxVersion,
		SupportedVersions:        supportedVersions,
		HeaderName:               cfg.Versioning.HeaderName,
		QueryParam:               cfg.Versioning.QueryParam,
		URLPrefix:                cfg.Versioning.URLPrefix,
		EnableDeprecationHeaders: cfg.Versioning.EnableDeprecationHeaders,
		EnableAutoMigration:      cfg.Versioning.EnableAutoMigration,
	}
}
