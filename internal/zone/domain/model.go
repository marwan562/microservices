package domain

import (
	"context"
	"errors"
	"time"
)

type Mode string

const (
	ModeTest Mode = "test"
	ModeLive Mode = "live"
)

type Zone struct {
	ID        string            `json:"id"`
	OrgID     string            `json:"org_id"`
	Name      string            `json:"name"`
	Mode      Mode              `json:"mode"`
	Metadata  map[string]string `json:"metadata"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
}

type CreateZoneParams struct {
	OrgID        string
	Name         string
	Mode         Mode
	TemplateName string
	Metadata     map[string]string
}

type Repository interface {
	Create(ctx context.Context, zone *Zone) error
	GetByID(ctx context.Context, id string) (*Zone, error)
	ListByOrgID(ctx context.Context, orgID string) ([]*Zone, error)
	UpdateMetadata(ctx context.Context, id string, metadata map[string]string) error
	Delete(ctx context.Context, id string) error
}

type Service interface {
	CreateZone(ctx context.Context, params CreateZoneParams) (*Zone, error)
	GetZone(ctx context.Context, id string) (*Zone, error)
	ListZones(ctx context.Context, orgID string) ([]*Zone, error)
	DeleteZone(ctx context.Context, id string) error
}

// Errors
var (
	ErrZoneNotFound = errors.New("zone not found")
	ErrInvalidZone  = errors.New("invalid zone")
)
