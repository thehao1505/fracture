package persistence

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lukenguyen/fracture/internal/domain"
	"github.com/lukenguyen/fracture/internal/infrastructure/persistence/sqlcgen"
	"github.com/lukenguyen/fracture/internal/repository"
)

type postgresBlockRepo struct {
	q    *sqlcgen.Queries
	pool *pgxpool.Pool // Transaction for reorder
}

var _ repository.BlockRepository = (*postgresBlockRepo)(nil)

func NewPostgresBlockRepo(db *pgxpool.Pool) repository.BlockRepository {
	return &postgresBlockRepo{q: sqlcgen.New(db), pool: db}
}

func blockToDomain(b sqlcgen.Block) *domain.Block {
	return &domain.Block{
		ID:         b.ID,
		ProfileID:  b.ProfileID,
		Type:       domain.BlockType(b.Type),
		Content:    json.RawMessage(b.Content),
		Position:   b.Position,
		IsActive:   b.IsActive,
		ClickCount: b.ClickCount,
		CreatedAt:  b.CreatedAt,
		UpdatedAt:  b.UpdatedAt,
	}
}

func (r *postgresBlockRepo) Create(ctx context.Context, b *domain.Block) error {
	return r.q.CreateBlock(ctx, sqlcgen.CreateBlockParams{
		ID:        b.ID,
		ProfileID: b.ProfileID,
		Type:      string(b.Type),
		Content:   jsonOrEmpty(b.Content),
		Position:  b.Position,
		IsActive:  b.IsActive,
		CreatedAt: b.CreatedAt,
		UpdatedAt: b.UpdatedAt,
	})
}

func (r *postgresBlockRepo) Delete(ctx context.Context, id uuid.UUID, profileID uuid.UUID) error {
	return r.q.DeleteBlock(ctx, sqlcgen.DeleteBlockParams{
		ID:        id,
		ProfileID: profileID,
	})
}

func (r *postgresBlockRepo) IncrementClick(ctx context.Context, id uuid.UUID) error {
	return r.q.IncrementBlockClick(ctx, id)
}

func (r *postgresBlockRepo) ListActiveByProfile(ctx context.Context, profileID uuid.UUID) ([]domain.Block, error) {
	rows, err := r.q.ListActiveBlocksByProfile(ctx, profileID)
	if err != nil {
		return nil, err
	}
	blocks := make([]domain.Block, 0, len(rows))
	for _, b := range rows {
		blocks = append(blocks, *blockToDomain(b))
	}
	return blocks, nil
}

func (r *postgresBlockRepo) ListByProfile(ctx context.Context, profileID uuid.UUID) ([]domain.Block, error) {
	rows, err := r.q.ListBlocksByProfile(ctx, profileID)
	if err != nil {
		return nil, err
	}
	blocks := make([]domain.Block, 0, len(rows))
	for _, b := range rows {
		blocks = append(blocks, *blockToDomain(b))
	}
	return blocks, nil
}

// Reorder cập nhật position cho từng block trong một transaction: hoặc tất cả
// thành công, hoặc rollback toàn bộ. position mới = vị trí của id trong orderedIDs.
func (r *postgresBlockRepo) Reorder(ctx context.Context, profileID uuid.UUID, orderedIDs []uuid.UUID) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) // no-op nếu đã Commit thành công

	qtx := r.q.WithTx(tx)
	now := time.Now().UTC()
	for i, id := range orderedIDs {
		if err := qtx.UpdateBlockPosition(ctx, sqlcgen.UpdateBlockPositionParams{
			ID:        id,
			Position:  int32(i),
			UpdatedAt: now,
			ProfileID: profileID,
		}); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (r *postgresBlockRepo) Update(ctx context.Context, b *domain.Block) error {
	return r.q.UpdateBlock(ctx, sqlcgen.UpdateBlockParams{
		ID:        b.ID,
		Type:      string(b.Type),
		Content:   jsonOrEmpty(b.Content),
		IsActive:  b.IsActive,
		UpdatedAt: b.UpdatedAt,
		ProfileID: b.ProfileID,
	})
}
