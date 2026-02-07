package domain

import (
	"context"
)

type Repository interface {
	// User methods
	CreateUser(ctx context.Context, email, passwordHash string) (*User, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	GetUserByID(ctx context.Context, id string) (*User, error)
	GetUserByExternalID(ctx context.Context, provider, providerUserID string) (*User, error)
	LinkExternalIdentity(ctx context.Context, userID, provider, providerUserID string) error
	UpdateUserPassword(ctx context.Context, userID, passwordHash string) error
	SetEmailVerified(ctx context.Context, userID string) error

	// Password Reset Token methods
	CreatePasswordResetToken(ctx context.Context, token *PasswordResetToken) error
	GetPasswordResetToken(ctx context.Context, tokenHash string) (*PasswordResetToken, error)
	MarkPasswordResetTokenUsed(ctx context.Context, tokenHash string) error

	// Email Verification Token methods
	CreateEmailVerificationToken(ctx context.Context, token *EmailVerificationToken) error
	GetEmailVerificationToken(ctx context.Context, tokenHash string) (*EmailVerificationToken, error)
	MarkEmailVerificationTokenUsed(ctx context.Context, tokenHash string) error

	// Organization methods
	CreateOrganization(ctx context.Context, name, domain string) (*Organization, error)
	GetOrganization(ctx context.Context, id string) (*Organization, error)
	AddMember(ctx context.Context, userID, orgID, role string) error
	RemoveMember(ctx context.Context, userID, orgID string) error
	UpdateMemberRole(ctx context.Context, userID, orgID, role string) error
	ListOrgMembers(ctx context.Context, orgID string) ([]Membership, error)
	GetUserMemberships(ctx context.Context, userID string) ([]Membership, error)
	GetMembership(ctx context.Context, userID, orgID string) (*Membership, error)

	// APIKey methods
	CreateAPIKey(ctx context.Context, key *APIKey) error
	GetAPIKeyByHash(ctx context.Context, hash string) (*APIKey, error)

	// OAuth methods
	GetClientByID(ctx context.Context, clientID string) (*OAuthClient, error)
	CreateOAuthClient(ctx context.Context, client *OAuthClient) error
	AddRedirectURI(ctx context.Context, clientID, redirectURI string) error
	ValidateRedirectURI(ctx context.Context, clientID, redirectURI string) (bool, error)
	CreateOAuthToken(ctx context.Context, token *OAuthToken) error
	ValidateOAuthToken(ctx context.Context, accessToken string) (*OAuthToken, error)
	CreateAuthorizationCode(ctx context.Context, code *AuthorizationCode) error
	GetAuthorizationCode(ctx context.Context, code string) (*AuthorizationCode, error)
	MarkAuthorizationCodeUsed(ctx context.Context, code string) error

	// SSO Provider methods
	CreateSSOProvider(ctx context.Context, p *SSOProvider) error
	GetSSOProviderByID(ctx context.Context, id string) (*SSOProvider, error)
	GetSSOProviderByDomain(ctx context.Context, domain string) (*SSOProvider, error)

	// Audit Log methods
	CreateAuditLog(ctx context.Context, log *AuditLog) error
	GetAuditLogs(ctx context.Context, orgID string, limit, offset int, action string) ([]AuditLog, int, error)
}
