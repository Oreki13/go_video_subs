package subscription

import (
	"context"
	"fmt"

	"github.com/go_video_subs/internal/domain/subscription"
	"github.com/jmoiron/sqlx"
)

type subscriptionRepository struct {
	db *sqlx.DB
}

func New(db *sqlx.DB) subscription.Repository {
	return &subscriptionRepository{db: db}
}

func (r *subscriptionRepository) FindActiveByUserID(ctx context.Context, userID uint64) (*subscription.Subscription, error) {
	var s subscription.Subscription
	query := `
		SELECT id, user_id, plan_id, tier, status, started_at, expired_at, created_at
		FROM subscriptions
		WHERE user_id = ?
		  AND status = 'active'
		  AND (expired_at IS NULL OR expired_at > NOW())
		LIMIT 1
	`
	if err := r.db.GetContext(ctx, &s, query, userID); err != nil {
		return nil, fmt.Errorf("repository: find active subscription by user_id: %w", err)
	}
	return &s, nil
}

func (r *subscriptionRepository) Upsert(ctx context.Context, s *subscription.Subscription) error {
	query := `
		INSERT INTO subscriptions (user_id, plan_id, tier, status, started_at, expired_at)
		VALUES (?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			plan_id    = VALUES(plan_id),
			tier       = VALUES(tier),
			status     = VALUES(status),
			started_at = VALUES(started_at),
			expired_at = VALUES(expired_at),
			updated_at = CURRENT_TIMESTAMP
	`
	if _, err := r.db.ExecContext(ctx, query,
		s.UserID, s.PlanID, s.Tier, s.Status, s.StartedAt, s.ExpiredAt,
	); err != nil {
		return fmt.Errorf("repository: upsert subscription for user_id %d: %w", s.UserID, err)
	}
	return nil
}
