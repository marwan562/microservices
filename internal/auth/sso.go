package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

type ExternalIdentity struct {
	ID             string `json:"id"`
	UserID         string `json:"user_id"`
	Provider       string `json:"provider"`
	ProviderUserID string `json:"provider_user_id"`
}

func (r *Repository) GetUserByExternalID(ctx context.Context, provider, providerUserID string) (*User, error) {
	var user User
	err := r.db.QueryRowContext(ctx,
		`SELECT u.id, u.email, u.org_id, u.created_at 
		 FROM users u
		 JOIN external_identities e ON u.id = e.user_id
		 WHERE e.provider = $1 AND e.provider_user_id = $2`,
		provider, providerUserID).Scan(&user.ID, &user.Email, &user.OrgID, &user.CreatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user by external id: %w", err)
	}
	return &user, nil
}

func (r *Repository) LinkExternalIdentity(ctx context.Context, userID, provider, providerUserID string) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO external_identities (user_id, provider, provider_user_id) 
		 VALUES ($1, $2, $3) ON CONFLICT (provider, provider_user_id) DO UPDATE SET user_id = EXCLUDED.user_id`,
		userID, provider, providerUserID)
	return err
}
