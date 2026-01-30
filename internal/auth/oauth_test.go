package auth

import (
	"context"
	"database/sql"
	"testing"
	"time"
)

// MockRow implements the Row interface for testing.
type MockRow struct {
	ScanFunc func(dest ...any) error
}

func (m *MockRow) Scan(dest ...any) error {
	return m.ScanFunc(dest...)
}

// MockDB implements the DB interface for testing.
type MockDB struct {
	QueryRowContextFunc func(ctx context.Context, query string, args ...any) Row
	ExecContextFunc     func(ctx context.Context, query string, args ...any) (sql.Result, error)
}

func (m *MockDB) QueryRowContext(ctx context.Context, query string, args ...any) Row {
	return m.QueryRowContextFunc(ctx, query, args...)
}
func (m *MockDB) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return m.ExecContextFunc(ctx, query, args...)
}

func TestHashString(t *testing.T) {
	s := "test-secret"
	h1 := HashString(s)
	h2 := HashString(s)

	if h1 == "" {
		t.Fatal("Hash should not be empty")
	}
	if h1 != h2 {
		t.Fatal("Hash should be deterministic")
	}
}

func TestValidateOAuthToken_Expiration(t *testing.T) {
	mockDB := &MockDB{}
	repo := &Repository{db: mockDB}

	t.Run("ExpiredToken", func(t *testing.T) {
		mockDB.QueryRowContextFunc = func(ctx context.Context, query string, args ...any) Row {
			return &MockRow{
				ScanFunc: func(dest ...any) error {
					// Scan indices: AccessToken, RefreshToken, ClientID, UserID, Scope, ExpiresAt, CreatedAt
					*(dest[5].(*time.Time)) = time.Now().Add(-1 * time.Hour)
					return nil
				},
			}
		}

		token, err := repo.ValidateOAuthToken(context.Background(), "expired-token")
		if err == nil || err.Error() != "token expired" {
			t.Errorf("Expected token expired error, got %v", err)
		}
		if token != nil {
			t.Error("Expected nil token")
		}
	})
}
