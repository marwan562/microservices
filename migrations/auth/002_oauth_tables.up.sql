CREATE TABLE IF NOT EXISTS oauth_clients (
    id VARCHAR(255) PRIMARY KEY,
    client_secret_hash TEXT NOT NULL,
    user_id VARCHAR(255) NOT NULL, -- or service_name for internal services
    name VARCHAR(255) NOT NULL,
    is_public BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS oauth_tokens (
    access_token VARCHAR(255) PRIMARY KEY,
    refresh_token VARCHAR(255) UNIQUE,
    client_id VARCHAR(255) REFERENCES oauth_clients(id) ON DELETE CASCADE,
    user_id VARCHAR(255), -- Optional, for Authorization Code flow later
    scope TEXT,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
