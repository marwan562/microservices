package auth

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

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

func (r *Repository) CreateSSOProvider(ctx context.Context, p *SSOProvider) error {
	err := r.db.QueryRowContext(ctx,
		`INSERT INTO sso_providers (org_id, name, provider_type, issuer_url, client_id, client_secret, metadata_url, sso_url, certificate)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		 RETURNING id, created_at, updated_at`,
		p.OrgID, p.Name, p.ProviderType, p.IssuerURL, p.ClientID, p.ClientSecret, p.MetadataURL, p.SSOURL, p.Certificate).
		Scan(&p.ID, &p.CreatedAt, &p.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create sso provider: %w", err)
	}
	return nil
}

func (r *Repository) GetSSOProviderByID(ctx context.Context, id string) (*SSOProvider, error) {
	var p SSOProvider
	err := r.db.QueryRowContext(ctx,
		`SELECT id, org_id, name, provider_type, issuer_url, client_id, client_secret, metadata_url, sso_url, certificate, active, created_at, updated_at
		 FROM sso_providers WHERE id = $1`,
		id).Scan(&p.ID, &p.OrgID, &p.Name, &p.ProviderType, &p.IssuerURL, &p.ClientID, &p.ClientSecret, &p.MetadataURL, &p.SSOURL, &p.Certificate, &p.Active, &p.CreatedAt, &p.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get sso provider: %w", err)
	}
	return &p, nil
}

func (r *Repository) GetSSOProviderByDomain(ctx context.Context, domain string) (*SSOProvider, error) {
	var p SSOProvider
	err := r.db.QueryRowContext(ctx,
		`SELECT p.id, p.org_id, p.name, p.provider_type, p.issuer_url, p.client_id, p.client_secret, p.metadata_url, p.sso_url, p.certificate, p.active, p.created_at, p.updated_at
		 FROM sso_providers p
		 JOIN organizations o ON p.org_id = o.id
		 WHERE o.domain = $1 AND p.active = TRUE
		 LIMIT 1`,
		domain).Scan(&p.ID, &p.OrgID, &p.Name, &p.ProviderType, &p.IssuerURL, &p.ClientID, &p.ClientSecret, &p.MetadataURL, &p.SSOURL, &p.Certificate, &p.Active, &p.CreatedAt, &p.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get sso provider by domain: %w", err)
	}
	return &p, nil
}
