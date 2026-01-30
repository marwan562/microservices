package auth

import (
	"context"
	"database/sql"
	"testing"
	"time"
)

func TestValidateOAuthToken_TableDriven(t *testing.T) {
	tests := []struct {
		name        string
		token       string
		mockSetup   func(*MockDB)
		expectedErr string
	}{
		{
			name:  "Expired Token",
			token: "expired",
			mockSetup: func(m *MockDB) {
				m.QueryRowContextFunc = func(ctx context.Context, query string, args ...any) Row {
					return &MockRow{
						ScanFunc: func(dest ...any) error {
							*(dest[5].(*time.Time)) = time.Now().Add(-1 * time.Hour)
							return nil
						},
					}
				}
			},
			expectedErr: "token expired",
		},
		{
			name:  "Valid Token",
			token: "valid",
			mockSetup: func(m *MockDB) {
				m.QueryRowContextFunc = func(ctx context.Context, query string, args ...any) Row {
					return &MockRow{
						ScanFunc: func(dest ...any) error {
							*(dest[5].(*time.Time)) = time.Now().Add(1 * time.Hour)
							return nil
						},
					}
				}
			},
			expectedErr: "",
		},
		{
			name:  "Not Found",
			token: "missing",
			mockSetup: func(m *MockDB) {
				m.QueryRowContextFunc = func(ctx context.Context, query string, args ...any) Row {
					return &MockRow{
						ScanFunc: func(dest ...any) error {
							return sql.ErrNoRows
						},
					}
				}
			},
			expectedErr: "", // ValidateOAuthToken returns nil, nil for sql.ErrNoRows
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mDB := &MockDB{}
			tt.mockSetup(mDB)
			repo := &Repository{db: mDB}

			_, err := repo.ValidateOAuthToken(context.Background(), tt.token)
			if tt.expectedErr != "" {
				if err == nil || err.Error() != tt.expectedErr {
					t.Errorf("Expected error '%s', got '%v'", tt.expectedErr, err)
				}
			} else if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}
