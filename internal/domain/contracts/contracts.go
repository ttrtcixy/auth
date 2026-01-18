package contracts

import (
	"context"

	entities "github.com/ttrtcixy/users/internal/domain/entities/auth"
)

type Repository interface {
	AuthRepository
}

type AuthRepository interface {
	SignupRepository
	SigninRepository
	VerifyRepository
	UpdateSessionRepository
	SignoutRepository
	CreateSessionRepository
}

type Tx interface {
	RunInTx(ctx context.Context, fn func(context.Context) error) error
}

type SignupRepository interface {
	CheckLoginExist(ctx context.Context, payload *entities.SignupRequest) (*entities.CheckLoginResponse, error)
	CreateUser(ctx context.Context, payload *entities.CreateUserRequest) error
	Tx
}

type SigninRepository interface {
	UserByEmail(ctx context.Context, email string) (*entities.User, error)
	UserByUsername(ctx context.Context, username string) (*entities.User, error)
}

type VerifyRepository interface {
	ActivateUser(ctx context.Context, email string) (*entities.TokenUserInfo, error)
}

type UpdateSessionRepository interface {
	UpdateSession(ctx context.Context, payload *entities.UpdateSessionRequest) (*entities.TokenUserInfo, error)
}
type SignoutRepository interface {
	DeleteSession(ctx context.Context, payload *entities.SignoutRequest) error
}

type CreateSessionRepository interface {
	CreateSession(ctx context.Context, payload *entities.Session) error
}
