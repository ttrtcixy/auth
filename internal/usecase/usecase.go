package usecase

import (
	"log/slog"

	"github.com/ttrtcixy/users/internal/domain/contracts"
	domainservice "github.com/ttrtcixy/users/internal/domain/service"
	infraservice "github.com/ttrtcixy/users/internal/infrastructure/service"
	authusecase "github.com/ttrtcixy/users/internal/usecase/auth"
)

type UseCase struct {
	*AuthUseCase
}

func NewUseCase(log *slog.Logger, repo contracts.Repository, infraService *infraservice.InfraServices, domainService *domainservice.DomainService) *UseCase {
	return &UseCase{
		NewAuthUseCase(log, repo, infraService, domainService),
	}
}

type AuthUseCase struct {
	*authusecase.SignoutUseCase
	*authusecase.SignupUseCase
	*authusecase.SigninUseCase
	*authusecase.VerifyEmailUseCase
	*authusecase.UpdateTokenUseCase
	*authusecase.VerifyTokensUseCase
}

func NewAuthUseCase(log *slog.Logger, repo contracts.Repository, infraService *infraservice.InfraServices, domainService *domainservice.DomainService) *AuthUseCase {
	return &AuthUseCase{
		SignoutUseCase: authusecase.NewSignout(log, repo),
		SignupUseCase: authusecase.NewSignup(log, repo, &authusecase.SignupServices{
			MailService:               infraService.Smtp,
			HashService:               infraService.PasswordHasher,
			EmailTokenVerifierService: infraService.Jwt,
		}),
		SigninUseCase: authusecase.NewSignin(log, repo, &authusecase.SigninServices{
			TokenCreator:         infraService.Jwt,
			PasswordComparer:     infraService.PasswordHasher,
			SigninSessionCreator: domainService.SessionService,
		}),
		VerifyEmailUseCase:  authusecase.NewVerifyEmail(log, repo, infraService.Jwt),
		UpdateTokenUseCase:  authusecase.NewRefreshUseCase(log, repo, infraService.Jwt),
		VerifyTokensUseCase: authusecase.NewTokensVerify(log, infraService.Jwt),
	}
}
