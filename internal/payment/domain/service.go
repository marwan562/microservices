package domain

import (
	"context"
)

type PaymentService struct {
	repo Repository
}

func NewPaymentService(repo Repository) *PaymentService {
	return &PaymentService{repo: repo}
}

func (s *PaymentService) CreatePaymentIntent(ctx context.Context, intent *PaymentIntent) error {
	return s.repo.CreatePaymentIntent(ctx, intent)
}

func (s *PaymentService) GetPaymentIntent(ctx context.Context, id string) (*PaymentIntent, error) {
	return s.repo.GetPaymentIntent(ctx, id)
}

func (s *PaymentService) UpdateStatus(ctx context.Context, id, status string) error {
	return s.repo.UpdateStatus(ctx, id, status)
}

func (s *PaymentService) GetIdempotencyKey(ctx context.Context, userID, key string) (*IdempotencyRecord, error) {
	return s.repo.GetIdempotencyKey(ctx, userID, key)
}

func (s *PaymentService) SaveIdempotencyKey(ctx context.Context, userID, key string, statusCode int, body string) error {
	return s.repo.SaveIdempotencyKey(ctx, userID, key, statusCode, body)
}
