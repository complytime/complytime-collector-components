package truthbeam

import (
	"testing"
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/collector/config/confighttp"

	"github.com/complytime/complybeacon/truthbeam/internal/consts"
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

// TestCacheTTLNormalization tests the cache TTL normalization logic
// as documented in truthbeam/internal/metadata/testdata/config.yaml.
// Multiple formats are expected: duration strings (1m, 5m, 10m, 30m, 1h, 6h, 12h,
// 24h, 72h, 168h) and 0 (no expiration - cache forever).
func TestCacheTTLNormalization(t *testing.T) {
	tests := []struct {
		name            string
		cacheTTL        time.Duration
		expectedAfter   time.Duration
		shouldNormalize bool
	}{
		{
			name:            "zero duration normalizes to NoExpiration",
			cacheTTL:        0,
			expectedAfter:   cache.NoExpiration,
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
		{
			name:            "NoExpiration preserved",
			cacheTTL:        cache.NoExpiration,
			expectedAfter:   cache.NoExpiration,
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
				assert.Equal(t, consts.DefaultCacheTTL, cfg.CacheTTL)
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
	assert.Equal(t, cache.NoExpiration, cfg.CacheTTL,
		"Zero cache TTL should be normalized to NoExpiration")
	assert.Equal(t, consts.DefaultCacheTTL, cfg.CacheTTL,
		"Normalized value should match DefaultCacheTTL")
}
