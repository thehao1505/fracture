package usecase

import (
	"context"
	"time"

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
	if user.Email == "" || user.Name == "" {
		return domain.ErrBadRequest
	}

	user.ID = uuid.New()
	now := time.Now().UTC()
	user.CreatedAt = now
	user.UpdatedAt = now

	if _, err := uc.userRepo.FindByEmail(ctx, user.Email); err == nil {
		return domain.ErrConflict
	} else if err != domain.ErrNotFound {
		return err
	}

	return uc.userRepo.Create(ctx, user)
}

func (uc *UserUseCase) UpdateUser(ctx context.Context, id string, user *domain.User) error {
	uid, err := uuid.Parse(id)
	if err != nil {
		return domain.ErrInvalidID
	}

	if user.Email == "" && user.Name == "" {
		return domain.ErrBadRequest
	}

	existingUser, err := uc.userRepo.FindByID(ctx, uid)
	if err != nil {
		return err
	}

	if user.Email != "" && user.Email != existingUser.Email {
		if _, err := uc.userRepo.FindByEmail(ctx, user.Email); err == nil {
			return domain.ErrConflict
		} else if err != domain.ErrNotFound {
			return err
		}
	}

	if user.Email != "" {
		existingUser.Email = user.Email
	}
	if user.Name != "" {
		existingUser.Name = user.Name
	}
	existingUser.UpdatedAt = time.Now().UTC()

	return uc.userRepo.Update(ctx, existingUser)
}

func (uc *UserUseCase) DeleteUser(ctx context.Context, id string) error {
	uid, err := uuid.Parse(id)
	if err != nil {
		return domain.ErrInvalidID
	}

	if _, err := uc.userRepo.FindByID(ctx, uid); err != nil {
		return err
	}

	return uc.userRepo.Delete(ctx, uid)
}

type ListUsersParams struct {
	Page    int    // bắt đầu từ 1
	Limit   int    // số item mỗi trang
	Keyword string
}

func (uc *UserUseCase) ListUsers(ctx context.Context, p ListUsersParams) ([]*domain.User, int64, error) {
	// Chuẩn hóa & chặn giá trị bất thường
	if p.Page < 1 {
		p.Page = 1
	}
	if p.Limit < 1 {
		p.Limit = 20 // default
	}
	if p.Limit > 100 {
		p.Limit = 100 // cap, tránh client xin 1 triệu record
	}

	offset := (p.Page - 1) * p.Limit

	users, err := uc.userRepo.List(ctx, p.Keyword, int32(p.Limit), int32(offset))
	if err != nil {
		return nil, 0, err
	}

	total, err := uc.userRepo.Count(ctx, p.Keyword)
	if err != nil {
		return nil, 0, err
	}

	return users, total, nil
}