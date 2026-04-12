package payment

// File ini berisi mock manual (tanpa library generator) untuk testing.
// Mock mengimplementasi interface dari domain layer.
// Sesuai dengan prinsip Clean Architecture:
//   domain -> usecase -> delivery
// Mock menggantikan dependency (repository) agar usecase bisa ditest secara isolated.

import (
	"context"
	"time"

	domainPayment "github.com/go_video_subs/internal/domain/payment"
	domainSub "github.com/go_video_subs/internal/domain/subscription"
)

// ---------------------------------------------------------------------------
// Mock: domainPayment.Repository
// ---------------------------------------------------------------------------

type mockPaymentRepo struct {
	// CreateFn dipanggil saat Create() dieksekusi; set sesuai kebutuhan test.
	CreateFn       func(ctx context.Context, tx *domainPayment.PaymentTransaction) error
	FindByIDFn     func(ctx context.Context, id uint64) (*domainPayment.PaymentTransaction, error)
	UpdateStatusFn func(ctx context.Context, id uint64, status domainPayment.Status, paidAt *time.Time) error
}

func (m *mockPaymentRepo) Create(ctx context.Context, tx *domainPayment.PaymentTransaction) error {
	return m.CreateFn(ctx, tx)
}

func (m *mockPaymentRepo) FindByID(ctx context.Context, id uint64) (*domainPayment.PaymentTransaction, error) {
	return m.FindByIDFn(ctx, id)
}

func (m *mockPaymentRepo) UpdateStatus(ctx context.Context, id uint64, status domainPayment.Status, paidAt *time.Time) error {
	return m.UpdateStatusFn(ctx, id, status, paidAt)
}

// ---------------------------------------------------------------------------
// Mock: domainSub.PlanRepository
// ---------------------------------------------------------------------------

type mockPlanRepo struct {
	FindActiveByTierFn func(ctx context.Context, tier domainSub.Tier) (*domainSub.SubscriptionPlan, error)
}

func (m *mockPlanRepo) FindActiveByTier(ctx context.Context, tier domainSub.Tier) (*domainSub.SubscriptionPlan, error) {
	return m.FindActiveByTierFn(ctx, tier)
}

// ---------------------------------------------------------------------------
// Mock: domainSub.Repository
// ---------------------------------------------------------------------------

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
