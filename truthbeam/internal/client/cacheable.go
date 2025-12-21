package client

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
)

// Cache defines a simple cache interface for storing Compliance data.
type Cache interface {
	// Get retrieves a value from the cache by key.
	Get(key string) (Compliance, bool)
	// Set stores a value in the cache with the given key.
	Set(key string, value Compliance) error
	// Delete removes a value from the cache by key.
	Delete(key string) error
}

// CacheableClient wraps the basic client with that leverages a caching mechanism.
type CacheableClient struct {
	client *Client
	cache  Cache
	logger *zap.Logger
}

// NewCacheableClient creates a new enriched client with caching capabilities.
// To use a different cache backend, use NewCacheableClientWithCache instead.
func NewCacheableClient(client *Client, logger *zap.Logger, ttl time.Duration, maxCacheSizeMB int) (*CacheableClient, error) {
	// Use default max cache size if not specified
	if maxCacheSizeMB == 0 {
		maxCacheSizeMB = DefaultMaxCacheSizeMB
	}

	cache, err := NewBigCacheStore(context.Background(), ttl, maxCacheSizeMB)
	if err != nil {
		return nil, fmt.Errorf("failed to create cache: %w", err)
	}

	return NewCacheableClientWithCache(client, logger, cache), nil
}

// NewCacheableClientWithCache creates a new cacheable client with a custom cache implementation.
func NewCacheableClientWithCache(client *Client, logger *zap.Logger, cache Cache) *CacheableClient {
	return &CacheableClient{
		client: client,
		cache:  cache,
		logger: logger,
	}
}

// cacheKey generates a composite cache key from policy engine name and policy rule id.
func cacheKey(policyEngineName, policyRuleId string) string {
	return policyEngineName + CacheKeySeparator + policyRuleId
}

// Retrieve gets compliance data for using policy data lookup values.
// Cached metadata is used by default.
func (c *CacheableClient) Retrieve(ctx context.Context, policy Policy) (Compliance, error) {
	key := cacheKey(policy.PolicyEngineName, policy.PolicyRuleId)

	// Cache implementation is already concurrent, so we can check directly
	compliance, found := c.cache.Get(key)
	if found {
		return compliance, nil
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
	compliance = resp.Compliance

	// Store in cache (errors are logged but don't fail the request)
	if setErr := c.cache.Set(key, compliance); setErr != nil {
		c.logger.Warn("failed to set cache value",
			zap.String("policy_rule_id", policy.PolicyRuleId),
			zap.String("policy_engine_name", policy.PolicyEngineName),
			zap.Error(setErr),
		)
	}

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
