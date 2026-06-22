package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/lukenguyen/fracture/internal/domain"
)

type UserRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	FindByEmail(ctx context.Context, email string) (*domain.User, error)
	Create(ctx context.Context, user *domain.User) error
	Update(ctx context.Context, user *domain.User) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, keyword string, limit, offset int32) ([]*domain.User, error)
	Count(ctx context.Context, keyword string) (int64, error)
}
