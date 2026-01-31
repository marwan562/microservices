package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Repository interface {
	CreateSubscription(ctx context.Context, sub *Subscription) error
	GetSubscription(ctx context.Context, id string) (*Subscription, error)
	UpdateSubscription(ctx context.Context, sub *Subscription) error
	ListDueSubscriptions(ctx context.Context) ([]*Subscription, error)
	GetPlan(ctx context.Context, id string) (*Plan, error)
}

type BillingService struct {
	repo Repository
}

func NewBillingService(repo Repository) *BillingService {
	return &BillingService{repo: repo}
}

func (s *BillingService) CreateSubscription(ctx context.Context, userID, orgID, planID string) (*Subscription, error) {
	plan, err := s.repo.GetPlan(ctx, planID)
	if err != nil {
		return nil, err
	}
	if plan == nil {
		return nil, ErrPlanNotFound
	}

	sub := &Subscription{
		ID:                 uuid.New().String(),
		UserID:             userID,
		OrgID:              orgID,
		PlanID:             planID,
		Status:             SubscriptionStatusActive,
		CurrentPeriodStart: time.Now(),
		CurrentPeriodEnd:   CalculateNextPeriod(time.Now(), plan.Interval),
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	if err := s.repo.CreateSubscription(ctx, sub); err != nil {
		return nil, err
	}

	return sub, nil
}

func (s *BillingService) CancelSubscription(ctx context.Context, id string) (*Subscription, error) {
	sub, err := s.repo.GetSubscription(ctx, id)
	if err != nil {
		return nil, err
	}
	if sub == nil {
		return nil, ErrSubscriptionNotFound
	}

	now := time.Now()
	sub.Status = SubscriptionStatusCanceled
	sub.CanceledAt = &now

	if err := s.repo.UpdateSubscription(ctx, sub); err != nil {
		return nil, err
	}

	return sub, nil
}

func CalculateNextPeriod(start time.Time, interval string) time.Time {
	if interval == "year" {
		return start.AddDate(1, 0, 0)
	}
	return start.AddDate(0, 1, 0) // Default to month
}
