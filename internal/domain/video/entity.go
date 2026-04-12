package video

import "time"

type Tier string

const (
	TierGold   Tier = "gold"
	TierSilver Tier = "silver"
	TierBronze Tier = "bronze"
)

type Video struct {
	ID              uint64    `db:"id"               json:"id"`
	CategoryID      uint64    `db:"category_id"      json:"category_id"`
	Title           string    `db:"title"            json:"title"`
	Description     string    `db:"description"      json:"description"`
	URL             string    `db:"url"              json:"url"`
	DurationSeconds uint64    `db:"duration_seconds" json:"duration_seconds"`
	CreatedAt       time.Time `db:"created_at"       json:"created_at"`
	UpdatedAt       time.Time `db:"updated_at"       json:"updated_at"`
}

type VideoCategory struct {
	ID          uint64    `db:"id"          json:"id"`
	Name        Tier      `db:"name"         json:"name"`
	Description string    `db:"description"  json:"description"`
	CreatedAt   time.Time `db:"created_at"   json:"created_at"`
}
