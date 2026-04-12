package subscription

import "context"

type Repository interface {
	FindActiveByUserID(ctx context.Context, userID uint64) (*Subscription, error)
	Upsert(ctx context.Context, s *Subscription) error
}
