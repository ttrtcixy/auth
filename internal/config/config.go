package config

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
	logger "github.com/ttrtcixy/color-slog-handler"
	storage "github.com/ttrtcixy/pgx-wrapper"
	closer "github.com/ttrtcixy/task-closer"
	"github.com/ttrtcixy/users/internal/delivery/grpc"
	domainservice "github.com/ttrtcixy/users/internal/domain/service"
	"github.com/ttrtcixy/users/internal/infrastructure/service"
)

type Config struct {
	Logger        logger.Config
	Closer        closer.Config
	DB            storage.Config
	InfraService  infraservice.Config
	DomainService domainservice.Config
	GRPC          grpc.Config
}

func (c *Config) Close(_ context.Context) error {
	os.Clearenv()
	return nil
}

// New load parameters from the .env file and return Config
func New() (cfg *Config, err error) {
	defer func() {
		if err != nil {
			os.Clearenv()
		}
	}()

	err = MustLoad(".env")
	if err != nil {
		return nil, err
	}
	cfg = &Config{}

	err = env.Parse(cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

// MustLoad loading parameters from the env file
func MustLoad(filename string) error {
	const op = "config.MustLoad()"
	_, err := os.Stat(filename)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("%s - file: %s does not exist", op, filename)
		}
		return err
	}

	err = godotenv.Load(filename)
	if err != nil {
		return fmt.Errorf("%s - load env file error -> %w", op, err)
	}

	return nil
}
