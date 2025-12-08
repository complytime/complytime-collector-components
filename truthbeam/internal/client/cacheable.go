package client

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/patrickmn/go-cache"
	"go.uber.org/zap"
)

// CacheableClient wraps the basic client with that leverages a caching mechanism.
type CacheableClient struct {
	client *Client
	cache  *cache.Cache
	mu     sync.Mutex
	logger *zap.Logger
}

// NewCacheableClient creates a new enriched client with caching capabilities.
// If ttl is -1, the cache will never expire (go-cache treats durations < 1ns as no expiration).
// Otherwise, items will expire after the specified TTL with cleanup
// happening at half the TTL interval.
func NewCacheableClient(client *Client, logger *zap.Logger, ttl time.Duration) *CacheableClient {
	var cleanupInterval time.Duration
	if ttl == cache.NoExpiration {
		cleanupInterval = cache.NoExpiration
	} else {
		cleanupInterval = ttl / 2
	}

	cacheInstance := cache.New(ttl, cleanupInterval)

	ec := &CacheableClient{
		client: client,
		cache:  cacheInstance,
		logger: logger,
	}
	return ec
}

// Retrieve gets compliance data for using policy data lookup values.
// Cached metadata is used by default.
func (c *CacheableClient) Retrieve(ctx context.Context, policy Policy) (Compliance, error) {
	// Uses double-checked locking: first check avoids lock overhead on cache hits (common case).
	// Second check prevents duplicate API calls when multiple goroutines concurrently miss the same key.
	if value, found := c.cache.Get(policy.PolicyRuleId); found {
		return value.(Compliance), nil
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// Second check (prevents duplicate API calls on concurrent misses)
	if value, found := c.cache.Get(policy.PolicyRuleId); found {
		return value.(Compliance), nil
	}

	// Fetch metadata from API on cache miss
	req := EnrichmentRequest{Policy: policy}
	resp, err := c.callEnrich(ctx, req)
	if err != nil {
		c.logger.Error("enrichment API call failed",
			zap.String("policy_rule_id", policy.PolicyRuleId),
			zap.String("policy_engine_name", policy.PolicyEngineName),
			zap.Error(err),
		)
		return Compliance{}, fmt.Errorf("failed to fetch metadata: %w", err)
	}
	compliance := resp.Compliance

	// Use the same expiration as configured for the cache
	c.cache.Set(policy.PolicyRuleId, compliance, cache.DefaultExpiration)

	return compliance, nil
}

func (c *CacheableClient) callEnrich(ctx context.Context, req EnrichmentRequest) (*EnrichmentResponse, error) {
	c.logger.Debug("calling compass enrich API",
		zap.String("policy_rule_id", req.Policy.PolicyRuleId),
		zap.String("policy_engine_name", req.Policy.PolicyEngineName),
	)

	resp, err := c.client.PostV1Enrich(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	parsedResp, err := ParsePostV1EnrichResponse(resp)
	if err != nil {
		return nil, err
	}

	if parsedResp.JSON200 != nil {
		return parsedResp.JSON200, nil
	}

	if parsedResp.JSONDefault != nil {
		return nil, fmt.Errorf("API call failed with status %d: %s", parsedResp.JSONDefault.Code, parsedResp.JSONDefault.Message)
	}

	return nil, fmt.Errorf("unexpected response status: %s", resp.Status)
}
