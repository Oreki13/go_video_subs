package subscription

import "time"

type Tier string
type Status string

const (
	TierGold   Tier = "gold"
	TierSilver Tier = "silver"
	TierBronze Tier = "bronze"

	StatusActive   Status = "active"
	StatusInactive Status = "inactive"
	StatusExpired  Status = "expired"
)

type Subscription struct {
	ID        uint64     `db:"id"         json:"id"`
	UserID    uint64     `db:"user_id"    json:"user_id"`
	PlanID    uint64     `db:"plan_id"    json:"plan_id"`
	Tier      Tier       `db:"tier"       json:"tier"`
	Status    Status     `db:"status"     json:"status"`
	StartedAt *time.Time `db:"started_at" json:"started_at"`
	ExpiredAt *time.Time `db:"expired_at" json:"expired_at"`
	CreatedAt time.Time  `db:"created_at" json:"created_at"`
}
