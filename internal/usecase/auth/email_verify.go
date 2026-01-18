package authusecase

import (
	"context"
	"errors"

	"log/slog"
	"strconv"
	"time"

	"github.com/ttrtcixy/users/internal/domain/contracts"
	entities "github.com/ttrtcixy/users/internal/domain/entities/auth"
	apperrors "github.com/ttrtcixy/users/internal/domain/errors"
)

type VerifyEmailServices struct {
	EmailVerifyTokenService
	EmailVerifySessionCreator
}

type EmailVerifyTokenService interface {
	AccessToken(user *entities.TokenUserInfo) (token string, err error)
	RefreshToken(clientID, tokenID string, exp time.Time) (token string, err error)
	ParseVerificationToken(jwtToken string) (email string, err error)
}

type EmailVerifySessionCreator interface {
	CreateSession(ctx context.Context, userID int) (*entities.CreateSessionResponse, error)
}

type VerifyEmailUseCase struct {
	log            *slog.Logger
	repo           contracts.VerifyRepository
	tokenService   EmailVerifyTokenService
	sessionService EmailVerifySessionCreator
}

func NewVerifyEmail(log *slog.Logger, repo contracts.VerifyRepository, tokenService EmailVerifyTokenService) *VerifyEmailUseCase {
	return &VerifyEmailUseCase{
		log:          log,
		repo:         repo,
		tokenService: tokenService,
	}
}

// EmailVerify - get jwtToken with email and activate user with that email.
func (u *VerifyEmailUseCase) EmailVerify(ctx context.Context, payload *entities.EmailVerifyRequest) (result *entities.EmailVerifyResponse, err error) {
	const op = "authusecase.EmailVerify()"

	defer func() {
		if err != nil {
			var userErr apperrors.UserError
			if errors.As(err, &userErr) {
				return
			}
			u.log.LogAttrs(ctx, slog.LevelError, "Email verify error", slog.String("op", op), slog.String("error", err.Error()))
			err = apperrors.ErrServer
		}
	}()

	userInfo, err := u.activateUser(ctx, payload.EmailToken)
	if err != nil {
		return nil, err
	}

	accessToken, err := u.tokenService.AccessToken(userInfo)
	if err != nil {
		return nil, err
	}

	userId, err := strconv.Atoi(userInfo.ID)
	if err != nil {
		return nil, err
	}

	sessionInfo, err := u.sessionService.CreateSession(ctx, userId)
	if err != nil {
		return nil, err
	}

	result = &entities.EmailVerifyResponse{
		AccessToken:  accessToken,
		RefreshToken: sessionInfo.RefreshToken,
		ClientUUID:   sessionInfo.ClientUUID,
	}

	return result, nil
}

func (u *VerifyEmailUseCase) activateUser(ctx context.Context, token string) (user *entities.TokenUserInfo, err error) {
	const op = "authusecase.activateUser()"

	email, err := u.tokenService.ParseVerificationToken(token)
	if err != nil {
		return nil, apperrors.Wrap(op, err)
	}

	if user, err = u.repo.ActivateUser(ctx, email); err != nil {
		return nil, apperrors.Wrap(op, err)
	}

	return user, nil
}

//func (u *VerifyEmailUseCase) createSession(ctx context.Context, userID int) (refreshToken, clientID string, err error) {
//	//const op = "authusecase.createSession()"
//	//
//	//clientUUID := uuid.NewString()
//	//
//	//tokenUUID := uuid.NewString()
//	//
//	//exp := time.Now().Add(u.cfg.RefreshJwtExpiry())
//	//
//	//if refreshToken, err = u.token.RefreshToken(clientUUID, tokenUUID, exp); err != nil {
//	//	return "", "", apperrors.Wrap(op, err)
//	//}
//	//
//	//createReq := &entities.CreateSession{
//	//	UserID:           userID,
//	//	ClientUUID:       clientUUID,
//	//	RefreshTokenUUID: tokenUUID,
//	//	ExpiresAt:        exp,
//	//}
//	//
//	//if err = u.repo.CreateSession(ctx, createReq); err != nil {
//	//	return "", "", apperrors.Wrap(op, err)
//	//}
//	//
//	//return refreshToken, clientID, nil
//}
