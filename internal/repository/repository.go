package repository

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5"
	storage "github.com/ttrtcixy/pgx-wrapper"
	"github.com/ttrtcixy/users/internal/domain"
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

func (r *Repository) RunInTx(ctx context.Context, txOptions domain.TxOptions, fn func(context.Context) error) error {
	var isoLevel pgx.TxIsoLevel
	switch txOptions.IsoLevel {
	case domain.Serializable:
		isoLevel = pgx.Serializable
	case domain.ReadCommitted:
		isoLevel = pgx.ReadCommitted
	case domain.RepeatableRead:
		isoLevel = pgx.RepeatableRead
	case domain.ReadUncommitted:
		isoLevel = pgx.ReadUncommitted
	default:
		r.log.LogAttrs(ctx, slog.LevelWarn, "transaction isolation level not specified, set by default: ReadCommitted")
		isoLevel = pgx.ReadCommitted
	}

	var accessMode pgx.TxAccessMode

	switch txOptions.AccessMode {
	case domain.ReadWrite:
		accessMode = pgx.ReadWrite
	case domain.ReadOnly:
		accessMode = pgx.ReadOnly
	default:
		accessMode = pgx.ReadWrite
	}

	return r.DB.RunInTx(ctx, pgx.TxOptions{IsoLevel: isoLevel, AccessMode: accessMode}, fn)
}
