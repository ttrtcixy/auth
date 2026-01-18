package authusecase

import (
	"context"
	"errors"
	"log/slog"

	entities "github.com/ttrtcixy/users/internal/domain/entities/auth"
	apperrors "github.com/ttrtcixy/users/internal/domain/errors"
)

type Refresh interface {
	Refresh(ctx context.Context, payload *entities.RefreshRequest) (*entities.RefreshResponse, error)
}

type TokenVerifier interface {
	ParseAccessToken(token string) (result *entities.TokenUserInfo, err error)
}

type VerifyTokensUseCase struct {
	//_              *config.UsecaseConfig
	log            *slog.Logger
	token          TokenVerifier
	refreshService Refresh
}

func NewTokensVerify(log *slog.Logger, tokenService TokenVerifier) *VerifyTokensUseCase {
	return &VerifyTokensUseCase{
		//_:     nil,
		log:   log,
		token: tokenService,
		//refreshUsecase: dep.RefreshUsecase,
	}
}

func (u *VerifyTokensUseCase) TokensVerify(ctx context.Context, payload *entities.TokensVerifyRequest) (response *entities.TokensVerifyResponse, err error) {
	const op = "ValidateUseCase.Validate"
	defer func() {
		if err != nil {
			var userErr apperrors.UserError
			if errors.As(err, &userErr) {
				return
			}

			u.log.LogAttrs(nil, slog.LevelError, "Token verify error", slog.String("op", op), slog.String("error", err.Error()))
			err = apperrors.ErrServer
		}
	}()

	userInfo, err := u.token.ParseAccessToken(payload.AccessToken)
	if err != nil {
		if errors.Is(err, apperrors.ErrAccessTokenExpired) {
			result, err := u.refresh(ctx, payload.RefreshToken)
			if err != nil {
				return nil, err
			}
			return &entities.TokensVerifyResponse{
				Tokens: &entities.Tokens{
					AccessToken:  &result.AccessToken,
					RefreshToken: &result.RefreshToken,
				},
				TokenUserInfo: result.TokenUserInfo,
			}, nil
		}
		return nil, err
	}

	return &entities.TokensVerifyResponse{TokenUserInfo: userInfo}, nil
}

func (u *VerifyTokensUseCase) refresh(ctx context.Context, refreshToken string) (*entities.RefreshResponse, error) {
	const op = "VerifyTokensUseCase.refresh"
	if refreshToken == "" {
		return nil, apperrors.ErrInvalidRefreshToken
	}

	response, err := u.refreshUsecase.Refresh(ctx, &entities.RefreshRequest{
		RefreshToken: refreshToken,
	})

	if err != nil {
		return nil, apperrors.Wrap(op, err)
	}

	return response, nil
}
