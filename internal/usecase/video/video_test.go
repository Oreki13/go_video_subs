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
func newTestUseCase(videoRepo *mockVideoRepo) *UseCase {
	return New(videoRepo)
}

// fakeVideos adalah data dummy video untuk dipakai antar test.
var fakeVideos = []domainVideo.Video{
	{ID: 1, Title: "Go Basics (Bronze)"},
	{ID: 2, Title: "Go Intermediate (Silver)"},
	{ID: 3, Title: "Go Advanced (Gold)"},
}

// =============================================================================
// GetVideosByTier – Happy Path berdasarkan tier
// =============================================================================

func TestGetVideosByTier_GoldUser_CanAccessAllTiers(t *testing.T) {
	// User Gold harus bisa mengakses konten gold, silver, DAN bronze.
	videoRepo := &mockVideoRepo{
		FindByTiersFn: func(ctx context.Context, tiers []string) ([]domainVideo.Video, error) {
			// Pastikan ketiga tier dikirim ke repository
			assert.ElementsMatch(t, []string{"gold", "silver", "bronze"}, tiers)
			return fakeVideos, nil
		},
	}

	uc := newTestUseCase(videoRepo)
	videos, err := uc.GetVideosByTier(context.Background(), domainSub.TierGold)

	require.NoError(t, err)
	assert.Len(t, videos, 3)
}

func TestGetVideosByTier_SilverUser_CanAccessSilverAndBronze(t *testing.T) {
	// User Silver hanya bisa akses silver dan bronze, tidak bisa gold.
	videoRepo := &mockVideoRepo{
		FindByTiersFn: func(ctx context.Context, tiers []string) ([]domainVideo.Video, error) {
			assert.ElementsMatch(t, []string{"silver", "bronze"}, tiers)
			assert.NotContains(t, tiers, "gold", "Silver user tidak boleh akses konten Gold")
			return fakeVideos[:2], nil
		},
	}

	uc := newTestUseCase(videoRepo)
	videos, err := uc.GetVideosByTier(context.Background(), domainSub.TierSilver)

	require.NoError(t, err)
	assert.Len(t, videos, 2)
}

func TestGetVideosByTier_BronzeUser_CanOnlyAccessBronze(t *testing.T) {
	// User Bronze hanya bisa akses bronze.
	videoRepo := &mockVideoRepo{
		FindByTiersFn: func(ctx context.Context, tiers []string) ([]domainVideo.Video, error) {
			assert.Equal(t, []string{"bronze"}, tiers)
			return fakeVideos[:1], nil
		},
	}

	uc := newTestUseCase(videoRepo)
	videos, err := uc.GetVideosByTier(context.Background(), domainSub.TierBronze)

	require.NoError(t, err)
	assert.Len(t, videos, 1)
}

func TestGetVideosByTier_ReturnsEmptyList_WhenNoVideosFound(t *testing.T) {
	// Subscription valid tapi belum ada video yang di-upload.
	videoRepo := &mockVideoRepo{
		FindByTiersFn: func(ctx context.Context, tiers []string) ([]domainVideo.Video, error) {
			return []domainVideo.Video{}, nil
		},
	}

	uc := newTestUseCase(videoRepo)
	videos, err := uc.GetVideosByTier(context.Background(), domainSub.TierGold)

	require.NoError(t, err)
	assert.Empty(t, videos, "Harus return slice kosong, bukan error")
}

// =============================================================================
// GetVideosByTier – Error Path
// =============================================================================

func TestGetVideosByTier_UnknownTier_ReturnsError(t *testing.T) {
	// Tier tidak dikenal harus return error (bukan panic).
	uc := newTestUseCase(&mockVideoRepo{})
	videos, err := uc.GetVideosByTier(context.Background(), domainSub.Tier("platinum"))

	assert.Error(t, err)
	assert.Nil(t, videos)
	assert.Contains(t, err.Error(), "unknown subscription tier")
}

func TestGetVideosByTier_VideoRepoError(t *testing.T) {
	// Tier valid tapi query video ke DB gagal.
	videoRepo := &mockVideoRepo{
		FindByTiersFn: func(ctx context.Context, tiers []string) ([]domainVideo.Video, error) {
			return nil, errors.New("db connection timeout")
		},
	}

	uc := newTestUseCase(videoRepo)
	videos, err := uc.GetVideosByTier(context.Background(), domainSub.TierGold)

	assert.Error(t, err)
	assert.Nil(t, videos)
	assert.Contains(t, err.Error(), "get videos by tier")
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
