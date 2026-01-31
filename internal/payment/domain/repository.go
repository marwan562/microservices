package domain

import (
	"context"
)

type Repository interface {
	CreatePaymentIntent(ctx context.Context, intent *PaymentIntent) error
	GetPaymentIntent(ctx context.Context, id string) (*PaymentIntent, error)
	UpdateStatus(ctx context.Context, id, status string) error
	GetIdempotencyKey(ctx context.Context, userID, key string) (*IdempotencyRecord, error)
	SaveIdempotencyKey(ctx context.Context, userID, key string, statusCode int, body string) error
}
