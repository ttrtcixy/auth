package grpcauthservise

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"unicode"
	"unicode/utf8"

	dtos "github.com/ttrtcixy/users-protos/gen/go/users"
	"github.com/ttrtcixy/users/internal/delivery/grpc/contracts"
	entities "github.com/ttrtcixy/users/internal/domain/entities/auth"
	apperrors "github.com/ttrtcixy/users/internal/domain/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type SigninService struct {
	log     *slog.Logger
	usecase contracts.SigninUseCase
}

func NewSignin(log *slog.Logger, usecase contracts.SigninUseCase) *SigninService {
	return &SigninService{
		log:     log,
		usecase: usecase,
	}
}

func (s *SigninService) Signin(ctx context.Context, payload *dtos.SigninRequest) (*dtos.SigninResponse, error) {
	if err := s.validate(payload); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	result, err := s.usecase.Signin(ctx, s.DTOToEntity(payload))
	if err != nil {
		return nil, s.errResponse(err)
	}

	return &dtos.SigninResponse{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		ClientId:     result.ClientUUID,
	}, nil
}

func (s *SigninService) DTOToEntity(payload *dtos.SigninRequest) *entities.SigninRequest {
	return &entities.SigninRequest{
		Username: payload.GetUsername(),
		Email:    payload.GetEmail(),
		Password: payload.GetPassword(),
	}
}

func (s *SigninService) errResponse(err error) error {
	var exists = &apperrors.ErrLoginExists{}
	switch {
	case errors.Is(err, apperrors.ErrEmailVerify):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, apperrors.ErrUserNotRegister):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, apperrors.ErrInvalidPassword):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.As(err, &exists):
		return status.Error(codes.InvalidArgument, err.Error())
	}
	return status.Error(codes.Internal, err.Error())
}

func (s *SigninService) validate(payload *dtos.SigninRequest) error {
	email := payload.GetEmail()
	username := payload.GetUsername()
	password := payload.GetPassword()

	if email == "" && username == "" {
		return errors.New("email or username is required")
	}

	if email != "" && username != "" {
		return errors.New("email and username cannot be used together")
	}

	if email != "" {
		if err := s.emailValidate(email); err != nil {
			return err
		}
	}

	if username != "" {
		if err := s.usernameValidate(username); err != nil {
			return err
		}
	}

	if err := s.passwordValidate(password); err != nil {
		return err
	}
	return nil
}

func (s *SigninService) emailValidate(email string) error {
	if utf8.RuneCountInString(email) > 254 {
		return errors.New("email is too long (max 254 characters)")
	}

	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return errors.New("email must contain exactly one @")
	}
	local, domain := parts[0], parts[1]

	if len(local) == 0 {
		return errors.New("local part (before @) cannot be empty")
	}
	for _, r := range local {
		// Allow: letters (a-z, A-Z), digits (0-9), and ., _, %, +, -
		if !(r >= 'a' && r <= 'z') && !(r >= 'A' && r <= 'Z') &&
			!(r >= '0' && r <= '9') && r != '.' && r != '_' && r != '%' && r != '+' && r != '-' {
			return errors.New("local part contains invalid characters")
		}
	}

	if len(domain) == 0 {
		return errors.New("domain part (after @) cannot be empty")
	}
	if strings.Contains(domain, "..") {
		return errors.New("domain cannot contain consecutive dots")
	}
	for _, r := range domain {
		// Allow: letters (a-z, A-Z), digits (0-9), ., and -
		if !(r >= 'a' && r <= 'z') && !(r >= 'A' && r <= 'Z') &&
			!(r >= '0' && r <= '9') && r != '.' && r != '-' {
			return errors.New("domain contains invalid characters")
		}
	}

	if !strings.Contains(domain, ".") {
		return errors.New("domain must contain a dot (e.g., example.com)")
	}

	if domain[0] == '.' || domain[0] == '-' ||
		domain[len(domain)-1] == '.' || domain[len(domain)-1] == '-' {
		return errors.New("domain cannot start or end with a dot or hyphen")
	}

	return nil
}

func (s *SigninService) usernameValidate(username string) error {
	if len(username) < 1 || len(username) > 20 {
		return errors.New("username must be 1-20 characters long")
	}

	firstChar := rune(username[0])
	if !unicode.IsLetter(firstChar) && !unicode.IsNumber(firstChar) {
		return errors.New("username must start with a letter or number")
	}

	for _, char := range username {
		switch {
		case unicode.IsLetter(char):
			continue
		case unicode.IsNumber(char):
			continue
		case char == '_' || char == '-':
			continue
		default:
			return errors.New("username contains invalid characters (only a-z, 0-9, _, -, allowed)")
		}
	}

	return nil
}

func (s *SigninService) passwordValidate(password string) error {
	if password == "" {
		return errors.New("password is required")
	}

	if len(password) < 8 || len(password) > 20 {
		return errors.New("password must be 8-20 characters long")
	}
	return nil
}
