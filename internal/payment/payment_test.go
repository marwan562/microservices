package payment

import (
	"context"
	"database/sql"
	"testing"
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

func TestCreatePaymentIntent_Defaults(t *testing.T) {
	mockDB := &MockDB{}
	repo := &Repository{db: mockDB}

	mockDB.QueryRowContextFunc = func(ctx context.Context, query string, args ...any) Row {
		return &MockRow{
			ScanFunc: func(dest ...any) error {
				// ID is at index 0, CreatedAt at index 1
				*(dest[0].(*string)) = "pi_123"
				return nil
			},
		}
	}

	intent := &PaymentIntent{
		Amount:   1000,
		Currency: "", // Should default to USD
		Status:   "requires_payment_method",
		UserID:   "user_123",
	}

	err := repo.CreatePaymentIntent(context.Background(), intent)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if intent.Currency != "USD" {
		t.Errorf("Expected default currency USD, got %s", intent.Currency)
	}
	if intent.ID != "pi_123" {
		t.Errorf("Expected ID pi_123, got %s", intent.ID)
	}
}

func TestUpdateStatus(t *testing.T) {
	mockDB := &MockDB{}
	repo := &Repository{db: mockDB}

	mockDB.ExecContextFunc = func(ctx context.Context, query string, args ...any) (sql.Result, error) {
		return nil, nil // Return nil for success
	}

	err := repo.UpdateStatus(context.Background(), "pi_123", "succeeded")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}
