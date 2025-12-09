package consts

import (
	"github.com/patrickmn/go-cache"
)

// DefaultCacheTTL is the default cache TTL for compliance metadata.
// It uses cache.NoExpiration to indicate that cached items should never expire
// by default, as compliance metadata changes infrequently.
const DefaultCacheTTL = cache.NoExpiration
