package grpcauthservise

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	dtos "github.com/ttrtcixy/users-protos/gen/go/users"
	"github.com/ttrtcixy/users/internal/delivery/grpc/contracts"
	entities "github.com/ttrtcixy/users/internal/domain/entities/auth"
	apperrors "github.com/ttrtcixy/users/internal/domain/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type RefreshService struct {
	log     *slog.Logger
	usecase contracts.RefreshUseCase
}

func NewRefresh(log *slog.Logger, usecase contracts.AuthUseCase) *RefreshService {
	return &RefreshService{
		log:     log,
		usecase: usecase,
	}
}

func (s *RefreshService) Refresh(ctx context.Context, payload *dtos.RefreshRequest) (*dtos.RefreshResponse, error) {
	if err := s.validate(payload); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	result, err := s.usecase.Refresh(ctx, s.DTOToEntity(payload))
	if err != nil {
		return nil, s.errResponse(err)
	}

	return &dtos.RefreshResponse{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
	}, nil
}

func (s *RefreshService) DTOToEntity(payload *dtos.RefreshRequest) *entities.RefreshRequest {
	return &entities.RefreshRequest{RefreshToken: payload.RefreshToken}
}

func (s *RefreshService) errResponse(err error) error {
	switch {
	case errors.Is(err, apperrors.ErrRefreshTokenExpired) || errors.Is(err, apperrors.ErrInvalidRefreshToken):
		return status.Error(codes.InvalidArgument, err.Error())
	}
	return status.Error(codes.Internal, err.Error())
}

func (s *RefreshService) validate(payload *dtos.RefreshRequest) error {
	token := payload.GetRefreshToken()
	if token == "" {
		return fmt.Errorf("token is required")
	}

	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return fmt.Errorf("invalid token format")
	}

	for _, part := range parts {
		if !isValidBase64URL(part) {
			return fmt.Errorf("invalid token format")
		}
	}

	return nil
}

func isValidBase64URL(s string) bool {
	const base64URLChars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_"

	for _, r := range s {
		if !strings.ContainsRune(base64URLChars, r) {
			return false
		}
	}
	return true
}
