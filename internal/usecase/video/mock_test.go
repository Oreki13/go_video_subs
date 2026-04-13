package video

import (
	"context"

	domainVideo "github.com/go_video_subs/internal/domain/video"
)

type mockVideoRepo struct {
	FindByTiersFn func(ctx context.Context, tiers []string) ([]domainVideo.Video, error)
}

func (m *mockVideoRepo) FindByTiers(ctx context.Context, tiers []string) ([]domainVideo.Video, error) {
	return m.FindByTiersFn(ctx, tiers)
}
