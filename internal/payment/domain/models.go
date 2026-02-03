package domain

import (
	"time"
)

// PaymentIntent represents a payment transaction intent.
type PaymentIntent struct {
	ID                   string    `json:"id"`
	ZoneID               string    `json:"zone_id"`
	Mode                 string    `json:"mode"`
	Amount               int64     `json:"amount"` // In cents
	Currency             string    `json:"currency"`
	Status               string    `json:"status"` // requires_payment_method, succeeded, failed
	Description          string    `json:"description,omitempty"`
	UserID               string    `json:"user_id"`
	ApplicationFeeAmount int64     `json:"application_fee_amount,omitempty"`
	OnBehalfOf           string    `json:"on_behalf_of,omitempty"`
	CreatedAt            time.Time `json:"created_at"`
}

// IdempotencyRecord keys response.
type IdempotencyRecord struct {
	UserID       string
	Key          string
	ResponseBody string
	StatusCode   int
}
