package domain

import (
	"context"
)

type MockRepository struct {
	CreateSubscriptionFunc   func(ctx context.Context, sub *Subscription) error
	GetSubscriptionFunc      func(ctx context.Context, id string) (*Subscription, error)
	UpdateSubscriptionFunc   func(ctx context.Context, sub *Subscription) error
	ListDueSubscriptionsFunc func(ctx context.Context) ([]*Subscription, error)
	GetPlanFunc              func(ctx context.Context, id string) (*Plan, error)
}

func (m *MockRepository) CreateSubscription(ctx context.Context, sub *Subscription) error {
	return m.CreateSubscriptionFunc(ctx, sub)
}

func (m *MockRepository) GetSubscription(ctx context.Context, id string) (*Subscription, error) {
	return m.GetSubscriptionFunc(ctx, id)
}

func (m *MockRepository) UpdateSubscription(ctx context.Context, sub *Subscription) error {
	return m.UpdateSubscriptionFunc(ctx, sub)
}

func (m *MockRepository) ListDueSubscriptions(ctx context.Context) ([]*Subscription, error) {
	return m.ListDueSubscriptionsFunc(ctx)
}

func (m *MockRepository) GetPlan(ctx context.Context, id string) (*Plan, error) {
	return m.GetPlanFunc(ctx, id)
}
