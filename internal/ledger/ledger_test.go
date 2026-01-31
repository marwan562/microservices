package ledger

import (
	"context"
	"database/sql"
	"errors"
	"testing"
)

func TestRecordTransaction_TableDriven(t *testing.T) {
	tests := []struct {
		name        string
		req         TransactionRequest
		mockSetup   func(*MockDB)
		expectedErr string
	}{
		{
			name: "Unbalanced Transaction",
			req: TransactionRequest{
				Entries: []EntryRequest{
					{Amount: 100},
					{Amount: -50},
				},
			},
			mockSetup:   func(m *MockDB) {},
			expectedErr: "transaction is not balanced (sum != 0)",
		},
		{
			name: "Account Not Found",
			req: TransactionRequest{
				Entries: []EntryRequest{
					{AccountID: "acc_1", Amount: 100},
					{AccountID: "acc_2", Amount: -100},
				},
			},
			mockSetup: func(m *MockDB) {
				m.QueryRowContextFunc = func(ctx context.Context, query string, args ...any) Row {
					return &MockRow{
						ScanFunc: func(dest ...any) error {
							return sql.ErrNoRows
						},
					}
				}
			},
			expectedErr: "account acc_1 not found",
		},
		// Add more cases here (Currency mismatch, Idempotency, etc)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := &MockDB{}
			tt.mockSetup(mockDB)
			repo := &Repository{db: mockDB}

			err := repo.RecordTransaction(context.Background(), tt.req)
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

func TestCreateAccount_TableDriven(t *testing.T) {
	tests := []struct {
		name        string
		accName     string
		mockSetup   func(*MockDB)
		expectedID  string
		expectedErr bool
	}{
		{
			name:    "Success",
			accName: "Checking",
			mockSetup: func(m *MockDB) {
				m.QueryRowContextFunc = func(ctx context.Context, query string, args ...any) Row {
					return &MockRow{
						ScanFunc: func(dest ...any) error {
							*(dest[0].(*string)) = "acc_123"
							return nil
						},
					}
				}
			},
			expectedID: "acc_123",
		},
		{
			name:    "Database Error",
			accName: "Savings",
			mockSetup: func(m *MockDB) {
				m.QueryRowContextFunc = func(ctx context.Context, query string, args ...any) Row {
					return &MockRow{
						ScanFunc: func(dest ...any) error {
							return errors.New("db error")
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

			acc, err := repo.CreateAccount(context.Background(), tt.accName, Asset, "USD", nil)
			if tt.expectedErr {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if acc.ID != tt.expectedID {
				t.Errorf("Expected ID %s, got %s", tt.expectedID, acc.ID)
			}
		})
	}
}
