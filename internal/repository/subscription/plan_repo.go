package subscription

import (
	"context"
	"fmt"

	"github.com/go_video_subs/internal/domain/subscription"
	"github.com/jmoiron/sqlx"
)

type planRepository struct {
	db *sqlx.DB
}

func NewPlan(db *sqlx.DB) subscription.PlanRepository {
	return &planRepository{db: db}
}

func (r *planRepository) FindActiveByTier(ctx context.Context, tier subscription.Tier) (*subscription.SubscriptionPlan, error) {
	var plan subscription.SubscriptionPlan
	query := `
		SELECT id, tier, price, duration_days, is_active, created_at, updated_at
		FROM subscription_plans
		WHERE tier = ? AND is_active = 1
		LIMIT 1
	`
	if err := r.db.GetContext(ctx, &plan, query, tier); err != nil {
		return nil, fmt.Errorf("repository: find active plan by tier %s: %w", tier, err)
	}
	return &plan, nil
}
