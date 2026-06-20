package persistence

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lukenguyen/fracture/internal/domain"
	"github.com/lukenguyen/fracture/internal/repository"
)

type postgresUserRepo struct {
    db *pgxpool.Pool
}

var _ repository.UserRepository = (*postgresUserRepo)(nil)

func NewPostgresUserRepo(db *pgxpool.Pool) repository.UserRepository {
    return &postgresUserRepo{db: db}
}

func (r *postgresUserRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	var user domain.User
	err := r.db.QueryRow(ctx,
		"SELECT id, email, name, created_at FROM users WHERE id = $1", id,
	).Scan(&user.ID, &user.Email, &user.Name, &user.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *postgresUserRepo) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	var user domain.User
	err := r.db.QueryRow(ctx,
		"SELECT id, email, name, created_at FROM users WHERE email = $1", email,
	).Scan(&user.ID, &user.Email, &user.Name, &user.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *postgresUserRepo) Create(ctx context.Context, user *domain.User) error {
	_, err := r.db.Exec(ctx,
		"INSERT INTO users (id, email, name, created_at, updated_at) VALUES ($1, $2, $3, $4, $5)",
		user.ID, user.Email, user.Name, user.CreatedAt, user.UpdatedAt,
	)
	return err
}

func (r *postgresUserRepo) Update(ctx context.Context, user *domain.User) error {
	_, err := r.db.Exec(ctx,
		"UPDATE users SET email = $1, name = $2, updated_at = $3 WHERE id = $4",
		user.Email, user.Name, user.UpdatedAt, user.ID,
	)
	return err
}

func (r *postgresUserRepo) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx,
		"DELETE FROM users WHERE id = $1", id,
	)
	return err
}