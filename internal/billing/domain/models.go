package domain

import (
	"time"
)

type SubscriptionStatus string

const (
	SubscriptionStatusActive     SubscriptionStatus = "active"
	SubscriptionStatusCanceled   SubscriptionStatus = "canceled"
	SubscriptionStatusPastDue    SubscriptionStatus = "past_due"
	SubscriptionStatusIncomplete SubscriptionStatus = "incomplete"
)

type Plan struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Amount    int64     `json:"amount"`
	Currency  string    `json:"currency"`
	Interval  string    `json:"interval"` // "month", "year"
	CreatedAt time.Time `json:"created_at"`
}

type Subscription struct {
	ID                 string             `json:"id"`
	UserID             string             `json:"user_id"`
	OrgID              string             `json:"org_id"`
	PlanID             string             `json:"plan_id"`
	Status             SubscriptionStatus `json:"status"`
	CurrentPeriodStart time.Time          `json:"current_period_start"`
	CurrentPeriodEnd   time.Time          `json:"current_period_end"`
	CanceledAt         *time.Time         `json:"canceled_at,omitempty"`
	CreatedAt          time.Time          `json:"created_at"`
	UpdatedAt          time.Time          `json:"updated_at"`
}

type Invoice struct {
	ID              string    `json:"id"`
	SubscriptionID  string    `json:"subscription_id"`
	UserID          string    `json:"user_id"`
	OrgID           string    `json:"org_id"`
	Amount          int64     `json:"amount"`
	Currency        string    `json:"currency"`
	Status          string    `json:"status"` // "paid", "unpaid", "void"
	PaymentIntentID string    `json:"payment_intent_id,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
}
