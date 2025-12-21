package client

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"time"

	"github.com/allegro/bigcache/v3"
)

// Interface Check
var _ Cache = (*bigCacheStore)(nil)

// bigCacheStore implements Cache using BigCache.
type bigCacheStore struct {
	cache *bigcache.BigCache
}

func (s *bigCacheStore) Get(key string) (Compliance, bool) {
	data, err := s.cache.Get(key)
	if err != nil {
		return Compliance{}, false
	}

	var compliance Compliance
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	if err := dec.Decode(&compliance); err != nil {
		return Compliance{}, false
	}

	return compliance, true
}

func (s *bigCacheStore) Set(key string, value Compliance) error {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(value); err != nil {
		return fmt.Errorf("failed to marshal compliance: %w", err)
	}

	return s.cache.Set(key, buf.Bytes())
}

func (s *bigCacheStore) Delete(key string) error {
	return s.cache.Delete(key)
}

// NewBigCacheStore creates a new BigCache-based cache store.
// If ttl is 0, the cache will never expire.
// maxCacheSizeMB specifies the maximum cache size in megabytes.
func NewBigCacheStore(ctx context.Context, ttl time.Duration, maxCacheSizeMB int) (Cache, error) {
	config := bigcache.DefaultConfig(ttl)

	// Configure cleanup interval (half of TTL if TTL is set)
	if ttl > 0 {
		config.CleanWindow = ttl / 2
	}

	if maxCacheSizeMB > 0 {
		config.HardMaxCacheSize = maxCacheSizeMB
	}

	cache, err := bigcache.New(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create bigcache: %w", err)
	}

	return &bigCacheStore{
		cache: cache,
	}, nil
}
