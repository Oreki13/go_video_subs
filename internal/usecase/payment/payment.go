package payment

import (
	"context"
	"fmt"
	"time"

	domainPayment "github.com/go_video_subs/internal/domain/payment"
	domainSub "github.com/go_video_subs/internal/domain/subscription"
)

type UseCase struct {
	paymentRepo domainPayment.Repository
	planRepo    domainSub.PlanRepository
	subRepo     domainSub.Repository
}

func New(paymentRepo domainPayment.Repository, planRepo domainSub.PlanRepository, subRepo domainSub.Repository) *UseCase {
	return &UseCase{
		paymentRepo: paymentRepo,
		planRepo:    planRepo,
		subRepo:     subRepo,
	}
}

type InitiatePaymentInput struct {
	UserID uint64
	Tier   string
}

type InitiatePaymentOutput struct {
	TransactionID     uint64  `json:"transaction_id"`
	PlanID            uint64  `json:"plan_id"`
	Tier              string  `json:"tier"`
	Amount            float64 `json:"amount"`
	ExternalPaymentID string  `json:"external_payment_id"`
	MockCallbackURL   string  `json:"mock_callback_url"`
}

type HandleCallbackInput struct {
	TransactionID uint64 `json:"transaction_id"`
	Status        string `json:"status"`
}

func (uc *UseCase) InitiatePayment(ctx context.Context, input InitiatePaymentInput) (*InitiatePaymentOutput, error) {
	tier := domainSub.Tier(input.Tier)
	if tier != domainSub.TierGold && tier != domainSub.TierSilver && tier != domainSub.TierBronze {
		return nil, fmt.Errorf("invalid tier: %s", input.Tier)
	}

	plan, err := uc.planRepo.FindActiveByTier(ctx, tier)
	if err != nil {
		return nil, fmt.Errorf("usecase: find active plan: %w", err)
	}

	externalID := fmt.Sprintf("mock-%d-%d", input.UserID, time.Now().UnixNano())

	tx := &domainPayment.PaymentTransaction{
		UserID:            input.UserID,
		PlanID:            plan.ID,
		ExternalPaymentID: externalID,
		Tier:              domainPayment.Tier(tier),
		Amount:            plan.Price,
		Status:            domainPayment.StatusPending,
	}

	if err := uc.paymentRepo.Create(ctx, tx); err != nil {
		return nil, fmt.Errorf("usecase: create payment transaction: %w", err)
	}

	return &InitiatePaymentOutput{
		TransactionID:     tx.ID,
		PlanID:            plan.ID,
		Tier:              string(tier),
		Amount:            plan.Price,
		ExternalPaymentID: externalID,
		MockCallbackURL:   "/api/v1/payments/callback",
	}, nil
}

func (uc *UseCase) HandleCallback(ctx context.Context, input HandleCallbackInput) error {
	status := domainPayment.Status(input.Status)
	if status != domainPayment.StatusSuccess && status != domainPayment.StatusFailed {
		return fmt.Errorf("invalid status: %s; accepted values: success, failed", input.Status)
	}

	tx, err := uc.paymentRepo.FindByID(ctx, input.TransactionID)
	if err != nil {
		return fmt.Errorf("usecase: find transaction: %w", err)
	}

	if tx.Status != domainPayment.StatusPending {
		return fmt.Errorf("transaction %d already processed with status: %s", tx.ID, tx.Status)
	}

	var paidAt *time.Time
	if status == domainPayment.StatusSuccess {
		now := time.Now()
		paidAt = &now
	}

	if err := uc.paymentRepo.UpdateStatus(ctx, tx.ID, status, paidAt); err != nil {
		return fmt.Errorf("usecase: update payment status: %w", err)
	}

	if status == domainPayment.StatusSuccess {
		plan, err := uc.planRepo.FindActiveByTier(ctx, domainSub.Tier(tx.Tier))
		if err != nil {
			return fmt.Errorf("usecase: find plan for subscription activation: %w", err)
		}

		now := time.Now()
		expiredAt := now.AddDate(0, 0, plan.DurationDays)

		sub := &domainSub.Subscription{
			UserID:    tx.UserID,
			PlanID:    plan.ID,
			Tier:      domainSub.Tier(tx.Tier),
			Status:    domainSub.StatusActive,
			StartedAt: &now,
			ExpiredAt: &expiredAt,
		}

		if err := uc.subRepo.Upsert(ctx, sub); err != nil {
			return fmt.Errorf("usecase: activate subscription: %w", err)
		}
	}

	return nil
}
