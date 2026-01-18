package authusecase

import (
	"context"
	"errors"
	"log/slog"
	"strconv"

	"github.com/ttrtcixy/users/internal/domain/contracts"
	entities "github.com/ttrtcixy/users/internal/domain/entities/auth"
	apperrors "github.com/ttrtcixy/users/internal/domain/errors"
)

type TokenGenerator interface {
	AccessToken(user *entities.TokenUserInfo) (token string, err error)
}

type PasswordChecker interface {
	ComparePasswords(storedHash, password string, salt string) (bool, error)
}

type SessionManager interface {
	CreateSession(ctx context.Context, userID int) (*entities.CreateSessionResponse, error)
}

type SigninUseCase struct {
	log            *slog.Logger
	repo           contracts.SigninRepository
	tokenService   TokenGenerator
	hashService    PasswordChecker
	sessionService SessionManager
}

func NewSignin(log *slog.Logger, repo contracts.Repository, tokenService TokenGenerator, hashService PasswordChecker, sessionManager SessionManager) *SigninUseCase {
	return &SigninUseCase{
		log:            log,
		repo:           repo,
		tokenService:   tokenService,
		hashService:    hashService,
		sessionService: sessionManager,
	}
}

func (u *SigninUseCase) Signin(ctx context.Context, payload *entities.SigninRequest) (result *entities.SigninResponse, err error) {
	const op = "authusecase.Signin()"
	defer func() {
		if err != nil {
			var userErr apperrors.UserError
			if errors.As(err, &userErr) {
				return
			}
			u.log.LogAttrs(ctx, slog.LevelError, "Signin error", slog.String("op", op), slog.String("error", err.Error()))
			err = apperrors.ErrServer
		}
	}()

	user, err := u.validateUser(ctx, payload)
	if err != nil {
		return nil, err
	}
	roleId := strconv.Itoa(user.RoleId)

	accessToken, err := u.tokenService.AccessToken(&entities.TokenUserInfo{
		ID:       strconv.Itoa(user.ID),
		Username: user.Username,
		Email:    user.Email,
		RoleID:   roleId,
	})
	if err != nil {
		return nil, err
	}

	sessionInfo, err := u.sessionService.CreateSession(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	result = &entities.SigninResponse{
		AccessToken:  accessToken,
		RefreshToken: sessionInfo.RefreshToken,
		ClientUUID:   sessionInfo.ClientUUID,
	}

	return result, nil
}

func (u *SigninUseCase) validateUser(ctx context.Context, payload *entities.SigninRequest) (user *entities.User, err error) {
	const op = "validateUser"

	if user, err = u.getUser(ctx, payload); err != nil {
		return nil, apperrors.Wrap(op, err)
	}

	if user.IsActive == false {
		return nil, apperrors.ErrEmailVerify
	}

	ok, err := u.hashService.ComparePasswords(user.PasswordHash, payload.Password, user.PasswordSalt)
	if err != nil {
		return nil, apperrors.Wrap(op, err)
	}

	if !ok {
		return nil, apperrors.ErrInvalidPassword
	}

	return user, nil
}

func (u *SigninUseCase) getUser(ctx context.Context, payload *entities.SigninRequest) (user *entities.User, err error) {
	const op = "user"

	if payload.Email != "" {
		user, err = u.repo.UserByEmail(ctx, payload.Email)
	} else {
		user, err = u.repo.UserByUsername(ctx, payload.Username)
	}
	if err != nil {
		return nil, apperrors.Wrap(op, err)
	}

	return user, nil
}

//func (u *SigninUseCase) createSession(ctx context.Context, userID int) (refreshToken, clientUUID string, err error) {
//	//const op = "createSession"
//	//
//	//clientUUID = uuid.NewString()
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
//	//return refreshToken, clientUUID, nil
//}
