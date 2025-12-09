package truthbeam

import (
	"errors"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/confighttp"

	"github.com/complytime/complybeacon/truthbeam/internal/consts"
)

// Config defines configuration for the truthbeam processor.
type Config struct {
	ClientConfig confighttp.ClientConfig `mapstructure:",squash"`   // squash ensures fields are correctly decoded in embedded struct.
	CacheTTL     time.Duration           `mapstructure:"cache_ttl"` // Cache TTL for compliance metadata
}

var _ component.Config = (*Config)(nil)

// Validate checks if the exporter configuration is valid
func (cfg *Config) Validate() error {
	if cfg.ClientConfig.Endpoint == "" {
		return errors.New("endpoint must be specified")
	}
	// Normalize cache TTL: 0 means no expiration (same as -1/NoExpiration)
	if cfg.CacheTTL == 0 {
		cfg.CacheTTL = consts.DefaultCacheTTL
	}
	return nil
}
