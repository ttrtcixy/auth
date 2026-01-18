package authrepo

import (
	"context"
	"errors"

	storage "github.com/ttrtcixy/pgx-wrapper"
	entities "github.com/ttrtcixy/users/internal/domain/entities/auth"
	apperrors "github.com/ttrtcixy/users/internal/domain/errors"
)

var userByEmail = `
select 
    u.user_id, 
    u.username, 
    u.is_active, 
    u.role_id, 
    up.hash, 
    up.salt  
from users u
	join user_password up using(user_id) 
where email = $1;
`

func (r *AuthRepository) UserByEmail(ctx context.Context, email string) (*entities.User, error) {
	const op = "AuthRepository.UserByEmail"

	q := storage.Query{
		Name:      "Get user by email",
		RawQuery:  userByEmail,
		Arguments: []any{email},
	}
	var user = &entities.User{}

	err := r.DB.QueryRow(ctx, q).Scan(
		&user.ID,
		&user.Username,
		&user.IsActive,
		&user.RoleId,
		&user.PasswordHash,
		&user.PasswordSalt)
	if err != nil {
		if errors.Is(err, storage.ErrNoRows) {
			return nil, apperrors.ErrUserNotRegister
		}
		return nil, apperrors.Wrap(op, err)
	}

	user.Email = email

	return user, nil
}
