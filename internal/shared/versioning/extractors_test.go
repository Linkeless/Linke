package versioning

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestURLPathExtractor(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name        string
		urlPrefix   string
		path        string
		expected    Version
		expectError bool
	}{
		{
			name:      "extract v1",
			urlPrefix: "/api",
			path:      "/api/v1/users",
			expected:  NewVersion(1, 0, 0),
		},
		{
			name:      "extract v2",
			urlPrefix: "/api",
			path:      "/api/v2/users",
			expected:  NewVersion(2, 0, 0),
		},
		{
			name:      "extract v2.1",
			urlPrefix: "/api",
			path:      "/api/v2.1/users",
			expected:  NewVersion(2, 1, 0),
		},
		{
			name:      "no prefix",
			urlPrefix: "",
			path:      "/v1/users",
			expected:  NewVersion(1, 0, 0),
		},
		{
			name:        "no version in path",
			urlPrefix:   "/api",
			path:        "/api/users",
			expectError: true,
		},
		{
			name:        "invalid version format",
			urlPrefix:   "/api",
			path:        "/api/vX/users",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			extractor := NewURLPathExtractor(tt.urlPrefix)

			// Create a test request
			req := httptest.NewRequest("GET", tt.path, nil)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req

			version, err := extractor.ExtractVersion(c)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error for path %s, but got none", tt.path)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error for path %s: %v", tt.path, err)
				return
			}

			if version.Compare(tt.expected) != 0 {
				t.Errorf("path %s: expected %s, got %s", tt.path, tt.expected.String(), version.String())
			}
		})
	}
}

func TestHeaderExtractor(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name        string
		headerName  string
		headerValue string
		expected    Version
		expectError bool
	}{
		{
			name:        "extract v1",
			headerName:  "X-API-Version",
			headerValue: "v1",
			expected:    NewVersion(1, 0, 0),
		},
		{
			name:        "extract 2.0.0",
			headerName:  "X-API-Version",
			headerValue: "2.0.0",
			expected:    NewVersion(2, 0, 0),
		},
		{
			name:        "custom header name",
			headerName:  "API-Version",
			headerValue: "v1",
			expected:    NewVersion(1, 0, 0),
		},
		{
			name:        "missing header",
			headerName:  "X-API-Version",
			headerValue: "",
			expectError: true,
		},
		{
			name:        "invalid version",
			headerName:  "X-API-Version",
			headerValue: "invalid",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			extractor := NewHeaderExtractor(tt.headerName)

			// Create a test request
			req := httptest.NewRequest("GET", "/api/users", nil)
			if tt.headerValue != "" {
				req.Header.Set(tt.headerName, tt.headerValue)
			}
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req

			version, err := extractor.ExtractVersion(c)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error for header %s=%s, but got none", tt.headerName, tt.headerValue)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error for header %s=%s: %v", tt.headerName, tt.headerValue, err)
				return
			}

			if version.Compare(tt.expected) != 0 {
				t.Errorf("header %s=%s: expected %s, got %s",
					tt.headerName, tt.headerValue, tt.expected.String(), version.String())
			}
		})
	}
}

func TestQueryExtractor(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name        string
		queryParam  string
		queryValue  string
		expected    Version
		expectError bool
	}{
		{
			name:       "extract v1",
			queryParam: "version",
			queryValue: "v1",
			expected:   NewVersion(1, 0, 0),
		},
		{
			name:       "extract 2.0.0",
			queryParam: "version",
			queryValue: "2.0.0",
			expected:   NewVersion(2, 0, 0),
		},
		{
			name:       "custom query param",
			queryParam: "api_version",
			queryValue: "v1",
			expected:   NewVersion(1, 0, 0),
		},
		{
			name:        "missing query param",
			queryParam:  "version",
			queryValue:  "",
			expectError: true,
		},
		{
			name:        "invalid version",
			queryParam:  "version",
			queryValue:  "invalid",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			extractor := NewQueryExtractor(tt.queryParam)

			// Create a test request
			path := "/api/users"
			if tt.queryValue != "" {
				path += "?" + tt.queryParam + "=" + tt.queryValue
			}
			req := httptest.NewRequest("GET", path, nil)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req

			version, err := extractor.ExtractVersion(c)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error for query %s=%s, but got none", tt.queryParam, tt.queryValue)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error for query %s=%s: %v", tt.queryParam, tt.queryValue, err)
				return
			}

			if version.Compare(tt.expected) != 0 {
				t.Errorf("query %s=%s: expected %s, got %s",
					tt.queryParam, tt.queryValue, tt.expected.String(), version.String())
			}
		})
	}
}

func TestContentTypeExtractor(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name        string
		mediaType   string
		acceptValue string
		expected    Version
		expectError bool
	}{
		{
			name:        "extract v1",
			mediaType:   "application/vnd.api+json",
			acceptValue: "application/vnd.api+json;version=1",
			expected:    NewVersion(1, 0, 0),
		},
		{
			name:        "extract v2",
			mediaType:   "application/vnd.api+json",
			acceptValue: "application/vnd.api+json;version=2",
			expected:    NewVersion(2, 0, 0),
		},
		{
			name:        "extract v2.1",
			mediaType:   "application/vnd.api+json",
			acceptValue: "application/vnd.api+json;version=2.1",
			expected:    NewVersion(2, 1, 0),
		},
		{
			name:        "custom media type",
			mediaType:   "application/json",
			acceptValue: "application/json;version=1",
			expected:    NewVersion(1, 0, 0),
		},
		{
			name:        "missing accept header",
			mediaType:   "application/vnd.api+json",
			acceptValue: "",
			expectError: true,
		},
		{
			name:        "no version in accept",
			mediaType:   "application/vnd.api+json",
			acceptValue: "application/vnd.api+json",
			expectError: true,
		},
		{
			name:        "wrong media type",
			mediaType:   "application/vnd.api+json",
			acceptValue: "application/json;version=1",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			extractor := NewContentTypeExtractor(tt.mediaType)

			// Create a test request
			req := httptest.NewRequest("GET", "/api/users", nil)
			if tt.acceptValue != "" {
				req.Header.Set("Accept", tt.acceptValue)
			}
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req

			version, err := extractor.ExtractVersion(c)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error for Accept=%s, but got none", tt.acceptValue)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error for Accept=%s: %v", tt.acceptValue, err)
				return
			}

			if version.Compare(tt.expected) != 0 {
				t.Errorf("Accept=%s: expected %s, got %s",
					tt.acceptValue, tt.expected.String(), version.String())
			}
		})
	}
}

func TestCompositeExtractor(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("tries extractors in order", func(t *testing.T) {
		urlExtractor := NewURLPathExtractor("/api")
		headerExtractor := NewHeaderExtractor("X-API-Version")
		queryExtractor := NewQueryExtractor("version")

		composite := NewCompositeExtractor(urlExtractor, headerExtractor, queryExtractor)

		// Test URL path extraction (first extractor succeeds)
		req := httptest.NewRequest("GET", "/api/v2/users", nil)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req

		version, err := composite.ExtractVersion(c)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		expected := NewVersion(2, 0, 0)
		if version.Compare(expected) != 0 {
			t.Errorf("expected %s, got %s", expected.String(), version.String())
		}
	})

	t.Run("falls back to second extractor", func(t *testing.T) {
		urlExtractor := NewURLPathExtractor("/api")
		headerExtractor := NewHeaderExtractor("X-API-Version")

		composite := NewCompositeExtractor(urlExtractor, headerExtractor)

		// Create request without version in URL but with header
		req := httptest.NewRequest("GET", "/api/users", nil)
		req.Header.Set("X-API-Version", "v1")
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req

		version, err := composite.ExtractVersion(c)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		expected := NewVersion(1, 0, 0)
		if version.Compare(expected) != 0 {
			t.Errorf("expected %s, got %s", expected.String(), version.String())
		}
	})

	t.Run("all extractors fail", func(t *testing.T) {
		urlExtractor := NewURLPathExtractor("/api")
		headerExtractor := NewHeaderExtractor("X-API-Version")

		composite := NewCompositeExtractor(urlExtractor, headerExtractor)

		// Create request without version anywhere
		req := httptest.NewRequest("GET", "/api/users", nil)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req

		_, err := composite.ExtractVersion(c)
		if err == nil {
			t.Error("expected error when all extractors fail")
		}
	})
}

func TestCreateExtractor(t *testing.T) {
	tests := []struct {
		name     string
		config   VersionConfig
		expected string // expected extractor type
	}{
		{
			name: "URL path strategy",
			config: VersionConfig{
				Strategy:  URLPathStrategy,
				URLPrefix: "/api",
			},
			expected: "*versioning.URLPathExtractor",
		},
		{
			name: "header strategy",
			config: VersionConfig{
				Strategy:   HeaderStrategy,
				HeaderName: "X-API-Version",
			},
			expected: "*versioning.HeaderExtractor",
		},
		{
			name: "query strategy",
			config: VersionConfig{
				Strategy:   QueryStrategy,
				QueryParam: "version",
			},
			expected: "*versioning.QueryExtractor",
		},
		{
			name: "content type strategy",
			config: VersionConfig{
				Strategy: ContentTypeStrategy,
			},
			expected: "*versioning.ContentTypeExtractor",
		},
		{
			name: "unknown strategy defaults to URL path",
			config: VersionConfig{
				Strategy:  "unknown",
				URLPrefix: "/api",
			},
			expected: "*versioning.URLPathExtractor",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			extractor := CreateExtractor(tt.config)

			// Check the type of the extractor
			extractorType := getExtractorType(extractor)
			if extractorType != tt.expected {
				t.Errorf("expected extractor type %s, got %s", tt.expected, extractorType)
			}
		})
	}
}

// Helper function to get extractor type name
func getExtractorType(extractor VersionExtractor) string {
	switch extractor.(type) {
	case *URLPathExtractor:
		return "*versioning.URLPathExtractor"
	case *HeaderExtractor:
		return "*versioning.HeaderExtractor"
	case *QueryExtractor:
		return "*versioning.QueryExtractor"
	case *ContentTypeExtractor:
		return "*versioning.ContentTypeExtractor"
	case *CompositeExtractor:
		return "*versioning.CompositeExtractor"
	default:
		return "unknown"
	}
}
