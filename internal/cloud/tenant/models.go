package tenant

import (
	"time"

	"github.com/google/uuid"
)

type TenantStatus string

const (
	StatusProvisioning TenantStatus = "PROVISIONING"
	StatusActive       TenantStatus = "ACTIVE"
	StatusSuspended    TenantStatus = "SUSPENDED"
	StatusDeleting     TenantStatus = "DELETING"
)

type Tenant struct {
	ID        uuid.UUID    `json:"id"`
	Name      string       `json:"name"`
	Slug      string       `json:"slug"` // URL-friendly identifier
	Status    TenantStatus `json:"status"`
	CreatedAt time.Time    `json:"created_at"`
	UpdatedAt time.Time    `json:"updated_at"`

	Config TenantConfig `json:"config"`
}

type TenantConfig struct {
	Region          string   `json:"region"`
	DedicatedDB     bool     `json:"dedicated_db"`
	RateLimitPerSec int      `json:"rate_limit_per_sec"`
	AllowedIPs      []string `json:"allowed_ips,omitempty"`
}

type Repository interface {
	Create(t *Tenant) error
	GetByID(id uuid.UUID) (*Tenant, error)
	GetBySlug(slug string) (*Tenant, error)
	Update(t *Tenant) error
	Delete(id uuid.UUID) error
}
