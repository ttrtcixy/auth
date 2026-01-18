package authusecase

import (
	"context"
	"errors"
	"log/slog"

	"github.com/ttrtcixy/users/internal/domain/contracts"
	entities "github.com/ttrtcixy/users/internal/domain/entities/auth"
	apperrors "github.com/ttrtcixy/users/internal/domain/errors"
)

type SignoutUseCase struct {
	log  *slog.Logger
	repo contracts.SignoutRepository
}

func NewSignout(log *slog.Logger, repo contracts.Repository) *SignoutUseCase {
	return &SignoutUseCase{
		log:  log,
		repo: repo,
	}
}

func (u *SignoutUseCase) Signout(ctx context.Context, payload *entities.SignoutRequest) (err error) {
	const op = "authusecase.Signout()"

	defer func() {
		if err != nil {
			var userErr apperrors.UserError
			if errors.As(err, &userErr) {
				return
			}

			u.log.LogAttrs(ctx, slog.LevelError, "Signout error", slog.String("op", op), slog.String("error", err.Error()))
			err = apperrors.ErrServer
		}
	}()

	if err := u.repo.DeleteSession(ctx, payload); err != nil {
		return apperrors.Wrap(op, err)
	}

	return nil
}
