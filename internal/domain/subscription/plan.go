package subscription

import "time"

type SubscriptionPlan struct {
	ID           uint64    `db:"id"            json:"id"`
	Tier         Tier      `db:"tier"          json:"tier"`
	Price        float64   `db:"price"         json:"price"`
	DurationDays int       `db:"duration_days" json:"duration_days"`
	IsActive     bool      `db:"is_active"     json:"is_active"`
	CreatedAt    time.Time `db:"created_at"    json:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"    json:"updated_at"`
}
