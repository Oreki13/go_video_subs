package user

import (
	"context"
	"errors"
	"fmt"

	"github.com/go_video_subs/internal/domain/user"
	appjwt "github.com/go_video_subs/pkg/jwt"
	"golang.org/x/crypto/bcrypt"
)

type UseCase struct {
	repo      user.Repository
	jwtSecret string
	jwtExpiry int
}

func New(repo user.Repository, jwtSecret string, jwtExpiry int) *UseCase {
	return &UseCase{repo: repo, jwtSecret: jwtSecret, jwtExpiry: jwtExpiry}
}

type CreateUserInput struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginOutput struct {
	Token string `json:"token"`
}

var ErrInvalidCredentials = errors.New("invalid email or password")

func (uc *UseCase) CreateUser(ctx context.Context, input CreateUserInput) error {
	hashed, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("usecase: hash password: %w", err)
	}

	u := &user.User{
		Name:     input.Name,
		Email:    input.Email,
		Password: string(hashed),
	}

	if err := uc.repo.Create(ctx, u); err != nil {
		return fmt.Errorf("usecase: create user: %w", err)
	}

	return nil
}

func (uc *UseCase) Login(ctx context.Context, input LoginInput) (*LoginOutput, error) {
	u, err := uc.repo.FindByEmail(ctx, input.Email)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(input.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	token, err := appjwt.GenerateToken(u.ID, u.Email, uc.jwtSecret, uc.jwtExpiry)
	if err != nil {
		return nil, fmt.Errorf("usecase: generate token: %w", err)
	}

	return &LoginOutput{Token: token}, nil
}
