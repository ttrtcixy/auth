package authrepo

import (
	"context"
	"errors"

	storage "github.com/ttrtcixy/pgx-wrapper"
	entities "github.com/ttrtcixy/users/internal/domain/entities/auth"
	apperrors "github.com/ttrtcixy/users/internal/domain/errors"
)

var refreshSession = `
with update_session as (
	update refresh_tokens
	set jti = $1
	where client_uuid = $2 and jti = $3 
	returning user_id
)
select 
	u.username,
	u.email,
	u.role_id 
from users u
where u.user_id = (select user_id from update_session);
`

func (r *AuthRepository) UpdateSession(ctx context.Context, payload *entities.UpdateSessionRequest) (*entities.TokenUserInfo, error) {
	const op = "AuthRepository.RefreshSession"

	q := storage.Query{
		Name:      "Refresh user session",
		RawQuery:  refreshSession,
		Arguments: []any{payload.NewRefreshTokenUUID, payload.ClientUUID, payload.OldRefreshTokenUUID},
	}

	userInfo := &entities.TokenUserInfo{}

	if err := r.DB.QueryRow(ctx, q).Scan(&userInfo.Username, &userInfo.Email, &userInfo.RoleID); err != nil {
		if errors.Is(err, storage.ErrNoRows) {
			return nil, apperrors.ErrInvalidRefreshToken
		}
		return nil, apperrors.Wrap(op, err)
	}

	return userInfo, nil
}
