package tenant

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strings"
)

// Tenant represents a tenant in the multi-tenant system.
type Tenant struct {
	ID             string
	OrganizationID string
	Name           string
	Tier           string
	EncryptionKey  []byte
	Settings       map[string]interface{}
}

// ContextKey is the type for context keys.
type ContextKey string

const (
	// TenantContextKey is the context key for the current tenant.
	TenantContextKey ContextKey = "tenant"

	// TenantIDHeader is the HTTP header for tenant ID.
	TenantIDHeader = "X-Tenant-ID"

	// ZoneIDHeader is the HTTP header for zone ID (alias for tenant).
	ZoneIDHeader = "X-Zone-ID"
)

// FromContext retrieves the tenant from context.
func FromContext(ctx context.Context) (*Tenant, bool) {
	t, ok := ctx.Value(TenantContextKey).(*Tenant)
	return t, ok
}

// MustFromContext retrieves the tenant from context or panics.
func MustFromContext(ctx context.Context) *Tenant {
	t, ok := FromContext(ctx)
	if !ok {
		panic("tenant not found in context")
	}
	return t
}

// WithTenant adds a tenant to the context.
func WithTenant(ctx context.Context, t *Tenant) context.Context {
	return context.WithValue(ctx, TenantContextKey, t)
}

// Resolver resolves tenant information from various sources.
type Resolver interface {
	// Resolve looks up a tenant by ID.
	Resolve(ctx context.Context, tenantID string) (*Tenant, error)
}

// CachingResolver wraps a resolver with caching.
type CachingResolver struct {
	resolver Resolver
	cache    map[string]*Tenant
}

// DatabaseResolver resolves tenants from the database.
type DatabaseResolver struct {
	db *sql.DB
}

// NewDatabaseResolver creates a new database-backed tenant resolver.
func NewDatabaseResolver(db *sql.DB) *DatabaseResolver {
	return &DatabaseResolver{db: db}
}

// Resolve looks up a tenant from the database.
func (r *DatabaseResolver) Resolve(ctx context.Context, tenantID string) (*Tenant, error) {
	query := `
		SELECT id, organization_id, name, tier
		FROM zones
		WHERE id = $1 AND deleted_at IS NULL
	`

	var t Tenant
	err := r.db.QueryRowContext(ctx, query, tenantID).Scan(
		&t.ID, &t.OrganizationID, &t.Name, &t.Tier,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("tenant %q not found", tenantID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to resolve tenant: %w", err)
	}

	return &t, nil
}

// Middleware creates an HTTP middleware that extracts and validates tenant.
func Middleware(resolver Resolver) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract tenant ID from header or API key
			tenantID := r.Header.Get(TenantIDHeader)
			if tenantID == "" {
				tenantID = r.Header.Get(ZoneIDHeader)
			}

			// Try to extract from API key if header not present
			if tenantID == "" {
				if apiKey := r.Header.Get("Authorization"); apiKey != "" {
					tenantID = extractTenantFromAPIKey(apiKey)
				}
			}

			if tenantID == "" {
				http.Error(w, "tenant ID required", http.StatusBadRequest)
				return
			}

			// Resolve tenant
			tenant, err := resolver.Resolve(r.Context(), tenantID)
			if err != nil {
				http.Error(w, "invalid tenant", http.StatusUnauthorized)
				return
			}

			// Add tenant to context
			ctx := WithTenant(r.Context(), tenant)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// extractTenantFromAPIKey extracts tenant ID from an API key.
// API keys format: sk_{mode}_{tenant_id}_{random}
func extractTenantFromAPIKey(apiKey string) string {
	// Remove "Bearer " prefix if present
	apiKey = strings.TrimPrefix(apiKey, "Bearer ")

	parts := strings.Split(apiKey, "_")
	if len(parts) >= 3 {
		return parts[2]
	}
	return ""
}

// IsolationLevel defines the level of tenant isolation.
type IsolationLevel int

const (
	// SharedDatabase - tenants share database with row-level security.
	SharedDatabase IsolationLevel = iota

	// SeparateSchema - each tenant has a separate database schema.
	SeparateSchema

	// SeparateDatabase - each tenant has a separate database.
	SeparateDatabase
)

// Config configures tenant isolation behavior.
type Config struct {
	// IsolationLevel determines how data is isolated.
	IsolationLevel IsolationLevel

	// EnforceRLS enables PostgreSQL row-level security.
	EnforceRLS bool

	// AllowCrossTenantAccess allows certain admin operations across tenants.
	AllowCrossTenantAccess bool
}
