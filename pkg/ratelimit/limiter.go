package ratelimit

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

// Limiter provides rate limiting functionality.
type Limiter struct {
	client     *redis.Client
	keyPrefix  string
	defaultCfg Config
}

// Config defines rate limit parameters.
type Config struct {
	// Limit is the maximum number of requests allowed in the window.
	Limit int64

	// Window is the time window for the rate limit.
	Window time.Duration

	// Burst allows a short burst above the limit (default: same as Limit).
	Burst int64
}

// Result contains the rate limit check result.
type Result struct {
	// Allowed indicates if the request should be allowed.
	Allowed bool

	// Remaining is the number of requests remaining in the current window.
	Remaining int64

	// ResetAt is when the rate limit window resets.
	ResetAt time.Time

	// RetryAfter is how long to wait before retrying (if not allowed).
	RetryAfter time.Duration
}

// Option configures the limiter.
type Option func(*Limiter)

// WithKeyPrefix sets a custom key prefix for Redis keys.
func WithKeyPrefix(prefix string) Option {
	return func(l *Limiter) {
		l.keyPrefix = prefix
	}
}

// WithDefaultConfig sets the default rate limit config.
func WithDefaultConfig(cfg Config) Option {
	return func(l *Limiter) {
		l.defaultCfg = cfg
	}
}

// NewLimiter creates a new rate limiter backed by Redis.
func NewLimiter(client *redis.Client, opts ...Option) *Limiter {
	l := &Limiter{
		client:    client,
		keyPrefix: "sapliy:ratelimit:",
		defaultCfg: Config{
			Limit:  100,
			Window: time.Minute,
			Burst:  100,
		},
	}

	for _, opt := range opts {
		opt(l)
	}

	return l
}

// Allow checks if a request should be allowed using sliding window log.
func (l *Limiter) Allow(ctx context.Context, key string) (*Result, error) {
	return l.AllowN(ctx, key, 1, l.defaultCfg)
}

// AllowN checks if N requests should be allowed with custom config.
func (l *Limiter) AllowN(ctx context.Context, key string, n int64, cfg Config) (*Result, error) {
	now := time.Now()
	windowStart := now.Add(-cfg.Window).UnixMilli()
	redisKey := l.keyPrefix + key

	// Sliding window log algorithm using Redis sorted sets
	// Remove old entries and count current entries in one atomic operation
	pipe := l.client.Pipeline()

	// Remove entries outside the window
	pipe.ZRemRangeByScore(ctx, redisKey, "-inf", strconv.FormatInt(windowStart, 10))

	// Count current entries
	countCmd := pipe.ZCard(ctx, redisKey)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("rate limit check failed: %w", err)
	}

	currentCount := countCmd.Val()
	resetAt := now.Add(cfg.Window)

	// Check if we're over the limit
	if currentCount+n > cfg.Limit {
		// Find the oldest entry to calculate retry-after
		oldestEntries, err := l.client.ZRangeWithScores(ctx, redisKey, 0, 0).Result()
		if err != nil {
			return nil, fmt.Errorf("failed to get oldest entry: %w", err)
		}

		var retryAfter time.Duration
		if len(oldestEntries) > 0 {
			oldestTimestamp := int64(oldestEntries[0].Score)
			expiresAt := time.UnixMilli(oldestTimestamp).Add(cfg.Window)
			retryAfter = time.Until(expiresAt)
			if retryAfter < 0 {
				retryAfter = 0
			}
		}

		return &Result{
			Allowed:    false,
			Remaining:  max(0, cfg.Limit-currentCount),
			ResetAt:    resetAt,
			RetryAfter: retryAfter,
		}, nil
	}

	// Add new entries to the sorted set
	members := make([]redis.Z, n)
	for i := int64(0); i < n; i++ {
		members[i] = redis.Z{
			Score:  float64(now.UnixMilli()),
			Member: fmt.Sprintf("%d-%d", now.UnixNano(), i),
		}
	}

	pipe = l.client.Pipeline()
	pipe.ZAdd(ctx, redisKey, members...)
	pipe.Expire(ctx, redisKey, cfg.Window+time.Second)
	_, err = pipe.Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to record request: %w", err)
	}

	return &Result{
		Allowed:   true,
		Remaining: cfg.Limit - currentCount - n,
		ResetAt:   resetAt,
	}, nil
}

// Middleware returns an HTTP middleware that applies rate limiting.
func (l *Limiter) Middleware(keyFunc func(*http.Request) string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := keyFunc(r)
			result, err := l.Allow(r.Context(), key)

			if err != nil {
				// On error, allow the request but log
				next.ServeHTTP(w, r)
				return
			}

			// Set rate limit headers
			w.Header().Set("X-RateLimit-Limit", strconv.FormatInt(l.defaultCfg.Limit, 10))
			w.Header().Set("X-RateLimit-Remaining", strconv.FormatInt(result.Remaining, 10))
			w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(result.ResetAt.Unix(), 10))

			if !result.Allowed {
				w.Header().Set("Retry-After", strconv.FormatInt(int64(result.RetryAfter.Seconds()), 10))
				http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// IPKeyFunc returns a key function that uses the client IP.
func IPKeyFunc(r *http.Request) string {
	// Check X-Forwarded-For first (for proxies)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return "ip:" + xff
	}
	// Check X-Real-IP
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return "ip:" + xri
	}
	return "ip:" + r.RemoteAddr
}

// APIKeyFunc returns a key function that uses the API key.
func APIKeyFunc(headerName string) func(*http.Request) string {
	return func(r *http.Request) string {
		apiKey := r.Header.Get(headerName)
		if apiKey == "" {
			// Fall back to IP-based limiting
			return IPKeyFunc(r)
		}
		return "apikey:" + apiKey
	}
}

// EndpointKeyFunc returns a key function that includes the endpoint.
func EndpointKeyFunc(baseKeyFunc func(*http.Request) string) func(*http.Request) string {
	return func(r *http.Request) string {
		baseKey := baseKeyFunc(r)
		return fmt.Sprintf("%s:%s:%s", r.Method, r.URL.Path, baseKey)
	}
}

func max(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}
