package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/lukenguyen/fracture/internal/domain"
)

type ProfileRepository interface {
	Create(ctx context.Context, p *domain.Profile) error
	GetByUserID(ctx context.Context, userID uuid.UUID) (*domain.Profile, error)
	GetPublishedByUsername(ctx context.Context, username string) (*domain.Profile, error)
	Update(ctx context.Context, p *domain.Profile) error
	// exclude là id profile đang sửa (lúc tạo mới truyền uuid.Nil).
	UsernameExists(ctx context.Context, username string, exclude uuid.UUID) (bool, error)
}

type BlockRepository interface {
	Create(ctx context.Context, b *domain.Block) error
	ListByProfile(ctx context.Context, profileID uuid.UUID) ([]domain.Block, error)
	ListActiveByProfile(ctx context.Context, profileID uuid.UUID) ([]domain.Block, error)
	Update(ctx context.Context, b *domain.Block) error
	Delete(ctx context.Context, id, profileID uuid.UUID) error
	IncrementClick(ctx context.Context, id uuid.UUID) error
	// Reorder cập nhật position cho nhiều block trong 1 transaction.
	Reorder(ctx context.Context, profileID uuid.UUID, orderedIDs []uuid.UUID) error
}
