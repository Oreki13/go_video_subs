package payment

import (
	"context"
	"fmt"
	"time"

	"github.com/go_video_subs/internal/domain/payment"
	"github.com/jmoiron/sqlx"
)

type paymentRepository struct {
	db *sqlx.DB
}

func New(db *sqlx.DB) payment.Repository {
	return &paymentRepository{db: db}
}

func (r *paymentRepository) Create(ctx context.Context, tx *payment.PaymentTransaction) error {
	query := `
		INSERT INTO payment_transactions (user_id, plan_id, external_payment_id, tier, amount, status)
		VALUES (?, ?, ?, ?, ?, ?)
	`
	res, err := r.db.ExecContext(ctx, query,
		tx.UserID, tx.PlanID, tx.ExternalPaymentID, tx.Tier, tx.Amount, tx.Status,
	)
	if err != nil {
		return fmt.Errorf("repository: create payment transaction: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("repository: get last insert id: %w", err)
	}
	tx.ID = uint64(id)
	return nil
}

func (r *paymentRepository) FindByID(ctx context.Context, id uint64) (*payment.PaymentTransaction, error) {
	var tx payment.PaymentTransaction
	query := `
		SELECT id, user_id, plan_id, subscription_id, external_payment_id,
		       tier, amount, status, payload_raw, paid_at, created_at, updated_at
		FROM payment_transactions
		WHERE id = ?
	`
	if err := r.db.GetContext(ctx, &tx, query, id); err != nil {
		return nil, fmt.Errorf("repository: find payment by id %d: %w", id, err)
	}
	return &tx, nil
}

func (r *paymentRepository) UpdateStatus(ctx context.Context, id uint64, status payment.Status, paidAt *time.Time) error {
	query := `UPDATE payment_transactions SET status = ?, paid_at = ?, updated_at = NOW() WHERE id = ?`
	if _, err := r.db.ExecContext(ctx, query, status, paidAt, id); err != nil {
		return fmt.Errorf("repository: update payment status for id %d: %w", id, err)
	}
	return nil
}
