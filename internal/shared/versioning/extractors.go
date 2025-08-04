package versioning

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
)

// URLPathExtractor extracts version from URL path
type URLPathExtractor struct {
	URLPrefix string // e.g., "/api"
}

// NewURLPathExtractor creates a new URL path extractor
func NewURLPathExtractor(urlPrefix string) *URLPathExtractor {
	if urlPrefix == "" {
		urlPrefix = "/api"
	}
	return &URLPathExtractor{
		URLPrefix: urlPrefix,
	}
}

// ExtractVersion 从 URL 路径中提取版本，如 /api/v1/users
func (e *URLPathExtractor) ExtractVersion(c *gin.Context) (Version, error) {
	path := c.Request.URL.Path

	// Remove the URL prefix
	if e.URLPrefix != "" && strings.HasPrefix(path, e.URLPrefix) {
		path = strings.TrimPrefix(path, e.URLPrefix)
	}

	// Remove leading slash
	path = strings.TrimPrefix(path, "/")

	// Extract version from path like "v1/users" or "v2.1/users"
	versionPattern := regexp.MustCompile(`^v(\d+(?:\.\d+)?(?:\.\d+)?)/`)
	matches := versionPattern.FindStringSubmatch(path)

	if len(matches) < 2 {
		return Version{}, fmt.Errorf("no version found in URL path: %s", c.Request.URL.Path)
	}

	return ParseVersion(matches[1])
}

// HeaderExtractor extracts version from HTTP header
type HeaderExtractor struct {
	HeaderName string // e.g., "X-API-Version"
}

// NewHeaderExtractor creates a new header extractor
func NewHeaderExtractor(headerName string) *HeaderExtractor {
	if headerName == "" {
		headerName = "X-API-Version"
	}
	return &HeaderExtractor{
		HeaderName: headerName,
	}
}

// ExtractVersion extracts version from HTTP header
func (e *HeaderExtractor) ExtractVersion(c *gin.Context) (Version, error) {
	versionStr := c.GetHeader(e.HeaderName)
	if versionStr == "" {
		return Version{}, fmt.Errorf("version header %s not found", e.HeaderName)
	}

	return ParseVersion(versionStr)
}

// QueryExtractor extracts version from query parameter
type QueryExtractor struct {
	QueryParam string // e.g., "version"
}

// NewQueryExtractor creates a new query extractor
func NewQueryExtractor(queryParam string) *QueryExtractor {
	if queryParam == "" {
		queryParam = "version"
	}
	return &QueryExtractor{
		QueryParam: queryParam,
	}
}

// ExtractVersion extracts version from query parameter
func (e *QueryExtractor) ExtractVersion(c *gin.Context) (Version, error) {
	versionStr := c.Query(e.QueryParam)
	if versionStr == "" {
		return Version{}, fmt.Errorf("version query parameter %s not found", e.QueryParam)
	}

	return ParseVersion(versionStr)
}

// ContentTypeExtractor extracts version from Accept header
type ContentTypeExtractor struct {
	MediaType string // e.g., "application/vnd.api+json"
}

// NewContentTypeExtractor creates a new content type extractor
func NewContentTypeExtractor(mediaType string) *ContentTypeExtractor {
	if mediaType == "" {
		mediaType = "application/vnd.api+json"
	}
	return &ContentTypeExtractor{
		MediaType: mediaType,
	}
}

// ExtractVersion 从 Accept 头部中提取版本，如 application/vnd.api+json;version=1
func (e *ContentTypeExtractor) ExtractVersion(c *gin.Context) (Version, error) {
	acceptHeader := c.GetHeader("Accept")
	if acceptHeader == "" {
		return Version{}, fmt.Errorf("Accept header not found")
	}

	// Parse media type with version parameter
	// 示例: application/vnd.api+json;version=1
	versionPattern := regexp.MustCompile(regexp.QuoteMeta(e.MediaType) + `;\s*version=(\d+(?:\.\d+)?(?:\.\d+)?)`)
	matches := versionPattern.FindStringSubmatch(acceptHeader)

	if len(matches) < 2 {
		return Version{}, fmt.Errorf("no version found in Accept header: %s", acceptHeader)
	}

	return ParseVersion(matches[1])
}

// CompositeExtractor tries multiple extractors in order
type CompositeExtractor struct {
	Extractors []VersionExtractor
}

// NewCompositeExtractor creates a new composite extractor
func NewCompositeExtractor(extractors ...VersionExtractor) *CompositeExtractor {
	return &CompositeExtractor{
		Extractors: extractors,
	}
}

// ExtractVersion tries extractors in order until one succeeds
func (e *CompositeExtractor) ExtractVersion(c *gin.Context) (Version, error) {
	var lastError error

	for _, extractor := range e.Extractors {
		if version, err := extractor.ExtractVersion(c); err == nil {
			return version, nil
		} else {
			lastError = err
		}
	}

	if lastError != nil {
		return Version{}, fmt.Errorf("all version extractors failed, last error: %w", lastError)
	}

	return Version{}, fmt.Errorf("no version extractors provided")
}

// CreateExtractor creates a version extractor based on strategy
func CreateExtractor(config VersionConfig) VersionExtractor {
	switch config.Strategy {
	case URLPathStrategy:
		return NewURLPathExtractor(config.URLPrefix)
	case HeaderStrategy:
		return NewHeaderExtractor(config.HeaderName)
	case QueryStrategy:
		return NewQueryExtractor(config.QueryParam)
	case ContentTypeStrategy:
		return NewContentTypeExtractor("application/vnd.api+json")
	default:
		// Default to URL path strategy
		return NewURLPathExtractor(config.URLPrefix)
	}
}

// CreateCompositeExtractor creates a composite extractor that tries multiple strategies
func CreateCompositeExtractor(config VersionConfig) VersionExtractor {
	extractors := []VersionExtractor{
		CreateExtractor(config),
	}

	// Add fallback extractors
	if config.Strategy != HeaderStrategy {
		extractors = append(extractors, NewHeaderExtractor(config.HeaderName))
	}
	if config.Strategy != QueryStrategy {
		extractors = append(extractors, NewQueryExtractor(config.QueryParam))
	}

	return NewCompositeExtractor(extractors...)
}
