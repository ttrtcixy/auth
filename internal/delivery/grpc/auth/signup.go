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
	"google.golang.org/protobuf/types/known/emptypb"
)

type SignupService struct {
	log     *slog.Logger
	usecase contracts.SignupUseCase
}

func NewSignup(log *slog.Logger, usecase contracts.SignupUseCase) *SignupService {
	return &SignupService{
		log:     log,
		usecase: usecase,
	}
}

func (s *SignupService) Signup(ctx context.Context, payload *dtos.SignupRequest) (*emptypb.Empty, error) {
	if err := s.validate(payload); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	err := s.usecase.Signup(ctx, s.DTOToEntity(payload))
	if err != nil {
		return nil, s.errResponse(err)
	}

	return nil, nil
}

func (s *SignupService) errResponse(err error) error {
	var exists = &apperrors.ErrLoginExists{}
	switch {
	case errors.As(err, &exists):
		return status.Error(codes.InvalidArgument, err.Error())
	}
	return status.Error(codes.Internal, err.Error())
}

func (s *SignupService) DTOToEntity(payload *dtos.SignupRequest) *entities.SignupRequest {
	return &entities.SignupRequest{
		Email:    payload.GetEmail(),
		Username: payload.GetUsername(),
		Password: payload.GetPassword(),
	}
}

func (s *SignupService) validate(payload *dtos.SignupRequest) error {
	email := payload.GetEmail()
	username := payload.GetUsername()
	password := payload.GetPassword()

	var validationErr = &apperrors.ValidationErrors{}
	if email == "" || username == "" || password == "" {
		return errors.New("email, username and password is required")
	}

	if err := s.emailValidate(email); err != nil {
		validationErr.Add("email", err.Error())
	}

	if err := s.usernameValidate(username); err != nil {
		validationErr.Add("username", err.Error())
	}

	if err := s.passwordValidate(password); err != nil {
		validationErr.Add("password", err.Error())
	}

	if len(*validationErr) > 0 {
		return validationErr
	}
	return nil
}

func (s *SignupService) emailValidate(email string) error {
	if utf8.RuneCountInString(email) > 254 {
		return errors.New("is too long (max 254 characters)")
	}

	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return errors.New("must contain exactly one @")
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

func (s *SignupService) usernameValidate(username string) error {
	if len(username) < 1 || len(username) > 20 {
		return errors.New("must be 1-20 characters long")
	}

	firstChar := rune(username[0])
	if !unicode.IsLetter(firstChar) && !unicode.IsNumber(firstChar) {
		return errors.New("must start with a letter or number")
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
			return errors.New("contains invalid characters (only a-z, 0-9, _, -, allowed)")
		}
	}

	return nil
}

func (s *SignupService) passwordValidate(password string) error {
	if len(password) < 8 || len(password) > 20 {
		return errors.New("must be 8-20 characters long")
	}
	return nil
}
