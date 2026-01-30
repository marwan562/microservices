package ledger

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

// MockTx implements the Tx interface for testing.
type MockTx struct {
	QueryRowContextFunc func(ctx context.Context, query string, args ...any) Row
	ExecContextFunc     func(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContextFunc    func(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	BeginTxFunc         func(ctx context.Context, opts *sql.TxOptions) (Tx, error)
	CommitFunc          func() error
	RollbackFunc        func() error
}

func (m *MockTx) QueryRowContext(ctx context.Context, query string, args ...any) Row {
	return m.QueryRowContextFunc(ctx, query, args...)
}
func (m *MockTx) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return m.ExecContextFunc(ctx, query, args...)
}
func (m *MockTx) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return m.QueryContextFunc(ctx, query, args...)
}
func (m *MockTx) BeginTx(ctx context.Context, opts *sql.TxOptions) (Tx, error) {
	return m.BeginTxFunc(ctx, opts)
}
func (m *MockTx) Commit() error   { return m.CommitFunc() }
func (m *MockTx) Rollback() error { return m.RollbackFunc() }

// MockDB implements the DB interface for testing.
type MockDB struct {
	QueryRowContextFunc func(ctx context.Context, query string, args ...any) Row
	ExecContextFunc     func(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContextFunc    func(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	BeginTxFunc         func(ctx context.Context, opts *sql.TxOptions) (Tx, error)
}

func (m *MockDB) QueryRowContext(ctx context.Context, query string, args ...any) Row {
	return m.QueryRowContextFunc(ctx, query, args...)
}
func (m *MockDB) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return m.ExecContextFunc(ctx, query, args...)
}
func (m *MockDB) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return m.QueryContextFunc(ctx, query, args...)
}
func (m *MockDB) BeginTx(ctx context.Context, opts *sql.TxOptions) (Tx, error) {
	return m.BeginTxFunc(ctx, opts)
}

func TestRecordTransaction_Validation(t *testing.T) {
	repo := &Repository{db: &MockDB{}}

	t.Run("UnbalancedTransaction", func(t *testing.T) {
		req := TransactionRequest{
			Entries: []EntryRequest{
				{Amount: 100},
				{Amount: -50},
			},
		}
		err := repo.RecordTransaction(context.Background(), req)
		if err == nil || err.Error() != "transaction is not balanced (sum != 0)" {
			t.Errorf("Expected unbalanced error, got %v", err)
		}
	})
}

func TestCreateAccount(t *testing.T) {
	mockDB := &MockDB{}
	repo := &Repository{db: mockDB}

	mockDB.QueryRowContextFunc = func(ctx context.Context, query string, args ...any) Row {
		return &MockRow{
			ScanFunc: func(dest ...any) error {
				// ID is at index 0, CreatedAt at index 1
				*(dest[0].(*string)) = "acc_123"
				return nil
			},
		}
	}

	acc, err := repo.CreateAccount(context.Background(), "Test", Asset, "USD", nil)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if acc.ID != "acc_123" {
		t.Errorf("Expected ID acc_123, got %s", acc.ID)
	}
}
