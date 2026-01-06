package client

import (
	"time"

	"github.com/maypok86/otter/v2"
	"github.com/maypok86/otter/v2/stats"
)

// Interface Check
var _ Cache = (*otterCacheStore)(nil)

// otterCacheStore implements Cache using Otter.
type otterCacheStore struct {
	cache *otter.Cache[string, Compliance]
}

func (s *otterCacheStore) Get(key string) (Compliance, bool) {
	return s.cache.GetIfPresent(key)
}

func (s *otterCacheStore) Set(key string, value Compliance) error {
	_, _ = s.cache.Set(key, value)
	return nil
}

func (s *otterCacheStore) Delete(key string) error {
	_, _ = s.cache.Invalidate(key)
	return nil
}

func NewOtterStore(ttl time.Duration, maxEntries int) (Cache, error) {
	opts := &otter.Options[string, Compliance]{
		MaximumSize:    maxEntries,
		ExpiryCalculator: otter.ExpiryWriting[string, Compliance](ttl),
		StatsRecorder:  stats.NewCounter(),
	}
	cache := otter.Must(opts)
	return &otterCacheStore{cache: cache}, nil
}
