package payment

import (
	"context"
	"time"
)

type Repository interface {
	Create(ctx context.Context, tx *PaymentTransaction) error
	FindByID(ctx context.Context, id uint64) (*PaymentTransaction, error)
	UpdateStatus(ctx context.Context, id uint64, status Status, paidAt *time.Time) error
}
