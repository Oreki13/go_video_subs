package video

import (
	"context"

	domainSub "github.com/go_video_subs/internal/domain/subscription"
	domainVideo "github.com/go_video_subs/internal/domain/video"
)

// mockSubRepo mengimplementasi domainSub.Repository untuk keperluan testing.
type mockSubRepo struct {
	FindActiveByUserIDFn func(ctx context.Context, userID uint64) (*domainSub.Subscription, error)
	UpsertFn             func(ctx context.Context, s *domainSub.Subscription) error
}

func (m *mockSubRepo) FindActiveByUserID(ctx context.Context, userID uint64) (*domainSub.Subscription, error) {
	return m.FindActiveByUserIDFn(ctx, userID)
}

func (m *mockSubRepo) Upsert(ctx context.Context, s *domainSub.Subscription) error {
	return m.UpsertFn(ctx, s)
}

// mockVideoRepo mengimplementasi domainVideo.Repository untuk keperluan testing.
type mockVideoRepo struct {
	FindByTiersFn func(ctx context.Context, tiers []string) ([]domainVideo.Video, error)
}

func (m *mockVideoRepo) FindByTiers(ctx context.Context, tiers []string) ([]domainVideo.Video, error) {
	return m.FindByTiersFn(ctx, tiers)
}
