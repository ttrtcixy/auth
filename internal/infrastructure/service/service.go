package infraservice

import (
	"github.com/ttrtcixy/users/internal/infrastructure/service/hash"
	token "github.com/ttrtcixy/users/internal/infrastructure/service/jwt"
	"github.com/ttrtcixy/users/internal/infrastructure/service/smtp"
)

type Config struct {
	tokenCfg token.Config
	smtpCfg  smtp.Config
}

type InfraServices struct {
	Jwt            *token.JwtTokenService
	PasswordHasher *hashservice.HasherService
	Smtp           *smtp.SenderService
}

func New(cfg *Config) *InfraServices {
	return &InfraServices{
		Jwt:  token.New(&cfg.tokenCfg),
		Smtp: smtp.New(&cfg.smtpCfg),
	}
}
