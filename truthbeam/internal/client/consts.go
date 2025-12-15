package client

import (
	"time"
)

// DefaultCacheTTL is the default cache TTL for compliance metadata.
// Zero value means no expiration, as compliance metadata changes infrequently.
const DefaultCacheTTL = time.Duration(0)

// DefaultMaxCacheSizeMB is the default maximum cache size in megabytes.
// Set to 512MB to provide reasonable capacity for compliance metadata caching
// while preventing unbounded cache growth.
const DefaultMaxCacheSizeMB = 512

// CacheKeySeparator is the separator used to create composite cache keys
// from policy engine name and policy rule id.
const CacheKeySeparator = ":"
