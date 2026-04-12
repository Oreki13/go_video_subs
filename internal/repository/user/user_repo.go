package user

import (
	"context"
	"fmt"

	"github.com/go_video_subs/internal/domain/user"
	"github.com/jmoiron/sqlx"
)

type userRepository struct {
	db *sqlx.DB
}

func New(db *sqlx.DB) user.Repository {
	return &userRepository{db: db}
}

func (r *userRepository) FindAll(ctx context.Context) ([]user.User, error) {
	users := make([]user.User, 0)
	query := `SELECT id, name, email, created_at, updated_at FROM users ORDER BY id DESC`
	if err := r.db.SelectContext(ctx, &users, query); err != nil {
		return nil, fmt.Errorf("repository: find all users: %w", err)
	}
	return users, nil
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*user.User, error) {
	var u user.User
	query := `SELECT id, name, email, password, created_at, updated_at FROM users WHERE email = ? LIMIT 1`
	if err := r.db.GetContext(ctx, &u, query, email); err != nil {
		return nil, fmt.Errorf("repository: find by email: %w", err)
	}
	return &u, nil
}

func (r *userRepository) Create(ctx context.Context, u *user.User) error {
	query := `INSERT INTO users (name, email, password) VALUES (:name, :email, :password)`
	if _, err := r.db.NamedExecContext(ctx, query, u); err != nil {
		return fmt.Errorf("repository: create user: %w", err)
	}
	return nil
}
