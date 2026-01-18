package grpcauthservise

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	dtos "github.com/ttrtcixy/users-protos/gen/go/users"
	"github.com/ttrtcixy/users/internal/delivery/grpc/contracts"
	entities "github.com/ttrtcixy/users/internal/domain/entities/auth"
	apperrors "github.com/ttrtcixy/users/internal/domain/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type TokensVerifyService struct {
	log     *slog.Logger
	usecase contracts.TokensVerifyUseCase
}

func NewVerifyTokens(log *slog.Logger, usecase contracts.UseCase) *TokensVerifyService {
	return &TokensVerifyService{
		log:     log,
		usecase: usecase,
	}
}

func (s *TokensVerifyService) VerifyToken(ctx context.Context, payload *dtos.VerifyTokensRequest) (*dtos.VerifyTokensResponse, error) {
	if err := s.validate(payload); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	result, err := s.usecase.TokensVerify(ctx, s.DTOToEntity(payload))
	if err != nil {
		return nil, s.errResponse(err)
	}

	return s.EntityToDTO(result), nil
}

func (s *TokensVerifyService) errResponse(err error) error {
	switch {
	case errors.Is(err, apperrors.ErrInvalidRefreshToken) || errors.Is(err, apperrors.ErrInvalidAccessToken):
		return status.Error(codes.InvalidArgument, apperrors.ErrPleaseLogin.Error())
	case errors.Is(err, apperrors.ErrRefreshTokenExpired) || errors.Is(err, apperrors.ErrAccessTokenExpired):
		return status.Error(codes.InvalidArgument, apperrors.ErrPleaseLogin.Error())
	default:
		return status.Error(codes.Internal, err.Error())
	}
}

func (s *TokensVerifyService) DTOToEntity(payload *dtos.VerifyTokensRequest) *entities.TokensVerifyRequest {
	refreshToken := payload.GetRefreshToken()
	return &entities.TokensVerifyRequest{
		AccessToken:  payload.GetAccessToken(),
		RefreshToken: refreshToken}
}

func (s *TokensVerifyService) validate(payload *dtos.VerifyTokensRequest) error {
	accessToken := payload.GetAccessToken()
	refreshToken := payload.GetRefreshToken()

	var vErr = &apperrors.ValidationErrors{}

	if err := s.validateAccessToken(accessToken); err != nil {
		vErr.Add("accessToken", err.Error())
	}

	if err := s.validateRefreshToken(refreshToken); err != nil {
		vErr.Add("refreshToken", err.Error())
	}

	if len(*vErr) > 0 {
		return vErr
	}
	return nil
}

func (s *TokensVerifyService) validateAccessToken(token string) error {
	if token == "" {
		return fmt.Errorf("access token is required")
	}

	if err := s.validateToken(token); err != nil {
		return err
	}

	return nil
}

func (s *TokensVerifyService) validateRefreshToken(token string) error {
	if token == "" {
		return nil
	}

	if err := s.validateToken(token); err != nil {
		return err
	}

	return nil
}

func (s *TokensVerifyService) validateToken(token string) error {
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

func (s *TokensVerifyService) EntityToDTO(result *entities.TokensVerifyResponse) *dtos.VerifyTokensResponse {
	var tokenUpdate *dtos.TokenUpdate

	if result.Tokens != nil {
		tokenUpdate = &dtos.TokenUpdate{}
		if result.Tokens.AccessToken != nil {
			tokenUpdate.NewAccessToken = result.Tokens.AccessToken
		}

		if result.Tokens.RefreshToken != nil {
			tokenUpdate.NewRefreshToken = result.Tokens.RefreshToken
		}
	}

	userId, _ := strconv.Atoi(result.ID)
	return &dtos.VerifyTokensResponse{
		UserData: &dtos.UserData{
			Id:       int64(userId),
			Username: result.Username,
			Email:    result.Email,
			RoleId:   result.RoleID,
		},
		TokenUpdate: tokenUpdate,
	}
}
