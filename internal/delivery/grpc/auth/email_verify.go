package grpcauthservise

import (
	"context"
	"errors"
	"log/slog"

	dtos "github.com/ttrtcixy/users-protos/gen/go/users"
	"github.com/ttrtcixy/users/internal/delivery/grpc/contracts"
	entities "github.com/ttrtcixy/users/internal/domain/entities/auth"
	apperrors "github.com/ttrtcixy/users/internal/domain/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type EmailVerifyService struct {
	log     *slog.Logger
	usecase contracts.VerifyUseCase
}

func NewVerifyEmail(log *slog.Logger, usecase contracts.UseCase) *EmailVerifyService {
	return &EmailVerifyService{
		log:     log,
		usecase: usecase,
	}
}

func (s *EmailVerifyService) VerifyEmail(ctx context.Context, payload *dtos.VerifyEmailRequest) (*dtos.VerifyEmailResponse, error) {
	if err := s.validate(payload); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	result, err := s.usecase.EmailVerify(ctx, s.DTOToEntity(payload))
	if err != nil {
		return nil, s.errResponse(err)
	}

	return s.EntityToDTO(result), nil
}

func (s *EmailVerifyService) DTOToEntity(payload *dtos.VerifyEmailRequest) *entities.EmailVerifyRequest {
	return &entities.EmailVerifyRequest{EmailToken: payload.JwtEmailVerifyToken}
}

func (s *EmailVerifyService) EntityToDTO(result *entities.EmailVerifyResponse) *dtos.VerifyEmailResponse {
	return &dtos.VerifyEmailResponse{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		ClientId:     result.ClientUUID,
	}
}

func (s *EmailVerifyService) errResponse(err error) error {
	switch {
	case errors.Is(err, apperrors.ErrEmailTokenExpired):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, apperrors.ErrInvalidEmailVerifyToken):
		return status.Error(codes.InvalidArgument, err.Error())
	default:
		return status.Error(codes.Internal, err.Error())
	}
}

func (s *EmailVerifyService) validate(payload *dtos.VerifyEmailRequest) error {
	jwt := payload.GetJwtEmailVerifyToken()

	if len(jwt) <= 0 {
		return errors.New("token required")
	}
	return nil
}
