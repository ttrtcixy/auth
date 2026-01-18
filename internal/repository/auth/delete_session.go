package authrepo

import (
	"context"

	storage "github.com/ttrtcixy/pgx-wrapper"
	entities "github.com/ttrtcixy/users/internal/domain/entities/auth"
	apperrors "github.com/ttrtcixy/users/internal/domain/errors"
)

var deleteSession = `
delete from refresh_tokens 
where client_uuid = $1;
`

func (r *AuthRepository) DeleteSession(ctx context.Context, payload *entities.SignoutRequest) error {
	const op = "AuthRepository.DeleteSession"

	q := storage.Query{
		Name:      "Delete user session by client_uuid",
		RawQuery:  deleteSession,
		Arguments: []any{payload.ClientUUID},
	}

	if _, err := r.DB.Exec(ctx, q); err != nil {
		return apperrors.Wrap(op, err)
	}

	return nil
}
