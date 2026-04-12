package user

import (
	"context"

	domainUser "github.com/go_video_subs/internal/domain/user"
)

// mockUserRepo mengimplementasi domainUser.Repository untuk keperluan testing.
// Mock ini memungkinkan pengujian UseCase tanpa koneksi database.
type mockUserRepo struct {
	FindAllFn      func(ctx context.Context) ([]domainUser.User, error)
	FindByEmailFn  func(ctx context.Context, email string) (*domainUser.User, error)
	CreateFn       func(ctx context.Context, u *domainUser.User) error
}

func (m *mockUserRepo) FindAll(ctx context.Context) ([]domainUser.User, error) {
	return m.FindAllFn(ctx)
}

func (m *mockUserRepo) FindByEmail(ctx context.Context, email string) (*domainUser.User, error) {
	return m.FindByEmailFn(ctx, email)
}

func (m *mockUserRepo) Create(ctx context.Context, u *domainUser.User) error {
	return m.CreateFn(ctx, u)
}
