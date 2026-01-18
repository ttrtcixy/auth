package domainservice

import (
	"github.com/ttrtcixy/users/internal/domain/contracts"
	sessionservice "github.com/ttrtcixy/users/internal/domain/service/session"
	infraservice "github.com/ttrtcixy/users/internal/infrastructure/service"
)

type Config struct {
	createSessionCfg sessionservice.Config
}

type DomainService struct {
	SessionService *sessionservice.CreateSessionService
}

func NewDomainService(cfg *Config, repo contracts.Repository, service *infraservice.InfraServices) *DomainService {
	return &DomainService{
		SessionService: sessionservice.New(&cfg.createSessionCfg, repo, service.Jwt),
	}
}
