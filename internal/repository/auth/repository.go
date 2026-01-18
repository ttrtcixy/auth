package authrepo

import (
	"context"
	"log/slog"

	storage "github.com/ttrtcixy/pgx-wrapper"
)

type AuthRepository struct {
	log *slog.Logger
	DB  storage.DB
}

func NewAuthRepository(_ context.Context, log *slog.Logger, db storage.DB) *AuthRepository {
	return &AuthRepository{
		log: log,
		DB:  db,
	}
}
