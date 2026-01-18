package grpcauthservise

import (
	"context"
	"log/slog"

	usersProtos "github.com/ttrtcixy/users-protos/gen/go/users"
	"github.com/ttrtcixy/users/internal/delivery/grpc/contracts"
)

type UserAuthService struct {
	*SigninService
	*SignupService
	*SignoutService
	*EmailVerifyService
	*RefreshService
	*TokensVerifyService
	usersProtos.UnsafeUsersAuthServer
}

func NewUserAuthService(_ context.Context, log *slog.Logger, usecase contracts.UseCase) usersProtos.UsersAuthServer {
	return &UserAuthService{
		SigninService:       NewSignin(log, usecase),
		SignupService:       NewSignup(log, usecase),
		SignoutService:      NewSignout(log, usecase),
		EmailVerifyService:  NewVerifyEmail(log, usecase),
		RefreshService:      NewRefresh(log, usecase),
		TokensVerifyService: NewVerifyTokens(log, usecase),
	}
}
