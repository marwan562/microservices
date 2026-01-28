package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"time"
)

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

// GetClientByID retrieves a client by its ID.
func (r *Repository) GetClientByID(ctx context.Context, clientID string) (*OAuthClient, error) {
	var client OAuthClient
	err := r.db.QueryRowContext(ctx,
		"SELECT id, client_secret_hash, user_id, name, is_public, created_at FROM oauth_clients WHERE id = $1",
		clientID).Scan(&client.ID, &client.ClientSecretHash, &client.UserID, &client.Name, &client.IsPublic, &client.CreatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // Not found
		}
		return nil, fmt.Errorf("failed to get client: %w", err)
	}
	return &client, nil
}

// CreateOAuthToken stores a new access token.
func (r *Repository) CreateOAuthToken(ctx context.Context, token *OAuthToken) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO oauth_tokens (access_token, refresh_token, client_id, user_id, scope, expires_at)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		token.AccessToken, token.RefreshToken, token.ClientID, token.UserID, token.Scope, token.ExpiresAt)

	if err != nil {
		return fmt.Errorf("failed to create token: %w", err)
	}
	return nil
}

// ValidateOAuthToken checks if an access token exists and is valid.
func (r *Repository) ValidateOAuthToken(ctx context.Context, accessToken string) (*OAuthToken, error) {
	var token OAuthToken
	err := r.db.QueryRowContext(ctx,
		"SELECT access_token, refresh_token, client_id, user_id, scope, expires_at, created_at FROM oauth_tokens WHERE access_token = $1",
		accessToken).Scan(&token.AccessToken, &token.RefreshToken, &token.ClientID, &token.UserID, &token.Scope, &token.ExpiresAt, &token.CreatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to validate token: %w", err)
	}

	if time.Now().After(token.ExpiresAt) {
		return nil, fmt.Errorf("token expired")
	}

	return &token, nil
}

// GenerateRandomString generates a secure random string (for tokens/secrets).
func GenerateRandomString(n int) (string, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// HashString creates a SHA-256 hash of the input string.
func HashString(s string) string {
	hash := sha256.Sum256([]byte(s))
	return fmt.Sprintf("%x", hash)
}
