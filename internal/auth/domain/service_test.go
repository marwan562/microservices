package domain

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestAuthService_CreateUser(t *testing.T) {
	ctx := context.Background()
	email := "test@example.com"
	passwordHash := "hashed-pwd"
	userID := "user-123"

	repo := &MockRepository{
		CreateUserFunc: func(ctx context.Context, e, p string) (*User, error) {
			if e != email || p != passwordHash {
				return nil, errors.New("unexpected arguments to CreateUser")
			}
			return &User{ID: userID, Email: e}, nil
		},
		CreateEmailVerificationTokenFunc: func(ctx context.Context, t *EmailVerificationToken) error {
			if t.UserID != userID {
				return errors.New("unexpected userID in verification token")
			}
			t.Token = "hashed-token" // Simulate hashing in repo if needed, but service returns raw
			return nil
		},
	}

	publisher := &MockPublisher{
		PublishFunc: func(ctx context.Context, topic string, event interface{}) error {
			evt := event.(map[string]interface{})
			if evt["type"] != "user.registered" {
				t.Errorf("expected event type user.registered, got %v", evt["type"])
			}
			data := evt["data"].(map[string]string)
			if data["user_id"] != userID || data["email"] != email {
				t.Errorf("unexpected event data: %v", data)
			}
			if data["token"] == "" || !strings.Contains(data["link"], "/verify-email?token=") {
				t.Errorf("missing or invalid token/link in event data: %v", data)
			}
			return nil
		},
	}

	service := NewAuthService(repo, publisher)
	user, err := service.CreateUser(ctx, email, passwordHash)

	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}
	if user.ID != userID {
		t.Errorf("expected userID %s, got %s", userID, user.ID)
	}
}

func TestAuthService_CreatePasswordResetToken(t *testing.T) {
	ctx := context.Background()
	userID := "user-123"
	email := "test@example.com"

	repo := &MockRepository{
		GetUserByIDFunc: func(ctx context.Context, id string) (*User, error) {
			if id != userID {
				return nil, errors.New("user not found")
			}
			return &User{ID: userID, Email: email}, nil
		},
		CreatePasswordResetTokenFunc: func(ctx context.Context, token *PasswordResetToken) error {
			if token.UserID != userID {
				return errors.New("unexpected userID in reset token")
			}
			return nil
		},
	}

	publisher := &MockPublisher{
		PublishFunc: func(ctx context.Context, topic string, event interface{}) error {
			evt := event.(map[string]interface{})
			if evt["type"] != "password.reset" {
				t.Errorf("expected event type password.reset, got %v", evt["type"])
			}
			data := evt["data"].(map[string]string)
			if data["email"] != email || data["user_id"] != userID {
				t.Errorf("unexpected event data: %v", data)
			}
			return nil
		},
	}

	service := NewAuthService(repo, publisher)
	token, err := service.CreatePasswordResetToken(ctx, userID)

	if err != nil {
		t.Fatalf("CreatePasswordResetToken failed: %v", err)
	}
	if token == "" {
		t.Error("expected non-empty token")
	}
}

func TestAuthService_VerifyEmail(t *testing.T) {
	ctx := context.Background()
	rawToken := "raw-verify-token"
	userID := "user-123"

	repo := &MockRepository{
		GetEmailVerificationTokenFunc: func(ctx context.Context, hash string) (*EmailVerificationToken, error) {
			return &EmailVerificationToken{
				UserID:    userID,
				ExpiresAt: time.Now().Add(1 * time.Hour),
			}, nil
		},
		SetEmailVerifiedFunc: func(ctx context.Context, id string) error {
			if id != userID {
				return errors.New("unexpected userID in SetEmailVerified")
			}
			return nil
		},
		MarkEmailVerificationTokenUsedFunc: func(ctx context.Context, hash string) error {
			return nil
		},
	}

	service := NewAuthService(repo, nil)
	err := service.VerifyEmail(ctx, rawToken)

	if err != nil {
		t.Fatalf("VerifyEmail failed: %v", err)
	}
}

func TestAuthService_VerifyEmail_Expired(t *testing.T) {
	ctx := context.Background()
	repo := &MockRepository{
		GetEmailVerificationTokenFunc: func(ctx context.Context, hash string) (*EmailVerificationToken, error) {
			return &EmailVerificationToken{
				ExpiresAt: time.Now().Add(-1 * time.Hour), // Expired
			}, nil
		},
	}
	service := NewAuthService(repo, nil)
	err := service.VerifyEmail(ctx, "token")
	if err == nil || !strings.Contains(err.Error(), "expired") {
		t.Errorf("expected expired error, got %v", err)
	}
}

func TestAuthService_VerifyEmail_AlreadyUsed(t *testing.T) {
	ctx := context.Background()
	now := time.Now()
	repo := &MockRepository{
		GetEmailVerificationTokenFunc: func(ctx context.Context, hash string) (*EmailVerificationToken, error) {
			return &EmailVerificationToken{
				ExpiresAt: time.Now().Add(1 * time.Hour),
				UsedAt:    &now,
			}, nil
		},
	}
	service := NewAuthService(repo, nil)
	err := service.VerifyEmail(ctx, "token")
	if err == nil || !strings.Contains(err.Error(), "already used") {
		t.Errorf("expected already used error, got %v", err)
	}
}

func TestAuthService_HasPermission(t *testing.T) {
	ctx := context.Background()
	userID := "u1"
	orgID := "o1"

	tests := []struct {
		name         string
		userRole     string
		requiredRole string
		want         bool
	}{
		{"Owner can do anything", RoleOwner, RoleOwner, true},
		{"Owner can do admin", RoleOwner, RoleAdmin, true},
		{"Admin can do admin", RoleAdmin, RoleAdmin, true},
		{"Admin can do member", RoleAdmin, RoleMember, true},
		{"Member cannot do admin", RoleMember, RoleAdmin, false},
		{"Guest/None cannot do anything", "none", RoleMember, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &MockRepository{
				GetMembershipFunc: func(ctx context.Context, u, o string) (*Membership, error) {
					return &Membership{Role: tt.userRole}, nil
				},
			}
			service := NewAuthService(repo, nil)
			got, err := service.HasPermission(ctx, userID, orgID, tt.requiredRole)
			if err != nil {
				t.Fatalf("HasPermission failed: %v", err)
			}
			if got != tt.want {
				t.Errorf("HasPermission(%s, %s) = %v; want %v", tt.userRole, tt.requiredRole, got, tt.want)
			}
		})
	}
}
func TestAuthService_ValidateOAuthToken_Expired(t *testing.T) {
	ctx := context.Background()
	repo := &MockRepository{
		ValidateOAuthTokenFunc: func(ctx context.Context, token string) (*OAuthToken, error) {
			return &OAuthToken{
				AccessToken: token,
				ExpiresAt:   time.Now().Add(-1 * time.Hour),
			}, nil
		},
	}
	service := NewAuthService(repo, nil)
	_, err := service.ValidateOAuthToken(ctx, "expired-token")
	if err == nil || !strings.Contains(err.Error(), "expired") {
		t.Errorf("expected expired error, got %v", err)
	}
}

func TestAuthService_ValidateRefreshToken_Revoked(t *testing.T) {
	ctx := context.Background()
	repo := &MockRepository{
		GetRefreshTokenFunc: func(ctx context.Context, hash string) (*RefreshToken, error) {
			return &RefreshToken{
				Revoked:   true,
				ExpiresAt: time.Now().Add(1 * time.Hour),
			}, nil
		},
	}
	service := NewAuthService(repo, nil)
	_, err := service.ValidateRefreshToken(ctx, "token")
	if err == nil || !strings.Contains(err.Error(), "revoked") {
		t.Errorf("expected revoked error, got %v", err)
	}
}

func TestAuthService_ValidateRefreshToken_Expired(t *testing.T) {
	ctx := context.Background()
	repo := &MockRepository{
		GetRefreshTokenFunc: func(ctx context.Context, hash string) (*RefreshToken, error) {
			return &RefreshToken{
				Revoked:   false,
				ExpiresAt: time.Now().Add(-1 * time.Hour),
			}, nil
		},
	}
	service := NewAuthService(repo, nil)
	_, err := service.ValidateRefreshToken(ctx, "token")
	if err == nil || !strings.Contains(err.Error(), "expired") {
		t.Errorf("expected expired error, got %v", err)
	}
}

func TestAuthService_CreateUser_VerificationTokenFail(t *testing.T) {
	ctx := context.Background()
	repo := &MockRepository{
		CreateUserFunc: func(ctx context.Context, email, pwd string) (*User, error) {
			return &User{ID: "u1", Email: email}, nil
		},
		CreateEmailVerificationTokenFunc: func(ctx context.Context, token *EmailVerificationToken) error {
			return errors.New("db error")
		},
	}
	service := NewAuthService(repo, nil)
	user, err := service.CreateUser(ctx, "test@example.com", "long-password")
	if err != nil {
		t.Fatalf("CreateUser should not fail if verification token fails currently: %v", err)
	}
	if user == nil {
		t.Error("expected user to be returned even if token fails")
	}
}
