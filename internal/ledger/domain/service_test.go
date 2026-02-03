package domain

import (
	"context"
	"errors"
	"testing"
)

func TestRecordTransaction_TableDriven(t *testing.T) {
	tests := []struct {
		name        string
		req         TransactionRequest
		mockSetup   func(*MockRepository)
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
			mockSetup:   func(m *MockRepository) {},
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
			mockSetup: func(m *MockRepository) {
				m.GetAccountFunc = func(ctx context.Context, id string) (*Account, error) {
					return nil, nil // Not found
				}
			},
			expectedErr: "account acc_1 not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockRepository{}
			tt.mockSetup(mockRepo)
			service := NewLedgerService(mockRepo, nil)

			err := service.RecordTransaction(context.Background(), tt.req, "zone_123", "test")
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
		mockSetup   func(*MockRepository)
		expectedID  string
		expectedErr bool
	}{
		{
			name:    "Success",
			accName: "Checking",
			mockSetup: func(m *MockRepository) {
				m.CreateAccountFunc = func(ctx context.Context, acc *Account) error {
					acc.ID = "acc_123"
					return nil
				}
			},
			expectedID: "acc_123",
		},
		{
			name:    "Database Error",
			accName: "Savings",
			mockSetup: func(m *MockRepository) {
				m.CreateAccountFunc = func(ctx context.Context, acc *Account) error {
					return errors.New("db error")
				}
			},
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockRepository{}
			tt.mockSetup(mockRepo)
			service := NewLedgerService(mockRepo, nil)

			acc, err := service.CreateAccount(context.Background(), tt.accName, Asset, "USD", nil, "zone_123", "test")
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
