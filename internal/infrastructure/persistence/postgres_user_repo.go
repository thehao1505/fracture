package persistence

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lukenguyen/fracture/internal/domain"
	"github.com/lukenguyen/fracture/internal/infrastructure/persistence/sqlcgen"
	"github.com/lukenguyen/fracture/internal/repository"
)

type postgresUserRepo struct {
	q *sqlcgen.Queries
}

var _ repository.UserRepository = (*postgresUserRepo)(nil)

func NewPostgresUserRepo(db *pgxpool.Pool) repository.UserRepository {
	return &postgresUserRepo{q: sqlcgen.New(db)}
}

// userToDomain maps the sqlc-generated row onto the domain model.
func userToDomain(u sqlcgen.User) *domain.User {
	return &domain.User{
		ID:        u.ID,
		Email:     u.Email,
		Password:  u.Password,
		Name:      u.Name,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

func (r *postgresUserRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	u, err := r.q.GetUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return userToDomain(u), nil
}

func (r *postgresUserRepo) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	u, err := r.q.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return userToDomain(u), nil
}

func (r *postgresUserRepo) Create(ctx context.Context, user *domain.User) error {
	return r.q.CreateUser(ctx, sqlcgen.CreateUserParams{
		ID:        user.ID,
		Email:     user.Email,
		Password:  user.Password,
		Name:      user.Name,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	})
}

func (r *postgresUserRepo) Update(ctx context.Context, user *domain.User) error {
	return r.q.UpdateUser(ctx, sqlcgen.UpdateUserParams{
		Email:     user.Email,
		Name:      user.Name,
		UpdatedAt: user.UpdatedAt,
		ID:        user.ID,
	})
}

func (r *postgresUserRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return r.q.DeleteUser(ctx, id)
}
