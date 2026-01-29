package connect

import (
	"time"
)

type Account struct {
	ID                 string         `json:"id"`
	UserID             string         `json:"user_id"`
	Type               string         `json:"type"`
	Country            string         `json:"country"`
	Email              string         `json:"email"`
	BusinessType       string         `json:"business_type"`
	Status             string         `json:"status"`
	PlatformFeePercent float64        `json:"platform_fee_percent"`
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at"`
	PayoutSettings     PayoutSettings `json:"payout_settings"`
}

type PayoutSettings struct {
	AccountID     string `json:"account_id"`
	Interval      string `json:"interval"`
	BankAccountID string `json:"bank_account_id"`
}

type Capabilities struct {
	Transfers    bool `json:"transfers"`
	CardPayments bool `json:"card_payments"`
}
