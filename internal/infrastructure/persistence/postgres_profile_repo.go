package persistence

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lukenguyen/fracture/internal/domain"
	"github.com/lukenguyen/fracture/internal/infrastructure/persistence/sqlcgen"
	"github.com/lukenguyen/fracture/internal/repository"
)

type postgresProfileRepo struct {
	q    *sqlcgen.Queries
	pool *pgxpool.Pool // Transaction for reorder
}

var _ repository.ProfileRepository = (*postgresProfileRepo)(nil)

func NewPostgresProfileRepo(db *pgxpool.Pool) repository.ProfileRepository {
	return &postgresProfileRepo{q: sqlcgen.New(db), pool: db}
}

func profileToDomain(p sqlcgen.Profile) *domain.Profile {
	return &domain.Profile{
		ID:          p.ID,
		UserID:      p.UserID,
		Username:    p.Username,
		DisplayName: p.DisplayName,
		Bio:         p.Bio,
		AvatarURL:   p.AvatarUrl,
		Appearance:  json.RawMessage(p.Appearance),
		IsPublished: p.IsPublished,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
}

// jsonOrEmpty đảm bảo không ghi NULL vào cột JSONB (NOT NULL DEFAULT '{}').
func jsonOrEmpty(raw json.RawMessage) []byte {
	if len(raw) == 0 {
		return []byte("{}")
	}
	return raw
}

func (r *postgresProfileRepo) Create(ctx context.Context, p *domain.Profile) error {
	return r.q.CreateProfile(ctx, sqlcgen.CreateProfileParams{
		ID:          p.ID,
		UserID:      p.UserID,
		Username:    p.Username,
		DisplayName: p.DisplayName,
		Bio:         p.Bio,
		AvatarUrl:   p.AvatarURL,
		Appearance:  jsonOrEmpty(p.Appearance),
		IsPublished: p.IsPublished,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	})
}

func (r *postgresProfileRepo) GetByUserID(ctx context.Context, userID uuid.UUID) (*domain.Profile, error) {
	p, err := r.q.GetProfileByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return profileToDomain(p), nil
}

func (r *postgresProfileRepo) GetPublishedByUsername(ctx context.Context, username string) (*domain.Profile, error) {
	p, err := r.q.GetPublishedProfileByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return profileToDomain(p), nil
}

func (r *postgresProfileRepo) Update(ctx context.Context, p *domain.Profile) error {
	return r.q.UpdateProfile(ctx, sqlcgen.UpdateProfileParams{
		ID:          p.ID,
		Username:    p.Username,
		DisplayName: p.DisplayName,
		Bio:         p.Bio,
		AvatarUrl:   p.AvatarURL,
		Appearance:  jsonOrEmpty(p.Appearance),
		IsPublished: p.IsPublished,
		UpdatedAt:   p.UpdatedAt,
	})
}

func (r *postgresProfileRepo) UsernameExists(ctx context.Context, username string, exclude uuid.UUID) (bool, error) {
	return r.q.ProfileUsernameExists(ctx, sqlcgen.ProfileUsernameExistsParams{
		Username: username,
		ID:       exclude,
	})
}
