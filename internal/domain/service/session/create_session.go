package sessionservice

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/ttrtcixy/users/internal/domain/contracts"
	entities "github.com/ttrtcixy/users/internal/domain/entities/auth"
	apperrors "github.com/ttrtcixy/users/internal/domain/errors"
)

type RefreshTokenGetter interface {
	RefreshToken(clientID, tokenID string, exp time.Time) (token string, err error)
}

type Config struct {
}

type CreateSessionService struct {
	cfg          *Config
	repo         contracts.CreateSessionRepository
	tokenService RefreshTokenGetter
}

func New(cfg *Config, repo contracts.Repository, tokenService RefreshTokenGetter) *CreateSessionService {
	return &CreateSessionService{
		cfg:          cfg,
		repo:         repo,
		tokenService: tokenService,
	}
}
func (s *CreateSessionService) CreateSession(ctx context.Context, userID int) (*entities.CreateSessionResponse, error) {
	const op = "sessionservice.CreateSession()"

	clientUUID := uuid.NewString()

	tokenUUID := uuid.NewString()

	exp := time.Now().Add(u.cfg.RefreshJwtExpiry())

	refreshToken, err := s.tokenService.RefreshToken(clientUUID, tokenUUID, exp)
	if err != nil {
		return nil, apperrors.Wrap(op, err)
	}

	createReq := &entities.Session{
		UserID:           userID,
		ClientUUID:       clientUUID,
		RefreshTokenUUID: tokenUUID,
		ExpiresAt:        exp,
	}

	if err = s.repo.CreateSession(ctx, createReq); err != nil {
		return nil, apperrors.Wrap(op, err)
	}

	return &entities.CreateSessionResponse{
			RefreshToken: refreshToken,
			ClientUUID:   clientUUID,
		},
		nil
}
