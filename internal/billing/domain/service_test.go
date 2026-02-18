package domain

import (
	"context"
	"errors"
	"testing"
)

func TestBillingService_CreateSubscription(t *testing.T) {
	ctx := context.Background()
	planID := "plan-123"
	userID := "user-123"
	orgID := "org-123"

	repo := &MockRepository{
		GetPlanFunc: func(ctx context.Context, id string) (*Plan, error) {
			if id != planID {
				return nil, errors.New("plan not found")
			}
			return &Plan{ID: id, Interval: "month"}, nil
		},
		CreateSubscriptionFunc: func(ctx context.Context, sub *Subscription) error {
			if sub.UserID != userID || sub.PlanID != planID {
				return errors.New("unexpected subscription data")
			}
			return nil
		},
	}

	service := NewBillingService(repo)
	sub, err := service.CreateSubscription(ctx, userID, orgID, planID)

	if err != nil {
		t.Fatalf("CreateSubscription failed: %v", err)
	}
	if sub.ID == "" {
		t.Error("expected non-empty subscription ID")
	}
	if sub.Status != SubscriptionStatusActive {
		t.Errorf("expected status active, got %s", sub.Status)
	}
}

func TestBillingService_CancelSubscription(t *testing.T) {
	ctx := context.Background()
	subID := "sub-123"

	repo := &MockRepository{
		GetSubscriptionFunc: func(ctx context.Context, id string) (*Subscription, error) {
			return &Subscription{ID: id, Status: SubscriptionStatusActive}, nil
		},
		UpdateSubscriptionFunc: func(ctx context.Context, sub *Subscription) error {
			if sub.Status != SubscriptionStatusCanceled {
				return errors.New("expected status to be canceled")
			}
			return nil
		},
	}

	service := NewBillingService(repo)
	sub, err := service.CancelSubscription(ctx, subID)

	if err != nil {
		t.Fatalf("CancelSubscription failed: %v", err)
	}
	if sub.Status != SubscriptionStatusCanceled {
		t.Errorf("expected status canceled, got %s", sub.Status)
	}
	if sub.CanceledAt == nil {
		t.Error("expected CanceledAt to be set")
	}
}
