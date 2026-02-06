package sso

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Provider represents an SSO identity provider.
type Provider interface {
	// GetAuthURL returns the URL to redirect users to for authentication.
	GetAuthURL(state string) string

	// ExchangeCode exchanges an authorization code for tokens.
	ExchangeCode(ctx context.Context, code string) (*Tokens, error)

	// GetUserInfo retrieves user information using the access token.
	GetUserInfo(ctx context.Context, accessToken string) (*UserInfo, error)

	// ValidateToken validates an ID token and returns claims.
	ValidateToken(ctx context.Context, idToken string) (*Claims, error)
}

// Tokens represents OAuth/OIDC tokens.
type Tokens struct {
	AccessToken  string    `json:"access_token"`
	IDToken      string    `json:"id_token,omitempty"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	TokenType    string    `json:"token_type"`
	ExpiresIn    int       `json:"expires_in"`
	ExpiresAt    time.Time `json:"-"`
}

// UserInfo represents user information from the identity provider.
type UserInfo struct {
	ID            string   `json:"id"`
	Email         string   `json:"email"`
	EmailVerified bool     `json:"email_verified"`
	Name          string   `json:"name"`
	GivenName     string   `json:"given_name"`
	FamilyName    string   `json:"family_name"`
	Picture       string   `json:"picture"`
	Locale        string   `json:"locale"`
	Groups        []string `json:"groups,omitempty"`
}

// Claims represents validated token claims.
type Claims struct {
	Subject   string   `json:"sub"`
	Email     string   `json:"email"`
	Name      string   `json:"name"`
	Groups    []string `json:"groups,omitempty"`
	IssuedAt  int64    `json:"iat"`
	ExpiresAt int64    `json:"exp"`
	Issuer    string   `json:"iss"`
	Audience  string   `json:"aud"`
}

// OIDCConfig configures an OIDC provider.
type OIDCConfig struct {
	// ProviderURL is the OIDC provider URL (e.g., https://accounts.google.com).
	ProviderURL string

	// ClientID is the OAuth client ID.
	ClientID string

	// ClientSecret is the OAuth client secret.
	ClientSecret string

	// RedirectURL is the callback URL.
	RedirectURL string

	// Scopes are the requested OAuth scopes.
	Scopes []string
}

// OIDCProvider implements OIDC authentication.
type OIDCProvider struct {
	config     OIDCConfig
	discovery  *OIDCDiscovery
	httpClient *http.Client
}

// OIDCDiscovery contains discovered OIDC endpoints.
type OIDCDiscovery struct {
	Issuer                string   `json:"issuer"`
	AuthorizationEndpoint string   `json:"authorization_endpoint"`
	TokenEndpoint         string   `json:"token_endpoint"`
	UserinfoEndpoint      string   `json:"userinfo_endpoint"`
	JwksURI               string   `json:"jwks_uri"`
	ScopesSupported       []string `json:"scopes_supported"`
}

// NewOIDCProvider creates a new OIDC provider.
func NewOIDCProvider(ctx context.Context, cfg OIDCConfig) (*OIDCProvider, error) {
	p := &OIDCProvider{
		config: cfg,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}

	// Discover OIDC endpoints
	if err := p.discover(ctx); err != nil {
		return nil, fmt.Errorf("OIDC discovery failed: %w", err)
	}

	return p, nil
}

func (p *OIDCProvider) discover(ctx context.Context) error {
	wellKnownURL := strings.TrimSuffix(p.config.ProviderURL, "/") + "/.well-known/openid-configuration"

	req, err := http.NewRequestWithContext(ctx, "GET", wellKnownURL, nil)
	if err != nil {
		return err
	}

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("discovery failed with status %d", resp.StatusCode)
	}

	var discovery OIDCDiscovery
	if err := json.NewDecoder(resp.Body).Decode(&discovery); err != nil {
		return err
	}

	p.discovery = &discovery
	return nil
}

// GetAuthURL returns the authorization URL for OIDC login.
func (p *OIDCProvider) GetAuthURL(state string) string {
	scopes := p.config.Scopes
	if len(scopes) == 0 {
		scopes = []string{"openid", "email", "profile"}
	}

	params := url.Values{
		"client_id":     {p.config.ClientID},
		"redirect_uri":  {p.config.RedirectURL},
		"response_type": {"code"},
		"scope":         {strings.Join(scopes, " ")},
		"state":         {state},
	}

	return p.discovery.AuthorizationEndpoint + "?" + params.Encode()
}

// ExchangeCode exchanges an authorization code for tokens.
func (p *OIDCProvider) ExchangeCode(ctx context.Context, code string) (*Tokens, error) {
	data := url.Values{
		"grant_type":    {"authorization_code"},
		"code":          {code},
		"redirect_uri":  {p.config.RedirectURL},
		"client_id":     {p.config.ClientID},
		"client_secret": {p.config.ClientSecret},
	}

	req, err := http.NewRequestWithContext(ctx, "POST", p.discovery.TokenEndpoint,
		strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("token exchange failed: %s", string(body))
	}

	var tokens Tokens
	if err := json.NewDecoder(resp.Body).Decode(&tokens); err != nil {
		return nil, err
	}

	tokens.ExpiresAt = time.Now().Add(time.Duration(tokens.ExpiresIn) * time.Second)
	return &tokens, nil
}

// GetUserInfo retrieves user information from the OIDC provider.
func (p *OIDCProvider) GetUserInfo(ctx context.Context, accessToken string) (*UserInfo, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", p.discovery.UserinfoEndpoint, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("userinfo request failed with status %d", resp.StatusCode)
	}

	var info UserInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, err
	}

	return &info, nil
}

// ValidateToken validates an ID token (simplified - production would verify JWT signature).
func (p *OIDCProvider) ValidateToken(ctx context.Context, idToken string) (*Claims, error) {
	// In production, this would:
	// 1. Fetch JWKS from p.discovery.JwksURI
	// 2. Parse and verify the JWT signature
	// 3. Validate claims (exp, iss, aud)
	// For now, we decode without verification (development only)

	parts := strings.Split(idToken, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid token format")
	}

	// This is a stub - real implementation would verify signature
	return nil, fmt.Errorf("token validation not implemented - use a JWT library")
}

// ProviderRegistry manages multiple SSO providers per tenant.
type ProviderRegistry struct {
	providers map[string]Provider // tenantID -> provider
}

// NewProviderRegistry creates a new provider registry.
func NewProviderRegistry() *ProviderRegistry {
	return &ProviderRegistry{
		providers: make(map[string]Provider),
	}
}

// Register registers a provider for a tenant.
func (r *ProviderRegistry) Register(tenantID string, provider Provider) {
	r.providers[tenantID] = provider
}

// Get retrieves the provider for a tenant.
func (r *ProviderRegistry) Get(tenantID string) (Provider, bool) {
	p, ok := r.providers[tenantID]
	return p, ok
}
