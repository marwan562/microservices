package secrets

import (
	"context"
	"fmt"
	"os"
	"strings"
)

// Config represents the secrets configuration.
type Config struct {
	// Provider is the secrets provider type: "vault", "aws", "env".
	Provider string `json:"provider" yaml:"provider"`

	// Vault configuration (when Provider is "vault").
	Vault VaultConfig `json:"vault" yaml:"vault"`

	// AWS configuration (when Provider is "aws").
	AWS AWSConfig `json:"aws" yaml:"aws"`

	// Env configuration (when Provider is "env").
	Env EnvConfig `json:"env" yaml:"env"`

	// Cache configuration.
	CacheTTL string `json:"cache_ttl" yaml:"cache_ttl"`
}

// LoadFromEnv loads secrets configuration from environment variables.
func LoadFromEnv() (*Config, error) {
	provider := getEnvOrDefault("SAPLIY_SECRETS_PROVIDER", "env")

	cfg := &Config{
		Provider: provider,
		CacheTTL: getEnvOrDefault("SAPLIY_SECRETS_CACHE_TTL", "5m"),
	}

	switch provider {
	case "vault":
		cfg.Vault = VaultConfig{
			Address:   os.Getenv("VAULT_ADDR"),
			Token:     os.Getenv("VAULT_TOKEN"),
			Namespace: os.Getenv("VAULT_NAMESPACE"),
			MountPath: getEnvOrDefault("VAULT_MOUNT_PATH", "secret"),
		}
	case "aws":
		cfg.AWS = AWSConfig{
			Region:   os.Getenv("AWS_REGION"),
			Endpoint: os.Getenv("AWS_SECRETSMANAGER_ENDPOINT"),
			Prefix:   getEnvOrDefault("AWS_SECRETS_PREFIX", "sapliy"),
		}
	case "env":
		cfg.Env = EnvConfig{
			Prefix: getEnvOrDefault("SAPLIY_ENV_PREFIX", "SAPLIY"),
		}
	}

	return cfg, nil
}

// NewManagerFromConfig creates a secrets Manager from configuration.
func NewManagerFromConfig(ctx context.Context, cfg *Config) (*Manager, error) {
	manager := NewManager(ManagerConfig{
		PrimaryProvider: cfg.Provider,
	})

	switch cfg.Provider {
	case "vault":
		vault, err := NewVaultProvider(cfg.Vault)
		if err != nil {
			return nil, fmt.Errorf("failed to create vault provider: %w", err)
		}
		manager.RegisterProvider("vault", vault)

	case "aws":
		aws, err := NewAWSProvider(ctx, cfg.AWS)
		if err != nil {
			return nil, fmt.Errorf("failed to create AWS provider: %w", err)
		}
		manager.RegisterProvider("aws", aws)

	case "env":
		env := NewEnvProvider(cfg.Env)
		manager.RegisterProvider("env", env)

	default:
		return nil, fmt.Errorf("unknown secrets provider: %s", cfg.Provider)
	}

	return manager, nil
}

// MustNewManager creates a secrets Manager or panics on error.
func MustNewManager(ctx context.Context) *Manager {
	cfg, err := LoadFromEnv()
	if err != nil {
		panic(fmt.Sprintf("failed to load secrets config: %v", err))
	}

	manager, err := NewManagerFromConfig(ctx, cfg)
	if err != nil {
		panic(fmt.Sprintf("failed to create secrets manager: %v", err))
	}

	return manager
}

func getEnvOrDefault(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}

// DatabaseSecrets represents database connection secrets.
type DatabaseSecrets struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Database string `json:"database"`
	Username string `json:"username"`
	Password string `json:"password"`
	SSLMode  string `json:"ssl_mode"`
}

// DSN returns the PostgreSQL connection string.
func (d *DatabaseSecrets) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		d.Host, d.Port, d.Username, d.Password, d.Database, d.SSLMode,
	)
}

// RedisSecrets represents Redis connection secrets.
type RedisSecrets struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Password string `json:"password"`
	DB       int    `json:"db"`
}

// Addr returns the Redis address.
func (r *RedisSecrets) Addr() string {
	return fmt.Sprintf("%s:%d", r.Host, r.Port)
}

// KafkaSecrets represents Kafka connection secrets.
type KafkaSecrets struct {
	Brokers  []string `json:"brokers"`
	Username string   `json:"username"`
	Password string   `json:"password"`
	SASL     bool     `json:"sasl"`
}

// BrokersString returns comma-separated brokers.
func (k *KafkaSecrets) BrokersString() string {
	return strings.Join(k.Brokers, ",")
}

// APIKeySecrets represents API key configuration.
type APIKeySecrets struct {
	EncryptionKey string `json:"encryption_key"`
	SigningKey    string `json:"signing_key"`
}
