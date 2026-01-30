package payment

import (
	"context"
	"database/sql"
	"errors"
	"testing"
)

func TestCreatePaymentIntent_TableDriven(t *testing.T) {
	tests := []struct {
		name        string
		intent      *PaymentIntent
		mockSetup   func(*MockDB)
		expectedID  string
		expectedErr bool
	}{
		{
			name: "Success with Defaults",
			intent: &PaymentIntent{
				Amount: 1000,
				UserID: "user_123",
			},
			mockSetup: func(m *MockDB) {
				m.QueryRowContextFunc = func(ctx context.Context, query string, args ...any) Row {
					return &MockRow{
						ScanFunc: func(dest ...any) error {
							*(dest[0].(*string)) = "pi_123"
							return nil
						},
					}
				}
			},
			expectedID: "pi_123",
		},
		{
			name: "Database Error",
			intent: &PaymentIntent{
				Amount: 500,
				UserID: "user_456",
			},
			mockSetup: func(m *MockDB) {
				m.QueryRowContextFunc = func(ctx context.Context, query string, args ...any) Row {
					return &MockRow{
						ScanFunc: func(dest ...any) error {
							return errors.New("insert failed")
						},
					}
				}
			},
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := &MockDB{}
			tt.mockSetup(mockDB)
			repo := &Repository{db: mockDB}

			err := repo.CreatePaymentIntent(context.Background(), tt.intent)
			if tt.expectedErr {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if tt.intent.Currency != "USD" {
				t.Errorf("Expected default currency USD, got %s", tt.intent.Currency)
			}
			if tt.intent.ID != tt.expectedID {
				t.Errorf("Expected ID %s, got %s", tt.expectedID, tt.intent.ID)
			}
		})
	}
}

func TestUpdateStatus_TableDriven(t *testing.T) {
	tests := []struct {
		name        string
		id          string
		status      string
		mockSetup   func(*MockDB)
		expectedErr bool
	}{
		{
			name:   "Success",
			id:     "pi_123",
			status: "succeeded",
			mockSetup: func(m *MockDB) {
				m.ExecContextFunc = func(ctx context.Context, query string, args ...any) (sql.Result, error) {
					return nil, nil
				}
			},
		},
		{
			name:   "Database Error",
			id:     "pi_456",
			status: "failed",
			mockSetup: func(m *MockDB) {
				m.ExecContextFunc = func(ctx context.Context, query string, args ...any) (sql.Result, error) {
					return nil, errors.New("update failed")
				}
			},
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := &MockDB{}
			tt.mockSetup(mockDB)
			repo := &Repository{db: mockDB}

			err := repo.UpdateStatus(context.Background(), tt.id, tt.status)
			if tt.expectedErr {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}
