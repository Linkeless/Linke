package versioning

import (
	"testing"
	"time"
)

func TestParseVersion(t *testing.T) {
	tests := []struct {
		name        string
		versionStr  string
		expected    Version
		expectError bool
	}{
		{
			name:       "parse v1",
			versionStr: "v1",
			expected:   NewVersion(1, 0, 0),
		},
		{
			name:       "parse 1.0.0",
			versionStr: "1.0.0",
			expected:   NewVersion(1, 0, 0),
		},
		{
			name:       "parse v2.1",
			versionStr: "v2.1",
			expected:   NewVersion(2, 1, 0),
		},
		{
			name:       "parse 2.1.3",
			versionStr: "2.1.3",
			expected:   NewVersion(2, 1, 3),
		},
		{
			name:       "parse 1",
			versionStr: "1",
			expected:   NewVersion(1, 0, 0),
		},
		{
			name:        "invalid format",
			versionStr:  "invalid",
			expectError: true,
		},
		{
			name:        "too many parts",
			versionStr:  "1.2.3.4",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseVersion(tt.versionStr)
			
			if tt.expectError {
				if err == nil {
					t.Errorf("expected error for version %s, but got none", tt.versionStr)
				}
				return
			}
			
			if err != nil {
				t.Errorf("unexpected error for version %s: %v", tt.versionStr, err)
				return
			}
			
			if result.Compare(tt.expected) != 0 {
				t.Errorf("version %s: expected %s, got %s", tt.versionStr, tt.expected.String(), result.String())
			}
		})
	}
}

func TestVersionCompare(t *testing.T) {
	tests := []struct {
		name     string
		v1       Version
		v2       Version
		expected int
	}{
		{
			name:     "equal versions",
			v1:       NewVersion(1, 0, 0),
			v2:       NewVersion(1, 0, 0),
			expected: 0,
		},
		{
			name:     "v1 < v2 (major)",
			v1:       NewVersion(1, 0, 0),
			v2:       NewVersion(2, 0, 0),
			expected: -1,
		},
		{
			name:     "v1 > v2 (major)",
			v1:       NewVersion(2, 0, 0),
			v2:       NewVersion(1, 0, 0),
			expected: 1,
		},
		{
			name:     "v1 < v2 (minor)",
			v1:       NewVersion(1, 1, 0),
			v2:       NewVersion(1, 2, 0),
			expected: -1,
		},
		{
			name:     "v1 > v2 (minor)",
			v1:       NewVersion(1, 2, 0),
			v2:       NewVersion(1, 1, 0),
			expected: 1,
		},
		{
			name:     "v1 < v2 (patch)",
			v1:       NewVersion(1, 1, 1),
			v2:       NewVersion(1, 1, 2),
			expected: -1,
		},
		{
			name:     "v1 > v2 (patch)",
			v1:       NewVersion(1, 1, 2),
			v2:       NewVersion(1, 1, 1),
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.v1.Compare(tt.v2)
			if result != tt.expected {
				t.Errorf("compare %s vs %s: expected %d, got %d", 
					tt.v1.String(), tt.v2.String(), tt.expected, result)
			}
		})
	}
}

func TestVersionCompatibility(t *testing.T) {
	tests := []struct {
		name     string
		v1       Version
		v2       Version
		expected bool
	}{
		{
			name:     "same version",
			v1:       NewVersion(1, 0, 0),
			v2:       NewVersion(1, 0, 0),
			expected: true,
		},
		{
			name:     "compatible minor version",
			v1:       NewVersion(1, 2, 0),
			v2:       NewVersion(1, 1, 0),
			expected: true,
		},
		{
			name:     "incompatible minor version",
			v1:       NewVersion(1, 1, 0),
			v2:       NewVersion(1, 2, 0),
			expected: false,
		},
		{
			name:     "incompatible major version",
			v1:       NewVersion(2, 0, 0),
			v2:       NewVersion(1, 0, 0),
			expected: false,
		},
		{
			name:     "compatible patch version",
			v1:       NewVersion(1, 1, 2),
			v2:       NewVersion(1, 1, 1),
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.v1.IsCompatibleWith(tt.v2)
			if result != tt.expected {
				t.Errorf("compatibility %s with %s: expected %t, got %t", 
					tt.v1.String(), tt.v2.String(), tt.expected, result)
			}
		})
	}
}

func TestVersionString(t *testing.T) {
	tests := []struct {
		name     string
		version  Version
		expected string
	}{
		{
			name:     "basic version",
			version:  NewVersion(1, 0, 0),
			expected: "1.0.0",
		},
		{
			name:     "version with minor",
			version:  NewVersion(2, 1, 0),
			expected: "2.1.0",
		},
		{
			name:     "version with patch",
			version:  NewVersion(1, 2, 3),
			expected: "1.2.3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.version.String()
			if result != tt.expected {
				t.Errorf("version string for %+v: expected %s, got %s", 
					tt.version, tt.expected, result)
			}
		})
	}
}

func TestVersionShortString(t *testing.T) {
	tests := []struct {
		name     string
		version  Version
		expected string
	}{
		{
			name:     "v1",
			version:  NewVersion(1, 0, 0),
			expected: "v1",
		},
		{
			name:     "v2",
			version:  NewVersion(2, 1, 3),
			expected: "v2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.version.ShortString()
			if result != tt.expected {
				t.Errorf("short string for %+v: expected %s, got %s", 
					tt.version, tt.expected, result)
			}
		})
	}
}

func TestVersionInfo(t *testing.T) {
	now := time.Now()
	futureDate := now.Add(30 * 24 * time.Hour) // 30 days from now
	pastDate := now.Add(-30 * 24 * time.Hour)  // 30 days ago

	tests := []struct {
		name           string
		versionInfo    VersionInfo
		expectDeprecated bool
		expectSunset   bool
		expectedDays   int
	}{
		{
			name: "active version",
			versionInfo: VersionInfo{
				Version:     NewVersion(2, 0, 0),
				Status:      "active",
				Description: "Current version",
				Released:    now,
			},
			expectDeprecated: false,
			expectSunset:     false,
			expectedDays:     -1,
		},
		{
			name: "deprecated version",
			versionInfo: VersionInfo{
				Version:     NewVersion(1, 0, 0),
				Status:      "deprecated",
				SunsetDate:  &futureDate,
				Description: "Deprecated version",
				Released:    now.Add(-365 * 24 * time.Hour),
			},
			expectDeprecated: true,
			expectSunset:     false,
			expectedDays:     30,
		},
		{
			name: "sunset version (future)",
			versionInfo: VersionInfo{
				Version:     NewVersion(1, 0, 0),
				Status:      "sunset",
				SunsetDate:  &futureDate,
				Description: "Version to be sunset",
				Released:    now.Add(-365 * 24 * time.Hour),
			},
			expectDeprecated: true,
			expectSunset:     false,
			expectedDays:     30,
		},
		{
			name: "sunset version (past)",
			versionInfo: VersionInfo{
				Version:     NewVersion(1, 0, 0),
				Status:      "sunset",
				SunsetDate:  &pastDate,
				Description: "Sunset version",
				Released:    now.Add(-365 * 24 * time.Hour),
			},
			expectDeprecated: true,
			expectSunset:     true,
			expectedDays:     -30,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.versionInfo.IsDeprecated() != tt.expectDeprecated {
				t.Errorf("IsDeprecated: expected %t, got %t", 
					tt.expectDeprecated, tt.versionInfo.IsDeprecated())
			}
			
			if tt.versionInfo.IsSunset() != tt.expectSunset {
				t.Errorf("IsSunset: expected %t, got %t", 
					tt.expectSunset, tt.versionInfo.IsSunset())
			}
			
			days := tt.versionInfo.DaysUntilSunset()
			// Allow some tolerance for test execution time
			if tt.expectedDays >= 0 && (days < tt.expectedDays-1 || days > tt.expectedDays+1) {
				t.Errorf("DaysUntilSunset: expected ~%d, got %d", 
					tt.expectedDays, days)
			} else if tt.expectedDays < 0 && days != -1 {
				t.Errorf("DaysUntilSunset: expected %d, got %d", 
					tt.expectedDays, days)
			}
		})
	}
}

func TestVersionConfigValidation(t *testing.T) {
	config := DefaultVersionConfig()
	
	t.Run("default config is valid", func(t *testing.T) {
		if len(config.SupportedVersions) == 0 {
			t.Error("default config should have supported versions")
		}
		
		if !config.IsVersionSupported(config.DefaultVersion) {
			t.Error("default version should be supported")
		}
		
		if config.DefaultVersion.Compare(config.MinVersion) < 0 {
			t.Error("default version should be >= min version")
		}
		
		if config.DefaultVersion.Compare(config.MaxVersion) > 0 {
			t.Error("default version should be <= max version")
		}
	})
	
	t.Run("version support check", func(t *testing.T) {
		v1 := NewVersion(1, 0, 0)
		v2 := NewVersion(2, 0, 0)
		v3 := NewVersion(3, 0, 0) // unsupported
		
		if !config.IsVersionSupported(v1) {
			t.Error("v1 should be supported")
		}
		
		if !config.IsVersionSupported(v2) {
			t.Error("v2 should be supported")
		}
		
		if config.IsVersionSupported(v3) {
			t.Error("v3 should not be supported")
		}
	})
	
	t.Run("latest version", func(t *testing.T) {
		latest := config.GetLatestVersion()
		if latest.Compare(config.MaxVersion) != 0 {
			t.Errorf("latest version should be max version, got %s, expected %s", 
				latest.String(), config.MaxVersion.String())
		}
	})
}