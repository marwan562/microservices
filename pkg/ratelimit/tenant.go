package ratelimit

import (
	"context"
	"sync"
	"time"
)

// TenantLimiter provides per-tenant rate limiting with different tiers.
type TenantLimiter struct {
	limiter     *Limiter
	tierConfigs map[string]Config
	tenantTiers map[string]string
	defaultTier string
	mu          sync.RWMutex
}

// TenantOption configures the tenant limiter.
type TenantOption func(*TenantLimiter)

// WithTierConfig adds a tier configuration.
func WithTierConfig(tier string, cfg Config) TenantOption {
	return func(tl *TenantLimiter) {
		tl.tierConfigs[tier] = cfg
	}
}

// WithDefaultTier sets the default tier for unknown tenants.
func WithDefaultTier(tier string) TenantOption {
	return func(tl *TenantLimiter) {
		tl.defaultTier = tier
	}
}

// NewTenantLimiter creates a tenant-aware rate limiter.
func NewTenantLimiter(limiter *Limiter, opts ...TenantOption) *TenantLimiter {
	tl := &TenantLimiter{
		limiter:     limiter,
		tierConfigs: make(map[string]Config),
		tenantTiers: make(map[string]string),
		defaultTier: "free",
	}

	// Default tier configs
	tl.tierConfigs = map[string]Config{
		"free": {
			Limit:  100,
			Window: time.Minute,
			Burst:  100,
		},
		"starter": {
			Limit:  1000,
			Window: time.Minute,
			Burst:  1000,
		},
		"pro": {
			Limit:  10000,
			Window: time.Minute,
			Burst:  10000,
		},
		"enterprise": {
			Limit:  100000,
			Window: time.Minute,
			Burst:  100000,
		},
	}

	for _, opt := range opts {
		opt(tl)
	}

	return tl
}

// SetTenantTier assigns a tier to a tenant.
func (tl *TenantLimiter) SetTenantTier(tenantID, tier string) {
	tl.mu.Lock()
	defer tl.mu.Unlock()
	tl.tenantTiers[tenantID] = tier
}

// GetTenantTier returns the tier for a tenant.
func (tl *TenantLimiter) GetTenantTier(tenantID string) string {
	tl.mu.RLock()
	defer tl.mu.RUnlock()
	if tier, ok := tl.tenantTiers[tenantID]; ok {
		return tier
	}
	return tl.defaultTier
}

// Allow checks rate limit for a tenant.
func (tl *TenantLimiter) Allow(ctx context.Context, tenantID, endpoint string) (*Result, error) {
	tier := tl.GetTenantTier(tenantID)
	cfg, ok := tl.tierConfigs[tier]
	if !ok {
		cfg = tl.tierConfigs[tl.defaultTier]
	}

	key := "tenant:" + tenantID + ":" + endpoint
	return tl.limiter.AllowN(ctx, key, 1, cfg)
}

// AllowGlobal checks global rate limit for a tenant (across all endpoints).
func (tl *TenantLimiter) AllowGlobal(ctx context.Context, tenantID string) (*Result, error) {
	tier := tl.GetTenantTier(tenantID)
	cfg, ok := tl.tierConfigs[tier]
	if !ok {
		cfg = tl.tierConfigs[tl.defaultTier]
	}

	key := "tenant:global:" + tenantID
	return tl.limiter.AllowN(ctx, key, 1, cfg)
}

// EndpointConfig provides per-endpoint rate limit configuration.
type EndpointConfig struct {
	// Path patterns (supports wildcards like /api/v1/*)
	Patterns []string

	// Config for this endpoint group
	Config Config

	// ApplyToAllTiers if true, overrides tier-based limits
	ApplyToAllTiers bool
}

// EndpointLimiter provides endpoint-specific rate limits.
type EndpointLimiter struct {
	limiter   *Limiter
	endpoints []EndpointConfig
	fallback  Config
}

// NewEndpointLimiter creates an endpoint-aware rate limiter.
func NewEndpointLimiter(limiter *Limiter, endpoints []EndpointConfig) *EndpointLimiter {
	return &EndpointLimiter{
		limiter:   limiter,
		endpoints: endpoints,
		fallback: Config{
			Limit:  1000,
			Window: time.Minute,
			Burst:  1000,
		},
	}
}

// Allow checks rate limit for a specific endpoint.
func (el *EndpointLimiter) Allow(ctx context.Context, key, path string) (*Result, error) {
	cfg := el.findConfig(path)
	return el.limiter.AllowN(ctx, "endpoint:"+key+":"+path, 1, cfg)
}

func (el *EndpointLimiter) findConfig(path string) Config {
	for _, ep := range el.endpoints {
		for _, pattern := range ep.Patterns {
			if matchPath(pattern, path) {
				return ep.Config
			}
		}
	}
	return el.fallback
}

// matchPath does simple path matching with * wildcard support.
func matchPath(pattern, path string) bool {
	if pattern == path {
		return true
	}

	// Handle trailing wildcard
	if len(pattern) > 0 && pattern[len(pattern)-1] == '*' {
		prefix := pattern[:len(pattern)-1]
		return len(path) >= len(prefix) && path[:len(prefix)] == prefix
	}

	return false
}

// DefaultEndpointConfigs returns sensible defaults for common API patterns.
func DefaultEndpointConfigs() []EndpointConfig {
	return []EndpointConfig{
		{
			// Auth endpoints - stricter limits to prevent brute force
			Patterns: []string{"/api/v1/auth/*", "/api/v1/login", "/api/v1/register"},
			Config: Config{
				Limit:  10,
				Window: time.Minute,
				Burst:  10,
			},
		},
		{
			// High-volume event ingestion - higher limits
			Patterns: []string{"/api/v1/events", "/api/v1/zones/*/events"},
			Config: Config{
				Limit:  10000,
				Window: time.Minute,
				Burst:  10000,
			},
		},
		{
			// Webhook endpoints - moderate limits
			Patterns: []string{"/api/v1/webhooks/*"},
			Config: Config{
				Limit:  1000,
				Window: time.Minute,
				Burst:  1000,
			},
		},
		{
			// Admin/management endpoints - lower limits
			Patterns: []string{"/api/v1/admin/*", "/api/v1/management/*"},
			Config: Config{
				Limit:  100,
				Window: time.Minute,
				Burst:  100,
			},
		},
	}
}
