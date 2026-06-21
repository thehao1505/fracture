package usecase

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/lukenguyen/fracture/internal/domain"
	"github.com/lukenguyen/fracture/internal/repository"
	"github.com/lukenguyen/fracture/pkg/hash"
	"github.com/lukenguyen/fracture/pkg/token"
)

type AuthUseCase struct {
	userRepo repository.UserRepository
	tokens   *token.Manager
}

func NewAuthUseCase(repo repository.UserRepository, tokens *token.Manager) *AuthUseCase {
	return &AuthUseCase{userRepo: repo, tokens: tokens}
}

// Register creates a new user with a bcrypt-hashed password. The plaintext
// password never leaves this method — only its hash is stored.
func (uc *AuthUseCase) Register(ctx context.Context, email, password, name string) (*domain.User, error) {
	if email == "" || password == "" || name == "" {
		return nil, domain.ErrBadRequest
	}

	if _, err := uc.userRepo.FindByEmail(ctx, email); err == nil {
		return nil, domain.ErrConflict
	} else if err != domain.ErrNotFound {
		return nil, err
	}

	hashed, err := hash.Password(password)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	user := &domain.User{
		ID:        uuid.New(),
		Email:     email,
		Password:  hashed,
		Name:      name,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := uc.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

// Login verifies the credentials and, on success, returns the user together
// with a signed access token. A missing user and a wrong password yield the
// same error so callers can't probe which emails are registered.
func (uc *AuthUseCase) Login(ctx context.Context, email, password string) (*domain.User, string, error) {
	user, err := uc.userRepo.FindByEmail(ctx, email)
	if err != nil {
		if err == domain.ErrNotFound {
			return nil, "", domain.ErrInvalidCredentials
		}
		return nil, "", err
	}

	if !hash.Compare(user.Password, password) {
		return nil, "", domain.ErrInvalidCredentials
	}

	accessToken, err := uc.tokens.Generate(user.ID, user.Email)
	if err != nil {
		return nil, "", err
	}
	return user, accessToken, nil
}
