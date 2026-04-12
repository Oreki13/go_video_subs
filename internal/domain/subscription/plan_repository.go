package subscription

import "context"

type PlanRepository interface {
	FindActiveByTier(ctx context.Context, tier Tier) (*SubscriptionPlan, error)
}
