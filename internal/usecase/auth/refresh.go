package authusecase

import (
	"context"
	"errors"

	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/ttrtcixy/users/internal/domain/contracts"
	entities "github.com/ttrtcixy/users/internal/domain/entities/auth"
	apperrors "github.com/ttrtcixy/users/internal/domain/errors"
)

type TokenRefresher interface {
	ParseRefreshToken(token string) (clientID, jtl string, err error)
	RefreshToken(clientID, tokenID string, exp time.Time) (token string, err error)
	AccessToken(user *entities.TokenUserInfo) (token string, err error)
}

type UpdateTokenUseCase struct {
	log   *slog.Logger
	repo  contracts.UpdateSessionRepository
	token TokenRefresher
}

func NewRefreshUseCase(log *slog.Logger, repo contracts.UpdateSessionRepository, tokenService TokenRefresher) *UpdateTokenUseCase {
	return &UpdateTokenUseCase{
		log:   log,
		repo:  repo,
		token: tokenService,
	}
}

func (u *UpdateTokenUseCase) Refresh(ctx context.Context, payload *entities.RefreshRequest) (result *entities.RefreshResponse, err error) {
	const op = "authusecase.Refresh()"
	defer func() {
		if err != nil {
			var userErr apperrors.UserError
			if errors.As(err, &userErr) {
				return
			}

			u.log.LogAttrs(ctx, slog.LevelError, "Refresh error", slog.String("op", op), slog.String("error", err.Error()))
			err = apperrors.ErrServer
		}
	}()

	// parse client token
	clientID, JTI, err := u.token.ParseRefreshToken(payload.RefreshToken)
	if err != nil {
		return nil, err
	}

	// new refresh token info
	exp := time.Now(RefreshJwtExpiry()) // todo how to get RefreshJwtExpiry()
	newRefreshTokenUUID := uuid.NewString()

	createReq := &entities.UpdateSessionRequest{
		ClientUUID:          clientID,
		OldRefreshTokenUUID: JTI,
		NewRefreshTokenUUID: newRefreshTokenUUID,
		ExpiresAt:           exp,
	}

	// check token jtl + check clientID if good add new refresh jti and return public user info
	userInfo, err := u.repo.RefreshSession(ctx, createReq)
	if err != nil {
		return nil, err
	}

	// new refresh token
	refreshToken, err := u.token.RefreshToken(clientID, newRefreshTokenUUID, exp)
	if err != nil {
		return nil, err
	}
	// new access token
	accessToken, err := u.token.AccessToken(userInfo)

	return &entities.RefreshResponse{
		AccessToken:   accessToken,
		RefreshToken:  refreshToken,
		ClientID:      clientID,
		TokenUserInfo: userInfo,
	}, nil
}
