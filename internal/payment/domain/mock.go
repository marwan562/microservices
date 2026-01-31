package domain

import (
	"context"
)

type MockRepository struct {
	CreatePaymentIntentFunc func(ctx context.Context, intent *PaymentIntent) error
	GetPaymentIntentFunc    func(ctx context.Context, id string) (*PaymentIntent, error)
	UpdateStatusFunc        func(ctx context.Context, id, status string) error
	GetIdempotencyKeyFunc   func(ctx context.Context, userID, key string) (*IdempotencyRecord, error)
	SaveIdempotencyKeyFunc  func(ctx context.Context, userID, key string, statusCode int, body string) error
}

func (m *MockRepository) CreatePaymentIntent(ctx context.Context, intent *PaymentIntent) error {
	return m.CreatePaymentIntentFunc(ctx, intent)
}

func (m *MockRepository) GetPaymentIntent(ctx context.Context, id string) (*PaymentIntent, error) {
	return m.GetPaymentIntentFunc(ctx, id)
}

func (m *MockRepository) UpdateStatus(ctx context.Context, id, status string) error {
	return m.UpdateStatusFunc(ctx, id, status)
}

func (m *MockRepository) GetIdempotencyKey(ctx context.Context, userID, key string) (*IdempotencyRecord, error) {
	return m.GetIdempotencyKeyFunc(ctx, userID, key)
}

func (m *MockRepository) SaveIdempotencyKey(ctx context.Context, userID, key string, statusCode int, body string) error {
	return m.SaveIdempotencyKeyFunc(ctx, userID, key, statusCode, body)
}
