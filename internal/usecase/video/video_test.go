package video

import (
	"context"
	"errors"
	"testing"

	domainSub "github.com/go_video_subs/internal/domain/subscription"
	domainVideo "github.com/go_video_subs/internal/domain/video"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestUseCase adalah helper untuk membuat UseCase dengan mock yang diberikan.
func newTestUseCase(subRepo *mockSubRepo, videoRepo *mockVideoRepo) *UseCase {
	return New(subRepo, videoRepo)
}

// fakeVideos adalah data dummy video untuk dipakai antar test.
var fakeVideos = []domainVideo.Video{
	{ID: 1, Title: "Go Basics (Bronze)"},
	{ID: 2, Title: "Go Intermediate (Silver)"},
	{ID: 3, Title: "Go Advanced (Gold)"},
}

// =============================================================================
// GetVideosByUserTier – Happy Path berdasarkan tier
// =============================================================================

func TestGetVideosByUserTier_GoldUser_CanAccessAllTiers(t *testing.T) {
	// User Gold harus bisa mengakses konten gold, silver, DAN bronze.
	subRepo := &mockSubRepo{
		FindActiveByUserIDFn: func(ctx context.Context, userID uint64) (*domainSub.Subscription, error) {
			assert.Equal(t, uint64(1), userID)
			return &domainSub.Subscription{UserID: 1, Tier: domainSub.TierGold}, nil
		},
	}
	videoRepo := &mockVideoRepo{
		FindByTiersFn: func(ctx context.Context, tiers []string) ([]domainVideo.Video, error) {
			// Pastikan ketiga tier dikirim ke repository
			assert.ElementsMatch(t, []string{"gold", "silver", "bronze"}, tiers)
			return fakeVideos, nil
		},
	}

	uc := newTestUseCase(subRepo, videoRepo)
	videos, err := uc.GetVideosByUserTier(context.Background(), 1)

	require.NoError(t, err)
	assert.Len(t, videos, 3)
}

func TestGetVideosByUserTier_SilverUser_CanAccessSilverAndBronze(t *testing.T) {
	// User Silver hanya bisa akses silver dan bronze, tidak bisa gold.
	subRepo := &mockSubRepo{
		FindActiveByUserIDFn: func(ctx context.Context, userID uint64) (*domainSub.Subscription, error) {
			return &domainSub.Subscription{UserID: 1, Tier: domainSub.TierSilver}, nil
		},
	}
	videoRepo := &mockVideoRepo{
		FindByTiersFn: func(ctx context.Context, tiers []string) ([]domainVideo.Video, error) {
			assert.ElementsMatch(t, []string{"silver", "bronze"}, tiers)
			assert.NotContains(t, tiers, "gold", "Silver user tidak boleh akses konten Gold")
			return fakeVideos[:2], nil
		},
	}

	uc := newTestUseCase(subRepo, videoRepo)
	videos, err := uc.GetVideosByUserTier(context.Background(), 1)

	require.NoError(t, err)
	assert.Len(t, videos, 2)
}

func TestGetVideosByUserTier_BronzeUser_CanOnlyAccessBronze(t *testing.T) {
	// User Bronze hanya bisa akses bronze.
	subRepo := &mockSubRepo{
		FindActiveByUserIDFn: func(ctx context.Context, userID uint64) (*domainSub.Subscription, error) {
			return &domainSub.Subscription{UserID: 1, Tier: domainSub.TierBronze}, nil
		},
	}
	videoRepo := &mockVideoRepo{
		FindByTiersFn: func(ctx context.Context, tiers []string) ([]domainVideo.Video, error) {
			assert.Equal(t, []string{"bronze"}, tiers)
			return fakeVideos[:1], nil
		},
	}

	uc := newTestUseCase(subRepo, videoRepo)
	videos, err := uc.GetVideosByUserTier(context.Background(), 1)

	require.NoError(t, err)
	assert.Len(t, videos, 1)
}

func TestGetVideosByUserTier_ReturnsEmptyList_WhenNoVideosFound(t *testing.T) {
	// Subscription valid tapi belum ada video yang di-upload.
	subRepo := &mockSubRepo{
		FindActiveByUserIDFn: func(ctx context.Context, userID uint64) (*domainSub.Subscription, error) {
			return &domainSub.Subscription{UserID: 1, Tier: domainSub.TierGold}, nil
		},
	}
	videoRepo := &mockVideoRepo{
		FindByTiersFn: func(ctx context.Context, tiers []string) ([]domainVideo.Video, error) {
			return []domainVideo.Video{}, nil
		},
	}

	uc := newTestUseCase(subRepo, videoRepo)
	videos, err := uc.GetVideosByUserTier(context.Background(), 1)

	require.NoError(t, err)
	assert.Empty(t, videos, "Harus return slice kosong, bukan error")
}

// =============================================================================
// GetVideosByUserTier – Error Path
// =============================================================================

func TestGetVideosByUserTier_NoActiveSubscription(t *testing.T) {
	// User belum subscribe atau subscription sudah expired.
	subRepo := &mockSubRepo{
		FindActiveByUserIDFn: func(ctx context.Context, userID uint64) (*domainSub.Subscription, error) {
			return nil, errors.New("record not found")
		},
	}

	uc := newTestUseCase(subRepo, &mockVideoRepo{})
	videos, err := uc.GetVideosByUserTier(context.Background(), 99)

	// Usecase harus mengembalikan ErrNoActiveSubscription (bukan expose error internal DB)
	assert.ErrorIs(t, err, ErrNoActiveSubscription)
	assert.Nil(t, videos)
}

func TestGetVideosByUserTier_VideoRepoError(t *testing.T) {
	// Subscription valid tapi query video ke DB gagal.
	subRepo := &mockSubRepo{
		FindActiveByUserIDFn: func(ctx context.Context, userID uint64) (*domainSub.Subscription, error) {
			return &domainSub.Subscription{UserID: 1, Tier: domainSub.TierGold}, nil
		},
	}
	videoRepo := &mockVideoRepo{
		FindByTiersFn: func(ctx context.Context, tiers []string) ([]domainVideo.Video, error) {
			return nil, errors.New("db connection timeout")
		},
	}

	uc := newTestUseCase(subRepo, videoRepo)
	videos, err := uc.GetVideosByUserTier(context.Background(), 1)

	assert.Error(t, err)
	assert.Nil(t, videos)
	assert.Contains(t, err.Error(), "get videos by user tier")
}

// =============================================================================
// Tier Hierarchy – validasi logika akses
// =============================================================================

func TestTierHierarchy_AllTiersAreDefined(t *testing.T) {
	// Memastikan semua tier subscription memiliki mapping di tierHierarchy.
	tiers := []domainSub.Tier{
		domainSub.TierGold,
		domainSub.TierSilver,
		domainSub.TierBronze,
	}

	for _, tier := range tiers {
		t.Run(string(tier), func(t *testing.T) {
			_, ok := tierHierarchy[tier]
			assert.True(t, ok, "Tier %q harus ada di tierHierarchy", tier)
		})
	}
}

func TestTierHierarchy_GoldHasHighestAccess(t *testing.T) {
	// Gold harus memiliki jumlah tier akses terbanyak.
	goldTiers := tierHierarchy[domainSub.TierGold]
	silverTiers := tierHierarchy[domainSub.TierSilver]
	bronzeTiers := tierHierarchy[domainSub.TierBronze]

	assert.Greater(t, len(goldTiers), len(silverTiers), "Gold harus punya lebih banyak akses dari Silver")
	assert.Greater(t, len(silverTiers), len(bronzeTiers), "Silver harus punya lebih banyak akses dari Bronze")
}

func TestTierHierarchy_LowerTierIsSubsetOfHigherTier(t *testing.T) {
	// Tier di bawah harus selalu jadi subset dari tier di atasnya.
	goldSet := toSet(tierHierarchy[domainSub.TierGold])
	silverSet := toSet(tierHierarchy[domainSub.TierSilver])
	bronzeSet := toSet(tierHierarchy[domainSub.TierBronze])

	// Semua silver tier harus ada di gold
	for tier := range silverSet {
		assert.Contains(t, goldSet, tier, "Silver tier %q harus inklusif dalam Gold", tier)
	}
	// Semua bronze tier harus ada di silver
	for tier := range bronzeSet {
		assert.Contains(t, silverSet, tier, "Bronze tier %q harus inklusif dalam Silver", tier)
	}
}

// toSet mengkonversi slice ke map untuk lookup O(1).
func toSet(s []string) map[string]struct{} {
	m := make(map[string]struct{}, len(s))
	for _, v := range s {
		m[v] = struct{}{}
	}
	return m
}
