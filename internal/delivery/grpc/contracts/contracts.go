package contracts

import (
	"context"

	entities "github.com/ttrtcixy/users/internal/domain/entities/auth"
)

type UseCase interface {
	AuthUseCase
}

type SigninUseCase interface {
	Signin(ctx context.Context, payload *entities.SigninRequest) (*entities.SigninResponse, error)
}

type SignupUseCase interface {
	Signup(ctx context.Context, payload *entities.SignupRequest) error
}

type SignoutUseCase interface {
	Signout(ctx context.Context, payload *entities.SignoutRequest) error
}

type VerifyUseCase interface {
	EmailVerify(ctx context.Context, payload *entities.EmailVerifyRequest) (*entities.EmailVerifyResponse, error)
}

type RefreshUseCase interface {
	Refresh(ctx context.Context, payload *entities.RefreshRequest) (*entities.RefreshResponse, error)
}

type TokensVerifyUseCase interface {
	TokensVerify(ctx context.Context, payload *entities.TokensVerifyRequest) (result *entities.TokensVerifyResponse, err error)
}

type AuthUseCase interface {
	SigninUseCase
	SignupUseCase
	SignoutUseCase
	VerifyUseCase
	RefreshUseCase
	TokensVerifyUseCase
}
