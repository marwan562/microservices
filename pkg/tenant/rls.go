package tenant

import (
	"context"
	"database/sql"
	"fmt"
)

// RLSPolicy represents a row-level security policy.
type RLSPolicy struct {
	Name       string
	Table      string
	Column     string
	Permission string // SELECT, INSERT, UPDATE, DELETE, ALL
}

// RLSManager manages PostgreSQL Row-Level Security policies.
type RLSManager struct {
	db *sql.DB
}

// NewRLSManager creates a new RLS manager.
func NewRLSManager(db *sql.DB) *RLSManager {
	return &RLSManager{db: db}
}

// EnableRLS enables row-level security on a table.
func (m *RLSManager) EnableRLS(ctx context.Context, table string) error {
	query := fmt.Sprintf("ALTER TABLE %s ENABLE ROW LEVEL SECURITY", table)
	_, err := m.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to enable RLS on %s: %w", table, err)
	}
	return nil
}

// ForceRLS forces RLS for table owners (prevents bypass).
func (m *RLSManager) ForceRLS(ctx context.Context, table string) error {
	query := fmt.Sprintf("ALTER TABLE %s FORCE ROW LEVEL SECURITY", table)
	_, err := m.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to force RLS on %s: %w", table, err)
	}
	return nil
}

// CreateTenantPolicy creates a policy that restricts rows to current tenant.
func (m *RLSManager) CreateTenantPolicy(ctx context.Context, policy RLSPolicy) error {
	// Create policy using current_setting('app.current_tenant')
	// This requires setting the tenant in the session before queries
	query := fmt.Sprintf(`
		CREATE POLICY %s ON %s
		FOR %s
		USING (%s = current_setting('app.current_tenant', true))
		WITH CHECK (%s = current_setting('app.current_tenant', true))
	`, policy.Name, policy.Table, policy.Permission, policy.Column, policy.Column)

	_, err := m.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to create policy %s: %w", policy.Name, err)
	}
	return nil
}

// CreateReadOnlyPolicy creates a policy that only allows SELECT.
func (m *RLSManager) CreateReadOnlyPolicy(ctx context.Context, table, column string) error {
	return m.CreateTenantPolicy(ctx, RLSPolicy{
		Name:       fmt.Sprintf("%s_tenant_select", table),
		Table:      table,
		Column:     column,
		Permission: "SELECT",
	})
}

// CreateFullAccessPolicy creates policies for all operations.
func (m *RLSManager) CreateFullAccessPolicy(ctx context.Context, table, column string) error {
	return m.CreateTenantPolicy(ctx, RLSPolicy{
		Name:       fmt.Sprintf("%s_tenant_all", table),
		Table:      table,
		Column:     column,
		Permission: "ALL",
	})
}

// DropPolicy removes a policy.
func (m *RLSManager) DropPolicy(ctx context.Context, policyName, table string) error {
	query := fmt.Sprintf("DROP POLICY IF EXISTS %s ON %s", policyName, table)
	_, err := m.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to drop policy %s: %w", policyName, err)
	}
	return nil
}

// SetCurrentTenant sets the current tenant in the database session.
// This must be called before any queries when using RLS.
func (m *RLSManager) SetCurrentTenant(ctx context.Context, tenantID string) error {
	query := "SELECT set_config('app.current_tenant', $1, false)"
	_, err := m.db.ExecContext(ctx, query, tenantID)
	if err != nil {
		return fmt.Errorf("failed to set current tenant: %w", err)
	}
	return nil
}

// TxWithTenant executes a function within a transaction with tenant context set.
func (m *RLSManager) TxWithTenant(ctx context.Context, tenantID string, fn func(tx *sql.Tx) error) error {
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Set tenant context for this transaction
	if _, err := tx.ExecContext(ctx, "SELECT set_config('app.current_tenant', $1, true)", tenantID); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to set tenant context: %w", err)
	}

	if err := fn(tx); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

// InitializeRLS sets up RLS for all core Sapliy tables.
func (m *RLSManager) InitializeRLS(ctx context.Context) error {
	tables := []struct {
		name   string
		column string
	}{
		{"events", "zone_id"},
		{"flows", "zone_id"},
		{"flow_executions", "zone_id"},
		{"webhooks", "zone_id"},
		{"webhook_deliveries", "zone_id"},
		{"audit_logs", "zone_id"},
		{"api_keys", "zone_id"},
		{"ledger_entries", "zone_id"},
		{"notifications", "zone_id"},
	}

	for _, t := range tables {
		// Enable RLS
		if err := m.EnableRLS(ctx, t.name); err != nil {
			return err
		}

		// Force RLS (prevent owner bypass)
		if err := m.ForceRLS(ctx, t.name); err != nil {
			return err
		}

		// Create tenant isolation policy
		if err := m.CreateFullAccessPolicy(ctx, t.name, t.column); err != nil {
			return err
		}
	}

	return nil
}

// ValidateTenantAccess checks if the current tenant can access a resource.
func ValidateTenantAccess(ctx context.Context, resourceTenantID string) error {
	tenant, ok := FromContext(ctx)
	if !ok {
		return fmt.Errorf("no tenant in context")
	}

	if tenant.ID != resourceTenantID {
		return fmt.Errorf("access denied: resource belongs to different tenant")
	}

	return nil
}
