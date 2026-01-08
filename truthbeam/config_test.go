package truthbeam

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/collector/config/confighttp"

	"github.com/complytime/complybeacon/truthbeam/internal/client"
)

// The config tests are table-driven tests to validate configuration validation
//and default values for the truthbeam processor.

// TestConfigValidate tests the Validate method of the Config struct
func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name        string
		config      *Config
		expectError bool
		errorMsg    string
	}{
		{
			name:        "empty config should fail",
			config:      &Config{},
			expectError: true,
			errorMsg:    "must be specified",
		},
		{
			name: "valid endpoint should pass",
			config: &Config{
				ClientConfig: confighttp.ClientConfig{
					Endpoint: "http://example.com",
				},
			},
			expectError: false,
		},
		{
			name: "https endpoint should pass",
			config: &Config{
				ClientConfig: confighttp.ClientConfig{
					Endpoint: "https://api.example.com:8080",
				},
			},
			expectError: false,
		},
		{
			name: "endpoint with path should pass",
			config: &Config{
				ClientConfig: confighttp.ClientConfig{
					Endpoint: "http://localhost:8081/v1",
				},
			},
			expectError: false,
		},
		{
			name: "empty string endpoint should fail",
			config: &Config{
				ClientConfig: confighttp.ClientConfig{
					Endpoint: "",
				},
			},
			expectError: true,
			errorMsg:    "must be specified",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.expectError {
				assert.Error(t, err, "Expected validation error")
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg, "Error message should contain expected text")
				}
			} else {
				assert.NoError(t, err, "Expected no validation error")
			}
		})
	}
}

func TestConfigStruct(t *testing.T) {
	// Test that Config struct can be created and accessed
	cfg := &Config{
		ClientConfig: confighttp.ClientConfig{
			Endpoint: "http://localhost:8081",
		},
	}

	// Test that we can access the embedded ClientConfig
	assert.Equal(t, "http://localhost:8081", cfg.ClientConfig.Endpoint)

	// Test that validation passes
	err := cfg.Validate()
	assert.NoError(t, err, "Config with valid endpoint should pass validation")
}

// TestCacheTTLNormalization tests the cache TTL normalization logic.
func TestCacheTTLNormalization(t *testing.T) {
	tests := []struct {
		name            string
		cacheTTL        time.Duration
		expectedAfter   time.Duration
		shouldNormalize bool
	}{
		{
			name:            "zero duration normalizes to DefaultCacheTTL",
			cacheTTL:        0,
			expectedAfter:   client.DefaultCacheTTL,
			shouldNormalize: true,
		},
		{
			name:            "1 minute duration preserved",
			cacheTTL:        1 * time.Minute,
			expectedAfter:   1 * time.Minute,
			shouldNormalize: false,
		},
		{
			name:            "5 minutes duration preserved",
			cacheTTL:        5 * time.Minute,
			expectedAfter:   5 * time.Minute,
			shouldNormalize: false,
		},
		{
			name:            "10 minutes duration preserved",
			cacheTTL:        10 * time.Minute,
			expectedAfter:   10 * time.Minute,
			shouldNormalize: false,
		},
		{
			name:            "30 minutes duration preserved",
			cacheTTL:        30 * time.Minute,
			expectedAfter:   30 * time.Minute,
			shouldNormalize: false,
		},
		{
			name:            "1 hour duration preserved",
			cacheTTL:        1 * time.Hour,
			expectedAfter:   1 * time.Hour,
			shouldNormalize: false,
		},
		{
			name:            "6 hours duration preserved",
			cacheTTL:        6 * time.Hour,
			expectedAfter:   6 * time.Hour,
			shouldNormalize: false,
		},
		{
			name:            "12 hours duration preserved",
			cacheTTL:        12 * time.Hour,
			expectedAfter:   12 * time.Hour,
			shouldNormalize: false,
		},
		{
			name:            "24 hours duration preserved",
			cacheTTL:        24 * time.Hour,
			expectedAfter:   24 * time.Hour,
			shouldNormalize: false,
		},
		{
			name:            "72 hours duration preserved",
			cacheTTL:        72 * time.Hour,
			expectedAfter:   72 * time.Hour,
			shouldNormalize: false,
		},
		{
			name:            "168 hours duration preserved",
			cacheTTL:        168 * time.Hour,
			expectedAfter:   168 * time.Hour,
			shouldNormalize: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				ClientConfig: confighttp.ClientConfig{
					Endpoint: "http://localhost:8081",
				},
				CacheTTL: tt.cacheTTL,
			}

			// Validate should normalize 0 to DefaultCacheTTL
			err := cfg.Validate()
			assert.NoError(t, err)

			if tt.shouldNormalize {
				assert.Equal(t, tt.expectedAfter, cfg.CacheTTL)
				assert.Equal(t, client.DefaultCacheTTL, cfg.CacheTTL)
			} else {
				assert.Equal(t, tt.expectedAfter, cfg.CacheTTL)
			}
		})
	}
}

// TestCacheTTLWithValidEndpoint tests that cache TTL normalization works
// correctly when a valid endpoint is provided.
func TestCacheTTLWithValidEndpoint(t *testing.T) {
	cfg := &Config{
		ClientConfig: confighttp.ClientConfig{
			Endpoint: "http://localhost:8081",
		},
		CacheTTL: 0,
	}

	err := cfg.Validate()
	assert.NoError(t, err,
		"Config with valid endpoint and zero cache TTL should pass validation")
	assert.Equal(t, client.DefaultCacheTTL, cfg.CacheTTL,
		"Zero cache TTL should be normalized to DefaultCacheTTL")
}

// TestCacheCapacityValidation tests the cache_capacity configuration validation
func TestCacheCapacityValidation(t *testing.T) {
	tests := []struct {
		name          string
		cacheCapacity int
		expectedAfter int
		expectError   bool
	}{
		{
			name:          "zero cache capacity normalizes to default",
			cacheCapacity: 0,
			expectedAfter: client.DefaultCacheCapacity,
			expectError:   false,
		},
		{
			name:          "positive cache capacity preserved",
			cacheCapacity: 50000,
			expectedAfter: 50000,
			expectError:   false,
		},
		{
			name:          "negative cache capacity should fail",
			cacheCapacity: -1,
			expectedAfter: 0,
			expectError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				ClientConfig: confighttp.ClientConfig{
					Endpoint: "http://localhost:8081",
				},
				CacheCapacity: tt.cacheCapacity,
			}

			err := cfg.Validate()
			if tt.expectError {
				assert.Error(t, err, "Expected validation error")
			} else {
				assert.NoError(t, err, "Expected no validation error")
				assert.Equal(t, tt.expectedAfter, cfg.CacheCapacity)
			}
		})
	}
}
