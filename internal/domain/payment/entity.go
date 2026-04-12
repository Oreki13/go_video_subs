package payment

import "time"

type Tier string
type Status string

const (
	TierGold   Tier = "gold"
	TierSilver Tier = "silver"
	TierBronze Tier = "bronze"

	StatusPending Status = "pending"
	StatusSuccess Status = "success"
	StatusFailed  Status = "failed"
)

type PaymentTransaction struct {
	ID                uint64     `db:"id"                  json:"id"`
	UserID            uint64     `db:"user_id"             json:"user_id"`
	PlanID            uint64     `db:"plan_id"             json:"plan_id"`
	SubscriptionID    *uint64    `db:"subscription_id"     json:"subscription_id"`
	ExternalPaymentID string     `db:"external_payment_id" json:"external_payment_id"`
	Tier              Tier       `db:"tier"                json:"tier"`
	Amount            float64    `db:"amount"              json:"amount"`
	Status            Status     `db:"status"              json:"status"`
	PayloadRaw        *string    `db:"payload_raw"         json:"payload_raw"`
	PaidAt            *time.Time `db:"paid_at"             json:"paid_at"`
	CreatedAt         time.Time  `db:"created_at"          json:"created_at"`
	UpdatedAt         time.Time  `db:"updated_at"          json:"updated_at"`
}
