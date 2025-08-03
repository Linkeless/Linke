package versioning

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// Version represents an API version
type Version struct {
	Major int `json:"major"`
	Minor int `json:"minor"`
	Patch int `json:"patch"`
}

// NewVersion creates a new version from major, minor, patch integers
func NewVersion(major, minor, patch int) Version {
	return Version{
		Major: major,
		Minor: minor,
		Patch: patch,
	}
}

// ParseVersion parses a version string like "1.0.0" or "v1" into a Version
func ParseVersion(versionStr string) (Version, error) {
	// Remove 'v' prefix if present
	versionStr = strings.TrimPrefix(versionStr, "v")
	
	// Handle simple version like "1" -> "1.0.0"
	if !strings.Contains(versionStr, ".") {
		major, err := strconv.Atoi(versionStr)
		if err != nil {
			return Version{}, fmt.Errorf("invalid version format: %s", versionStr)
		}
		return NewVersion(major, 0, 0), nil
	}
	
	parts := strings.Split(versionStr, ".")
	if len(parts) < 2 || len(parts) > 3 {
		return Version{}, fmt.Errorf("invalid version format: %s", versionStr)
	}
	
	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return Version{}, fmt.Errorf("invalid major version: %s", parts[0])
	}
	
	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		return Version{}, fmt.Errorf("invalid minor version: %s", parts[1])
	}
	
	patch := 0
	if len(parts) == 3 {
		patch, err = strconv.Atoi(parts[2])
		if err != nil {
			return Version{}, fmt.Errorf("invalid patch version: %s", parts[2])
		}
	}
	
	return NewVersion(major, minor, patch), nil
}

// String returns the string representation of the version
func (v Version) String() string {
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
}

// ShortString returns the short string representation (e.g., "v1")
func (v Version) ShortString() string {
	return fmt.Sprintf("v%d", v.Major)
}

// Compare compares two versions. Returns:
// -1 if v < other
//  0 if v == other
//  1 if v > other
func (v Version) Compare(other Version) int {
	if v.Major != other.Major {
		if v.Major < other.Major {
			return -1
		}
		return 1
	}
	
	if v.Minor != other.Minor {
		if v.Minor < other.Minor {
			return -1
		}
		return 1
	}
	
	if v.Patch != other.Patch {
		if v.Patch < other.Patch {
			return -1
		}
		return 1
	}
	
	return 0
}

// IsCompatibleWith checks if this version is compatible with another version
// Compatible means same major version and this minor >= other minor
func (v Version) IsCompatibleWith(other Version) bool {
	return v.Major == other.Major && v.Minor >= other.Minor
}

// VersioningStrategy defines how API version is determined from request
type VersioningStrategy string

const (
	// URLPathStrategy 从 URL 路径中提取版本，如 /api/v1/users
	URLPathStrategy VersioningStrategy = "url_path"
	// HeaderStrategy extracts version from HTTP header like X-API-Version: v1
	HeaderStrategy VersioningStrategy = "header"
	// QueryStrategy extracts version from query parameter like ?version=v1
	QueryStrategy VersioningStrategy = "query"
	// ContentTypeStrategy 从 Accept 头部中提取版本，如 application/vnd.api+json;version=1
	ContentTypeStrategy VersioningStrategy = "content_type"
)

// VersionExtractor extracts version from HTTP request
type VersionExtractor interface {
	ExtractVersion(c *gin.Context) (Version, error)
}

// VersionInfo contains metadata about a version
type VersionInfo struct {
	Version     Version    `json:"version"`
	Status      string     `json:"status"`      // active, deprecated, sunset
	SunsetDate  *time.Time `json:"sunset_date,omitempty"`
	Description string     `json:"description"`
	Released    time.Time  `json:"released"`
}

// IsDeprecated returns true if the version is deprecated
func (vi VersionInfo) IsDeprecated() bool {
	return vi.Status == "deprecated" || vi.Status == "sunset"
}

// IsSunset returns true if the version is past its sunset date
func (vi VersionInfo) IsSunset() bool {
	if vi.Status != "sunset" || vi.SunsetDate == nil {
		return false
	}
	return time.Now().After(*vi.SunsetDate)
}

// DaysUntilSunset returns the number of days until sunset, or -1 if no sunset date
func (vi VersionInfo) DaysUntilSunset() int {
	if vi.SunsetDate == nil {
		return -1
	}
	duration := time.Until(*vi.SunsetDate)
	return int(duration.Hours() / 24)
}

// VersionConfig contains configuration for the versioning system
type VersionConfig struct {
	// Strategy defines how version is extracted from requests
	Strategy VersioningStrategy `json:"strategy"`
	
	// DefaultVersion is used when no version is specified
	DefaultVersion Version `json:"default_version"`
	
	// SupportedVersions lists all supported versions with their metadata
	SupportedVersions []VersionInfo `json:"supported_versions"`
	
	// MinVersion is the minimum supported version
	MinVersion Version `json:"min_version"`
	
	// MaxVersion is the maximum supported version
	MaxVersion Version `json:"max_version"`
	
	// HeaderName is the header name for header strategy (default: "X-API-Version")
	HeaderName string `json:"header_name,omitempty"`
	
	// QueryParam is the query parameter name for query strategy (default: "version")
	QueryParam string `json:"query_param,omitempty"`
	
	// URLPrefix is the URL prefix for path strategy (default: "/api")
	URLPrefix string `json:"url_prefix,omitempty"`
	
	// EnableDeprecationHeaders adds deprecation headers to responses
	EnableDeprecationHeaders bool `json:"enable_deprecation_headers"`
	
	// EnableAutoMigration enables automatic version migration for compatible versions
	EnableAutoMigration bool `json:"enable_auto_migration"`
}

// DefaultVersionConfig returns a default configuration
func DefaultVersionConfig() VersionConfig {
	now := time.Now()
	v1 := NewVersion(1, 0, 0)
	v2 := NewVersion(2, 0, 0)
	
	return VersionConfig{
		Strategy:       URLPathStrategy,
		DefaultVersion: v1,
		MinVersion:     v1,
		MaxVersion:     v2,
		SupportedVersions: []VersionInfo{
			{
				Version:     v1,
				Status:      "active",
				Description: "Initial API version",
				Released:    now.AddDate(0, -6, 0), // Released 6 months ago
			},
			{
				Version:     v2,
				Status:      "active",
				Description: "Enhanced API with improved features",
				Released:    now,
			},
		},
		HeaderName:               "X-API-Version",
		QueryParam:               "version",
		URLPrefix:                "/api",
		EnableDeprecationHeaders: true,
		EnableAutoMigration:      false,
	}
}

// GetVersionInfo returns version info for a specific version
func (vc VersionConfig) GetVersionInfo(version Version) *VersionInfo {
	for _, vi := range vc.SupportedVersions {
		if vi.Version.Compare(version) == 0 {
			return &vi
		}
	}
	return nil
}

// IsVersionSupported checks if a version is supported
func (vc VersionConfig) IsVersionSupported(version Version) bool {
	return vc.GetVersionInfo(version) != nil &&
		   version.Compare(vc.MinVersion) >= 0 &&
		   version.Compare(vc.MaxVersion) <= 0
}

// GetLatestVersion returns the latest supported version
func (vc VersionConfig) GetLatestVersion() Version {
	latest := vc.MinVersion
	for _, vi := range vc.SupportedVersions {
		if vi.Version.Compare(latest) > 0 {
			latest = vi.Version
		}
	}
	return latest
}

// VersionContext represents version information in the request context
type VersionContext struct {
	RequestedVersion Version     `json:"requested_version"`
	ResolvedVersion  Version     `json:"resolved_version"`
	VersionInfo      VersionInfo `json:"version_info"`
	Strategy         VersioningStrategy `json:"strategy"`
}

// Context keys for storing version information
const (
	VersionContextKey = "api_version"
)

// GetVersionFromContext extracts version context from gin context
func GetVersionFromContext(c *gin.Context) (*VersionContext, bool) {
	if value, exists := c.Get(VersionContextKey); exists {
		if versionCtx, ok := value.(*VersionContext); ok {
			return versionCtx, true
		}
	}
	return nil, false
}

// SetVersionInContext sets version context in gin context
func SetVersionInContext(c *gin.Context, versionCtx *VersionContext) {
	c.Set(VersionContextKey, versionCtx)
}

// GetVersionFromGoContext extracts version context from Go context
func GetVersionFromGoContext(ctx context.Context) (*VersionContext, bool) {
	if value := ctx.Value(VersionContextKey); value != nil {
		if versionCtx, ok := value.(*VersionContext); ok {
			return versionCtx, true
		}
	}
	return nil, false
}

// SetVersionInGoContext sets version context in Go context
func SetVersionInGoContext(ctx context.Context, versionCtx *VersionContext) context.Context {
	return context.WithValue(ctx, VersionContextKey, versionCtx)
}