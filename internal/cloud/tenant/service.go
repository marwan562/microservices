package tenant

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Provisioner interface {
	ProvisionResources(ctx context.Context, tenantID uuid.UUID, config TenantConfig) error
	DeprovisionResources(ctx context.Context, tenantID uuid.UUID) error
}

type Service struct {
	repo        Repository
	provisioner Provisioner
}

func NewService(repo Repository, provisioner Provisioner) *Service {
	return &Service{
		repo:        repo,
		provisioner: provisioner,
	}
}

func (s *Service) OnboardTenant(ctx context.Context, name, slug string, config TenantConfig) (*Tenant, error) {
	// 1. Check if slug exists
	if _, err := s.repo.GetBySlug(slug); err == nil {
		return nil, fmt.Errorf("tenant slug already exists")
	}

	t := &Tenant{
		ID:        uuid.New(),
		Name:      name,
		Slug:      slug,
		Status:    StatusProvisioning,
		Config:    config,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	// 2. Persist initial state
	if err := s.repo.Create(t); err != nil {
		return nil, err
	}

	// 3. Trigger async provisioning (simplified synchronous call here for v1)
	go func() {
		// In a real system, this would be a workflow/temporal activity
		// provisioningCtx := context.Background()
		if err := s.provisioner.ProvisionResources(context.Background(), t.ID, t.Config); err != nil {
			// Handle failure (mark tenant as failed)
			fmt.Printf("Failed to provision tenant %s: %v\n", t.ID, err)
			return
		}

		t.Status = StatusActive
		t.UpdatedAt = time.Now().UTC()
		_ = s.repo.Update(t)
	}()

	return t, nil
}

func (s *Service) GetTenant(ctx context.Context, id uuid.UUID) (*Tenant, error) {
	return s.repo.GetByID(id)
}

func (s *Service) SuspendTenant(ctx context.Context, id uuid.UUID) error {
	t, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}

	if t.Status != StatusActive {
		return errors.New("only active tenants can be suspended")
	}

	t.Status = StatusSuspended
	t.UpdatedAt = time.Now().UTC()
	return s.repo.Update(t)
}
