package secrets

import (
	"context"
	"fmt"
	"os"
	"sync"
)

// EnvProvider implements secrets.Provider using environment variables.
// Useful for local development and simple deployments.
type EnvProvider struct {
	prefix string
	cache  map[string]string
	mu     sync.RWMutex
}

// EnvConfig configures the environment variable provider.
type EnvConfig struct {
	// Prefix is prepended to all environment variable names.
	// e.g., prefix "SAPLIY" would look for "SAPLIY_DATABASE_URL".
	Prefix string
}

// NewEnvProvider creates a new environment variable secrets provider.
func NewEnvProvider(cfg EnvConfig) *EnvProvider {
	return &EnvProvider{
		prefix: cfg.Prefix,
		cache:  make(map[string]string),
	}
}

// Get retrieves a secret from environment variables.
func (p *EnvProvider) Get(ctx context.Context, key string) (*Secret, error) {
	envKey := p.formatKey(key)

	value := os.Getenv(envKey)
	if value == "" {
		return nil, fmt.Errorf("environment variable %q not set", envKey)
	}

	return &Secret{
		Key:   key,
		Value: value,
	}, nil
}

// Put stores a secret (in memory only for env provider).
// Note: This doesn't persist across restarts.
func (p *EnvProvider) Put(ctx context.Context, key string, value string, opts ...PutOption) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	envKey := p.formatKey(key)
	p.cache[envKey] = value

	// Also set in actual environment for current process
	return os.Setenv(envKey, value)
}

// Delete removes a secret from memory cache.
func (p *EnvProvider) Delete(ctx context.Context, key string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	envKey := p.formatKey(key)
	delete(p.cache, envKey)

	return os.Unsetenv(envKey)
}

// List returns all cached keys matching the prefix.
func (p *EnvProvider) List(ctx context.Context, prefix string) ([]string, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	var keys []string
	for k := range p.cache {
		keys = append(keys, k)
	}
	return keys, nil
}

// Health always returns nil for EnvProvider.
func (p *EnvProvider) Health(ctx context.Context) error {
	return nil
}

func (p *EnvProvider) formatKey(key string) string {
	if p.prefix == "" {
		return key
	}
	return fmt.Sprintf("%s_%s", p.prefix, key)
}

// MustGet retrieves a secret or panics if not found.
// Useful for required configuration at startup.
func (p *EnvProvider) MustGet(key string) string {
	secret, err := p.Get(context.Background(), key)
	if err != nil {
		panic(fmt.Sprintf("required secret %q not found: %v", key, err))
	}
	return secret.Value
}

// GetOrDefault retrieves a secret or returns the default value.
func (p *EnvProvider) GetOrDefault(key, defaultValue string) string {
	secret, err := p.Get(context.Background(), key)
	if err != nil {
		return defaultValue
	}
	return secret.Value
}
