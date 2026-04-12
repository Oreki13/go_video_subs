package user

import (
	"context"
	"errors"
	"testing"

	domainUser "github.com/go_video_subs/internal/domain/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

// jwtSecret dan expiry yang dipakai seluruh test
const (
	testJWTSecret = "test-secret-key"
	testJWTExpiry = 60 // menit
)

// newTestUserUseCase adalah helper untuk membuat UseCase dengan mock repo.
func newTestUserUseCase(repo *mockUserRepo) *UseCase {
	return New(repo, testJWTSecret, testJWTExpiry)
}

// =============================================================================
// CreateUser Tests
// =============================================================================

func TestCreateUser_Success(t *testing.T) {
	// Arrange
	var storedUser *domainUser.User
	repo := &mockUserRepo{
		CreateFn: func(ctx context.Context, u *domainUser.User) error {
			storedUser = u
			return nil
		},
	}

	uc := newTestUserUseCase(repo)

	// Act
	err := uc.CreateUser(context.Background(), CreateUserInput{
		Name:     "Alice",
		Email:    "alice@example.com",
		Password: "secret123",
	})

	// Assert
	require.NoError(t, err)
	require.NotNil(t, storedUser)
	assert.Equal(t, "Alice", storedUser.Name)
	assert.Equal(t, "alice@example.com", storedUser.Email)

	// Password harus sudah di-hash, bukan plaintext
	assert.NotEqual(t, "secret123", storedUser.Password)

	// Verifikasi hash benar
	err = bcrypt.CompareHashAndPassword([]byte(storedUser.Password), []byte("secret123"))
	assert.NoError(t, err, "Stored password harus berupa bcrypt hash dari 'secret123'")
}

func TestCreateUser_RepoCreateError(t *testing.T) {
	// Arrange – repo gagal (misal email duplikat)
	repo := &mockUserRepo{
		CreateFn: func(ctx context.Context, u *domainUser.User) error {
			return errors.New("email already exists")
		},
	}

	uc := newTestUserUseCase(repo)

	// Act
	err := uc.CreateUser(context.Background(), CreateUserInput{
		Name:     "Alice",
		Email:    "alice@example.com",
		Password: "secret123",
	})

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "create user")
}

// =============================================================================
// Login Tests
// =============================================================================

func TestLogin_Success(t *testing.T) {
	// Arrange – buat hash terlebih dahulu seperti yang disimpan di DB
	plainPassword := "correctpassword"
	hashed, _ := bcrypt.GenerateFromPassword([]byte(plainPassword), bcrypt.MinCost)

	repo := &mockUserRepo{
		FindByEmailFn: func(ctx context.Context, email string) (*domainUser.User, error) {
			assert.Equal(t, "alice@example.com", email)
			return &domainUser.User{
				ID:       1,
				Email:    "alice@example.com",
				Password: string(hashed),
			}, nil
		},
	}

	uc := newTestUserUseCase(repo)

	// Act
	out, err := uc.Login(context.Background(), LoginInput{
		Email:    "alice@example.com",
		Password: plainPassword,
	})

	// Assert
	require.NoError(t, err)
	require.NotNil(t, out)
	assert.NotEmpty(t, out.Token, "Token JWT harus dihasilkan")
}

func TestLogin_WrongPassword(t *testing.T) {
	// Arrange
	hashed, _ := bcrypt.GenerateFromPassword([]byte("correctpassword"), bcrypt.MinCost)

	repo := &mockUserRepo{
		FindByEmailFn: func(ctx context.Context, email string) (*domainUser.User, error) {
			return &domainUser.User{
				ID:       1,
				Email:    "alice@example.com",
				Password: string(hashed),
			}, nil
		},
	}

	uc := newTestUserUseCase(repo)

	// Act
	out, err := uc.Login(context.Background(), LoginInput{
		Email:    "alice@example.com",
		Password: "wrongpassword",
	})

	// Assert
	assert.ErrorIs(t, err, ErrInvalidCredentials)
	assert.Nil(t, out)
}

func TestLogin_EmailNotFound(t *testing.T) {
	// Arrange – email tidak ditemukan di DB
	repo := &mockUserRepo{
		FindByEmailFn: func(ctx context.Context, email string) (*domainUser.User, error) {
			return nil, errors.New("record not found")
		},
	}

	uc := newTestUserUseCase(repo)

	// Act
	out, err := uc.Login(context.Background(), LoginInput{
		Email:    "notexist@example.com",
		Password: "anypassword",
	})

	// Assert
	// Usecase harus mengembalikan ErrInvalidCredentials (bukan expose error internal DB)
	assert.ErrorIs(t, err, ErrInvalidCredentials)
	assert.Nil(t, out)
}
