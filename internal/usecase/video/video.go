package video

import (
	"context"
	"errors"
	"fmt"

	"github.com/go_video_subs/internal/domain/subscription"
	"github.com/go_video_subs/internal/domain/video"
)

var ErrNoActiveSubscription = errors.New("no active subscription found")

var tierHierarchy = map[subscription.Tier][]string{
	subscription.TierGold:   {"gold", "silver", "bronze"},
	subscription.TierSilver: {"silver", "bronze"},
	subscription.TierBronze: {"bronze"},
}

type UseCase struct {
	subRepo   subscription.Repository
	videoRepo video.Repository
}

func New(subRepo subscription.Repository, videoRepo video.Repository) *UseCase {
	return &UseCase{subRepo: subRepo, videoRepo: videoRepo}
}

func (uc *UseCase) GetVideosByUserTier(ctx context.Context, userID uint64) ([]video.Video, error) {
	sub, err := uc.subRepo.FindActiveByUserID(ctx, userID)
	if err != nil {
		return nil, ErrNoActiveSubscription
	}

	allowedTiers, ok := tierHierarchy[sub.Tier]
	if !ok {
		return nil, fmt.Errorf("usecase: unknown subscription tier: %s", sub.Tier)
	}

	videos, err := uc.videoRepo.FindByTiers(ctx, allowedTiers)
	if err != nil {
		return nil, fmt.Errorf("usecase: get videos by user tier: %w", err)
	}

	return videos, nil
}
