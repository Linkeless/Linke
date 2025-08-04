package versioning

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"linke/internal/shared/logger"
)

func TestVersionMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create a test logger
	testLogger, _ := zap.NewDevelopment()
	log := logger.NewZapLogger(testLogger)

	// Create version config
	now := time.Now()
	sunsetDate := now.Add(30 * 24 * time.Hour) // 30 days from now

	config := VersionConfig{
		Strategy:       URLPathStrategy,
		DefaultVersion: NewVersion(1, 0, 0),
		MinVersion:     NewVersion(1, 0, 0),
		MaxVersion:     NewVersion(2, 0, 0),
		SupportedVersions: []VersionInfo{
			{
				Version:     NewVersion(1, 0, 0),
				Status:      "deprecated",
				SunsetDate:  &sunsetDate,
				Description: "Deprecated version",
				Released:    now.Add(-365 * 24 * time.Hour),
			},
			{
				Version:     NewVersion(2, 0, 0),
				Status:      "active",
				Description: "Current version",
				Released:    now,
			},
		},
		URLPrefix:                "/api",
		EnableDeprecationHeaders: true,
		EnableAutoMigration:      false,
	}

	middleware := NewVersionMiddleware(config, log)

	t.Run("extracts version from URL", func(t *testing.T) {
		router := gin.New()
		router.Use(middleware.Middleware())
		router.GET("/api/v2/users", func(c *gin.Context) {
			versionCtx, exists := GetVersionFromContext(c)
			if !exists {
				t.Error("version context not found")
				return
			}

			expected := NewVersion(2, 0, 0)
			if versionCtx.ResolvedVersion.Compare(expected) != 0 {
				t.Errorf("expected version %s, got %s",
					expected.String(), versionCtx.ResolvedVersion.String())
			}

			c.JSON(http.StatusOK, gin.H{"version": versionCtx.ResolvedVersion.String()})
		})

		req := httptest.NewRequest("GET", "/api/v2/users", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		// Check version headers
		if w.Header().Get("X-API-Version") != "2.0.0" {
			t.Errorf("expected X-API-Version header to be '2.0.0', got '%s'",
				w.Header().Get("X-API-Version"))
		}
	})

	t.Run("uses default version when extraction fails", func(t *testing.T) {
		router := gin.New()
		router.Use(middleware.Middleware())
		router.GET("/api/users", func(c *gin.Context) {
			versionCtx, exists := GetVersionFromContext(c)
			if !exists {
				t.Error("version context not found")
				return
			}

			expected := config.DefaultVersion
			if versionCtx.ResolvedVersion.Compare(expected) != 0 {
				t.Errorf("expected default version %s, got %s",
					expected.String(), versionCtx.ResolvedVersion.String())
			}

			c.JSON(http.StatusOK, gin.H{"version": versionCtx.ResolvedVersion.String()})
		})

		req := httptest.NewRequest("GET", "/api/users", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}
	})

	t.Run("rejects unsupported version", func(t *testing.T) {
		router := gin.New()
		router.Use(middleware.Middleware())
		router.GET("/api/v3/users", func(c *gin.Context) {
			t.Error("handler should not be called for unsupported version")
		})

		req := httptest.NewRequest("GET", "/api/v3/users", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})

	t.Run("adds deprecation headers for deprecated version", func(t *testing.T) {
		router := gin.New()
		router.Use(middleware.Middleware())
		router.GET("/api/v1/users", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "ok"})
		})

		req := httptest.NewRequest("GET", "/api/v1/users", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		// Check deprecation headers
		warning := w.Header().Get("Warning")
		if warning == "" {
			t.Error("expected Warning header for deprecated version")
		}

		sunset := w.Header().Get("Sunset")
		if sunset == "" {
			t.Error("expected Sunset header for deprecated version")
		}

		sunsetDays := w.Header().Get("X-API-Sunset-Days")
		if sunsetDays == "" {
			t.Error("expected X-API-Sunset-Days header for deprecated version")
		}
	})

	t.Run("rejects sunset version", func(t *testing.T) {
		// Create config with past sunset date
		pastDate := now.Add(-30 * 24 * time.Hour)
		sunsetConfig := config
		sunsetConfig.SupportedVersions[0].SunsetDate = &pastDate

		sunsetMiddleware := NewVersionMiddleware(sunsetConfig, log)

		router := gin.New()
		router.Use(sunsetMiddleware.Middleware())
		router.GET("/api/v1/users", func(c *gin.Context) {
			t.Error("handler should not be called for sunset version")
		})

		req := httptest.NewRequest("GET", "/api/v1/users", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusGone {
			t.Errorf("expected status 410 (Gone), got %d", w.Code)
		}
	})

	t.Run("version info endpoint", func(t *testing.T) {
		router := gin.New()
		router.GET("/api/version", middleware.VersionInfo())

		req := httptest.NewRequest("GET", "/api/version", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		// Check that response contains version information
		body := w.Body.String()
		if body == "" {
			t.Error("expected response body with version information")
		}
	})

	t.Run("health check endpoint", func(t *testing.T) {
		router := gin.New()
		router.Use(middleware.Middleware())
		router.GET("/api/health", middleware.HealthCheck())

		req := httptest.NewRequest("GET", "/api/health", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		// Check that response contains health and version information
		body := w.Body.String()
		if body == "" {
			t.Error("expected response body with health information")
		}
	})
}

func TestVersionContext(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("set and get version context", func(t *testing.T) {
		router := gin.New()
		router.GET("/test", func(c *gin.Context) {
			// Set version context
			versionCtx := &VersionContext{
				RequestedVersion: NewVersion(1, 0, 0),
				ResolvedVersion:  NewVersion(1, 0, 0),
				Strategy:         URLPathStrategy,
			}
			SetVersionInContext(c, versionCtx)

			// Get version context
			retrievedCtx, exists := GetVersionFromContext(c)
			if !exists {
				t.Error("version context not found after setting")
				return
			}

			if retrievedCtx.ResolvedVersion.Compare(versionCtx.ResolvedVersion) != 0 {
				t.Errorf("version context mismatch: expected %s, got %s",
					versionCtx.ResolvedVersion.String(), retrievedCtx.ResolvedVersion.String())
			}

			if retrievedCtx.Strategy != versionCtx.Strategy {
				t.Errorf("strategy mismatch: expected %s, got %s",
					versionCtx.Strategy, retrievedCtx.Strategy)
			}

			c.JSON(http.StatusOK, gin.H{"ok": true})
		})

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}
	})

	t.Run("get non-existent version context", func(t *testing.T) {
		router := gin.New()
		router.GET("/test", func(c *gin.Context) {
			_, exists := GetVersionFromContext(c)
			if exists {
				t.Error("expected no version context, but found one")
			}
			c.JSON(http.StatusOK, gin.H{"ok": true})
		})

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}
	})
}

func TestHeaderStrategyMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create a test logger
	testLogger, _ := zap.NewDevelopment()
	log := logger.NewZapLogger(testLogger)

	// Create config with header strategy
	config := VersionConfig{
		Strategy:       HeaderStrategy,
		DefaultVersion: NewVersion(1, 0, 0),
		MinVersion:     NewVersion(1, 0, 0),
		MaxVersion:     NewVersion(2, 0, 0),
		SupportedVersions: []VersionInfo{
			{
				Version:     NewVersion(1, 0, 0),
				Status:      "active",
				Description: "Version 1",
				Released:    time.Now(),
			},
			{
				Version:     NewVersion(2, 0, 0),
				Status:      "active",
				Description: "Version 2",
				Released:    time.Now(),
			},
		},
		HeaderName:               "X-API-Version",
		EnableDeprecationHeaders: false,
		EnableAutoMigration:      false,
	}

	middleware := NewVersionMiddleware(config, log)

	t.Run("extracts version from header", func(t *testing.T) {
		router := gin.New()
		router.Use(middleware.Middleware())
		router.GET("/users", func(c *gin.Context) {
			versionCtx, exists := GetVersionFromContext(c)
			if !exists {
				t.Error("version context not found")
				return
			}

			expected := NewVersion(2, 0, 0)
			if versionCtx.ResolvedVersion.Compare(expected) != 0 {
				t.Errorf("expected version %s, got %s",
					expected.String(), versionCtx.ResolvedVersion.String())
			}

			c.JSON(http.StatusOK, gin.H{"version": versionCtx.ResolvedVersion.String()})
		})

		req := httptest.NewRequest("GET", "/users", nil)
		req.Header.Set("X-API-Version", "v2")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}
	})
}
