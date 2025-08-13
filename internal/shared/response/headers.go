package response

import (
	"crypto/md5"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// Standard HTTP headers for RESTful APIs
const (
	HeaderContentType        = "Content-Type"
	HeaderContentLocation    = "Content-Location"
	HeaderLocation           = "Location"
	HeaderETag               = "ETag"
	HeaderIfMatch            = "If-Match"
	HeaderIfNoneMatch        = "If-None-Match"
	HeaderIfModifiedSince    = "If-Modified-Since"
	HeaderIfUnmodifiedSince  = "If-Unmodified-Since"
	HeaderLastModified       = "Last-Modified"
	HeaderCacheControl       = "Cache-Control"
	HeaderVary               = "Vary"
	HeaderExpires            = "Expires"
	HeaderIdempotencyKey     = "Idempotency-Key"
	HeaderRequestID          = "X-Request-ID"
	HeaderRateLimit          = "X-RateLimit-Limit"
	HeaderRateLimitRemaining = "X-RateLimit-Remaining"
	HeaderRateLimitReset     = "X-RateLimit-Reset"
	HeaderRetryAfter         = "Retry-After"
	HeaderAPIVersion         = "API-Version"
	HeaderDeprecation        = "Deprecation"
	HeaderSunset             = "Sunset"
	HeaderWWWAuthenticate    = "WWW-Authenticate"
)

// MediaTypes for content negotiation
const (
	MediaTypeJSON        = "application/json"
	MediaTypeHALJSON     = "application/hal+json"
	MediaTypeProblemJSON = "application/problem+json"
	MediaTypeFormData    = "application/x-www-form-urlencoded"
	MediaTypeMultipart   = "multipart/form-data"
)

// Cache directives
const (
	CacheNoStore        = "no-store"
	CacheNoCache        = "no-cache"
	CacheMustRevalidate = "must-revalidate"
	CachePrivate        = "private"
	CachePublic         = "public"
)

// HeaderManager provides standardized header management for RESTful APIs
type HeaderManager struct {
	c *gin.Context
}

// NewHeaderManager creates a new header manager
func NewHeaderManager(c *gin.Context) *HeaderManager {
	return &HeaderManager{c: c}
}

// SetContentType sets the Content-Type header
func (h *HeaderManager) SetContentType(mediaType string) {
	h.c.Header(HeaderContentType, mediaType)
}

// SetJSONContentType sets Content-Type to application/json
func (h *HeaderManager) SetJSONContentType() {
	h.SetContentType(MediaTypeJSON)
}

// SetHALContentType sets Content-Type to application/hal+json
func (h *HeaderManager) SetHALContentType() {
	h.SetContentType(MediaTypeHALJSON)
}

// SetProblemContentType sets Content-Type to application/problem+json
func (h *HeaderManager) SetProblemContentType() {
	h.SetContentType(MediaTypeProblemJSON)
}

// SetLocation sets the Location header for created resources
func (h *HeaderManager) SetLocation(url string) {
	h.c.Header(HeaderLocation, url)
}

// SetContentLocation sets the Content-Location header
func (h *HeaderManager) SetContentLocation(url string) {
	h.c.Header(HeaderContentLocation, url)
}

// SetETag sets the ETag header with proper quoting
func (h *HeaderManager) SetETag(etag string, weak bool) {
	if weak {
		h.c.Header(HeaderETag, fmt.Sprintf("W/\"%s\"", etag))
	} else {
		h.c.Header(HeaderETag, fmt.Sprintf("\"%s\"", etag))
	}
}

// SetLastModified sets the Last-Modified header
func (h *HeaderManager) SetLastModified(t time.Time) {
	h.c.Header(HeaderLastModified, t.UTC().Format(http.TimeFormat))
}

// SetCacheControl sets cache control directives
func (h *HeaderManager) SetCacheControl(directives ...string) {
	h.c.Header(HeaderCacheControl, strings.Join(directives, ", "))
}

// SetNoCache sets no-cache directives
func (h *HeaderManager) SetNoCache() {
	h.SetCacheControl(CacheNoCache, CacheNoStore, CacheMustRevalidate)
	h.c.Header(HeaderExpires, "0")
}

// SetPrivateCache sets private cache with max-age
func (h *HeaderManager) SetPrivateCache(maxAge int) {
	h.SetCacheControl(CachePrivate, CacheMustRevalidate, fmt.Sprintf("max-age=%d", maxAge))
}

// SetPublicCache sets public cache with max-age
func (h *HeaderManager) SetPublicCache(maxAge int) {
	h.SetCacheControl(CachePublic, fmt.Sprintf("max-age=%d", maxAge))
}

// SetVary sets the Vary header for content negotiation
func (h *HeaderManager) SetVary(fields ...string) {
	h.c.Header(HeaderVary, strings.Join(fields, ", "))
}

// SetStandardVary sets common Vary headers for RESTful APIs
func (h *HeaderManager) SetStandardVary() {
	h.SetVary("Accept", "Accept-Encoding", "Authorization")
}

// SetRateLimit sets rate limiting headers
func (h *HeaderManager) SetRateLimit(limit, remaining int, resetTime int64) {
	h.c.Header(HeaderRateLimit, strconv.Itoa(limit))
	h.c.Header(HeaderRateLimitRemaining, strconv.Itoa(remaining))
	h.c.Header(HeaderRateLimitReset, strconv.FormatInt(resetTime, 10))
}

// SetRetryAfter sets the Retry-After header
func (h *HeaderManager) SetRetryAfter(seconds int) {
	h.c.Header(HeaderRetryAfter, strconv.Itoa(seconds))
}

// SetAPIVersion sets the API version header
func (h *HeaderManager) SetAPIVersion(version string) {
	h.c.Header(HeaderAPIVersion, version)
}

// SetDeprecation sets deprecation warning headers
func (h *HeaderManager) SetDeprecation(deprecationDate time.Time) {
	h.c.Header(HeaderDeprecation, deprecationDate.UTC().Format(http.TimeFormat))
}

// SetSunset sets the sunset header for API retirement
func (h *HeaderManager) SetSunset(sunsetDate time.Time) {
	h.c.Header(HeaderSunset, sunsetDate.UTC().Format(http.TimeFormat))
}

// SetWWWAuthenticate sets WWW-Authenticate header for 401 responses
func (h *HeaderManager) SetWWWAuthenticate(scheme, realm string) {
	h.c.Header(HeaderWWWAuthenticate, fmt.Sprintf("%s realm=\"%s\"", scheme, realm))
}

// SetRequestID sets the request ID header if available
func (h *HeaderManager) SetRequestID() {
	if requestID := h.c.GetString("request_id"); requestID != "" {
		h.c.Header(HeaderRequestID, requestID)
	}
}

// Conditional request handling

// CheckIfMatch checks the If-Match header for optimistic locking
func (h *HeaderManager) CheckIfMatch(etag string) bool {
	ifMatch := h.c.GetHeader(HeaderIfMatch)
	if ifMatch == "" {
		return true // No condition means proceed
	}

	if ifMatch == "*" {
		return true // Wildcard matches any existing resource
	}

	// Parse comma-separated ETags
	etags := strings.Split(ifMatch, ",")
	for _, tag := range etags {
		tag = strings.TrimSpace(tag)
		if tag == fmt.Sprintf("\"%s\"", etag) {
			return true
		}
	}

	return false
}

// CheckIfNoneMatch checks the If-None-Match header for conditional requests
func (h *HeaderManager) CheckIfNoneMatch(etag string) bool {
	ifNoneMatch := h.c.GetHeader(HeaderIfNoneMatch)
	if ifNoneMatch == "" {
		return true // No condition means proceed
	}

	if ifNoneMatch == "*" {
		return false // Wildcard means don't proceed if resource exists
	}

	// Parse comma-separated ETags
	etags := strings.Split(ifNoneMatch, ",")
	for _, tag := range etags {
		tag = strings.TrimSpace(tag)
		if tag == fmt.Sprintf("\"%s\"", etag) {
			return false // ETag matches, don't proceed
		}
	}

	return true
}

// CheckIfModifiedSince checks the If-Modified-Since header
func (h *HeaderManager) CheckIfModifiedSince(lastModified time.Time) bool {
	ifModifiedSince := h.c.GetHeader(HeaderIfModifiedSince)
	if ifModifiedSince == "" {
		return true // No condition means proceed
	}

	since, err := time.Parse(http.TimeFormat, ifModifiedSince)
	if err != nil {
		return true // Invalid date means proceed
	}

	// Return true if resource was modified after the given time
	return lastModified.After(since)
}

// CheckIfUnmodifiedSince checks the If-Unmodified-Since header
func (h *HeaderManager) CheckIfUnmodifiedSince(lastModified time.Time) bool {
	ifUnmodifiedSince := h.c.GetHeader(HeaderIfUnmodifiedSince)
	if ifUnmodifiedSince == "" {
		return true // No condition means proceed
	}

	since, err := time.Parse(http.TimeFormat, ifUnmodifiedSince)
	if err != nil {
		return true // Invalid date means proceed
	}

	// Return true if resource was NOT modified after the given time
	return !lastModified.After(since)
}

// Utility functions

// GenerateETag generates an ETag from content
func GenerateETag(content []byte) string {
	hash := md5.Sum(content)
	return fmt.Sprintf("%x", hash)
}

// GenerateResourceETag generates an ETag for a resource based on ID and modification time
func GenerateResourceETag(resourceID string, modTime time.Time) string {
	data := fmt.Sprintf("%s-%d", resourceID, modTime.Unix())
	hash := md5.Sum([]byte(data))
	return fmt.Sprintf("%x", hash)
}

// SetStandardRESTHeaders sets common headers for RESTful responses
func SetStandardRESTHeaders(c *gin.Context) {
	hm := NewHeaderManager(c)

	// Set standard content type
	hm.SetJSONContentType()

	// Set vary headers for content negotiation
	hm.SetStandardVary()

	// Set no-cache by default (can be overridden by specific handlers)
	hm.SetNoCache()

	// Set request ID if available
	hm.SetRequestID()

	// Set API version if available
	if version := c.GetString("api_version"); version != "" {
		hm.SetAPIVersion(version)
	}
}

// WithHeaders is a middleware that sets standard RESTful headers
func WithHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		SetStandardRESTHeaders(c)
		c.Next()
	}
}

// WithETag is a middleware that handles ETag generation and conditional requests
func WithETag(etagFunc func(c *gin.Context) string) gin.HandlerFunc {
	return func(c *gin.Context) {
		hm := NewHeaderManager(c)

		// Generate ETag
		etag := etagFunc(c)
		if etag != "" {
			hm.SetETag(etag, false)

			// Check If-None-Match for GET/HEAD requests
			if c.Request.Method == "GET" || c.Request.Method == "HEAD" {
				if !hm.CheckIfNoneMatch(etag) {
					c.Status(http.StatusNotModified)
					c.Abort()
					return
				}
			}

			// Check If-Match for PUT/PATCH/DELETE requests
			if c.Request.Method == "PUT" || c.Request.Method == "PATCH" || c.Request.Method == "DELETE" {
				if !hm.CheckIfMatch(etag) {
					ProblemDetail(c, http.StatusPreconditionFailed,
						"/problems/precondition-failed", "Precondition Failed",
						"The resource has been modified by another request")
					c.Abort()
					return
				}
			}
		}

		c.Next()
	}
}

// WithLastModified is a middleware that handles Last-Modified and conditional requests
func WithLastModified(lastModifiedFunc func(c *gin.Context) time.Time) gin.HandlerFunc {
	return func(c *gin.Context) {
		hm := NewHeaderManager(c)

		// Get last modified time
		lastModified := lastModifiedFunc(c)
		if !lastModified.IsZero() {
			hm.SetLastModified(lastModified)

			// Check If-Modified-Since for GET/HEAD requests
			if c.Request.Method == "GET" || c.Request.Method == "HEAD" {
				if !hm.CheckIfModifiedSince(lastModified) {
					c.Status(http.StatusNotModified)
					c.Abort()
					return
				}
			}

			// Check If-Unmodified-Since for PUT/PATCH/DELETE requests
			if c.Request.Method == "PUT" || c.Request.Method == "PATCH" || c.Request.Method == "DELETE" {
				if !hm.CheckIfUnmodifiedSince(lastModified) {
					ProblemDetail(c, http.StatusPreconditionFailed,
						"/problems/precondition-failed", "Precondition Failed",
						"The resource has been modified since the specified time")
					c.Abort()
					return
				}
			}
		}

		c.Next()
	}
}
