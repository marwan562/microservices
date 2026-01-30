package payment

import (
	"context"
	"database/sql"
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
