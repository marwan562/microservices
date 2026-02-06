package secrets

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// VaultConfig configures the HashiCorp Vault provider.
type VaultConfig struct {
	// Address is the Vault server address (e.g., "https://vault.example.com:8200").
	Address string

	// Token is the authentication token.
	Token string

	// Namespace is the Vault namespace (enterprise feature).
	Namespace string

	// MountPath is the secrets engine mount path (default: "secret").
	MountPath string

	// Timeout for HTTP requests.
	Timeout time.Duration
}

// VaultProvider implements secrets.Provider for HashiCorp Vault.
type VaultProvider struct {
	config     VaultConfig
	httpClient *http.Client
}

// NewVaultProvider creates a new Vault secrets provider.
func NewVaultProvider(cfg VaultConfig) (*VaultProvider, error) {
	if cfg.Address == "" {
		return nil, fmt.Errorf("vault address is required")
	}
	if cfg.Token == "" {
		return nil, fmt.Errorf("vault token is required")
	}
	if cfg.MountPath == "" {
		cfg.MountPath = "secret"
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = 30 * time.Second
	}

	return &VaultProvider{
		config: cfg,
		httpClient: &http.Client{
			Timeout: cfg.Timeout,
		},
	}, nil
}

// Get retrieves a secret from Vault KV v2.
func (p *VaultProvider) Get(ctx context.Context, key string) (*Secret, error) {
	url := fmt.Sprintf("%s/v1/%s/data/%s", p.config.Address, p.config.MountPath, key)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	p.setHeaders(req)

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("vault request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("secret %q not found", key)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("vault returned status %d", resp.StatusCode)
	}

	var vaultResp vaultReadResponse
	if err := json.NewDecoder(resp.Body).Decode(&vaultResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Extract the value (KV v2 stores data under "data.data.value")
	value, _ := vaultResp.Data.Data["value"].(string)

	return &Secret{
		Key:     key,
		Value:   value,
		Version: fmt.Sprintf("%d", vaultResp.Data.Metadata.Version),
	}, nil
}

// Put stores a secret in Vault KV v2.
func (p *VaultProvider) Put(ctx context.Context, key string, value string, opts ...PutOption) error {
	options := &PutOptions{}
	for _, opt := range opts {
		opt(options)
	}

	url := fmt.Sprintf("%s/v1/%s/data/%s", p.config.Address, p.config.MountPath, key)

	payload := map[string]interface{}{
		"data": map[string]interface{}{
			"value": value,
		},
	}

	if options.Metadata != nil {
		payload["options"] = map[string]interface{}{
			"custom_metadata": options.Metadata,
		}
	}

	body, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(string(body)))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	p.setHeaders(req)
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("vault request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("vault returned status %d", resp.StatusCode)
	}

	return nil
}

// Delete removes a secret from Vault.
func (p *VaultProvider) Delete(ctx context.Context, key string) error {
	url := fmt.Sprintf("%s/v1/%s/metadata/%s", p.config.Address, p.config.MountPath, key)

	req, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	p.setHeaders(req)

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("vault request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("vault returned status %d", resp.StatusCode)
	}

	return nil
}

// List returns all secret keys under the given prefix.
func (p *VaultProvider) List(ctx context.Context, prefix string) ([]string, error) {
	url := fmt.Sprintf("%s/v1/%s/metadata/%s", p.config.Address, p.config.MountPath, prefix)

	req, err := http.NewRequestWithContext(ctx, "LIST", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	p.setHeaders(req)

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("vault request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return []string{}, nil
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("vault returned status %d", resp.StatusCode)
	}

	var listResp vaultListResponse
	if err := json.NewDecoder(resp.Body).Decode(&listResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return listResp.Data.Keys, nil
}

// Health checks Vault connectivity.
func (p *VaultProvider) Health(ctx context.Context) error {
	url := fmt.Sprintf("%s/v1/sys/health", p.config.Address)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("vault health check failed: %w", err)
	}
	defer resp.Body.Close()

	// Vault health returns 200 for initialized, unsealed, active
	// 429 for standby, 472 for DR secondary, 473 for performance standby
	if resp.StatusCode >= 500 {
		return fmt.Errorf("vault is unhealthy: status %d", resp.StatusCode)
	}

	return nil
}

func (p *VaultProvider) setHeaders(req *http.Request) {
	req.Header.Set("X-Vault-Token", p.config.Token)
	if p.config.Namespace != "" {
		req.Header.Set("X-Vault-Namespace", p.config.Namespace)
	}
}

// Vault response types
type vaultReadResponse struct {
	Data struct {
		Data     map[string]interface{} `json:"data"`
		Metadata struct {
			Version int `json:"version"`
		} `json:"metadata"`
	} `json:"data"`
}

type vaultListResponse struct {
	Data struct {
		Keys []string `json:"keys"`
	} `json:"data"`
}
