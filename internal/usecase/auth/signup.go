package authusecase

import (
	"context"
	"encoding/base64"
	"errors"
	"log/slog"

	"github.com/ttrtcixy/users/internal/domain/contracts"
	entities "github.com/ttrtcixy/users/internal/domain/entities/auth"
	apperrors "github.com/ttrtcixy/users/internal/domain/errors"
)

type SignupServices struct {
	MailService
	HashService
	EmailTokenVerifierService
}

type MailService interface {
	Send(to string, token string) error
}

type HashService interface {
	Salt() ([]byte, error)
	HashWithSalt(str string, salt []byte) (hash string, err error)
}

type EmailTokenVerifierService interface {
	EmailVerificationToken(email string) (token string, err error)
}

type SignupUseCase struct {
	log   *slog.Logger
	repo  contracts.SignupRepository
	mail  MailService
	hash  HashService
	token EmailTokenVerifierService
}

func NewSignup(log *slog.Logger, repo contracts.Repository, service *SignupServices) *SignupUseCase {
	return &SignupUseCase{
		log:   log,
		repo:  repo,
		mail:  service.MailService,
		hash:  service.HashService,
		token: service.EmailTokenVerifierService,
	}
}

// todo validate как работает fmt.Errorf("%s: %w", на большом стеке вызова)

func (u *SignupUseCase) Signup(ctx context.Context, payload *entities.SignupRequest) (err error) {
	const op = "authusecase.Signup()"

	err = u.repo.RunInTx(ctx, func(ctx context.Context) error {
		if err := u.validPayload(ctx, payload); err != nil {
			return err
		}

		hash, salt, err := u.passwordHashing(payload.Password)
		if err != nil {
			return err
		}

		emailToken, err := u.token.EmailVerificationToken(payload.Email)
		if err != nil {
			return err
		}

		createReq := &entities.CreateUserRequest{
			Username:     payload.Username,
			Email:        payload.Email,
			RoleID:       2, // "default user"
			PasswordHash: hash,
			PasswordSalt: salt,
		}

		if err = u.repo.CreateUser(ctx, createReq); err != nil {
			return err
		}

		if err = u.mail.Send(payload.Email, emailToken); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		var ue apperrors.UserError
		if errors.As(err, &ue) {
			return err
		}

		u.log.LogAttrs(ctx, slog.LevelError, "", slog.String("op", op), slog.String("error", err.Error()))
		return apperrors.ErrServer
	}
	return nil
}

// validPayload - UserError: *apperrors.ErrLoginExists
func (u *SignupUseCase) validPayload(ctx context.Context, payload *entities.SignupRequest) error {
	const op = "authusecase.validPayload()"

	exists, err := u.repo.CheckLoginExist(ctx, payload)
	if err != nil {
		return apperrors.Wrap(op, err)
	}

	if exists.Status {
		var err = &apperrors.ErrLoginExists{}
		if exists.UsernameExists {
			err.Username = payload.Username
		}
		if exists.EmailExists {
			err.Email = payload.Email
		}
		// todo test
		return apperrors.Wrap(op, err)
	}

	return nil
}

func (u *SignupUseCase) passwordHashing(password string) (hash string, salt string, err error) {
	const op = "authusecase.passwordHashing()"

	byteSalt, err := u.hash.Salt()
	if err != nil {
		return "", "", apperrors.Wrap(op, err)
	}

	if hash, err = u.hash.HashWithSalt(password, byteSalt); err != nil {
		return "", "", apperrors.Wrap(op, err)
	}

	return hash, base64.StdEncoding.EncodeToString(byteSalt), nil
}
