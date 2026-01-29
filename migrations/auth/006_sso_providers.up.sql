-- Migration: SSO Providers Configuration
CREATE TABLE IF NOT EXISTS sso_providers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    provider_type TEXT NOT NULL, -- 'saml' or 'oidc'
    issuer_url TEXT,
    client_id TEXT,
    client_secret TEXT,
    metadata_url TEXT,
    sso_url TEXT,
    certificate TEXT,
    active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_sso_providers_org ON sso_providers(org_id);
