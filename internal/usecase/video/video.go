package video

import (
	"context"
	"fmt"

	"github.com/go_video_subs/internal/domain/subscription"
	"github.com/go_video_subs/internal/domain/video"
)

var tierHierarchy = map[subscription.Tier][]string{
	subscription.TierGold:   {"gold", "silver", "bronze"},
	subscription.TierSilver: {"silver", "bronze"},
	subscription.TierBronze: {"bronze"},
}

type UseCase struct {
	videoRepo video.Repository
}

func New(videoRepo video.Repository) *UseCase {
	return &UseCase{videoRepo: videoRepo}
}

func (uc *UseCase) GetVideosByTier(ctx context.Context, tier subscription.Tier) ([]video.Video, error) {
	allowedTiers, ok := tierHierarchy[tier]
	if !ok {
		return nil, fmt.Errorf("usecase: unknown subscription tier: %s", tier)
	}

	videos, err := uc.videoRepo.FindByTiers(ctx, allowedTiers)
	if err != nil {
		return nil, fmt.Errorf("usecase: get videos by tier: %w", err)
	}

	return videos, nil
}
