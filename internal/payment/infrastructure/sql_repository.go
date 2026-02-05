package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/sapliy/fintech-ecosystem/internal/payment/domain"
)

type SQLRepository struct {
	db *sql.DB
}

func NewSQLRepository(db *sql.DB) *SQLRepository {
	return &SQLRepository{db: db}
}

func (r *SQLRepository) CreatePaymentIntent(ctx context.Context, intent *domain.PaymentIntent) error {
	if intent.Currency == "" {
		intent.Currency = "USD"
	}
	err := r.db.QueryRowContext(ctx,
		`INSERT INTO payment_intents (amount, currency, status, description, user_id, application_fee_amount, on_behalf_of, zone_id, mode) 
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING id, created_at`,
		intent.Amount, intent.Currency, intent.Status, intent.Description, intent.UserID, intent.ApplicationFeeAmount, intent.OnBehalfOf, intent.ZoneID, intent.Mode).
		Scan(&intent.ID, &intent.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create payment intent: %w", err)
	}
	return nil
}

func (r *SQLRepository) GetPaymentIntent(ctx context.Context, id string) (*domain.PaymentIntent, error) {
	var intent domain.PaymentIntent
	err := r.db.QueryRowContext(ctx,
		"SELECT id, amount, currency, status, description, user_id, application_fee_amount, on_behalf_of, zone_id, mode, created_at FROM payment_intents WHERE id = $1",
		id).Scan(&intent.ID, &intent.Amount, &intent.Currency, &intent.Status, &intent.Description, &intent.UserID, &intent.ApplicationFeeAmount, &intent.OnBehalfOf, &intent.ZoneID, &intent.Mode, &intent.CreatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // Not found
		}
		return nil, fmt.Errorf("failed to get payment intent: %w", err)
	}
	return &intent, nil
}

func (r *SQLRepository) UpdateStatus(ctx context.Context, id, status string) error {
	_, err := r.db.ExecContext(ctx, "UPDATE payment_intents SET status = $1 WHERE id = $2", status, id)
	if err != nil {
		return fmt.Errorf("failed to update payment status: %w", err)
	}
	return nil
}

func (r *SQLRepository) GetIdempotencyKey(ctx context.Context, userID, key string) (*domain.IdempotencyRecord, error) {
	var rec domain.IdempotencyRecord
	err := r.db.QueryRowContext(ctx, "SELECT user_id, key, response_body, status_code FROM idempotency_keys WHERE user_id = $1 AND key = $2", userID, key).
		Scan(&rec.UserID, &rec.Key, &rec.ResponseBody, &rec.StatusCode)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &rec, nil
}

func (r *SQLRepository) SaveIdempotencyKey(ctx context.Context, userID, key string, statusCode int, body string) error {
	_, err := r.db.ExecContext(ctx, "INSERT INTO idempotency_keys (user_id, key, response_body, status_code) VALUES ($1, $2, $3, $4)", userID, key, body, statusCode)
	return err
}
