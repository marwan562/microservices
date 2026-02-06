package secrets

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Secret represents a secret value with optional metadata.
type Secret struct {
	Key       string            `json:"key"`
	Value     string            `json:"value"`
	Version   string            `json:"version,omitempty"`
	ExpiresAt *time.Time        `json:"expires_at,omitempty"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

// Provider defines the interface for secret storage backends.
type Provider interface {
	// Get retrieves a secret by key.
	Get(ctx context.Context, key string) (*Secret, error)

	// Put stores a secret.
	Put(ctx context.Context, key string, value string, opts ...PutOption) error

	// Delete removes a secret.
	Delete(ctx context.Context, key string) error

	// List returns all secret keys matching the prefix.
	List(ctx context.Context, prefix string) ([]string, error)

	// Health checks if the provider is operational.
	Health(ctx context.Context) error
}

// PutOptions configures secret storage options.
type PutOptions struct {
	TTL      time.Duration
	Metadata map[string]string
}

// PutOption is a functional option for Put operations.
type PutOption func(*PutOptions)

// WithTTL sets the time-to-live for a secret.
func WithTTL(ttl time.Duration) PutOption {
	return func(o *PutOptions) {
		o.TTL = ttl
	}
}

// WithMetadata sets metadata for a secret.
func WithMetadata(m map[string]string) PutOption {
	return func(o *PutOptions) {
		o.Metadata = m
	}
}

// Manager handles secret retrieval with caching and rotation.
type Manager struct {
	providers map[string]Provider
	primary   string
	cache     *secretCache
	mu        sync.RWMutex
}

// ManagerConfig configures the secret manager.
type ManagerConfig struct {
	// PrimaryProvider is the name of the main secrets provider.
	PrimaryProvider string

	// CacheTTL is how long to cache secrets in memory (0 = no caching).
	CacheTTL time.Duration

	// RefreshInterval is how often to refresh cached secrets.
	RefreshInterval time.Duration
}

// NewManager creates a new secret manager.
func NewManager(cfg ManagerConfig) *Manager {
	m := &Manager{
		providers: make(map[string]Provider),
		primary:   cfg.PrimaryProvider,
	}

	if cfg.CacheTTL > 0 {
		m.cache = newSecretCache(cfg.CacheTTL)
	}

	return m
}

// RegisterProvider adds a secrets provider.
func (m *Manager) RegisterProvider(name string, p Provider) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.providers[name] = p
	if m.primary == "" {
		m.primary = name
	}
}

// Get retrieves a secret using the primary provider.
func (m *Manager) Get(ctx context.Context, key string) (*Secret, error) {
	return m.GetFrom(ctx, m.primary, key)
}

// GetFrom retrieves a secret from a specific provider.
func (m *Manager) GetFrom(ctx context.Context, providerName, key string) (*Secret, error) {
	// Check cache first
	if m.cache != nil {
		if secret, found := m.cache.get(key); found {
			return secret, nil
		}
	}

	m.mu.RLock()
	p, ok := m.providers[providerName]
	m.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("provider %q not found", providerName)
	}

	secret, err := p.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	// Update cache
	if m.cache != nil {
		m.cache.set(key, secret)
	}

	return secret, nil
}

// GetString retrieves a secret value as a string.
func (m *Manager) GetString(ctx context.Context, key string) (string, error) {
	secret, err := m.Get(ctx, key)
	if err != nil {
		return "", err
	}
	return secret.Value, nil
}

// GetOrDefault retrieves a secret or returns a default value.
func (m *Manager) GetOrDefault(ctx context.Context, key, defaultValue string) string {
	val, err := m.GetString(ctx, key)
	if err != nil {
		return defaultValue
	}
	return val
}

// Put stores a secret using the primary provider.
func (m *Manager) Put(ctx context.Context, key, value string, opts ...PutOption) error {
	m.mu.RLock()
	p, ok := m.providers[m.primary]
	m.mu.RUnlock()

	if !ok {
		return fmt.Errorf("primary provider %q not found", m.primary)
	}

	if err := p.Put(ctx, key, value, opts...); err != nil {
		return err
	}

	// Invalidate cache
	if m.cache != nil {
		m.cache.delete(key)
	}

	return nil
}

// Health checks all registered providers.
func (m *Manager) Health(ctx context.Context) map[string]error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	results := make(map[string]error)
	for name, p := range m.providers {
		results[name] = p.Health(ctx)
	}
	return results
}

// secretCache provides in-memory caching for secrets.
type secretCache struct {
	data map[string]*cacheEntry
	ttl  time.Duration
	mu   sync.RWMutex
}

type cacheEntry struct {
	secret    *Secret
	expiresAt time.Time
}

func newSecretCache(ttl time.Duration) *secretCache {
	return &secretCache{
		data: make(map[string]*cacheEntry),
		ttl:  ttl,
	}
}

func (c *secretCache) get(key string) (*Secret, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, ok := c.data[key]
	if !ok || time.Now().After(entry.expiresAt) {
		return nil, false
	}
	return entry.secret, true
}

func (c *secretCache) set(key string, secret *Secret) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data[key] = &cacheEntry{
		secret:    secret,
		expiresAt: time.Now().Add(c.ttl),
	}
}

func (c *secretCache) delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.data, key)
}
