package domain

import (
	"context"
)

type MockRepository struct {
	CreateUserFunc                func(ctx context.Context, email, passwordHash string) (*User, error)
	GetUserByEmailFunc            func(ctx context.Context, email string) (*User, error)
	GetUserByIDFunc               func(ctx context.Context, id string) (*User, error)
	GetUserByExternalIDFunc       func(ctx context.Context, provider, providerUserID string) (*User, error)
	LinkExternalIdentityFunc      func(ctx context.Context, userID, provider, providerUserID string) error
	CreateOrganizationFunc        func(ctx context.Context, name, domain string) (*Organization, error)
	GetOrganizationFunc           func(ctx context.Context, id string) (*Organization, error)
	AddMemberFunc                 func(ctx context.Context, userID, orgID, role string) error
	RemoveMemberFunc              func(ctx context.Context, userID, orgID string) error
	UpdateMemberRoleFunc          func(ctx context.Context, userID, orgID, role string) error
	ListOrgMembersFunc            func(ctx context.Context, orgID string) ([]Membership, error)
	GetUserMembershipsFunc        func(ctx context.Context, userID string) ([]Membership, error)
	GetMembershipFunc             func(ctx context.Context, userID, orgID string) (*Membership, error)
	CreateAPIKeyFunc              func(ctx context.Context, key *APIKey) error
	GetAPIKeyByHashFunc           func(ctx context.Context, hash string) (*APIKey, error)
	GetClientByIDFunc             func(ctx context.Context, clientID string) (*OAuthClient, error)
	CreateOAuthClientFunc         func(ctx context.Context, client *OAuthClient) error
	AddRedirectURIFunc            func(ctx context.Context, clientID, redirectURI string) error
	ValidateRedirectURIFunc       func(ctx context.Context, clientID, redirectURI string) (bool, error)
	CreateOAuthTokenFunc          func(ctx context.Context, token *OAuthToken) error
	ValidateOAuthTokenFunc        func(ctx context.Context, accessToken string) (*OAuthToken, error)
	CreateAuthorizationCodeFunc   func(ctx context.Context, code *AuthorizationCode) error
	GetAuthorizationCodeFunc      func(ctx context.Context, code string) (*AuthorizationCode, error)
	MarkAuthorizationCodeUsedFunc func(ctx context.Context, code string) error
	CreateSSOProviderFunc         func(ctx context.Context, p *SSOProvider) error
	GetSSOProviderByIDFunc        func(ctx context.Context, id string) (*SSOProvider, error)
	GetSSOProviderByDomainFunc    func(ctx context.Context, domain string) (*SSOProvider, error)
	CreateAuditLogFunc            func(ctx context.Context, log *AuditLog) error
	GetAuditLogsFunc              func(ctx context.Context, orgID string, limit, offset int, action string) ([]AuditLog, int, error)
}

func (m *MockRepository) CreateUser(ctx context.Context, email, passwordHash string) (*User, error) {
	return m.CreateUserFunc(ctx, email, passwordHash)
}

func (m *MockRepository) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	return m.GetUserByEmailFunc(ctx, email)
}

func (m *MockRepository) GetUserByID(ctx context.Context, id string) (*User, error) {
	return m.GetUserByIDFunc(ctx, id)
}

func (m *MockRepository) GetUserByExternalID(ctx context.Context, provider, providerUserID string) (*User, error) {
	return m.GetUserByExternalIDFunc(ctx, provider, providerUserID)
}

func (m *MockRepository) LinkExternalIdentity(ctx context.Context, userID, provider, providerUserID string) error {
	return m.LinkExternalIdentityFunc(ctx, userID, provider, providerUserID)
}

func (m *MockRepository) CreateOrganization(ctx context.Context, name, domain string) (*Organization, error) {
	return m.CreateOrganizationFunc(ctx, name, domain)
}

func (m *MockRepository) GetOrganization(ctx context.Context, id string) (*Organization, error) {
	return m.GetOrganizationFunc(ctx, id)
}

func (m *MockRepository) AddMember(ctx context.Context, userID, orgID, role string) error {
	return m.AddMemberFunc(ctx, userID, orgID, role)
}

func (m *MockRepository) RemoveMember(ctx context.Context, userID, orgID string) error {
	return m.RemoveMemberFunc(ctx, userID, orgID)
}

func (m *MockRepository) UpdateMemberRole(ctx context.Context, userID, orgID, role string) error {
	return m.UpdateMemberRoleFunc(ctx, userID, orgID, role)
}

func (m *MockRepository) ListOrgMembers(ctx context.Context, orgID string) ([]Membership, error) {
	return m.ListOrgMembersFunc(ctx, orgID)
}

func (m *MockRepository) GetUserMemberships(ctx context.Context, userID string) ([]Membership, error) {
	return m.GetUserMembershipsFunc(ctx, userID)
}

func (m *MockRepository) GetMembership(ctx context.Context, userID, orgID string) (*Membership, error) {
	return m.GetMembershipFunc(ctx, userID, orgID)
}

func (m *MockRepository) CreateAPIKey(ctx context.Context, key *APIKey) error {
	return m.CreateAPIKeyFunc(ctx, key)
}

func (m *MockRepository) GetAPIKeyByHash(ctx context.Context, hash string) (*APIKey, error) {
	return m.GetAPIKeyByHashFunc(ctx, hash)
}

func (m *MockRepository) GetClientByID(ctx context.Context, clientID string) (*OAuthClient, error) {
	return m.GetClientByIDFunc(ctx, clientID)
}

func (m *MockRepository) CreateOAuthClient(ctx context.Context, client *OAuthClient) error {
	return m.CreateOAuthClientFunc(ctx, client)
}

func (m *MockRepository) AddRedirectURI(ctx context.Context, clientID, redirectURI string) error {
	return m.AddRedirectURIFunc(ctx, clientID, redirectURI)
}

func (m *MockRepository) ValidateRedirectURI(ctx context.Context, clientID, redirectURI string) (bool, error) {
	return m.ValidateRedirectURIFunc(ctx, clientID, redirectURI)
}

func (m *MockRepository) CreateOAuthToken(ctx context.Context, token *OAuthToken) error {
	return m.CreateOAuthTokenFunc(ctx, token)
}

func (m *MockRepository) ValidateOAuthToken(ctx context.Context, accessToken string) (*OAuthToken, error) {
	return m.ValidateOAuthTokenFunc(ctx, accessToken)
}

func (m *MockRepository) CreateAuthorizationCode(ctx context.Context, code *AuthorizationCode) error {
	return m.CreateAuthorizationCodeFunc(ctx, code)
}

func (m *MockRepository) GetAuthorizationCode(ctx context.Context, code string) (*AuthorizationCode, error) {
	return m.GetAuthorizationCodeFunc(ctx, code)
}

func (m *MockRepository) MarkAuthorizationCodeUsed(ctx context.Context, code string) error {
	return m.MarkAuthorizationCodeUsedFunc(ctx, code)
}

func (m *MockRepository) CreateSSOProvider(ctx context.Context, p *SSOProvider) error {
	return m.CreateSSOProviderFunc(ctx, p)
}

func (m *MockRepository) GetSSOProviderByID(ctx context.Context, id string) (*SSOProvider, error) {
	return m.GetSSOProviderByIDFunc(ctx, id)
}

func (m *MockRepository) GetSSOProviderByDomain(ctx context.Context, domain string) (*SSOProvider, error) {
	return m.GetSSOProviderByDomainFunc(ctx, domain)
}

func (m *MockRepository) CreateAuditLog(ctx context.Context, log *AuditLog) error {
	return m.CreateAuditLogFunc(ctx, log)
}

func (m *MockRepository) GetAuditLogs(ctx context.Context, orgID string, limit, offset int, action string) ([]AuditLog, int, error) {
	return m.GetAuditLogsFunc(ctx, orgID, limit, offset, action)
}
