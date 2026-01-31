package domain

import (
	"encoding/json"
	"time"
)

// User represents a registered user in the system.
type User struct {
	ID        string    `json:"id"`
	OrgID     string    `json:"org_id,omitempty"` // Primary or current organization
	Email     string    `json:"email"`
	Password  string    `json:"-"` // Never return password
	CreatedAt time.Time `json:"created_at"`
}

// Organization represents a team or company.
type Organization struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Domain    string    `json:"domain,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// Membership represents a user's role in an organization.
type Membership struct {
	UserID    string    `json:"user_id"`
	OrgID     string    `json:"org_id"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

// APIKey represents a secret key used for API authentication.
type APIKey struct {
	ID           string     `json:"id"`
	UserID       string     `json:"user_id"`
	OrgID        string     `json:"org_id,omitempty"`
	KeyPrefix    string     `json:"key_prefix"`
	KeyHash      string     `json:"-"`
	TruncatedKey string     `json:"truncated_key"`
	Environment  string     `json:"environment"`
	Scopes       string     `json:"scopes"` // Space-separated scopes, "*" for all
	CreatedAt    time.Time  `json:"created_at"`
	RevokedAt    *time.Time `json:"revoked_at,omitempty"`
}

// OAuthClient represents a registered OAuth client application.
type OAuthClient struct {
	ID               string    `json:"id"`
	ClientSecretHash string    `json:"-"`
	UserID           string    `json:"user_id"`
	Name             string    `json:"name"`
	IsPublic         bool      `json:"is_public"`
	CreatedAt        time.Time `json:"created_at"`
}

// OAuthToken represents an access token issued to a client.
type OAuthToken struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ClientID     string    `json:"client_id"`
	UserID       string    `json:"user_id,omitempty"`
	Scope        string    `json:"scope"`
	ExpiresAt    time.Time `json:"expires_at"`
	CreatedAt    time.Time `json:"created_at"`
}

// AuthorizationCode represents an OAuth2 authorization code with PKCE.
type AuthorizationCode struct {
	Code                string    `json:"code"`
	ClientID            string    `json:"client_id"`
	UserID              string    `json:"user_id"`
	RedirectURI         string    `json:"redirect_uri"`
	Scope               string    `json:"scope"`
	CodeChallenge       string    `json:"code_challenge,omitempty"`
	CodeChallengeMethod string    `json:"code_challenge_method,omitempty"`
	ExpiresAt           time.Time `json:"expires_at"`
	Used                bool      `json:"used"`
	CreatedAt           time.Time `json:"created_at"`
}

// ExternalIdentity represents a linked external account (SSO).
type ExternalIdentity struct {
	ID             string `json:"id"`
	UserID         string `json:"user_id"`
	Provider       string `json:"provider"`
	ProviderUserID string `json:"provider_user_id"`
}

// SSOProvider represents an external identity provider configuration.
type SSOProvider struct {
	ID           string    `json:"id"`
	OrgID        string    `json:"org_id"`
	Name         string    `json:"name"`
	ProviderType string    `json:"provider_type"` // 'saml' or 'oidc'
	IssuerURL    string    `json:"issuer_url,omitempty"`
	ClientID     string    `json:"client_id,omitempty"`
	ClientSecret string    `json:"-"`
	MetadataURL  string    `json:"metadata_url,omitempty"`
	SSOURL       string    `json:"sso_url,omitempty"`
	Certificate  string    `json:"certificate,omitempty"`
	Active       bool      `json:"active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// AuditLog represents a single security or business event.
type AuditLog struct {
	ID           string          `json:"id"`
	OrgID        string          `json:"org_id"`
	UserID       string          `json:"user_id"`
	Action       string          `json:"action"`
	ResourceType string          `json:"resource_type"`
	ResourceID   string          `json:"resource_id"`
	Metadata     json.RawMessage `json:"metadata"`
	IPAddress    string          `json:"ip_address"`
	UserAgent    string          `json:"user_agent"`
	CreatedAt    time.Time       `json:"created_at"`
}

const (
	RoleOwner     = "owner"
	RoleAdmin     = "admin"
	RoleMember    = "member"
	RoleDeveloper = "developer"
)
