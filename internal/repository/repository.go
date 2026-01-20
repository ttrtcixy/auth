package repository

import (
	"context"
	"log/slog"

	storage "github.com/ttrtcixy/pgx-wrapper"
	authrepo "github.com/ttrtcixy/users/internal/repository/auth"
)

type Repository struct {
	log *slog.Logger
	*authrepo.AuthRepository
}

func NewRepository(ctx context.Context, log *slog.Logger, db *storage.Postgres) *Repository {
	return &Repository{
		log,
		authrepo.NewAuthRepository(ctx, log, db),
	}
}

func (r *Repository) RunInTx(ctx context.Context, fn func(context.Context) error) error {
	err := r.DB.RunInTx(ctx, fn)
	if err != nil {
		return err
	}
	return nil
}
