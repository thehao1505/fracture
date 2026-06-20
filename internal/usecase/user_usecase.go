package usecase

import (
	"context"

	"github.com/google/uuid"
	"github.com/lukenguyen/fracture/internal/domain"
	"github.com/lukenguyen/fracture/internal/repository"
)

type UserUseCase struct {
	userRepo repository.UserRepository
}

func NewUserUseCase(repo repository.UserRepository) *UserUseCase {
	return &UserUseCase{userRepo: repo}
}

func (uc *UserUseCase) GetUser(ctx context.Context, id string) (*domain.User, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, domain.ErrInvalidID
	}

	user, err := uc.userRepo.FindByID(ctx, uid)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (uc *UserUseCase) CreateUser(ctx context.Context, user *domain.User) error {
	return uc.userRepo.Create(ctx, user)
}