package payment

import (
	"context"
	"errors"
	"testing"
	"time"

	domainPayment "github.com/go_video_subs/internal/domain/payment"
	domainSub "github.com/go_video_subs/internal/domain/subscription"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestUseCase adalah helper untuk membuat UseCase dengan mock yang diberikan.
func newTestUseCase(
	payRepo *mockPaymentRepo,
	planRepo *mockPlanRepo,
	subRepo *mockSubRepo,
) *UseCase {
	return New(payRepo, planRepo, subRepo)
}

// =============================================================================
// InitiatePayment Tests
// =============================================================================

func TestInitiatePayment_Success(t *testing.T) {
	// Arrange
	fakePlan := &domainSub.SubscriptionPlan{
		ID:           10,
		Tier:         domainSub.TierGold,
		Price:        150_000,
		DurationDays: 30,
	}

	planRepo := &mockPlanRepo{
		FindActiveByTierFn: func(ctx context.Context, tier domainSub.Tier) (*domainSub.SubscriptionPlan, error) {
			assert.Equal(t, domainSub.TierGold, tier)
			return fakePlan, nil
		},
	}

	payRepo := &mockPaymentRepo{
		CreateFn: func(ctx context.Context, tx *domainPayment.PaymentTransaction) error {
			// Pastikan data yang dikirim ke repository benar
			assert.Equal(t, uint64(1), tx.UserID)
			assert.Equal(t, fakePlan.ID, tx.PlanID)
			assert.Equal(t, fakePlan.Price, tx.Amount)
			assert.Equal(t, domainPayment.StatusPending, tx.Status)
			assert.NotEmpty(t, tx.ExternalPaymentID)
			return nil
		},
	}

	subRepo := &mockSubRepo{}

	uc := newTestUseCase(payRepo, planRepo, subRepo)

	// Act
	out, err := uc.InitiatePayment(context.Background(), InitiatePaymentInput{
		UserID: 1,
		Tier:   "gold",
	})

	// Assert
	require.NoError(t, err)
	require.NotNil(t, out)
	assert.Equal(t, "gold", out.Tier)
	assert.Equal(t, fakePlan.ID, out.PlanID)
	assert.Equal(t, fakePlan.Price, out.Amount)
	assert.NotEmpty(t, out.ExternalPaymentID)
	assert.Equal(t, "/api/v1/payments/callback", out.MockCallbackURL)
}

func TestInitiatePayment_InvalidTier(t *testing.T) {
	// Arrange – tidak butuh dependency apapun karena gagal sebelum memanggil repo
	uc := newTestUseCase(
		&mockPaymentRepo{},
		&mockPlanRepo{},
		&mockSubRepo{},
	)

	testCases := []struct {
		name string
		tier string
	}{
		{"empty tier", ""},
		{"unknown tier", "platinum"},
		{"typo tier", "GOLD"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			out, err := uc.InitiatePayment(context.Background(), InitiatePaymentInput{
				UserID: 1,
				Tier:   tc.tier,
			})

			assert.Error(t, err)
			assert.Nil(t, out)
			assert.Contains(t, err.Error(), "invalid tier")
		})
	}
}

func TestInitiatePayment_AllValidTiers(t *testing.T) {
	// Memastikan ketiga tier yang valid (gold, silver, bronze) dapat diproses
	validTiers := []string{"gold", "silver", "bronze"}

	for _, tierStr := range validTiers {
		tierStr := tierStr // capture untuk closure
		t.Run(tierStr, func(t *testing.T) {
			fakePlan := &domainSub.SubscriptionPlan{ID: 1, Price: 50_000, DurationDays: 30}

			planRepo := &mockPlanRepo{
				FindActiveByTierFn: func(ctx context.Context, tier domainSub.Tier) (*domainSub.SubscriptionPlan, error) {
					return fakePlan, nil
				},
			}
			payRepo := &mockPaymentRepo{
				CreateFn: func(ctx context.Context, tx *domainPayment.PaymentTransaction) error {
					return nil
				},
			}

			uc := newTestUseCase(payRepo, planRepo, &mockSubRepo{})
			out, err := uc.InitiatePayment(context.Background(), InitiatePaymentInput{UserID: 1, Tier: tierStr})

			require.NoError(t, err)
			assert.Equal(t, tierStr, out.Tier)
		})
	}
}

func TestInitiatePayment_PlanRepoError(t *testing.T) {
	// Arrange – planRepo gagal (misal plan tidak aktif)
	planRepo := &mockPlanRepo{
		FindActiveByTierFn: func(ctx context.Context, tier domainSub.Tier) (*domainSub.SubscriptionPlan, error) {
			return nil, errors.New("plan not found")
		},
	}

	uc := newTestUseCase(&mockPaymentRepo{}, planRepo, &mockSubRepo{})

	// Act
	out, err := uc.InitiatePayment(context.Background(), InitiatePaymentInput{UserID: 1, Tier: "gold"})

	// Assert
	assert.Error(t, err)
	assert.Nil(t, out)
	assert.Contains(t, err.Error(), "find active plan")
}

func TestInitiatePayment_PaymentRepoCreateError(t *testing.T) {
	// Arrange – create ke DB gagal
	planRepo := &mockPlanRepo{
		FindActiveByTierFn: func(ctx context.Context, tier domainSub.Tier) (*domainSub.SubscriptionPlan, error) {
			return &domainSub.SubscriptionPlan{ID: 1, Price: 50_000}, nil
		},
	}
	payRepo := &mockPaymentRepo{
		CreateFn: func(ctx context.Context, tx *domainPayment.PaymentTransaction) error {
			return errors.New("db connection error")
		},
	}

	uc := newTestUseCase(payRepo, planRepo, &mockSubRepo{})

	out, err := uc.InitiatePayment(context.Background(), InitiatePaymentInput{UserID: 1, Tier: "bronze"})

	assert.Error(t, err)
	assert.Nil(t, out)
	assert.Contains(t, err.Error(), "create payment transaction")
}

// =============================================================================
// HandleCallback Tests
// =============================================================================

func TestHandleCallback_SuccessStatus(t *testing.T) {
	// Arrange – transaksi pending, callback success -> buat subscription
	fakeTx := &domainPayment.PaymentTransaction{
		ID:     42,
		UserID: 1,
		PlanID: 10,
		Tier:   domainPayment.TierGold,
		Amount: 150_000,
		Status: domainPayment.StatusPending,
	}
	fakePlan := &domainSub.SubscriptionPlan{
		ID:           10,
		Tier:         domainSub.TierGold,
		DurationDays: 30,
	}

	payRepo := &mockPaymentRepo{
		FindByIDFn: func(ctx context.Context, id uint64) (*domainPayment.PaymentTransaction, error) {
			assert.Equal(t, uint64(42), id)
			return fakeTx, nil
		},
		UpdateStatusFn: func(ctx context.Context, id uint64, status domainPayment.Status, paidAt *time.Time) error {
			assert.Equal(t, uint64(42), id)
			assert.Equal(t, domainPayment.StatusSuccess, status)
			assert.NotNil(t, paidAt)
			return nil
		},
	}
	planRepo := &mockPlanRepo{
		FindActiveByTierFn: func(ctx context.Context, tier domainSub.Tier) (*domainSub.SubscriptionPlan, error) {
			return fakePlan, nil
		},
	}

	var capturedSub *domainSub.Subscription
	subRepo := &mockSubRepo{
		UpsertFn: func(ctx context.Context, s *domainSub.Subscription) error {
			capturedSub = s
			return nil
		},
	}

	uc := newTestUseCase(payRepo, planRepo, subRepo)

	// Act
	err := uc.HandleCallback(context.Background(), HandleCallbackInput{
		TransactionID: 42,
		Status:        "success",
	})

	// Assert
	require.NoError(t, err)
	require.NotNil(t, capturedSub)
	assert.Equal(t, fakeTx.UserID, capturedSub.UserID)
	assert.Equal(t, fakePlan.ID, capturedSub.PlanID)
	assert.Equal(t, domainSub.StatusActive, capturedSub.Status)
	assert.NotNil(t, capturedSub.StartedAt)
	assert.NotNil(t, capturedSub.ExpiredAt)
}

func TestHandleCallback_FailedStatus(t *testing.T) {
	// Arrange – callback failed; tidak boleh membuat subscription
	fakeTx := &domainPayment.PaymentTransaction{
		ID:     42,
		UserID: 1,
		PlanID: 10,
		Tier:   domainPayment.TierGold,
		Status: domainPayment.StatusPending,
	}

	upsertCalled := false
	payRepo := &mockPaymentRepo{
		FindByIDFn: func(ctx context.Context, id uint64) (*domainPayment.PaymentTransaction, error) {
			return fakeTx, nil
		},
		UpdateStatusFn: func(ctx context.Context, id uint64, status domainPayment.Status, paidAt *time.Time) error {
			assert.Equal(t, domainPayment.StatusFailed, status)
			assert.Nil(t, paidAt)
			return nil
		},
	}
	subRepo := &mockSubRepo{
		UpsertFn: func(ctx context.Context, s *domainSub.Subscription) error {
			upsertCalled = true
			return nil
		},
	}

	uc := newTestUseCase(payRepo, &mockPlanRepo{}, subRepo)

	// Act
	err := uc.HandleCallback(context.Background(), HandleCallbackInput{
		TransactionID: 42,
		Status:        "failed",
	})

	// Assert
	require.NoError(t, err)
	assert.False(t, upsertCalled, "Upsert subscription tidak boleh dipanggil saat status failed")
}

func TestHandleCallback_InvalidStatus(t *testing.T) {
	// Arrange – status tidak valid, harus ditolak sebelum memanggil repo
	uc := newTestUseCase(
		&mockPaymentRepo{},
		&mockPlanRepo{},
		&mockSubRepo{},
	)

	testCases := []struct {
		name   string
		status string
	}{
		{"empty status", ""},
		{"unknown status", "cancelled"},
		{"pending status", "pending"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := uc.HandleCallback(context.Background(), HandleCallbackInput{
				TransactionID: 1,
				Status:        tc.status,
			})

			assert.Error(t, err)
			assert.Contains(t, err.Error(), "invalid status")
		})
	}
}

func TestHandleCallback_TransactionNotFound(t *testing.T) {
	payRepo := &mockPaymentRepo{
		FindByIDFn: func(ctx context.Context, id uint64) (*domainPayment.PaymentTransaction, error) {
			return nil, errors.New("record not found")
		},
	}

	uc := newTestUseCase(payRepo, &mockPlanRepo{}, &mockSubRepo{})

	err := uc.HandleCallback(context.Background(), HandleCallbackInput{
		TransactionID: 99,
		Status:        "success",
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "find transaction")
}

func TestHandleCallback_AlreadyProcessed(t *testing.T) {
	// Arrange – transaksi sudah berstatus success; tidak boleh diproses ulang
	fakeTx := &domainPayment.PaymentTransaction{
		ID:     42,
		Status: domainPayment.StatusSuccess, // bukan pending
	}
	payRepo := &mockPaymentRepo{
		FindByIDFn: func(ctx context.Context, id uint64) (*domainPayment.PaymentTransaction, error) {
			return fakeTx, nil
		},
	}

	uc := newTestUseCase(payRepo, &mockPlanRepo{}, &mockSubRepo{})

	err := uc.HandleCallback(context.Background(), HandleCallbackInput{
		TransactionID: 42,
		Status:        "success",
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already processed")
}

func TestHandleCallback_UpdateStatusError(t *testing.T) {
	fakeTx := &domainPayment.PaymentTransaction{
		ID:     42,
		Status: domainPayment.StatusPending,
	}
	payRepo := &mockPaymentRepo{
		FindByIDFn: func(ctx context.Context, id uint64) (*domainPayment.PaymentTransaction, error) {
			return fakeTx, nil
		},
		UpdateStatusFn: func(ctx context.Context, id uint64, status domainPayment.Status, paidAt *time.Time) error {
			return errors.New("db write error")
		},
	}

	uc := newTestUseCase(payRepo, &mockPlanRepo{}, &mockSubRepo{})

	err := uc.HandleCallback(context.Background(), HandleCallbackInput{
		TransactionID: 42,
		Status:        "success",
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "update payment status")
}

func TestHandleCallback_SubscriptionUpsertError(t *testing.T) {
	// Arrange – update status berhasil tapi upsert subscription gagal
	fakeTx := &domainPayment.PaymentTransaction{
		ID:     42,
		UserID: 1,
		PlanID: 10,
		Tier:   domainPayment.TierGold,
		Status: domainPayment.StatusPending,
	}

	payRepo := &mockPaymentRepo{
		FindByIDFn: func(ctx context.Context, id uint64) (*domainPayment.PaymentTransaction, error) {
			return fakeTx, nil
		},
		UpdateStatusFn: func(ctx context.Context, id uint64, status domainPayment.Status, paidAt *time.Time) error {
			return nil
		},
	}
	planRepo := &mockPlanRepo{
		FindActiveByTierFn: func(ctx context.Context, tier domainSub.Tier) (*domainSub.SubscriptionPlan, error) {
			return &domainSub.SubscriptionPlan{ID: 10, DurationDays: 30}, nil
		},
	}
	subRepo := &mockSubRepo{
		UpsertFn: func(ctx context.Context, s *domainSub.Subscription) error {
			return errors.New("subscription upsert failed")
		},
	}

	uc := newTestUseCase(payRepo, planRepo, subRepo)

	err := uc.HandleCallback(context.Background(), HandleCallbackInput{
		TransactionID: 42,
		Status:        "success",
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "activate subscription")
}
